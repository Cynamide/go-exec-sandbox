package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gexec-sandbox/internal/api"
	"gexec-sandbox/internal/config"
	"gexec-sandbox/internal/metrics"
	"gexec-sandbox/internal/sandbox"
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

func main() {
	cfg := config.LoadConfig()

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

	mux.HandleFunc("/execute", executeHandler(cfg))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
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
