package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
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
)

func init() {
	_, debugMode := os.LookupEnv("DEBUG_MODE")
	if debugMode {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger := slog.New(handler)
		slog.SetDefault(logger)
	}
}

func main() {
	config, err := conf.Load(os.Getenv("IRME_CONF_FILE_PATH"))
	if err != nil {
		slog.Error(err.Error())
		return
	}

	tags, err := metrics.New()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to generate metrics struct: %v", err))
		os.Exit(1)
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
				slog.Info("Metrics server closed")
			} else {
				slog.Error(fmt.Sprintf("Failed to start metrics server %v", err))
			}
		}
	}()
	go func() {
		defer waitGroup.Done()
		if err := healthSrv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				slog.Info("Http server closed")
			} else {
				slog.Error(fmt.Sprintf("Failed to start http server %v", err))
			}
		}
	}()
	controllers.UpdateHealth(true)
	controllers.UpdateReady(true)

	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		slog.Error(fmt.Sprintf("failed to load location: %v", err))
		os.Exit(1)
	}

	scheduler := gocron.NewScheduler(location)
	_, err = scheduler.Cron(config.Cron).StartImmediately().Do(func() {
		slog.Info("Starting scraping metrics")
		if err := scrapper.Scrape(config.Registries, tags.Queue); err != nil {
			slog.Error(fmt.Sprintf("failed to scrape images metadata: %v", err))
			os.Exit(1)
		}
		slog.Info(fmt.Sprintf("Getting metrics is done, next schedule at %s",
			cronexpr.MustParse(config.Cron).Next(time.Now())))
	})
	if err != nil {
		slog.Error(fmt.Sprintf("failed to run cron job: %v", err))
		os.Exit(1)
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
	slog.Info("Shutting down")
}
