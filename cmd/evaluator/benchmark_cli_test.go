package main

import (
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
