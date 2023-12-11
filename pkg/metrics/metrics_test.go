package metrics_test

import (
	"testing"
	"time"

	"github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateMetricsOn(t *testing.T) { //nolint:paralleltest // cannot output metrics in parallels
	tag, err := metrics.New()
	require.NoError(t, err)

	dataset := []struct {
		title  string
		images map[string]map[string]metrics.TagMetadata
	}{
		{
			title: "test with an image",
			images: map[string]map[string]metrics.TagMetadata{
				"image1": {
					"0.0.1": {
						Created:  time.Date(2020, time.February, 0o1, 1, 3, 4, 5, time.UTC),
						Uploaded: time.Date(2021, time.February, 0o1, 1, 3, 4, 5, time.UTC),
					},
				},
			},
		},
		{
			title: "test with multiple images and multiple tags",
			images: map[string]map[string]metrics.TagMetadata{
				"image1": {
					"0.0.1": {
						Created:  time.Date(2020, time.February, 0o1, 1, 3, 4, 5, time.UTC),
						Uploaded: time.Date(2021, time.February, 0o1, 1, 3, 4, 5, time.UTC),
					},
					"0.0.2": {
						Created:  time.Date(2020, time.February, 0o1, 1, 3, 4, 5, time.UTC),
						Uploaded: time.Date(2021, time.February, 0o1, 1, 3, 4, 5, time.UTC),
					},
				},
				"image2": {
					"0.0.1": {
						Created:  time.Date(2020, time.February, 0o1, 1, 3, 4, 5, time.UTC),
						Uploaded: time.Date(2021, time.February, 0o1, 1, 3, 4, 5, time.UTC),
					},
					"0.0.2": {
						Created:  time.Date(2020, time.February, 0o1, 1, 3, 4, 5, time.UTC),
						Uploaded: time.Date(2021, time.February, 0o1, 1, 3, 4, 5, time.UTC),
					},
				},
			},
		},
	}
	for _, data := range dataset { //nolint:paralleltest // cannot output metrics in parallels
		data := data
		t.Run(data.title, func(t *testing.T) {
			go func() {
				tag.GenerateMetricsOn()
			}()

			for imageName, imageMetadata := range data.images {
				for tagName, tagMetadata := range imageMetadata {
					tag.Queue <- metrics.Job{
						ImageName: imageName,
						TagName:   tagName,
						Metadata:  tagMetadata,
					}
				}
			}
			time.Sleep(1 * time.Second)

			for imageName, imageMetadata := range data.images {
				for tagName, tagMetadata := range imageMetadata {
					assert.InEpsilon(t, float64(tagMetadata.Created.Unix()), testutil.ToFloat64(
						tag.MetricsCreatedTime.WithLabelValues(imageName, tagName)), 0.1, "Defines when image was Created")
					assert.InEpsilon(t, float64(tagMetadata.Uploaded.Unix()), testutil.ToFloat64(
						tag.MetricsUploadedTime.WithLabelValues(imageName, tagName)), 0.1, "Defines when image was Uploaded")
				}
			}
		})
	}
}
