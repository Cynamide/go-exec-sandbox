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

var benchmarkScaffolds = benchmark.ScaffoldCatalog{
	Scaffolds: []benchmark.Scaffold{
		{
			Baseline: true,
			Name:     "baseline",
		},
		{
			Name:         "tool-assisted",
			Description:  "Ask for a compact plan before coding and return only executable code.",
			PromptPrefix: "First think through the solution briefly, then write the final code. Return only executable code.\n\n",
		},
	},
}

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
	tasks, err := loadBenchmarkTasks()
	if err != nil {
		return benchmark.BenchmarkService{}, err
	}

	client, err := llm.NewClient()
	if err != nil {
		return benchmark.BenchmarkService{}, fmt.Errorf("create benchmark llm client: %w", err)
	}

	return benchmark.BenchmarkService{
		Tasks:     tasks,
		Scaffolds: benchmarkScaffolds,
		Client:    client,
		Executor:  benchmark.NewCodeExecutionAdapter(),
		Grader:    benchmark.DefaultGrader{},
		Config:    cfg,
	}, nil
}

func loadBenchmarkTasks() (benchmark.TaskCatalog, error) {
	problemsPath, err := benchmarkProblemsPath()
	if err != nil {
		return benchmark.TaskCatalog{}, err
	}

	payload, err := os.ReadFile(problemsPath)
	if err != nil {
		return benchmark.TaskCatalog{}, fmt.Errorf("read benchmark tasks: %w", err)
	}

	var tasks []benchmark.Task
	if err := json.Unmarshal(payload, &tasks); err != nil {
		return benchmark.TaskCatalog{}, fmt.Errorf("decode benchmark tasks: %w", err)
	}

	return benchmark.TaskCatalog{Tasks: tasks}, nil
}

func benchmarkProblemsPath() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve evaluator source path")
	}

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	return filepath.Join(repoRoot, "data", "problems.json"), nil
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
