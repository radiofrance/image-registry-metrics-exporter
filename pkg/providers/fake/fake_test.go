package fake_test

import (
	"testing"
	"time"

	"github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/providers/fake"

	"github.com/stretchr/testify/assert"
)

func TestFake_GetImagesList(t *testing.T) {
	t.Parallel()
	dataset := []struct {
		title          string
		fake           fake.OCI
		data           []string
		expectedResult interface{}
	}{{
		title: "test simple object",
		fake: *fake.New(
			nil, []string{"image1", "image2"},
		),
		expectedResult: []string{"image1", "image2"},
	}}
	for _, data := range dataset {
		data := data
		t.Run(data.title, func(t *testing.T) {
			t.Parallel()
			result, _ := data.fake.GetImagesList("random-string")
			assert.Equal(t, data.expectedResult, result)
		})
	}
}

func TestFake_ListImageTag(t *testing.T) {
	t.Parallel()
	dataset := []struct {
		title          string
		fake           fake.OCI
		image          string
		expectedResult interface{}
	}{
		{
			title: "test simple object with / in image",
			image: "repo/image1",
			fake: fake.OCI{
				Images: map[string]map[string]metrics.TagMetadata{"repo/image1": {"tag1": metrics.TagMetadata{
					Created:  time.Date(1970, time.January, 1, 1, 0, 0, 3000000, time.Local),
					Uploaded: time.Date(1970, time.January, 1, 1, 0, 0, 4000000, time.Local),
				}}},
			},
			expectedResult: map[string]metrics.TagMetadata{"tag1": {
				Created:  time.Date(1970, time.January, 1, 1, 0, 0, 3000000, time.Local),
				Uploaded: time.Date(1970, time.January, 1, 1, 0, 0, 4000000, time.Local),
			}},
		},
		{
			title: "test simple object with no / in image",
			image: "image1",
			fake: fake.OCI{
				Images: map[string]map[string]metrics.TagMetadata{"image1": {"tag1": metrics.TagMetadata{
					Created:  time.Date(1970, time.January, 1, 1, 0, 0, 3000000, time.Local),
					Uploaded: time.Date(1970, time.January, 1, 1, 0, 0, 4000000, time.Local),
				}}},
			},
			expectedResult: map[string]metrics.TagMetadata{"tag1": {
				Created:  time.Date(1970, time.January, 1, 1, 0, 0, 3000000, time.Local),
				Uploaded: time.Date(1970, time.January, 1, 1, 0, 0, 4000000, time.Local),
			}},
		},
	}
	for _, data := range dataset {
		data := data
		t.Run(data.title, func(t *testing.T) {
			t.Parallel()
			result, _ := data.fake.ListImageTag(data.image)
			assert.Equal(t, data.expectedResult, result)
		})
	}
}
