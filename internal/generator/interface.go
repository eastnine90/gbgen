package generator

import (
	"context"
	"net/http"

	"github.com/eastnine90/gbgen/internal/growthbookapi"
)

// APIClient is a narrow interface used by the generator package so we can mock GrowthBook calls in tests.
//
// Note: this is intentionally NOT identical to the concrete openapi-generator `growthbookapi.APIClient`.
// It's the minimal shape the generator logic should depend on.
type APIClient interface {
	Features() FeaturesAPI
}

// FeaturesAPI is the subset of the GrowthBook Features API used by gbgen.
// Implemented by a thin adapter around `*growthbookapi.FeaturesAPIService` in production, and mocks in tests.
type FeaturesAPI interface {
	// ListFeatures corresponds to GET /features.
	ListFeatures(ctx context.Context, limit int32, offset int32, projectID *string) (*growthbookapi.ListFeatures200Response, *http.Response, error)

	// GetFeatureKeys corresponds to GET /feature-keys.
	GetFeatureKeys(ctx context.Context, projectID *string) ([]string, *http.Response, error)

	// GetFeature corresponds to GET /features/{id}.
	GetFeature(ctx context.Context, id string) (*growthbookapi.GetFeature200Response, *http.Response, error)
}
