package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/benchmark"
	"gexec-sandbox/internal/config"
	"gexec-sandbox/internal/httpapi"
	"gexec-sandbox/internal/llm"
	"gexec-sandbox/internal/manifest"
	"gexec-sandbox/internal/metrics"
	"gexec-sandbox/internal/middleware"
	"gexec-sandbox/internal/modeladapter"
	"gexec-sandbox/internal/sandbox"
	"golang.org/x/time/rate"
)

func executeHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics.IncrementRequest()

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			metrics.IncrementError()
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			metrics.IncrementError()
			return
		}

		var req api.ExecutionRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			metrics.IncrementError()
			return
		}

		if req.SourceCode == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ExecutionResponse{
				Error: "source_code cannot be empty",
			})
			metrics.IncrementError()
			return
		}

		if req.TimeoutMS == 0 {
			req.TimeoutMS = cfg.DefaultTimeoutMS
		}

		response, err := sandbox.RunCodeInSandbox(r.Context(), req, cfg)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(api.ExecutionResponse{
				Error: err.Error(),
			})
			metrics.IncrementError()
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func buildMux(cfg config.Config, benchmarkService benchmark.BenchmarkServiceAPI) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics.GetMetrics())
	})

	mux.Handle("/benchmark/run", httpapi.BenchmarkRunHandler{
		Service: benchmarkService,
	})

	mux.Handle("/execute", middleware.RateLimitMiddleware(rate.Every(6*time.Second), 10)(http.HandlerFunc(executeHandler(cfg))))

	return mux
}

func newBenchmarkService(loaded manifest.Loaded) (benchmark.BenchmarkService, error) {
	if len(loaded.Models) == 0 {
		return benchmark.BenchmarkService{}, fmt.Errorf("create benchmark service: at least one enabled model is required")
	}

	models := make([]benchmark.ModelClient, 0, len(loaded.Models))
	for _, modelConfig := range loaded.Models {
		adapter, err := modeladapter.New(modelConfig)
		if err != nil {
			return benchmark.BenchmarkService{}, fmt.Errorf("create adapter for model %q: %w", modelConfig.ID, err)
		}
		client, err := llm.NewClientWithAdapter(adapter)
		if err != nil {
			return benchmark.BenchmarkService{}, fmt.Errorf("create benchmark llm client for model %q: %w", modelConfig.ID, err)
		}
		models = append(models, benchmark.ModelClient{ID: modelConfig.ID, Client: client})
	}

	return benchmark.BenchmarkService{
		Tasks:     loaded.Tasks,
		Scaffolds: loaded.Scaffolds,
		Models:    models,
		Executor:  benchmark.NewCodeExecutionAdapter(),
		Grader:    benchmark.DefaultGrader{},
		Config:    loaded.Runtime,
	}, nil
}

func loadBenchmarkManifest() (manifest.Loaded, error) {
	path, err := repoDataPath("benchmark.yaml")
	if err != nil {
		return manifest.Loaded{}, err
	}

	loaded, err := manifest.Load(path)
	if err != nil {
		return manifest.Loaded{}, fmt.Errorf("load benchmark manifest: %w", err)
	}

	return loaded, nil
}

func loadBenchmarkCatalogs() (benchmark.TaskCatalog, benchmark.ScaffoldCatalog, error) {
	tasksPath, err := benchmarkDataPath("tasks.json")
	if err != nil {
		return benchmark.TaskCatalog{}, benchmark.ScaffoldCatalog{}, err
	}

	scaffoldsPath, err := benchmarkDataPath("scaffolds.json")
	if err != nil {
		return benchmark.TaskCatalog{}, benchmark.ScaffoldCatalog{}, err
	}

	tasks, err := benchmark.LoadTaskCatalog(tasksPath)
	if err != nil {
		return benchmark.TaskCatalog{}, benchmark.ScaffoldCatalog{}, fmt.Errorf("load benchmark tasks: %w", err)
	}

	scaffolds, err := benchmark.LoadScaffoldCatalog(scaffoldsPath)
	if err != nil {
		return benchmark.TaskCatalog{}, benchmark.ScaffoldCatalog{}, fmt.Errorf("load benchmark scaffolds: %w", err)
	}

	return tasks, scaffolds, nil
}

func benchmarkDataPath(name string) (string, error) {
	return repoDataPath(filepath.Join("data", name))
}

func repoDataPath(name string) (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve evaluator source path")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	return filepath.Join(repoRoot, name), nil
}

func runBenchmarkCLI(args []string, service benchmark.BenchmarkServiceAPI) (string, error) {
	return runBenchmarkCLIWithContext(context.Background(), args, service)
}

func runBenchmarkCLIWithContext(ctx context.Context, args []string, service benchmark.BenchmarkServiceAPI) (string, error) {
	if len(args) == 0 || args[0] != "benchmark" {
		return "", fmt.Errorf("unsupported command")
	}

	if err := ctx.Err(); err != nil {
		return "", err
	}

	report, err := service.Run(ctx)
	if err != nil {
		return "", err
	}

	raw, err := json.Marshal(report)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func main() {
	loaded, err := loadBenchmarkManifest()
	if err != nil {
		log.Fatalf("Failed to load benchmark manifest: %v", err)
	}
	cfg := loaded.Runtime
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	benchmarkService, err := newBenchmarkService(loaded)
	if err != nil {
		log.Fatalf("Failed to initialize benchmark service: %v", err)
	}
	log.Printf("Initialized %d benchmark model adapter(s)", len(benchmarkService.Models))

	if len(os.Args) > 1 && os.Args[1] == "benchmark" {
		output, err := runBenchmarkCLIWithContext(rootCtx, os.Args[1:], benchmarkService)
		if err != nil {
			log.Fatalf("Failed to run benchmark CLI: %v", err)
		}
		fmt.Println(output)
		return
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: buildMux(cfg, benchmarkService),
	}

	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-rootCtx.Done()

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		sandbox.CleanupAllContainers()
	}

	log.Println("Server exited")
}
