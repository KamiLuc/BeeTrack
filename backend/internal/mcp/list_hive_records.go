package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

var hiveRecordTypes = []string{"inspection", "treatment", "harvest", "feeding"}

type HiveRecords struct {
	Inspections []InspectionSummary `json:"inspections,omitempty"`
	Treatments  []TreatmentSummary  `json:"treatments,omitempty"`
	Harvests    []HarvestSummary    `json:"harvests,omitempty"`
	Feedings    []FeedingSummary    `json:"feedings,omitempty"`
}

// ListHiveRecords returns one hive's records of the requested types (all
// four if recordTypes is empty), each within the given lookback window (or
// all of history, if days is nil).
func (t *HiveTools) ListHiveRecords(ctx context.Context, userID, hiveID int64, recordTypes []string, days *int) (*HiveRecords, error) {
	if _, err := authorizeHive(ctx, t.apiaries, t.hives, userID, hiveID); err != nil {
		return nil, err
	}
	types, err := validateAllowedValues(recordTypes, hiveRecordTypes, "record_type")
	if err != nil {
		return nil, err
	}
	from, to := daysToRange(days)

	var result HiveRecords
	for _, rt := range types {
		switch rt {
		case "inspection":
			result.Inspections, err = t.listInspections(ctx, hiveID, from, to)
		case "treatment":
			result.Treatments, err = t.listTreatments(ctx, hiveID, from, to)
		case "harvest":
			result.Harvests, err = t.listHarvests(ctx, hiveID, from, to)
		case "feeding":
			result.Feedings, err = t.listFeedings(ctx, hiveID, from, to)
		}
		if err != nil {
			return nil, err
		}
	}
	return &result, nil
}

type listHiveRecordsInput struct {
	HiveID      int64    `json:"hive_id"`
	RecordTypes []string `json:"record_types,omitempty"`
	Days        *int     `json:"days,omitempty"`
}

func (t *HiveTools) ListHiveRecordsTool() Tool {
	return Tool{
		Name:        "list_hive_records",
		Description: "List one hive's inspections, treatments, harvests, and/or feedings, optionally limited to the last N days. Pass record_types to get only some of them (e.g. [\"inspection\",\"feeding\"]); omit it for all four.",
		InputSchema: InputSchema{
			Properties: map[string]any{
				"hive_id": map[string]any{
					"type":        "integer",
					"description": "The hive to look up.",
				},
				"record_types": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string", "enum": hiveRecordTypes},
					"description": "Which record types to include; omit for all of them.",
				},
				"days": map[string]any{
					"type":        "integer",
					"description": "Only include records from the last N days; omit for all of history.",
				},
			},
			Required: []string{"hive_id"},
		},
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			var in listHiveRecordsInput
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, fmt.Errorf("decode input: %w", err)
				}
			}
			if in.HiveID <= 0 {
				return nil, errHiveIDRequired
			}
			return t.ListHiveRecords(ctx, userID, in.HiveID, in.RecordTypes, in.Days)
		},
	}
}
