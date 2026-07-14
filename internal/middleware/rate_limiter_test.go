package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestIPRateLimiterIgnoresForwardedFor(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Minute), 1)

	calls := 0
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRequest(http.MethodGet, "http://example.com/execute", nil)
	first.RemoteAddr = "203.0.113.10:12345"
	first.Header.Set("X-Forwarded-For", "198.51.100.1")

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, first)

	if rr1.Code != http.StatusNoContent {
		t.Fatalf("first request status = %d, want %d", rr1.Code, http.StatusNoContent)
	}

	second := httptest.NewRequest(http.MethodGet, "http://example.com/execute", nil)
	second.RemoteAddr = "203.0.113.10:12345"
	second.Header.Set("X-Forwarded-For", "198.51.100.2")

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, second)

	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want %d", rr2.Code, http.StatusTooManyRequests)
	}

	if calls != 1 {
		t.Fatalf("handler called %d times, want 1", calls)
	}
}
