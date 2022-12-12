package metrics

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartRecordingMetrics(h *mux.Router) {
	// reg := prometheus.NewRegistry()
	h.Handle("/metrics", promhttp.Handler())
	// h.Handle("/api/metrics", promhttp.HandlerFor(
	// 	reg,
	// 	promhttp.HandlerOpts{
	// 		Registry: reg,
	// 	},
	// ))
	return
}
