package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gexec-sandbox/internal/benchmark"
	"gexec-sandbox/internal/config"
)

func TestBuildMuxRegistersBenchmarkRunRoute(t *testing.T) {
	service := &fakeBenchmarkService{
		report: benchmark.BenchmarkReport{TotalTasks: 1},
	}

	req := httptest.NewRequest(http.MethodPost, "/benchmark/run", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	buildMux(config.Config{}, service).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	if service.calls != 1 {
		t.Fatalf("service calls = %d, want 1", service.calls)
	}

	if !strings.Contains(rr.Body.String(), "\"total_tasks\":1") {
		t.Fatalf("body = %s, want benchmark report JSON", rr.Body.String())
	}
}

type fakeBenchmarkService struct {
	report benchmark.BenchmarkReport
	calls  int
}

func (f *fakeBenchmarkService) Run(ctx context.Context) (benchmark.BenchmarkReport, error) {
	f.calls++
	return f.report, nil
}
