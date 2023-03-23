package controllers

import (
	"net/http"
	"sync"
)

const (
	HealthyMessage   = "OK"
	UnhealthyMessage = "Unhealthy"
)

var (
	mutexHealthCheck sync.RWMutex
	healthy          = false
)

func UpdateHealth(isHealthy bool) {
	mutexHealthCheck.Lock()
	healthy = isHealthy
	mutexHealthCheck.Unlock()
}

func HealthCheck(writer http.ResponseWriter, _ *http.Request) {
	mutexHealthCheck.RLock()
	isHealthy := healthy
	mutexHealthCheck.RUnlock()
	if isHealthy {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte(HealthyMessage))
	} else {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(UnhealthyMessage))
	}
}
