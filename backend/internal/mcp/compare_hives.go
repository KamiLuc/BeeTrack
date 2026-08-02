package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beetrack/backend/internal/model"
)

var errHiveIDsRequired = fmt.Errorf("hive_ids is required")

type HiveComparison struct {
	HiveSummary
	LastInspectedAt *time.Time        `json:"last_inspected_at"`
	BroodPattern    string            `json:"brood_pattern"`
	ColonyStrength  string            `json:"colony_strength"`
	FramesBrood     *int              `json:"frames_brood"`
	FramesFeed      *int              `json:"frames_feed"`
	FramesPollen    *int              `json:"frames_pollen"`
	LastTreatment   *TreatmentSummary `json:"last_treatment"`
	LastFeeding     *FeedingSummary   `json:"last_feeding"`
	LastHarvest     *HarvestSummary   `json:"last_harvest"`
}

// CompareHives returns a side-by-side comparison of the given hives: status
// flags, diseases, and each one's latest inspection, treatment, feeding, and
// harvest. Every hiveID must belong to userID; the whole call fails if any
// doesn't, same as authorizeHive's behavior elsewhere.
func (t *HiveTools) CompareHives(ctx context.Context, userID int64, hiveIDs []int64) ([]HiveComparison, error) {
	if len(hiveIDs) == 0 {
		return nil, errHiveIDsRequired
	}

	hives := make([]*model.Hive, len(hiveIDs))
	apiaryNames := make(map[int64]string, len(hiveIDs))
	for i, id := range hiveIDs {
		hive, apiaryName, err := authorizeHive(ctx, t.apiaries, t.hives, userID, id)
		if err != nil {
			return nil, err
		}
		hives[i] = hive
		apiaryNames[hive.ApiaryID] = apiaryName
	}

	summaries, err := t.hiveSummaries(ctx, hives, apiaryNames)
	if err != nil {
		return nil, err
	}

	from, to := daysToRange(nil)
	result := make([]HiveComparison, len(summaries))
	for i, s := range summaries {
		inspections, err := t.listInspections(ctx, s.ID, from, to)
		if err != nil {
			return nil, err
		}
		treatments, err := t.listTreatments(ctx, s.ID, from, to)
		if err != nil {
			return nil, err
		}
		feedings, err := t.listFeedings(ctx, s.ID, from, to)
		if err != nil {
			return nil, err
		}
		harvests, err := t.listHarvests(ctx, s.ID, from, to)
		if err != nil {
			return nil, err
		}

		result[i] = HiveComparison{HiveSummary: s}
		if len(inspections) > 0 {
			latest := inspections[0]
			result[i].LastInspectedAt = &latest.InspectedAt
			result[i].BroodPattern = latest.BroodPattern
			result[i].ColonyStrength = latest.ColonyStrength
			result[i].FramesBrood = latest.FramesBrood
			result[i].FramesFeed = latest.FramesFeed
			result[i].FramesPollen = latest.FramesPollen
		}
		if len(treatments) > 0 {
			result[i].LastTreatment = &treatments[0]
		}
		if len(feedings) > 0 {
			result[i].LastFeeding = &feedings[0]
		}
		if len(harvests) > 0 {
			result[i].LastHarvest = &harvests[0]
		}
	}
	return result, nil
}

type compareHivesInput struct {
	HiveIDs []int64 `json:"hive_ids"`
}

func (t *HiveTools) CompareHivesTool() Tool {
	return Tool{
		Name:        "compare_hives",
		Description: "Compare hives side by side: status flags, diseases, and each one's latest inspection, treatment, feeding, and harvest.",
		InputSchema: InputSchema{
			Properties: map[string]any{
				"hive_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "integer"},
					"description": "The hive IDs to compare.",
				},
			},
			Required: []string{"hive_ids"},
		},
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			var in compareHivesInput
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, fmt.Errorf("hive_ids must be the numeric ids from list_hives, not hive names: %w", err)
				}
			}
			return t.CompareHives(ctx, userID, in.HiveIDs)
		},
	}
}
