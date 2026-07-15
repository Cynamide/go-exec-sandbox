package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gexec-sandbox/internal/benchmark"
)

func TestBenchmarkRunHandlerReturnsJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/benchmark/run", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	handler := BenchmarkRunHandler{
		Service: fakeBenchmarkService{report: benchmark.BenchmarkReport{TotalTasks: 1}},
	}
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "\"total_tasks\":1") {
		t.Fatalf("body = %s, want benchmark report JSON", rr.Body.String())
	}
}

type fakeBenchmarkService struct {
	report benchmark.BenchmarkReport
}

func (f fakeBenchmarkService) Run(ctx context.Context) (benchmark.BenchmarkReport, error) {
	return f.report, nil
}
