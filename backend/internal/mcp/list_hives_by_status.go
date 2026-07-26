package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

var hiveStatuses = []string{"queenless", "needs_food", "sick", "ready_for_harvest"}

func matchesStatus(h HiveSummary, status string) bool {
	switch status {
	case "queenless":
		return h.Queenless
	case "needs_food":
		return h.NeedsFood
	case "sick":
		return len(h.Diseases) > 0
	case "ready_for_harvest":
		return h.ReadyForHarvest
	default:
		return false
	}
}

// ListHivesByStatus returns hives across userID's apiaries (or just
// apiaryID's, if non-nil) matching any of statuses (all four if empty):
// queenless, needs_food, sick (has an active disease), ready_for_harvest.
func (t *HiveTools) ListHivesByStatus(ctx context.Context, userID int64, apiaryID *int64, statuses []string) ([]HiveSummary, error) {
	types, err := validateAllowedValues(statuses, hiveStatuses, "status")
	if err != nil {
		return nil, err
	}
	hives, err := resolveHives(ctx, t.apiaries, t.hives, userID, apiaryID)
	if err != nil {
		return nil, err
	}
	summaries, err := t.hiveSummaries(ctx, hives)
	if err != nil {
		return nil, err
	}

	var result []HiveSummary
	for _, s := range summaries {
		for _, status := range types {
			if matchesStatus(s, status) {
				result = append(result, s)
				break
			}
		}
	}
	return result, nil
}

type listHivesByStatusInput struct {
	ApiaryID *int64   `json:"apiary_id,omitempty"`
	Statuses []string `json:"statuses,omitempty"`
}

func (t *HiveTools) ListHivesByStatusTool() Tool {
	return Tool{
		Name:        "list_hives_by_status",
		Description: "List hives matching any of the given statuses (queenless, needs_food, sick, ready_for_harvest). Omit statuses to match hives with any of these flags set. Accepts an optional apiary_id to filter to one apiary.",
		InputSchema: InputSchema{
			Properties: map[string]any{
				"apiary_id": map[string]any{
					"type":        "integer",
					"description": "Restrict results to this apiary's hives.",
				},
				"statuses": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string", "enum": hiveStatuses},
					"description": "Which statuses to match (any one qualifies); omit to match any of them.",
				},
			},
		},
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			var in listHivesByStatusInput
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, fmt.Errorf("decode input: %w", err)
				}
			}
			return t.ListHivesByStatus(ctx, userID, in.ApiaryID, in.Statuses)
		},
	}
}
