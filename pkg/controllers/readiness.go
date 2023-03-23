package controllers

import (
	"net/http"
	"sync"
)

var (
	mutexReadiness sync.RWMutex
	ready          = false
)

func UpdateReady(isReady bool) {
	mutexReadiness.Lock()
	ready = isReady
	mutexReadiness.Unlock()
}

func Ready(writer http.ResponseWriter, _ *http.Request) {
	mutexReadiness.RLock()
	isReady := ready
	mutexReadiness.RUnlock()
	if isReady {
		writer.WriteHeader(http.StatusOK)
	} else {
		writer.WriteHeader(http.StatusInternalServerError)
	}
}
