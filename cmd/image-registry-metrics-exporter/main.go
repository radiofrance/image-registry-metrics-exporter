package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/radiofrance/image-registry-metrics-exporter/pkg/conf"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/controllers"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/scrapper"

	"github.com/aptible/supercronic/cronexpr"
	"github.com/go-co-op/gocron"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func init() {
	_, debugMode := os.LookupEnv("DEBUG_MODE")
	if debugMode {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	config, err := conf.Load(os.Getenv("IRME_CONF_FILE_PATH"))
	if err != nil {
		log.Error(err)
		return
	}

	tags, err := metrics.New()
	if err != nil {
		log.Fatalf("failed to generate metrics struct: %s", err)
	}
	tags.GenerateMetricsOn()

	bindAddrHealth := flag.String("bind-address-health", ":8080", "address:port to bind status endpoints to")
	bindAddrMetrics := flag.String("bind-address-metrics", ":9252", "address:port to bind /metrics endpoint to")

	flag.Parse()

	// Activating routes for HTTP server
	router := mux.NewRouter().StrictSlash(true)
	router.Use(
		mux.CORSMethodMiddleware(router), // Handle CORS requests
	)
	router.HandleFunc("/health", controllers.HealthCheck)
	router.HandleFunc("/readiness", controllers.Ready)

	// Activating routes for Metrics server
	routerMetrics := mux.NewRouter().StrictSlash(true)
	routerMetrics.Use(
		mux.CORSMethodMiddleware(routerMetrics), // Handle CORS requests
	)
	routerMetrics.Handle("/metrics", metrics.Handler())

	timeoutDuration := 30 * time.Second
	metricsSrv := &http.Server{
		Addr:         *bindAddrMetrics,
		Handler:      http.TimeoutHandler(routerMetrics, timeoutDuration, "Server Timeout"),
		ReadTimeout:  timeoutDuration,
		WriteTimeout: timeoutDuration,
	}
	healthSrv := &http.Server{
		Addr:         *bindAddrHealth,
		Handler:      http.TimeoutHandler(router, timeoutDuration, "Server Timeout"),
		ReadTimeout:  timeoutDuration,
		WriteTimeout: timeoutDuration,
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		if err := metricsSrv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("Metrics server closed")
			} else {
				log.Fatalf("Failed to start metrics server %v", err)
			}
		}
	}()
	go func() {
		defer waitGroup.Done()
		if err := healthSrv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("Http server closed")
			} else {
				log.Fatalf("Failed to start http server %v", err)
			}
		}
	}()
	controllers.UpdateHealth(true)
	controllers.UpdateReady(true)

	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}

	scheduler := gocron.NewScheduler(location)
	_, err = scheduler.Cron(config.Cron).StartImmediately().Do(func() {
		log.Info("Starting scraping metrics")
		if err := scrapper.Scrape(config.Registries, tags.Queue); err != nil {
			log.Fatalf("failed to scrape images metadata: %v", err)
		}
		log.Infof("Getting metrics is done, next schedule at %s",
			cronexpr.MustParse(config.Cron).Next(time.Now()))
	})
	if err != nil {
		log.Fatalf("failed to run cron job: %v", err)
	}
	scheduler.SetMaxConcurrentJobs(1, gocron.RescheduleMode)
	scheduler.StartAsync()

	// Graceful shutdown, inspired by https://github.com/gorilla/mux#graceful-shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.Signal(0xf)) // syscall.Signal(0xf) == SIGTERM
	<-c
	time.Sleep(15 * time.Second) // We give time to the readiness probe to be down
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	_ = metricsSrv.Shutdown(ctx)
	_ = healthSrv.Shutdown(ctx)
	waitGroup.Wait()
	log.Info("Shutting down")
}
