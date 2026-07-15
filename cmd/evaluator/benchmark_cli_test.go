package main

import (
	"context"
	"strings"
	"testing"

	"gexec-sandbox/internal/benchmark"
)

func TestBenchmarkCLIPrintsJSON(t *testing.T) {
	out, err := runBenchmarkCLI([]string{"benchmark"}, &fakeBenchmarkService{report: benchmark.BenchmarkReport{TotalTasks: 1}})
	if err != nil {
		t.Fatalf("runBenchmarkCLI() error = %v", err)
	}
	if !strings.Contains(out, "\"total_tasks\"") {
		t.Fatalf("CLI output = %s, want benchmark JSON", out)
	}
}

func TestBenchmarkCLIRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	service := &fakeBenchmarkService{report: benchmark.BenchmarkReport{TotalTasks: 1}}

	out, err := runBenchmarkCLIWithContext(ctx, []string{"benchmark"}, service)
	if err == nil {
		t.Fatal("runBenchmarkCLIWithContext() error = nil, want context cancellation")
	}

	if out != "" {
		t.Fatalf("CLI output = %q, want empty output on cancellation", out)
	}

	if service.calls != 0 {
		t.Fatalf("service calls = %d, want 0", service.calls)
	}
}
