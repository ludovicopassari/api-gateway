package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
)

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		RequestsCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gateway_http_requests_total",
				Help: "Number of recived requests",
			},
			[]string{"method", "path", "status"},
		),
		HttpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "gateway_http_request_duration_seconds",
				Help:    "HTTP requests duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
	}

	reg.MustRegister(m.RequestsCount)
	reg.MustRegister(m.HttpRequestDuration)

	return m
}
