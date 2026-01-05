package generator

import (
	"context"
	"go/format"
	"strings"
	"testing"

	"github.com/eastnine90/gbgen/internal/config"
	"github.com/eastnine90/gbgen/internal/growthbookapi"
)

var _ growthbookapi.ClientWithResponsesInterface = (*mockFeaturesAPI)(nil)

type mockFeaturesAPI struct {
	t *testing.T

	listCalls []listCall
	keysCalls []keysCall

	featuresRespByOffset map[int32]*growthbookapi.ListFeaturesResponse
}

func (m *mockFeaturesAPI) ListFeaturesWithResponse(ctx context.Context, params *growthbookapi.ListFeaturesParams, reqEditors ...growthbookapi.RequestEditorFn) (*growthbookapi.ListFeaturesResponse, error) {
	m.listCalls = append(m.listCalls, listCall{limit: params.Limit, offset: params.Offset, projectID: params.ProjectId})

	var offset int32
	if params.Offset != nil {
		offset = int32(*params.Offset)
	}
	resp := m.featuresRespByOffset[offset]
	if resp == nil {
		m.t.Fatalf("unexpected offset %d", offset)
	}
	return resp, nil
}

type listCall struct {
	limit     *int
	offset    *int
	projectID *string
}

type keysCall struct {
	projectID *string
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
		featuresRespByOffset: map[int32]*growthbookapi.ListFeaturesResponse{
			0: {
				JSON200: &struct {
					Count      int                     `json:"count"`
					Features   []growthbookapi.Feature `json:"features"`
					HasMore    bool                    `json:"hasMore"`
					Limit      int                     `json:"limit"`
					NextOffset *int                    `json:"nextOffset"`
					Offset     int                     `json:"offset"`
					Total      int                     `json:"total"`
				}{
					HasMore: false,
					Features: []growthbookapi.Feature{
						{
							Id:          "checkout-redesign",
							Description: "Checkout redesign flag",
							Environments: map[string]growthbookapi.FeatureEnvironment{
								"production": {Enabled: true},
							},
							ValueType: growthbookapi.Boolean,
						},
						{
							Id:          "disabled-feature",
							Description: "Disabled everywhere",
							Environments: map[string]growthbookapi.FeatureEnvironment{
								"production": {Enabled: false},
							},
							ValueType: growthbookapi.Boolean,
						},
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
	assertGofmtIdempotent(t, src)
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
	if *mock.listCalls[0].limit != 100 || *mock.listCalls[0].offset != 0 {
		t.Fatalf("unexpected list call %+v", mock.listCalls[0])
	}
	if mock.listCalls[0].projectID == nil || *mock.listCalls[0].projectID != project {
		t.Fatalf("expected projectID %q, got %#v", project, mock.listCalls[0].projectID)
	}
}

func TestGeneratorGenerate_KeysOnly_NoFeatureList(t *testing.T) {
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
			EmitFeatureList:   false,
		},
	}

	mock := &mockFeaturesAPI{
		t: t,
		featuresRespByOffset: map[int32]*growthbookapi.ListFeaturesResponse{
			0: {
				JSON200: &struct {
					Count      int                     `json:"count"`
					Features   []growthbookapi.Feature `json:"features"`
					HasMore    bool                    `json:"hasMore"`
					Limit      int                     `json:"limit"`
					NextOffset *int                    `json:"nextOffset"`
					Offset     int                     `json:"offset"`
					Total      int                     `json:"total"`
				}{
					HasMore: false,
					Features: []growthbookapi.Feature{
						{
							Id:          "checkout-redesign",
							Description: "Checkout redesign flag",
							Environments: map[string]growthbookapi.FeatureEnvironment{
								"production": {Enabled: true},
							},
							ValueType: growthbookapi.Boolean,
						},
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
	assertGofmtIdempotent(t, src)
	out := string(src)

	assertContains(t, out, "package features")
	assertContains(t, out, "type FeatureKey string")
	assertContains(t, out, "FeatureCheckoutRedesign FeatureKey = \"checkout-redesign\"")
	assertNotContains(t, out, "FeatureList")
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
		featuresRespByOffset: map[int32]*growthbookapi.ListFeaturesResponse{
			0: {
				JSON200: &struct {
					Count      int                     `json:"count"`
					Features   []growthbookapi.Feature `json:"features"`
					HasMore    bool                    `json:"hasMore"`
					Limit      int                     `json:"limit"`
					NextOffset *int                    `json:"nextOffset"`
					Offset     int                     `json:"offset"`
					Total      int                     `json:"total"`
				}{
					HasMore: false,
					Features: []growthbookapi.Feature{
						{
							Id:          "theme-name",
							Description: "Theme name",
							Environments: map[string]growthbookapi.FeatureEnvironment{
								"production": {Enabled: true},
							},
							ValueType: growthbookapi.String,
						},
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
	assertGofmtIdempotent(t, src)
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

func TestGeneratorGenerate_Typed_NoFeatureList(t *testing.T) {
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
			EmitFeatureList:   false,
		},
	}

	mock := &mockFeaturesAPI{
		t: t,
		featuresRespByOffset: map[int32]*growthbookapi.ListFeaturesResponse{
			0: {
				JSON200: &struct {
					Count      int                     `json:"count"`
					Features   []growthbookapi.Feature `json:"features"`
					HasMore    bool                    `json:"hasMore"`
					Limit      int                     `json:"limit"`
					NextOffset *int                    `json:"nextOffset"`
					Offset     int                     `json:"offset"`
					Total      int                     `json:"total"`
				}{
					HasMore: false,
					Features: []growthbookapi.Feature{
						{
							Id:          "theme-name",
							Description: "Theme name",
							Environments: map[string]growthbookapi.FeatureEnvironment{
								"production": {Enabled: true},
							},
							ValueType: growthbookapi.String,
						},
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
	assertGofmtIdempotent(t, src)
	out := string(src)

	assertContains(t, out, "import (")
	assertContains(t, out, "\"github.com/eastnine90/gbgen/types\"")
	assertContains(t, out, "FeatureThemeName = types.StringFeature(\"theme-name\")")
	assertNotContains(t, out, "FeatureList")
	assertNotContains(t, out, "type FeatureKey string")
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected output to contain %q\n---\n%s\n---", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("expected output to NOT contain %q\n---\n%s\n---", substr, s)
	}
}

func assertGofmtIdempotent(t *testing.T, src []byte) {
	t.Helper()

	formatted, err := format.Source(src)
	if err != nil {
		t.Fatalf("generated output is not valid go/format input: %v", err)
	}
	if string(formatted) != string(src) {
		t.Fatalf("generated output is not gofmt-idempotent (format.Source would change it)\n--- before ---\n%s\n--- after ---\n%s\n", string(src), string(formatted))
	}
}
