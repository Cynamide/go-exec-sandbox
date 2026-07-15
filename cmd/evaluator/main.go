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
	"gexec-sandbox/internal/metrics"
	"gexec-sandbox/internal/middleware"
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

		response, err := sandbox.RunCodeInSandbox(req, cfg)
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

func newBenchmarkService(cfg config.Config) (benchmark.BenchmarkService, error) {
	tasks, scaffolds, err := loadBenchmarkCatalogs()
	if err != nil {
		return benchmark.BenchmarkService{}, err
	}

	client, err := llm.NewClient()
	if err != nil {
		return benchmark.BenchmarkService{}, fmt.Errorf("create benchmark llm client: %w", err)
	}

	return benchmark.BenchmarkService{
		Tasks:     tasks,
		Scaffolds: scaffolds,
		Client:    client,
		Executor:  benchmark.NewCodeExecutionAdapter(),
		Grader:    benchmark.DefaultGrader{},
		Config:    cfg,
	}, nil
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

	taskPayload, err := os.ReadFile(tasksPath)
	if err != nil {
		return benchmark.TaskCatalog{}, benchmark.ScaffoldCatalog{}, fmt.Errorf("read benchmark tasks: %w", err)
	}

	scaffoldPayload, err := os.ReadFile(scaffoldsPath)
	if err != nil {
		return benchmark.TaskCatalog{}, benchmark.ScaffoldCatalog{}, fmt.Errorf("read benchmark scaffolds: %w", err)
	}

	var tasks benchmark.TaskCatalog
	if err := json.Unmarshal(taskPayload, &tasks); err != nil {
		return benchmark.TaskCatalog{}, benchmark.ScaffoldCatalog{}, fmt.Errorf("decode benchmark tasks: %w", err)
	}

	var scaffolds benchmark.ScaffoldCatalog
	if err := json.Unmarshal(scaffoldPayload, &scaffolds); err != nil {
		return benchmark.TaskCatalog{}, benchmark.ScaffoldCatalog{}, fmt.Errorf("decode benchmark scaffolds: %w", err)
	}

	return tasks, scaffolds, nil
}

func benchmarkDataPath(name string) (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve evaluator source path")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	return filepath.Join(repoRoot, "data", name), nil
}

func runBenchmarkCLI(args []string, service benchmark.BenchmarkServiceAPI) (string, error) {
	if len(args) == 0 || args[0] != "benchmark" {
		return "", fmt.Errorf("unsupported command")
	}

	report, err := service.Run(context.Background())
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
	cfg := config.LoadConfig()

	log.Println("Checking Ollama availability...")
	if err := llm.WaitForOllama(); err != nil {
		log.Fatalf("Failed to connect to Ollama: %v", err)
	}

	log.Printf("Model %s should be available from Ollama service", cfg.OLLAMAModel)

	benchmarkService, err := newBenchmarkService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize benchmark service: %v", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "benchmark" {
		output, err := runBenchmarkCLI(os.Args[1:], benchmarkService)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		sandbox.CleanupAllContainers()
	}

	log.Println("Server exited")
}
