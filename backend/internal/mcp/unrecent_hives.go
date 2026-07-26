package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type UntreatedHive struct {
	HiveSummary
	LastTreatedAt *time.Time `json:"last_treated_at"`
}

type UninspectedHive struct {
	HiveSummary
	LastInspectedAt *time.Time `json:"last_inspected_at"`
}

type UnfedHive struct {
	HiveSummary
	LastFedAt *time.Time `json:"last_fed_at"`
}

// cutoffFromDays turns an optional lookback window into a cutoff time; nil
// means "no cutoff — only hives with no record at all qualify."
func cutoffFromDays(days *int) *time.Time {
	if days == nil {
		return nil
	}
	cutoff := time.Now().AddDate(0, 0, -*days)
	return &cutoff
}

// lacksRecentRecord reports whether last (a hive's most recent record of
// some type, or nil if it's never had one) fails to satisfy cutoff. A nil
// cutoff means "only hives that have never had one qualify."
func lacksRecentRecord(last, cutoff *time.Time) bool {
	if cutoff == nil {
		return last == nil
	}
	return last == nil || last.Before(*cutoff)
}

// ListUntreatedHives returns hives across userID's apiaries (or just
// apiaryID's, if non-nil) with no treatment in the last days days, or never
// treated at all if days is nil.
func (t *HiveTools) ListUntreatedHives(ctx context.Context, userID int64, apiaryID *int64, days *int) ([]UntreatedHive, error) {
	hives, err := resolveHives(ctx, t.apiaries, t.hives, userID, apiaryID)
	if err != nil {
		return nil, err
	}
	summaries, err := t.hiveSummaries(ctx, hives)
	if err != nil {
		return nil, err
	}
	lastTreated, err := t.treatments.LastTreatmentDatesByHiveIDs(ctx, hiveIDsOf(hives))
	if err != nil {
		return nil, fmt.Errorf("get last treatment dates: %w", err)
	}

	cutoff := cutoffFromDays(days)
	var result []UntreatedHive
	for _, s := range summaries {
		last := lastTreated[s.ID]
		if !lacksRecentRecord(last, cutoff) {
			continue
		}
		result = append(result, UntreatedHive{HiveSummary: s, LastTreatedAt: last})
	}
	return result, nil
}

// ListUninspectedHives returns hives with no inspection in the last days
// days, or never inspected at all if days is nil.
func (t *HiveTools) ListUninspectedHives(ctx context.Context, userID int64, apiaryID *int64, days *int) ([]UninspectedHive, error) {
	hives, err := resolveHives(ctx, t.apiaries, t.hives, userID, apiaryID)
	if err != nil {
		return nil, err
	}
	summaries, err := t.hiveSummaries(ctx, hives)
	if err != nil {
		return nil, err
	}
	lastInspected, err := t.inspections.LastInspectionDatesByHiveIDs(ctx, hiveIDsOf(hives))
	if err != nil {
		return nil, fmt.Errorf("get last inspection dates: %w", err)
	}

	cutoff := cutoffFromDays(days)
	var result []UninspectedHive
	for _, s := range summaries {
		last := lastInspected[s.ID]
		if !lacksRecentRecord(last, cutoff) {
			continue
		}
		result = append(result, UninspectedHive{HiveSummary: s, LastInspectedAt: last})
	}
	return result, nil
}

// ListUnfedHives returns hives with no feeding in the last days days, or
// never fed at all if days is nil.
func (t *HiveTools) ListUnfedHives(ctx context.Context, userID int64, apiaryID *int64, days *int) ([]UnfedHive, error) {
	hives, err := resolveHives(ctx, t.apiaries, t.hives, userID, apiaryID)
	if err != nil {
		return nil, err
	}
	summaries, err := t.hiveSummaries(ctx, hives)
	if err != nil {
		return nil, err
	}
	lastFed, err := t.feedings.LastFeedingDatesByHiveIDs(ctx, hiveIDsOf(hives))
	if err != nil {
		return nil, fmt.Errorf("get last feeding dates: %w", err)
	}

	cutoff := cutoffFromDays(days)
	var result []UnfedHive
	for _, s := range summaries {
		last := lastFed[s.ID]
		if !lacksRecentRecord(last, cutoff) {
			continue
		}
		result = append(result, UnfedHive{HiveSummary: s, LastFedAt: last})
	}
	return result, nil
}

type listUnrecentHivesInput struct {
	ApiaryID *int64 `json:"apiary_id,omitempty"`
	Days     *int   `json:"days,omitempty"`
}

func decodeListUnrecentHivesInput(input json.RawMessage) (listUnrecentHivesInput, error) {
	var in listUnrecentHivesInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return in, fmt.Errorf("decode input: %w", err)
		}
	}
	return in, nil
}

func unrecentHivesSchema(recordKind string) InputSchema {
	return InputSchema{
		Properties: map[string]any{
			"apiary_id": map[string]any{
				"type":        "integer",
				"description": "Restrict results to this apiary's hives.",
			},
			"days": map[string]any{
				"type":        "integer",
				"description": fmt.Sprintf("Only include hives with no %s in the last N days; omit to mean 'never %s'.", recordKind, recordKind),
			},
		},
	}
}

func (t *HiveTools) ListUntreatedHivesTool() Tool {
	return Tool{
		Name:        "list_untreated_hives",
		Description: "List hives with no treatment in the last N days, or never treated at all if days is omitted. Accepts an optional apiary_id to filter to one apiary.",
		InputSchema: unrecentHivesSchema("treatment"),
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			in, err := decodeListUnrecentHivesInput(input)
			if err != nil {
				return nil, err
			}
			return t.ListUntreatedHives(ctx, userID, in.ApiaryID, in.Days)
		},
	}
}

func (t *HiveTools) ListUninspectedHivesTool() Tool {
	return Tool{
		Name:        "list_uninspected_hives",
		Description: "List hives with no inspection in the last N days, or never inspected at all if days is omitted. Accepts an optional apiary_id to filter to one apiary.",
		InputSchema: unrecentHivesSchema("inspection"),
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			in, err := decodeListUnrecentHivesInput(input)
			if err != nil {
				return nil, err
			}
			return t.ListUninspectedHives(ctx, userID, in.ApiaryID, in.Days)
		},
	}
}

func (t *HiveTools) ListUnfedHivesTool() Tool {
	return Tool{
		Name:        "list_unfed_hives",
		Description: "List hives with no feeding in the last N days, or never fed at all if days is omitted. Accepts an optional apiary_id to filter to one apiary.",
		InputSchema: unrecentHivesSchema("feeding"),
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			in, err := decodeListUnrecentHivesInput(input)
			if err != nil {
				return nil, err
			}
			return t.ListUnfedHives(ctx, userID, in.ApiaryID, in.Days)
		},
	}
}
