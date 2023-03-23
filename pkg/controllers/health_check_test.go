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
func TestUnhealthyIfNotSet(t *testing.T) {
	t.Run("healthcheck returns 500 initially", func(t *testing.T) {
		request, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil)
		response := httptest.NewRecorder()

		controllers.HealthCheck(response, request)

		assert.EqualValues(t, http.StatusInternalServerError, response.Code)
		assert.EqualValues(t, controllers.UnhealthyMessage, response.Body.String())
	})
}

//nolint:paralleltest // Tests fail randomly when parallelized
func TestHealthCheckIfSetHealthy(t *testing.T) {
	t.Run("healthcheck returns 200 if set healthy", func(t *testing.T) {
		request, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil)
		response := httptest.NewRecorder()

		controllers.UpdateHealth(true)
		controllers.HealthCheck(response, request)

		assert.EqualValues(t, http.StatusOK, response.Code)
		assert.EqualValues(t, controllers.HealthyMessage, response.Body.String())
	})
}

//nolint:paralleltest // Tests fail randomly when parallelized
func TestHealthCheckIfSetNotHealthy(t *testing.T) {
	t.Run("healthcheck returns 500 if set not healthy", func(t *testing.T) {
		request, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil)
		response := httptest.NewRecorder()

		controllers.UpdateHealth(false)
		controllers.HealthCheck(response, request)

		assert.EqualValues(t, http.StatusInternalServerError, response.Code)
		assert.EqualValues(t, controllers.UnhealthyMessage, response.Body.String())
	})
}
