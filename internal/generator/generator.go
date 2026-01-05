package generator

import (
	"context"
	"net/http"
	"strings"

	"github.com/eastnine90/gbgen/internal/config"
	"github.com/eastnine90/gbgen/internal/growthbookapi"
)

type Generator struct {
	api    growthbookapi.ClientWithResponsesInterface
	config config.Config
}

func NewGenerator(cfg config.Config) (*Generator, error) {
	base := strings.TrimRight(cfg.GrowthBook.APIBaseURL, "/")
	if !strings.HasSuffix(base, "/api/v1") {
		base = base + "/api/v1"
	}

	api, err := growthbookapi.NewClientWithResponses(base, growthbookapi.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+cfg.GrowthBook.APIKey)
		return nil
	}))

	if err != nil {
		return nil, err
	}

	return &Generator{
		api:    api,
		config: cfg,
	}, nil
}

func (g *Generator) Generate(ctx context.Context) ([]byte, error) {
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
