package config

import (
	"strings"
	"testing"
)

func TestConfigValidate_OK(t *testing.T) {
	cfg := Config{
		GrowthBook: GrowthBookConfig{
			APIBaseURL: "https://api.growthbook.io",
			APIKey:     "secret_abc123",
			ProjectID:  nil,
		},
		Generator: GeneratorConfig{
			OutputDir:   "./out",
			PackageName: "growthbooktypes",
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestConfigValidate_Invalid_ReturnsFriendlyMessages(t *testing.T) {
	cfg := Config{
		GrowthBook: GrowthBookConfig{
			APIBaseURL: "not-a-url",
			APIKey:     "",
		},
		Generator: GeneratorConfig{
			OutputDir:   "",
			PackageName: "",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	verr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T (%v)", err, err)
	}

	got := strings.Join(verr.Problems, "\n")
	assertContains(t, got, "growthbook.apiKey is required")
	assertContains(t, got, "growthbook.apiBaseURL must be a valid URL")
	assertContains(t, got, "generator.outputDir is required")
	assertContains(t, got, "generator.packageName is required")
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q to contain %q", s, substr)
	}
}
