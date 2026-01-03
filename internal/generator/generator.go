package generator

import (
	"context"
	"strings"

	"github.com/eastnine90/gbgen/internal/config"
	"github.com/eastnine90/gbgen/internal/growthbookapi"
)

type Generator struct {
	api    FeaturesAPI
	config config.Config
}

func NewGenerator(cfg config.Config) *Generator {
	clientConfig := growthbookapi.NewConfiguration()
	// Apply server URL from config.
	// Users provide apiBaseURL like "https://api.growthbook.io" (we append /api/v1),
	// but if they already include "/api/v1" we keep it as-is.
	base := strings.TrimRight(cfg.GrowthBook.APIBaseURL, "/")
	if !strings.HasSuffix(base, "/api/v1") {
		base = base + "/api/v1"
	}
	clientConfig.Servers = growthbookapi.ServerConfigurations{
		{URL: base},
	}
	client := growthbookapi.NewAPIClient(clientConfig)

	return &Generator{
		api:    featuresAPIAdapter{svc: client.FeaturesAPI},
		config: cfg,
	}
}

func (g *Generator) Validate(ctx context.Context) error {
	ctx = context.WithValue(ctx, growthbookapi.ContextAccessToken, g.config.GrowthBook.APIKey)

	// One cheap request to verify base URL + auth + (optional) project filter.
	_, _, err := g.api.GetFeatureKeys(ctx, g.config.GrowthBook.ProjectID)
	return err
}

func (g *Generator) Generate(ctx context.Context) ([]byte, error) {
	ctx = context.WithValue(ctx, growthbookapi.ContextAccessToken, g.config.GrowthBook.APIKey)

	features, err := g.fetchAllFeatureMeta(ctx)
	if err != nil {
		return nil, err
	}

	var src []byte
	if g.config.Generator.EmitTypedFeatures {
		src, err = renderTypedFeaturesGo(g.config.Generator.PackageName, features, g.config.Generator.EmitFeatureList)
	} else {
		src, err = renderFeatureKeysGo(g.config.Generator.PackageName, features, g.config.Generator.EmitFeatureList)
	}

	return src, err
}
