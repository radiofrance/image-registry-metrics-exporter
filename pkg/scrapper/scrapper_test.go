package scrapper_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/radiofrance/image-registry-metrics-exporter/pkg/conf"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/providers/fake"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/scrapper"

	"github.com/stretchr/testify/assert"
)

func TestRegistry_Scrape(t *testing.T) {
	t.Parallel()

	dataset := []struct {
		title          string
		registry       conf.Registry
		images         []string
		expectedResult map[string]map[string]metrics.TagMetadata
		expectedError  string
	}{
		{
			title:  "test valid case with one image",
			images: []string{"image1"},
			registry: conf.Registry{
				Domain: "eu.gcr.io",
				TagsFilters: []string{
					".*",
				},
				ImagesFilters: []string{
					".*",
				},
				RateLimitAPI:      1,
				MaxConcurrentJobs: 1,
				Provider:          "fake",
				ProviderObject: &fake.OCI{
					Images: map[string]map[string]metrics.TagMetadata{
						"eu.gcr.io/image1": {
							"tag1": {
								Created:  time.Time{},
								Uploaded: time.Time{},
							},
						},
					},
					ImagesList: []string{"image1"},
				},
			},
			expectedResult: map[string]map[string]metrics.TagMetadata{
				"image1": {
					"tag1": {
						Created:  time.Time{},
						Uploaded: time.Time{},
					},
				},
			},
		},
		{
			title:  "test without domain",
			images: []string{"image1"},
			registry: conf.Registry{
				Domain: "",
				TagsFilters: []string{
					".*",
				},
				ImagesFilters: []string{
					".*",
				},
				RateLimitAPI:      1,
				MaxConcurrentJobs: 1,
				Provider:          "fake",
				ProviderObject: &fake.OCI{
					Images: map[string]map[string]metrics.TagMetadata{
						"image1": {
							"tag1": {
								Created:  time.Time{},
								Uploaded: time.Time{},
							},
						},
					},
					ImagesList: []string{"image1"},
				},
			},
			expectedResult: map[string]map[string]metrics.TagMetadata{
				"image1": {
					"tag1": {
						Created:  time.Time{},
						Uploaded: time.Time{},
					},
				},
			},
			expectedError: "domain is empty, cannot get images from provider",
		},
		{
			title:  "test without tags filters",
			images: []string{"image1"},
			registry: conf.Registry{
				Domain:      "eu.gcr.io",
				TagsFilters: []string{},
				ImagesFilters: []string{
					".*",
				},
				RateLimitAPI:      1,
				MaxConcurrentJobs: 1,
				Provider:          "fake",
				ProviderObject: &fake.OCI{
					Images: map[string]map[string]metrics.TagMetadata{
						"eu.gcr.io/image1": {
							"tag1": {
								Created:  time.Time{},
								Uploaded: time.Time{},
							},
						},
					},
					ImagesList: []string{"image1"},
				},
			},
			expectedResult: map[string]map[string]metrics.TagMetadata{
				"image1": {
					"tag1": {
						Created:  time.Time{},
						Uploaded: time.Time{},
					},
				},
			},
		},
		{
			title:  "test without images filters",
			images: []string{"image1"},
			registry: conf.Registry{
				Domain: "eu.gcr.io",
				TagsFilters: []string{
					".*",
				},
				ImagesFilters:     []string{},
				RateLimitAPI:      1,
				MaxConcurrentJobs: 1,
				Provider:          "fake",
				ProviderObject: &fake.OCI{
					Images: map[string]map[string]metrics.TagMetadata{
						"eu.gcr.io/image1": {
							"tag1": {
								Created:  time.Time{},
								Uploaded: time.Time{},
							},
						},
					},
					ImagesList: []string{"image1"},
				},
			},
			expectedResult: map[string]map[string]metrics.TagMetadata{
				"image1": {
					"tag1": {
						Created:  time.Time{},
						Uploaded: time.Time{},
					},
				},
			},
		},
	}
	for _, data := range dataset {
		data := data
		t.Run(data.title, func(t *testing.T) {
			t.Parallel()
			catalog := make(map[string]map[string]metrics.TagMetadata)
			var registries []conf.Registry
			registries = append(registries, data.registry)
			images := make(chan metrics.Job)
			go func() {
				for image := range images {
					if _, ok := catalog[image.ImageName]; !ok {
						catalog[image.ImageName] = make(map[string]metrics.TagMetadata)
					}
					catalog[image.ImageName][image.TagName] = image.Metadata
				}
			}()
			if err := scrapper.Scrape(registries, images); err != nil {
				assert.Regexp(t, regexp.MustCompile(data.expectedError), err)
				return
			}
			close(images)
			time.Sleep(1 * time.Second)
			assert.Equal(t, data.expectedResult, catalog)
		})
	}
}
