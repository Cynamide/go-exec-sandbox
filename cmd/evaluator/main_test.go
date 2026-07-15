package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestLoadBenchmarkCatalogsUsesReusableFixtureFiles(t *testing.T) {
	tasks, scaffolds, err := loadBenchmarkCatalogs()
	if err != nil {
		t.Fatalf("loadBenchmarkCatalogs() error = %v", err)
	}

	if len(tasks.Tasks) == 0 {
		t.Fatal("tasks.Tasks is empty")
	}

	if tasks.Tasks[0].TaskFamily == "" {
		t.Fatal("first task missing TaskFamily")
	}

	var baselineCount int
	for _, scaffold := range scaffolds.Scaffolds {
		if scaffold.Baseline {
			baselineCount++
		}
	}

	if baselineCount != 1 {
		t.Fatalf("baseline scaffold count = %d, want 1", baselineCount)
	}

	problemsPath, err := benchmarkDataPath("problems.json")
	if err != nil {
		t.Fatalf("benchmarkProblemsPath() error = %v", err)
	}

	rawProblems, err := os.ReadFile(problemsPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", problemsPath, err)
	}

	var toyProblems []benchmark.Task
	if err := json.Unmarshal(rawProblems, &toyProblems); err != nil {
		t.Fatalf("json.Unmarshal(%q) error = %v", filepath.Base(problemsPath), err)
	}

	if len(toyProblems) == 0 {
		t.Fatal("toy problems fixture is empty")
	}

	if tasks.Tasks[0].ID == toyProblems[0].ID {
		t.Fatalf("first benchmark task ID = %q, unexpectedly matches toy problem fixture", tasks.Tasks[0].ID)
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
