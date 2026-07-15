package metrics

import "testing"

func TestMetricsCountersIncrementAndReadBack(t *testing.T) {
	original := globalMetrics
	globalMetrics = &Metrics{}
	t.Cleanup(func() {
		globalMetrics = original
	})

	IncrementRequest()
	IncrementRequest()
	IncrementError()

	got := GetMetrics()

	if got.TotalRequests != 2 {
		t.Fatalf("TotalRequests = %d, want 2", got.TotalRequests)
	}

	if got.TotalErrors != 1 {
		t.Fatalf("TotalErrors = %d, want 1", got.TotalErrors)
	}
}
