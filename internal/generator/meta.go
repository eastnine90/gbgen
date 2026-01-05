package generator

import (
	"context"
	"fmt"
	"sort"

	"github.com/eastnine90/gbgen/internal/growthbookapi"
)

type featureMeta struct {
	ID           string
	Description  string
	NoActiveEnvs bool
	ValueType    growthbookapi.FeatureValueType
}

type namedFeature struct {
	Name         string
	ID           string
	Description  string
	NoActiveEnvs bool
	ValueType    growthbookapi.FeatureValueType
}

func (g *Generator) fetchAllFeatureMeta(ctx context.Context) ([]featureMeta, error) {
	limit := 100
	offset := 0

	var out []featureMeta
	for {
		resp, err := g.api.ListFeaturesWithResponse(ctx, &growthbookapi.ListFeaturesParams{
			Limit:     &limit,
			Offset:    &offset,
			ProjectId: g.config.GrowthBook.ProjectID,
			ClientKey: nil,
		})
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return nil, fmt.Errorf("list features: empty response")
		}

		for _, f := range resp.JSON200.Features {
			if f.Id == "" {
				continue
			}
			out = append(out, featureMeta{
				ID:           f.Id,
				Description:  f.Description,
				NoActiveEnvs: featureHasNoActiveEnvironments(f.Environments),
				ValueType:    f.ValueType,
			})
		}

		if !resp.JSON200.HasMore {
			break
		}
		if resp.JSON200.NextOffset != nil {
			offset = *resp.JSON200.NextOffset
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func nameAndDedupe(features []featureMeta) []namedFeature {
	nameCounts := map[string]int{}
	out := make([]namedFeature, 0, len(features))

	for _, f := range features {
		baseName := "Feature" + toExportedIdentifier(f.ID)
		nameCounts[baseName]++
		name := baseName
		if nameCounts[baseName] > 1 {
			name = fmt.Sprintf("%s_%d", baseName, nameCounts[baseName])
		}
		out = append(out, namedFeature{
			Name:         name,
			ID:           f.ID,
			Description:  f.Description,
			NoActiveEnvs: f.NoActiveEnvs,
			ValueType:    f.ValueType,
		})
	}

	return out
}

func featureHasNoActiveEnvironments(envs map[string]growthbookapi.FeatureEnvironment) bool {
	if len(envs) == 0 {
		return true
	}
	for _, e := range envs {
		if e.Enabled {
			return false
		}
	}
	return true
}
