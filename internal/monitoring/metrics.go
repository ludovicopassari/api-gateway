package monitoring

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	RequestsCount       *prometheus.CounterVec
	HttpRequestDuration *prometheus.HistogramVec
}
