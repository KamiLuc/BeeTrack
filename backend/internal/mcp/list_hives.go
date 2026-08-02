package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
)

type ApiaryLister interface {
	GetMembership(ctx context.Context, apiaryID, userID int64) (*model.Apiary, string, error)
	ListByUserID(ctx context.Context, userID int64) ([]model.ApiaryMembership, error)
}

type HiveLister interface {
	GetByID(ctx context.Context, hiveID int64) (*model.Hive, error)
	ListByApiaryID(ctx context.Context, apiaryID int64) ([]*model.Hive, error)
	ListDiseasesByHiveIDs(ctx context.Context, ids []int64) ([]*model.HiveDisease, error)
}

type HiveSummary struct {
	ID                    int64    `json:"id"`
	ApiaryID              int64    `json:"apiary_id"`
	ApiaryName            string   `json:"apiary_name"`
	Name                  string   `json:"name"`
	Type                  string   `json:"type"`
	Active                bool     `json:"active"`
	QueenNeedsReplacement bool     `json:"queen_needs_replacement"`
	NeedsFood             bool     `json:"needs_food"`
	BoxNeedsAdding        bool     `json:"box_needs_adding"`
	ReadyForHarvest       bool     `json:"ready_for_harvest"`
	Diseases              []string `json:"diseases"`
}

// HiveTools groups the read-only hive tools; list_hives is also called
// directly (bypassing the registry) by the voice worker's hive-name
// resolution, filtered to the apiary it's recording in.
type HiveTools struct {
	apiaries    ApiaryLister
	hives       HiveLister
	inspections InspectionLister
	treatments  TreatmentLister
	harvests    HarvestLister
	feedings    FeedingLister
}

func NewHiveTools(apiaries ApiaryLister, hives HiveLister, inspections InspectionLister, treatments TreatmentLister, harvests HarvestLister, feedings FeedingLister) *HiveTools {
	return &HiveTools{
		apiaries:    apiaries,
		hives:       hives,
		inspections: inspections,
		treatments:  treatments,
		harvests:    harvests,
		feedings:    feedings,
	}
}

// resolveHives returns active hives across every apiary userID belongs to,
// or just apiaryID's, if it's non-nil, along with each involved apiary's
// name (keyed by apiary ID) so callers can label hives by apiary without a
// second lookup. Inactive hives are excluded, same as the app's own
// bulk-action hive selection.
func resolveHives(ctx context.Context, apiaries ApiaryLister, hives HiveLister, userID int64, apiaryID *int64) ([]*model.Hive, map[int64]string, error) {
	apiaryNames := make(map[int64]string)
	var apiaryIDs []int64
	if apiaryID != nil {
		apiary, _, err := apiaries.GetMembership(ctx, *apiaryID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, ErrApiaryNotFound
			}
			return nil, nil, fmt.Errorf("get apiary membership: %w", err)
		}
		apiaryIDs = []int64{*apiaryID}
		apiaryNames[*apiaryID] = apiary.Name
	} else {
		memberships, err := apiaries.ListByUserID(ctx, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("list apiaries: %w", err)
		}
		for _, m := range memberships {
			apiaryIDs = append(apiaryIDs, m.Apiary.ID)
			apiaryNames[m.Apiary.ID] = m.Apiary.Name
		}
	}

	var result []*model.Hive
	for _, id := range apiaryIDs {
		apiaryHives, err := hives.ListByApiaryID(ctx, id)
		if err != nil {
			return nil, nil, fmt.Errorf("list hives: %w", err)
		}
		for _, h := range apiaryHives {
			if h.Active {
				result = append(result, h)
			}
		}
	}
	return result, apiaryNames, nil
}

func hiveIDsOf(hives []*model.Hive) []int64 {
	ids := make([]int64, len(hives))
	for i, h := range hives {
		ids[i] = h.ID
	}
	return ids
}

// hiveSummaries builds a HiveSummary per hive, attaching each one's diseases
// in a single batch lookup. apiaryNames labels each hive's ApiaryName; a
// missing entry just leaves it blank rather than failing the whole call.
func (t *HiveTools) hiveSummaries(ctx context.Context, hives []*model.Hive, apiaryNames map[int64]string) ([]HiveSummary, error) {
	diseases, err := t.hives.ListDiseasesByHiveIDs(ctx, hiveIDsOf(hives))
	if err != nil {
		return nil, fmt.Errorf("list hive diseases: %w", err)
	}
	diseasesByHiveID := make(map[int64][]string, len(hives))
	for _, d := range diseases {
		diseasesByHiveID[d.HiveID] = append(diseasesByHiveID[d.HiveID], d.Disease)
	}

	summaries := make([]HiveSummary, len(hives))
	for i, h := range hives {
		summaries[i] = HiveSummary{
			ID:                    h.ID,
			ApiaryID:              h.ApiaryID,
			ApiaryName:            apiaryNames[h.ApiaryID],
			Name:                  h.Name,
			Type:                  h.Type,
			Active:                h.Active,
			QueenNeedsReplacement: h.QueenNeedsReplacement,
			NeedsFood:             h.NeedsFood,
			BoxNeedsAdding:        h.BoxNeedsAdding,
			ReadyForHarvest:       h.ReadyForHarvest,
			Diseases:              diseasesByHiveID[h.ID],
		}
	}
	return summaries, nil
}

// ListHives returns hives across every apiary userID belongs to, or just
// apiaryID's hives if it's non-nil.
func (t *HiveTools) ListHives(ctx context.Context, userID int64, apiaryID *int64) ([]HiveSummary, error) {
	hives, apiaryNames, err := resolveHives(ctx, t.apiaries, t.hives, userID, apiaryID)
	if err != nil {
		return nil, err
	}
	return t.hiveSummaries(ctx, hives, apiaryNames)
}

type listHivesInput struct {
	ApiaryID *int64 `json:"apiary_id,omitempty"`
}

// ListHivesTool adapts ListHives into a registry Tool: it JSON-decodes the
// model's arguments and calls straight through to ListHives.
func (t *HiveTools) ListHivesTool() Tool {
	return Tool{
		Name:        "list_hives",
		Description: "List active hives across the caller's apiaries, with status flags (queen_needs_replacement, needs_food, box_needs_adding, ready_for_harvest) and active diseases. Accepts an optional apiary_id to filter to one apiary.",
		InputSchema: InputSchema{
			Properties: map[string]any{
				"apiary_id": map[string]any{
					"type":        "integer",
					"description": "Restrict results to this apiary's hives.",
				},
			},
		},
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			var in listHivesInput
			if len(input) > 0 {
				if err := json.Unmarshal(input, &in); err != nil {
					return nil, fmt.Errorf("decode list_hives input: %w", err)
				}
			}
			return t.ListHives(ctx, userID, in.ApiaryID)
		},
	}
}
