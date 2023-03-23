package fake

import "github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"

type OCI struct {
	Images     map[string]map[string]metrics.TagMetadata
	ImagesList []string
}

// New generates a oci struct object.
func New(images map[string]map[string]metrics.TagMetadata, imagesList []string) *OCI {
	return &OCI{Images: images, ImagesList: imagesList}
}

// GetImagesList fakes expecting result from GetImagesList from other providers.
func (o OCI) GetImagesList(url string) ([]string, error) {
	return o.ImagesList, nil
}

// ListImageTag fakes expecting result from ListImageTag from other providers.
func (o OCI) ListImageTag(repo string) (map[string]metrics.TagMetadata, error) {
	return o.Images[repo], nil
}
