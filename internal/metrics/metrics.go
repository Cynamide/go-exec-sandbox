package metrics

import (
	"sync/atomic"
)

type Metrics struct {
	TotalRequests uint64 `json:"total_requests"`
	TotalErrors   uint64 `json:"total_errors"`
}

var (
	globalMetrics = &Metrics{}
)

func IncrementRequest() {
	atomic.AddUint64(&globalMetrics.TotalRequests, 1)
}

func IncrementError() {
	atomic.AddUint64(&globalMetrics.TotalErrors, 1)
}

func GetMetrics() Metrics {
	return Metrics{
		TotalRequests: atomic.LoadUint64(&globalMetrics.TotalRequests),
		TotalErrors:   atomic.LoadUint64(&globalMetrics.TotalErrors),
	}
}
