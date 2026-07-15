package httpapi

import (
	"encoding/json"
	"net/http"

	"gexec-sandbox/internal/benchmark"
)

type BenchmarkRunHandler struct {
	Service benchmark.BenchmarkServiceAPI
}

func (h BenchmarkRunHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	report, err := h.Service.Run(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
