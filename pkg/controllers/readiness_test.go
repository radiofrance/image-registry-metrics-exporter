package controllers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/radiofrance/image-registry-metrics-exporter/pkg/controllers"

	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest // Tests fail randomly when parallelized
func TestNotReadyIfNotSet(t *testing.T) {
	t.Run("returns 500 initially", func(t *testing.T) {
		request, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/ready", nil)
		response := httptest.NewRecorder()

		controllers.Ready(response, request)

		assert.EqualValues(t, http.StatusInternalServerError, response.Code)
	})
}

//nolint:paralleltest // Tests fail randomly when parallelized
func TestReadyIfSet(t *testing.T) {
	t.Run("returns 200 when set to ready", func(t *testing.T) {
		request, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/ready", nil)
		response := httptest.NewRecorder()

		controllers.UpdateReady(true)
		controllers.Ready(response, request)

		assert.EqualValues(t, http.StatusOK, response.Code)
	})
}
