package metrics

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartRecordingMetrics(h *mux.Router) {
	h.Handle("/metrics", promhttp.Handler())
	return
}
