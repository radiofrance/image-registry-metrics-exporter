package google

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"
)

// Gcr implements conf.Provider interface
// to interact with Google Container Registry.
type Gcr struct {
	Auth any
}

func New() *Gcr {
	auth, err := google.NewEnvAuthenticator(context.Background())
	if err != nil {
		slog.Info(err.Error())
	}

	return &Gcr{
		Auth: auth,
	}
}

// GetImagesList returns a list of images from Google Container Registry.
func (gcr Gcr) GetImagesList(url string) ([]string, error) {
	switch v := gcr.Auth.(type) {
	case authn.Authenticator:
		catalog, err := crane.Catalog(url, crane.WithAuth(v))
		if err != nil {
			return []string{}, fmt.Errorf("cannot get images from google registry on %s : %w", url, err)
		}

		return catalog, nil
	default:
		catalog, err := crane.Catalog(url, crane.WithAuthFromKeychain(authn.DefaultKeychain))
		if err != nil {
			return []string{}, fmt.Errorf("cannot get images from google registry on %s : %w", url, err)
		}

		return catalog, nil
	}
}

// ListImageTag returns a map of tags names with metrics.TagMetadata for specified image from Google.
func (gcr Gcr) ListImageTag(imageName string) (map[string]metrics.TagMetadata, error) {
	imageInfos := map[string]metrics.TagMetadata{}

	var catalog *google.Tags

	rep, err := name.NewRepository(imageName)
	if err != nil {
		return map[string]metrics.TagMetadata{}, fmt.Errorf("cannot create repository on %s : %w", imageName, err)
	}

	switch v := gcr.Auth.(type) {
	case authn.Authenticator:
		catalog, err = google.List(rep, google.WithAuth(v))
		if err != nil {
			return map[string]metrics.TagMetadata{}, fmt.Errorf("cannot get images list from %s : %w", imageName, err)
		}
	default:
		catalog, err = google.List(rep, google.WithAuthFromKeychain(authn.DefaultKeychain))
		if err != nil {
			return map[string]metrics.TagMetadata{}, fmt.Errorf("cannot get images list from %s : %w", imageName, err)
		}
	}

	// Parse data from google to a app standardized struct
	for _, sha256 := range catalog.Manifests {
		metadata := metrics.TagMetadata{
			Created:  sha256.Created,
			Uploaded: sha256.Uploaded,
		}
		for _, tag := range sha256.Tags {
			imageInfos[tag] = metadata
		}
	}

	return imageInfos, nil
}
