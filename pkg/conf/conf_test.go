package conf_test

import (
	"regexp"
	"testing"

	"github.com/radiofrance/image-registry-metrics-exporter/pkg/conf"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Parallel()
	t.Run("load a valid configuration file", func(t *testing.T) {
		t.Parallel()

		config, err := conf.Load("./tests/valid")
		if err != nil {
			assert.Equal(t, nil, err)
		}

		fake := conf.Registry{
			Domain:            "eu.gcr.io",
			ImagesFilters:     []string{"filter"},
			ImagesRegex:       []*regexp.Regexp{regexp.MustCompile("filter")},
			TagsFilters:       []string{"latest"},
			TagsRegex:         []*regexp.Regexp{regexp.MustCompile("latest")},
			RateLimitAPI:      5,
			MaxConcurrentJobs: 5,
			Provider:          "fake",
		}
		var registries []conf.Registry
		conf.AddProvider(&fake)
		if err != nil {
			assert.Equal(t, nil, err)
		}

		registries = append(registries, fake)
		assert.Equal(t, conf.Config{Registries: registries, Cron: "0 * * * *"}, config)
	})
	t.Run("load a wrong type field registry in configuration file", func(t *testing.T) {
		t.Parallel()
		_, err := conf.Load("./tests/wrongtyperegistry/")
		assert.Regexp(t, regexp.MustCompile(".*failed to unmarshal configuration file:.*"), err)
	})
	t.Run("load a wrong type field cron in configuration file", func(t *testing.T) {
		t.Parallel()
		_, err := conf.Load("./tests/wrongtypecron/")
		assert.Regexp(t, regexp.MustCompile(".*failed to unmarshal configuration file:.*"), err)
	})
	t.Run("load a malformed yaml configuration file", func(t *testing.T) {
		t.Parallel()
		_, err := conf.Load("./tests/malformed/")
		assert.Regexp(t, regexp.MustCompile(".*could not find expected.*"), err)
	})
}
