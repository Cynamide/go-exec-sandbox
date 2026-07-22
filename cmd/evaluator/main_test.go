package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gexec-sandbox/internal/benchmark"
	"gexec-sandbox/internal/config"
	"gexec-sandbox/internal/manifest"
	"gexec-sandbox/internal/modeladapter"
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

func TestLoadBenchmarkManifestUsesReusableFixtureFiles(t *testing.T) {
	loaded, err := loadBenchmarkManifest()
	if err != nil {
		t.Fatalf("loadBenchmarkManifest() error = %v", err)
	}
	tasks := loaded.Tasks
	scaffolds := loaded.Scaffolds

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

	if loaded.Runtime.OLLAMAModel == "" {
		t.Fatal("manifest runtime missing OLLAMAModel")
	}

	service, err := newBenchmarkService(loaded)
	if err != nil {
		t.Fatalf("newBenchmarkService(repository manifest) error = %v", err)
	}
	if len(service.Models) != len(loaded.Models) {
		t.Fatalf("service model count = %d, want %d", len(service.Models), len(loaded.Models))
	}
}

func TestNewBenchmarkServiceConstructsClientsFromLoadedModelsWithoutOllama(t *testing.T) {
	loaded := manifest.Loaded{
		Models: []modeladapter.Config{{
			ID:           "remote",
			ProviderKind: "openai_compatible",
			ModelName:    "remote-model",
			BaseURL:      "https://models.example.test/v1",
		}},
		DefaultModelRoles: map[string]string{"planner": "remote"},
	}

	service, err := newBenchmarkService(loaded)
	if err != nil {
		t.Fatalf("newBenchmarkService() error = %v", err)
	}
	if len(service.Models) != 1 {
		t.Fatalf("len(service.Models) = %d, want 1", len(service.Models))
	}
	if service.Models[0].ID != "remote" {
		t.Fatalf("service.Models[0].ID = %q, want remote", service.Models[0].ID)
	}
	if got := service.DefaultModelRoles["planner"]; got != "remote" {
		t.Fatalf("service.DefaultModelRoles[planner] = %q, want remote", got)
	}
}

func TestNewBenchmarkServiceRejectsInvalidOpenAIModelWithoutConstructingOllama(t *testing.T) {
	const apiKeyEnv = "EVALUATOR_TEST_MISSING_OPENAI_KEY"
	t.Setenv(apiKeyEnv, "")

	_, err := newBenchmarkService(manifest.Loaded{Models: []modeladapter.Config{{
		ID:           "remote",
		ProviderKind: "openai_compatible",
		ModelName:    "remote-model",
		BaseURL:      "https://models.example.test/v1",
		APIKeyEnv:    apiKeyEnv,
	}}})
	if err == nil {
		t.Fatal("newBenchmarkService() error = nil, want missing API key error")
	}
	if !strings.Contains(err.Error(), apiKeyEnv) {
		t.Fatalf("newBenchmarkService() error = %v, want missing key env name", err)
	}
}

func TestNewBenchmarkServiceRejectsNoEnabledModels(t *testing.T) {
	_, err := newBenchmarkService(manifest.Loaded{})
	if err == nil || !strings.Contains(err.Error(), "enabled model") {
		t.Fatalf("newBenchmarkService() error = %v, want enabled model error", err)
	}
}

func TestNewBenchmarkServiceConstructsBeforeSeparateModelHealthCheck(t *testing.T) {
	service, err := newBenchmarkService(manifest.Loaded{Models: []modeladapter.Config{{
		ID:           "remote",
		ProviderKind: "openai_compatible",
		ModelName:    "remote-model",
		BaseURL:      "http://model-health.test/v1",
	}}})
	if err != nil {
		t.Fatalf("newBenchmarkService() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if service.Models[0].HealthCheck == nil {
		t.Fatal("constructed model is missing health check callback")
	}
	if err := service.Models[0].HealthCheck(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("model HealthCheck() error = %v, want context canceled", err)
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
