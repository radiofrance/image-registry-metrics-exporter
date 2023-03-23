package scrapper

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/radiofrance/image-registry-metrics-exporter/pkg/conf"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/controllers"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"

	"go.uber.org/ratelimit"

	"github.com/sirupsen/logrus"
)

// Scrape work with []conf.Registry to scrape image tags metadata from OCI Registry.
// Then add to chan metrics.Job image tags to generate metrics on.
func Scrape(registries []conf.Registry, tags chan<- metrics.Job) error {
	for _, reg := range registries {
		filteredImageList, err := GetFilteredImagesList(reg)
		if err != nil {
			controllers.UpdateHealth(false)
			return err
		}

		GetImagesTags(reg, filteredImageList, tags)
	}
	return nil
}

// filterImagesList is used to filter a list of image names based on a list of filters.
func filterImagesList(filters []*regexp.Regexp, imagesNames *[]string) []string {
	var filteredCatalog []string

	if len(filters) == 0 {
		return *imagesNames
	}

	for _, image := range *imagesNames {
		for _, filterRegistry := range filters {
			matched := filterRegistry.MatchString(image)
			if matched {
				filteredCatalog = append(filteredCatalog, image)
			}
		}
	}
	return filteredCatalog
}

// filterTagsList is used to filter a map of provider.metrics.TagMetadata based on a list of filters.
func filterTagsList(
	tags map[string]metrics.TagMetadata,
	filters []*regexp.Regexp,
) map[string]metrics.TagMetadata {
	filteredTags := make(map[string]metrics.TagMetadata)

	if len(filters) == 0 {
		return tags
	}

	for tag, tagMetadata := range tags {
		for _, filter := range filters {
			matched := filter.MatchString(tag)
			if matched {
				filteredTags[tag] = tagMetadata
			}
		}
	}
	return filteredTags
}

// GetFilteredImagesList requests a Provider to get a list of images.
// It filters it based on Registry.ImagesFilters.
func GetFilteredImagesList(reg conf.Registry) ([]string, error) {
	var filteredImagesNames []string

	if len(reg.Domain) == 0 {
		return []string{}, fmt.Errorf("domain is empty, cannot get images from provider")
	}

	imageNames, err := reg.ProviderObject.GetImagesList(reg.Domain)
	if err != nil {
		return []string{}, fmt.Errorf("cannot get image list from provider: %w", err)
	}

	filteredImagesNames = filterImagesList(reg.ImagesRegex, &imageNames)

	if len(filteredImagesNames) == 0 {
		return []string{}, fmt.Errorf("filtered image list is empty")
	}

	return filteredImagesNames, nil
}

// GetImageTags request OCI compatible Registry through conf.Provider interface.
// For each image and tags filtered, it generates a metrics.Job and send it to chan.
// It is a worker function called by GetImagesTags.
func GetImageTags(
	reg conf.Registry,
	wg *sync.WaitGroup,
	images <-chan string,
	tags chan<- metrics.Job,
	rate ratelimit.Limiter,
) {
	defer wg.Done()

	for image := range images {
		logrus.Info("Updating image infos: ", image)
		rate.Take()

		extractedTags, err := reg.ProviderObject.ListImageTag(reg.Domain + "/" + image)
		if err != nil {
			logrus.Errorf("cannot list tags of %s : %s", reg.Domain+"/"+image, err)
		}
		filteredTags := filterTagsList(extractedTags, reg.TagsRegex)
		if err != nil {
			logrus.Errorf("cannot filter tags of %s : %s", reg.Domain+"/"+image, err)
		}

		for k, v := range filteredTags {
			tags <- metrics.Job{
				ImageName: image,
				TagName:   k,
				Metadata:  v,
			}
		}
	}
}

// GetImagesTags is a job wrapper which works on a list of image names. For each image, it calls GetImageTags.
func GetImagesTags(reg conf.Registry, imagesNames []string, tags chan<- metrics.Job) {
	max := reg.MaxConcurrentJobs
	images := make(chan string, max)
	wgCatalog := &sync.WaitGroup{}
	rate := ratelimit.New(reg.RateLimitAPI)
	for i := 0; i < max; i++ {
		wgCatalog.Add(1)
		go GetImageTags(reg, wgCatalog, images, tags, rate)
	}

	for _, imageName := range imagesNames {
		images <- imageName
	}
	close(images)
	wgCatalog.Wait()
}
