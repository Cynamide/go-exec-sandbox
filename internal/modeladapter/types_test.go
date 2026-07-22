package modeladapter_test

import (
	"strings"
	"testing"

	"gexec-sandbox/internal/modeladapter"
)

func TestConfigRequiresModelName(t *testing.T) {
	cfg := modeladapter.Config{ID: "local"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing model name error")
	}
}

func TestConfigRequiresID(t *testing.T) {
	cfg := modeladapter.Config{ModelName: "gpt-4.1"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing ID error")
	}
}

func TestCustomHTTPProviderRequiresMappings(t *testing.T) {
	cfg := modeladapter.Config{
		ID:           "custom",
		ProviderKind: "custom_http",
		ModelName:    "custom-model",
		BaseURL:      "http://localhost:8080",
	}

	if err := modeladapter.ValidateMappings(cfg); err == nil {
		t.Fatal("ValidateMappings() error = nil, want missing mapping error")
	}
}

func TestCustomHTTPProviderMissingMappingsErrorListsFields(t *testing.T) {
	cfg := modeladapter.Config{
		ID:           "custom",
		ProviderKind: "custom_http",
		ModelName:    "custom-model",
		RequestMapping: modeladapter.RequestMapping{
			Method: "POST",
		},
	}

	err := modeladapter.ValidateMappings(cfg)
	if err == nil {
		t.Fatal("ValidateMappings() error = nil, want missing mapping error")
	}

	want := `model adapter config "custom" custom_http provider requires request_mapping.path, request_mapping.body_template, response_mapping.text_path`
	if err.Error() != want {
		t.Fatalf("ValidateMappings() error = %q, want %q", err.Error(), want)
	}
	if strings.Contains(err.Error(), "%!s") {
		t.Fatalf("ValidateMappings() error = %q, want clean formatting", err.Error())
	}
}
