package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Inspections/Treatments/Feedings/Harvests are counted within GetDashboardSummary's day window, not all-time.
type DashboardSummary struct {
	ApiaryCount      int     `json:"apiary_count"`
	HiveCount        int     `json:"hive_count"`
	Queenless        int     `json:"queenless"`
	NeedsFood        int     `json:"needs_food"`
	Sick             int     `json:"sick"`
	ReadyForHarvest  int     `json:"ready_for_harvest"`
	Inspections      int     `json:"inspections"`
	Treatments       int     `json:"treatments"`
	Feedings         int     `json:"feedings"`
	Harvests         int     `json:"harvests"`
	HarvestKilograms float64 `json:"harvest_kilograms"`
}

// apiaryID nil means every apiary userID belongs to; days nil means all-time.
func (t *HiveTools) GetDashboardSummary(ctx context.Context, userID int64, apiaryID *int64, days *int) (DashboardSummary, error) {
	hives, apiaryNames, err := resolveHives(ctx, t.apiaries, t.hives, userID, apiaryID)
	if err != nil {
		return DashboardSummary{}, err
	}

	apiaryCount := 1
	if apiaryID == nil {
		memberships, err := t.apiaries.ListByUserID(ctx, userID)
		if err != nil {
			return DashboardSummary{}, fmt.Errorf("list apiaries: %w", err)
		}
		apiaryCount = len(memberships)
	}

	summaries, err := t.hiveSummaries(ctx, hives, apiaryNames)
	if err != nil {
		return DashboardSummary{}, err
	}

	result := DashboardSummary{ApiaryCount: apiaryCount, HiveCount: len(hives)}
	for _, s := range summaries {
		if s.Queenless {
			result.Queenless++
		}
		if s.NeedsFood {
			result.NeedsFood++
		}
		if len(s.Diseases) > 0 {
			result.Sick++
		}
		if s.ReadyForHarvest {
			result.ReadyForHarvest++
		}
	}

	hiveIDs := hiveIDsOf(hives)
	from, to := daysToRange(days)

	inspections, err := t.inspections.ListByHiveIDsAndRange(ctx, hiveIDs, from, to)
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("list inspections: %w", err)
	}
	result.Inspections = len(inspections)

	treatments, err := t.treatments.ListByHiveIDsAndRange(ctx, hiveIDs, from, to)
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("list treatments: %w", err)
	}
	result.Treatments = len(treatments)

	feedings, err := t.feedings.ListByHiveIDsAndRange(ctx, hiveIDs, from, to)
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("list feedings: %w", err)
	}
	result.Feedings = len(feedings)

	harvests, err := t.harvests.ListByHiveIDsAndRange(ctx, hiveIDs, from, to)
	if err != nil {
		return DashboardSummary{}, fmt.Errorf("list harvests: %w", err)
	}
	result.Harvests = len(harvests)
	for _, h := range harvests {
		result.HarvestKilograms += h.Kilograms
	}

	return result, nil
}

type dashboardSummaryInput struct {
	ApiaryID *int64 `json:"apiary_id,omitempty"`
	Days     *int   `json:"days,omitempty"`
}

func (t *HiveTools) GetDashboardSummaryTool() Tool {
	return Tool{
		Name:        "get_dashboard_summary",
		Description: "Cross-apiary overview: hive counts by status (queenless, needs_food, sick, ready_for_harvest) and recent inspection/treatment/feeding/harvest counts plus total kilograms harvested. Accepts an optional apiary_id to scope to one apiary and days to limit the recent-activity window (omit for all-time).",
		InputSchema: InputSchema{
			Properties: map[string]any{
				"apiary_id": map[string]any{
					"type":        "integer",
					"description": "Restrict to this apiary; omit for a rollup across all of the caller's apiaries.",
				},
				"days": map[string]any{
					"type":        "integer",
					"description": "Only count recent activity from the last N days; omit for all of it.",
				},
			},
		},
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			var in dashboardSummaryInput
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, fmt.Errorf("decode get_dashboard_summary input: %w", err)
				}
			}
			return t.GetDashboardSummary(ctx, userID, in.ApiaryID, in.Days)
		},
	}
}
