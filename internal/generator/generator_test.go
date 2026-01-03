package generator

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/eastnine90/gbgen/internal/config"
	"github.com/eastnine90/gbgen/internal/growthbookapi"
)

type mockFeaturesAPI struct {
	t *testing.T

	listCalls []listCall
	keysCalls []keysCall

	featuresRespByOffset map[int32]*growthbookapi.ListFeatures200Response
}

type listCall struct {
	limit     int32
	offset    int32
	projectID *string
}

type keysCall struct {
	projectID *string
}

func (m *mockFeaturesAPI) ListFeatures(ctx context.Context, limit int32, offset int32, projectID *string) (*growthbookapi.ListFeatures200Response, *http.Response, error) {
	m.listCalls = append(m.listCalls, listCall{limit: limit, offset: offset, projectID: projectID})
	resp := m.featuresRespByOffset[offset]
	if resp == nil {
		m.t.Fatalf("unexpected offset %d", offset)
	}
	return resp, nil, nil
}

func (m *mockFeaturesAPI) GetFeatureKeys(ctx context.Context, projectID *string) ([]string, *http.Response, error) {
	m.keysCalls = append(m.keysCalls, keysCall{projectID: projectID})
	return []string{"a", "b"}, nil, nil
}

func (m *mockFeaturesAPI) GetFeature(ctx context.Context, id string) (*growthbookapi.GetFeature200Response, *http.Response, error) {
	m.t.Fatalf("GetFeature should not be called in these tests")
	return nil, nil, nil
}

func TestGeneratorGenerate_KeysOnly(t *testing.T) {
	project := "proj_123"
	cfg := config.Config{
		GrowthBook: config.GrowthBookConfig{
			APIBaseURL: "https://api.growthbook.io",
			APIKey:     "secret_x",
			ProjectID:  &project,
		},
		Generator: config.GeneratorConfig{
			OutputDir:         "./ignored",
			PackageName:       "features",
			EmitTypedFeatures: false,
			EmitFeatureList:   true,
		},
	}

	mock := &mockFeaturesAPI{
		t: t,
		featuresRespByOffset: map[int32]*growthbookapi.ListFeatures200Response{
			0: {
				HasMore:    false,
				NextOffset: 0,
				Features: []growthbookapi.Feature{
					{
						Id:          "checkout-redesign",
						Description: "Checkout redesign flag",
						Environments: map[string]growthbookapi.FeatureEnvironment{
							"production": {Enabled: true},
						},
						ValueType: growthbookapi.FEATUREVALUETYPE_BOOLEAN,
					},
					{
						Id:          "disabled-feature",
						Description: "Disabled everywhere",
						Environments: map[string]growthbookapi.FeatureEnvironment{
							"production": {Enabled: false},
						},
						ValueType: growthbookapi.FEATUREVALUETYPE_BOOLEAN,
					},
				},
			},
		},
	}

	g := &Generator{api: mock, config: cfg}
	src, err := g.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	out := string(src)

	assertContains(t, out, "package features")
	assertContains(t, out, "type FeatureKey string")
	assertContains(t, out, "FeatureCheckoutRedesign FeatureKey = \"checkout-redesign\"")
	assertContains(t, out, "FeatureDisabledFeature FeatureKey = \"disabled-feature\"")
	assertContains(t, out, "var FeatureList = []FeatureKey")
	assertContains(t, out, "// Deprecated: no active environments")

	if len(mock.listCalls) != 1 {
		t.Fatalf("expected 1 list call, got %d", len(mock.listCalls))
	}
	if mock.listCalls[0].limit != 100 || mock.listCalls[0].offset != 0 {
		t.Fatalf("unexpected list call %+v", mock.listCalls[0])
	}
	if mock.listCalls[0].projectID == nil || *mock.listCalls[0].projectID != project {
		t.Fatalf("expected projectID %q, got %#v", project, mock.listCalls[0].projectID)
	}
}

func TestGeneratorGenerate_Typed(t *testing.T) {
	cfg := config.Config{
		GrowthBook: config.GrowthBookConfig{
			APIBaseURL: "https://api.growthbook.io",
			APIKey:     "secret_x",
			ProjectID:  nil,
		},
		Generator: config.GeneratorConfig{
			OutputDir:         "./ignored",
			PackageName:       "features",
			EmitTypedFeatures: true,
			EmitFeatureList:   true,
		},
	}

	mock := &mockFeaturesAPI{
		t: t,
		featuresRespByOffset: map[int32]*growthbookapi.ListFeatures200Response{
			0: {
				HasMore:    false,
				NextOffset: 0,
				Features: []growthbookapi.Feature{
					{
						Id:          "theme-name",
						Description: "Theme name",
						Environments: map[string]growthbookapi.FeatureEnvironment{
							"production": {Enabled: true},
						},
						ValueType: growthbookapi.FEATUREVALUETYPE_STRING,
					},
				},
			},
		},
	}

	g := &Generator{api: mock, config: cfg}
	src, err := g.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	out := string(src)

	assertContains(t, out, "import (")
	assertContains(t, out, "\"github.com/eastnine90/gbgen/types\"")
	assertContains(t, out, "FeatureThemeName = types.StringFeature(\"theme-name\")")
	assertContains(t, out, "var FeatureList")

	if len(mock.listCalls) != 1 {
		t.Fatalf("expected 1 list call, got %d", len(mock.listCalls))
	}
	if mock.listCalls[0].projectID != nil {
		t.Fatalf("expected nil projectID, got %#v", mock.listCalls[0].projectID)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected output to contain %q\n---\n%s\n---", substr, s)
	}
}
