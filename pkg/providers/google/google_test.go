package google_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/metrics"
	"github.com/radiofrance/image-registry-metrics-exporter/pkg/providers/google"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoogleProvider_ImagesCatalog(t *testing.T) {
	t.Parallel()

	dataset := []struct {
		title           string
		activateBackend bool
		expectedErr     error
		expectedResult  interface{}
	}{
		{
			title:           "test with fake crane",
			activateBackend: true,
			expectedErr:     nil,
			expectedResult:  []string{},
		},
		{
			title:           "no backend",
			activateBackend: false,
			expectedErr:     errors.New("(.*401*)|(.*UNAUTHORIZED.*)"),
			expectedResult:  []string{},
		},
	}
	for _, data := range dataset {
		data := data
		t.Run(data.title, func(t *testing.T) {
			t.Parallel()
			gProv := google.New()
			repo := ""
			if data.activateBackend {
				s := httptest.NewServer(registry.New())
				defer s.Close()
				u, err := url.Parse(s.URL)
				if err != nil {
					t.Fatal(err)
				}
				repo = u.Host
			}
			list, err := gProv.GetImagesList(repo)
			assert.Equal(t, data.expectedResult, list)
			if data.expectedErr == nil {
				require.NoError(t, err)
				return
			}
			assert.Regexp(t, regexp.MustCompile(data.expectedErr.Error()), err)
		})
	}
}

func TestGoogleProvider_ListImageTag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		responseBody []byte
		wantErr      bool
		want         map[string]metrics.TagMetadata
	}{
		{
			name:         "success",
			responseBody: []byte(`{"tags":["foo","bar"]}`),
			wantErr:      false,
			want:         map[string]metrics.TagMetadata{},
		},
		{
			name: "gcr success",
			responseBody: []byte(`{"child":["hello", "world"],"manifest":{"digest1":{"imageSizeBytes":"1",
			"mediaType":"mainstream","timeCreatedms":"1","timeUploadedMs":"2","tag":["foo"]},
			"digest2":{"imageSizeBytes":"2","mediaType":"indie","timeCreatedMs":"3","timeUploadedMs":"4",
			"tag":["bar","baz"]}},"tags":["foo","bar","baz"]}`),
			wantErr: false,
			want: map[string]metrics.TagMetadata{
				"bar": {
					Created:  time.Date(1970, time.January, 1, 1, 0, 0, 3000000, time.Local),
					Uploaded: time.Date(1970, time.January, 1, 1, 0, 0, 4000000, time.Local),
				},
				"baz": {
					Created:  time.Date(1970, time.January, 1, 1, 0, 0, 3000000, time.Local),
					Uploaded: time.Date(1970, time.January, 1, 1, 0, 0, 4000000, time.Local),
				},
				"foo": {
					Created:  time.Date(1970, time.January, 1, 1, 0, 0, 1000000, time.Local),
					Uploaded: time.Date(1970, time.January, 1, 1, 0, 0, 2000000, time.Local),
				},
			},
		},
		{
			name:         "just children",
			responseBody: []byte(`{"child":["hello", "world"]}`),
			wantErr:      false,
			want:         map[string]metrics.TagMetadata{},
		},
		{
			name:         "not json",
			responseBody: []byte("notjson"),
			want:         map[string]metrics.TagMetadata{},
			wantErr:      true,
		},
	}

	for _, tc := range cases {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			gProv := google.New()
			repoName := "ubuntu"
			tagsPath := fmt.Sprintf("/v2/%s/tags/list", repoName)
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				switch request.URL.Path {
				case "/v2/":
					writer.WriteHeader(http.StatusOK)
				case tagsPath:
					if request.Method != http.MethodGet {
						t.Errorf("Method; got %v, want %v", request.Method, http.MethodGet)
					}

					writer.Write(testCase.responseBody) //nolint:errcheck
				default:
					t.Fatalf("Unexpected path: %v", request.URL.Path)
				}
			}))
			defer server.Close()
			u, err := url.Parse(server.URL)
			if err != nil {
				t.Fatalf("url.Parse(%v) = %v", server.URL, err)
			}

			tags, err := gProv.ListImageTag(fmt.Sprintf("%s/%s", u.Host, repoName))
			if (err != nil) != testCase.wantErr {
				t.Errorf("List() wrong error: %v, want %v: %v\n", err != nil, testCase.wantErr, err)
			}
			assert.Equal(t, testCase.want, tags)
		})
	}
}
