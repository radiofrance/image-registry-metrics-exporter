package scrapper

import (
	"regexp"
	"testing"
	"time"

	"github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"

	"github.com/stretchr/testify/assert"
)

func Test_filterTagsList(t *testing.T) {
	type args struct {
		filters []*regexp.Regexp
		tags    map[string]metrics.TagMetadata
	}

	t.Parallel()

	dataset := []struct {
		title          string
		args           args
		expectedResult map[string]metrics.TagMetadata
	}{
		{
			title: "test with empty filters",
			args: args{
				filters: []*regexp.Regexp{},
				tags: map[string]metrics.TagMetadata{
					"v.0.0.1": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
					"v.0.0.2": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
					"v.0.0.3": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
				},
			},
			expectedResult: map[string]metrics.TagMetadata{
				"v.0.0.1": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
				"v.0.0.2": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
				"v.0.0.3": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
			},
		},
		{
			title: "test with filter unmatched",
			args: args{
				filters: []*regexp.Regexp{regexp.MustCompile("-")},
				tags: map[string]metrics.TagMetadata{
					"v.0.0.1": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
					"v.0.0.2": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
					"v.0.0.3": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
				},
			},
			expectedResult: map[string]metrics.TagMetadata{},
		},
		{
			title: "test with multiple filters",
			args: args{
				filters: []*regexp.Regexp{regexp.MustCompile("v0.0.1"), regexp.MustCompile("v0.0.2")},
				tags: map[string]metrics.TagMetadata{
					"v0.0.1": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
					"v0.0.2": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
					"v0.0.3": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
				},
			},
			expectedResult: map[string]metrics.TagMetadata{
				"v0.0.1": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
				"v0.0.2": {Created: time.Unix(1, 1), Uploaded: time.Unix(1, 1)},
			},
		},
	}
	for _, data := range dataset {
		data := data
		t.Run(data.title, func(t *testing.T) {
			t.Parallel()
			result := filterTagsList(data.args.tags, data.args.filters)
			assert.Equal(t, data.expectedResult, result)
		})
	}
}

func Test_filterImagesList(t *testing.T) {
	type args struct {
		filters []*regexp.Regexp
		catalog *[]string
	}

	t.Parallel()

	dataset := []struct {
		title          string
		args           args
		expectedResult []string
	}{
		{
			title: "test with empty filters",
			args: args{
				filters: []*regexp.Regexp{},
				catalog: &[]string{
					"example.org/ovibos-moschatus",
					"example.org/mirounga-angustirostris",
					"example.org/psophia-viridis",
					"example.org/antechinus-flavipes",
					"example.org/bubalus-arnee",
					"example.org/lama-guanicoe",
				},
			},
			expectedResult: []string{
				"example.org/ovibos-moschatus",
				"example.org/mirounga-angustirostris",
				"example.org/psophia-viridis",
				"example.org/antechinus-flavipes",
				"example.org/bubalus-arnee",
				"example.org/lama-guanicoe",
			},
		},
		{
			title: "test with simple filter",
			args: args{
				filters: []*regexp.Regexp{regexp.MustCompile("mirounga-angustirostris")},
				catalog: &[]string{
					"example.org/ovibos-moschatus",
					"example.org/mirounga-angustirostris",
					"example.org/psophia-viridis",
					"example.org/antechinus-flavipes",
					"example.org/bubalus-arnee",
					"example.org/lama-guanicoe",
				},
			},
			expectedResult: []string{"example.org/mirounga-angustirostris"},
		},
		{
			title: "test with multiple filters",
			args: args{
				filters: []*regexp.Regexp{
					regexp.MustCompile("mirounga-angustirostris"),
					regexp.MustCompile("antechinus-flavipes"),
				},
				catalog: &[]string{
					"example.org/ovibos-moschatus",
					"example.org/mirounga-angustirostris",
					"example.org/psophia-viridis",
					"example.org/antechinus-flavipes",
					"example.org/bubalus-arnee",
					"example.org/lama-guanicoe",
				},
			},
			expectedResult: []string{
				"example.org/mirounga-angustirostris",
				"example.org/antechinus-flavipes",
			},
		},
		{
			title: "test with regex filter",
			args: args{
				filters: []*regexp.Regexp{regexp.MustCompile("/[abl].*")},
				catalog: &[]string{
					"example.org/ovibos-moschatus",
					"example.org/mirounga-angustirostris",
					"example.org/psophia-viridis",
					"example.org/antechinus-flavipes",
					"example.org/bubalus-arnee",
					"example.org/lama-guanicoe",
				},
			},
			expectedResult: []string{
				"example.org/antechinus-flavipes",
				"example.org/bubalus-arnee",
				"example.org/lama-guanicoe",
			},
		},
	}
	for _, data := range dataset {
		data := data
		t.Run(data.title, func(t *testing.T) {
			t.Parallel()
			result := filterImagesList(data.args.filters, data.args.catalog)
			assert.Equal(t, data.expectedResult, result)
		})
	}
}
