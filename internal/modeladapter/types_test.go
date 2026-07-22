package modeladapter_test

import (
	"testing"

	"gexec-sandbox/internal/modeladapter"
)

func TestConfigRequiresModelName(t *testing.T) {
	cfg := modeladapter.Config{ID: "local"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing model name error")
	}
}
