package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

var unrecentRecordTypes = []string{"inspection", "treatment", "feeding"}

type HiveMissingRecords struct {
	HiveSummary
	LastRecordedAt map[string]*time.Time `json:"last_recorded_at"`
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

// ListHivesMissingRecords returns hives across userID's apiaries (or just
// apiaryID's, if non-nil) missing at least one of recordTypes (all three if
// empty) in the last days days, or ever if days is nil. Each returned hive's
// LastRecordedAt only lists the requested types it's actually missing.
func (t *HiveTools) ListHivesMissingRecords(ctx context.Context, userID int64, apiaryID *int64, recordTypes []string, days *int) ([]HiveMissingRecords, error) {
	types, err := validateRecordTypes(recordTypes, unrecentRecordTypes)
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
	hiveIDs := hiveIDsOf(hives)

	lastByType := make(map[string]map[int64]*time.Time, len(types))
	for _, rt := range types {
		switch rt {
		case "treatment":
			lastByType[rt], err = t.treatments.LastTreatmentDatesByHiveIDs(ctx, hiveIDs)
		case "inspection":
			lastByType[rt], err = t.inspections.LastInspectionDatesByHiveIDs(ctx, hiveIDs)
		case "feeding":
			lastByType[rt], err = t.feedings.LastFeedingDatesByHiveIDs(ctx, hiveIDs)
		}
		if err != nil {
			return nil, fmt.Errorf("get last %s dates: %w", rt, err)
		}
	}

	cutoff := cutoffFromDays(days)
	var result []HiveMissingRecords
	for _, s := range summaries {
		missing := map[string]*time.Time{}
		for _, rt := range types {
			last := lastByType[rt][s.ID]
			if lacksRecentRecord(last, cutoff) {
				missing[rt] = last
			}
		}
		if len(missing) > 0 {
			result = append(result, HiveMissingRecords{HiveSummary: s, LastRecordedAt: missing})
		}
	}
	return result, nil
}

type listHivesMissingRecordsInput struct {
	ApiaryID    *int64   `json:"apiary_id,omitempty"`
	RecordTypes []string `json:"record_types,omitempty"`
	Days        *int     `json:"days,omitempty"`
}

func (t *HiveTools) ListHivesMissingRecordsTool() Tool {
	return Tool{
		Name:        "list_hives_missing_records",
		Description: "List hives missing at least one of the given record types (inspection/treatment/feeding) in the last N days, or that have never had one at all if days is omitted. Pass record_types to check only some of them (e.g. [\"treatment\",\"feeding\"]); omit it for all three. Accepts an optional apiary_id to filter to one apiary.",
		InputSchema: InputSchema{
			Properties: map[string]any{
				"apiary_id": map[string]any{
					"type":        "integer",
					"description": "Restrict results to this apiary's hives.",
				},
				"record_types": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string", "enum": unrecentRecordTypes},
					"description": "Which record types to check; omit for all of them.",
				},
				"days": map[string]any{
					"type":        "integer",
					"description": "Only include hives missing a record in the last N days; omit to mean 'never had one'.",
				},
			},
		},
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			var in listHivesMissingRecordsInput
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, fmt.Errorf("decode input: %w", err)
				}
			}
			return t.ListHivesMissingRecords(ctx, userID, in.ApiaryID, in.RecordTypes, in.Days)
		},
	}
}
