//go:build integration

package integration_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/eastnine90/gbgen/internal/config"
	"github.com/eastnine90/gbgen/internal/generator"
)

func TestGenerateIntegration_KeysOnly_Compiles(t *testing.T) {
	apiBaseURL := getenvOrSkip(t, "GBGEN_API_BASE_URL")
	apiKey := getenvOrSkip(t, "GBGEN_API_KEY")
	projectID := getenvPtr("GBGEN_PROJECT_ID")
	expectFeatureID := os.Getenv("GBGEN_IT_EXPECT_FEATURE_ID")

	root := findRepoRoot(t)
	outDir := mkdirTempDir(t, filepath.Join(root, "tmp"), "gbgen-it-keys-")
	outFile := filepath.Join(outDir, "features.gen.go")

	cfg := config.Config{
		GrowthBook: config.GrowthBookConfig{
			APIBaseURL: apiBaseURL,
			APIKey:     apiKey,
			ProjectID:  projectID,
		},
		Generator: config.GeneratorConfig{
			OutputDir:         outDir,
			PackageName:       "itfeatures",
			EmitTypedFeatures: false,
			EmitFeatureList:   true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gen := generator.NewGenerator(cfg)
	if err := gen.Validate(ctx); err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	src, err := gen.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	out := string(src)

	if !strings.Contains(out, "type FeatureKey string") {
		t.Fatalf("expected keys-only output to contain FeatureKey type\n---\n%s\n---", out)
	}
	if strings.Contains(out, "\"github.com/eastnine90/gbgen/types\"") {
		t.Fatalf("did not expect keys-only output to import types\n---\n%s\n---", out)
	}
	if expectFeatureID != "" && !strings.Contains(out, "\""+expectFeatureID+"\"") {
		t.Fatalf("expected output to contain feature id %q\n---\n%s\n---", expectFeatureID, out)
	}

	if err := os.WriteFile(outFile, src, 0o644); err != nil {
		t.Fatalf("write %s: %v", outFile, err)
	}

	goTestPackageDir(t, root, outDir)
}

func TestGenerateIntegration_Typed_Compiles(t *testing.T) {
	apiBaseURL := getenvOrSkip(t, "GBGEN_API_BASE_URL")
	apiKey := getenvOrSkip(t, "GBGEN_API_KEY")
	projectID := getenvPtr("GBGEN_PROJECT_ID")
	expectFeatureID := os.Getenv("GBGEN_IT_EXPECT_FEATURE_ID")

	root := findRepoRoot(t)
	outDir := mkdirTempDir(t, filepath.Join(root, "tmp"), "gbgen-it-typed-")
	outFile := filepath.Join(outDir, "features.gen.go")

	cfg := config.Config{
		GrowthBook: config.GrowthBookConfig{
			APIBaseURL: apiBaseURL,
			APIKey:     apiKey,
			ProjectID:  projectID,
		},
		Generator: config.GeneratorConfig{
			OutputDir:         outDir,
			PackageName:       "itfeatures",
			EmitTypedFeatures: true,
			EmitFeatureList:   true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gen := generator.NewGenerator(cfg)
	if err := gen.Validate(ctx); err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	src, err := gen.Generate(ctx)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	out := string(src)

	if !strings.Contains(out, "\"github.com/eastnine90/gbgen/types\"") {
		t.Fatalf("expected typed output to import types\n---\n%s\n---", out)
	}
	if expectFeatureID != "" && !strings.Contains(out, "\""+expectFeatureID+"\"") {
		t.Fatalf("expected output to contain feature id %q\n---\n%s\n---", expectFeatureID, out)
	}

	if err := os.WriteFile(outFile, src, 0o644); err != nil {
		t.Fatalf("write %s: %v", outFile, err)
	}

	goTestPackageDir(t, root, outDir)
}

func getenvOrSkip(t *testing.T, key string) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		t.Skipf("skipping integration test: missing %s", key)
	}
	return v
}

func getenvPtr(key string) *string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return nil
	}
	return &v
}

func findRepoRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	dir := filepath.Dir(thisFile)

	for i := 0; i < 8; i++ {
		candidate := dir
		if _, err := os.Stat(filepath.Join(candidate, "go.mod")); err == nil {
			return candidate
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	t.Fatalf("could not locate repo root (go.mod) from %s", thisFile)
	return ""
}

func mkdirTempDir(t *testing.T, baseDir, prefix string) string {
	t.Helper()
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", baseDir, err)
	}
	dir, err := os.MkdirTemp(baseDir, prefix)
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func goTestPackageDir(t *testing.T, repoRoot, pkgDir string) {
	t.Helper()
	rel, err := filepath.Rel(repoRoot, pkgDir)
	if err != nil {
		t.Fatalf("rel: %v", err)
	}

	cmd := exec.Command("go", "test", "./"+filepath.ToSlash(rel))
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test %s failed: %v\n%s", rel, err, string(out))
	}
}
