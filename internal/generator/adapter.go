package generator

import (
	"context"
	"net/http"

	"github.com/eastnine90/gbgen/internal/growthbookapi"
)

// apiClientAdapter adapts the generated GrowthBook client to generator interfaces.
type apiClientAdapter struct {
	c *growthbookapi.APIClient
}

func (a apiClientAdapter) Features() FeaturesAPI {
	return featuresAPIAdapter{svc: a.c.FeaturesAPI}
}

type featuresAPIAdapter struct {
	svc *growthbookapi.FeaturesAPIService
}

func (a featuresAPIAdapter) ListFeatures(ctx context.Context, limit int32, offset int32, projectID *string) (*growthbookapi.ListFeatures200Response, *http.Response, error) {
	req := a.svc.ListFeatures(ctx).Limit(limit).Offset(offset)
	if projectID != nil && *projectID != "" {
		req = req.ProjectId(*projectID)
	}
	return req.Execute()
}

func (a featuresAPIAdapter) GetFeatureKeys(ctx context.Context, projectID *string) ([]string, *http.Response, error) {
	req := a.svc.GetFeatureKeys(ctx)
	if projectID != nil && *projectID != "" {
		req = req.ProjectId(*projectID)
	}
	return req.Execute()
}

func (a featuresAPIAdapter) GetFeature(ctx context.Context, id string) (*growthbookapi.GetFeature200Response, *http.Response, error) {
	return a.svc.GetFeature(ctx, id).Execute()
}
