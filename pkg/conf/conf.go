package conf

import (
	"regexp"

	"github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/providers/fake"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/providers/google"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config contains a list of Registry and a Cron formatted string.
type Config struct {
	Registries []Registry
	Cron       string
}

// Registry provide configuration for image metrics registry exporter.
type Registry struct {
	// Domain is URL for OCI Registry.
	Domain string `mapstructure:"domain"`
	// TagsFilters gives a way to work only on a subsets of tags (regex).
	TagsFilters []string `mapstructure:"tagsFilters"`
	// TagRegex defines a type regex for TagFilters
	TagsRegex []*regexp.Regexp
	// ImageFilters is used to filter images. It is mandatory as none will output no images.
	ImagesFilters []string `mapstructure:"imagesFilters"`
	// ImagesRegex defines a type regex for ImagesFilters
	ImagesRegex []*regexp.Regexp
	// RateLimitAPI permits to not overflow backend Provider.
	RateLimitAPI int `mapstructure:"rateLimitAPI"`
	// MaxConcurrentJobs permit to not overflow IRME container on CPU/RAM.
	MaxConcurrentJobs int `mapstructure:"maxConcurrentJobs"`
	// Provider is used to import from yaml a Provider backend type.
	Provider string
	// ProviderObject is interpreted Provider field by AddProvider.
	ProviderObject Provider
}

// Provider works as a interface to OCI compatible registry backend.
type Provider interface {
	// GetImagesList that give a list of images ;
	GetImagesList(string) ([]string, error)
	// ListImageTag that give a map of tags with provider.metrics.TagMetadata on each of it.
	ListImageTag(string) (map[string]metrics.TagMetadata, error)
}

// Load a file with which contains a struct that can be unmarshal to a Config.
func Load(path string) (Config, error) {
	vip := viper.New()
	vip.SetConfigName("config")
	vip.SetConfigType("yaml")
	vip.AddConfigPath("/etc/irme/")
	vip.AddConfigPath("$HOME/.irme")
	vip.AddConfigPath(".")
	vip.AddConfigPath(path)

	if err := vip.ReadInConfig(); err != nil {
		return Config{}, errors.New("failed to open configuration file: " + err.Error())
	}

	var registries []Registry
	if err := vip.UnmarshalKey("registries", &registries); err != nil {
		return Config{}, errors.New("failed to unmarshal configuration file: " + err.Error())
	}

	for reg := range registries {
		AddProvider(&registries[reg])
		for _, filter := range registries[reg].ImagesFilters {
			r, err := regexp.Compile(filter)
			if err != nil {
				log.Warnf("image filter : %s is not a regex", filter)
			} else {
				registries[reg].ImagesRegex = append(registries[reg].ImagesRegex, r)
			}
		}
		for _, filter := range registries[reg].TagsFilters {
			r, err := regexp.Compile(filter)
			if err != nil {
				log.Warnf("tag filter : %s is not a regex", filter)
			} else {
				registries[reg].TagsRegex = append(registries[reg].TagsRegex, r)
			}
		}
	}

	var cron string
	if err := vip.UnmarshalKey("cron", &cron); err != nil {
		return Config{}, errors.New("failed to unmarshal configuration file: " + err.Error())
	}

	log.Info("Loaded configuration from ", vip.ConfigFileUsed())
	config := Config{Registries: registries, Cron: cron}

	return config, nil
}

// AddProvider interprets Provider to add ProviderObject on Registry.
func AddProvider(r *Registry) {
	switch r.Provider {
	case "google":
		r.ProviderObject = google.New()
	default:
		r.ProviderObject = fake.New(nil, nil)
	}
}
