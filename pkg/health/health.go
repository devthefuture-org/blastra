package health

import (
	"net/http"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

type HealthChecker struct {
	ready atomic.Value
}

func NewHealthChecker() *HealthChecker {
	log.Debug("Initializing health checker")
	hc := &HealthChecker{}
	hc.ready.Store(false)
	return hc
}

func (hc *HealthChecker) SetReady() {
	hc.ready.Store(true)
	log.Info("Service marked as ready")
}

func (hc *HealthChecker) IsReady() bool {
	return hc.ready.Load().(bool)
}

func (hc *HealthChecker) LivenessProbeHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Liveness probe request received")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (hc *HealthChecker) ReadinessProbeHandler(w http.ResponseWriter, r *http.Request) {
	log.Debug("Readiness probe request received")
	if hc.IsReady() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
		return
	}
	log.Debug("Service not ready yet")
	http.Error(w, "Not Ready", http.StatusServiceUnavailable)
}
