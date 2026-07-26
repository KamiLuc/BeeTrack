package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/beetrack/backend/internal/model"
)

type InspectionLister interface {
	ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Inspection, error)
	LastInspectionDatesByHiveIDs(ctx context.Context, ids []int64) (map[int64]*time.Time, error)
}

type TreatmentLister interface {
	ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Treatment, error)
	LastTreatmentDatesByHiveIDs(ctx context.Context, ids []int64) (map[int64]*time.Time, error)
}

type HarvestLister interface {
	ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Harvest, error)
}

type FeedingLister interface {
	ListByHiveIDsAndRange(ctx context.Context, hiveIDs []int64, from, to time.Time) ([]*model.Feeding, error)
	LastFeedingDatesByHiveIDs(ctx context.Context, ids []int64) (map[int64]*time.Time, error)
}

type InspectionSummary struct {
	ID           int64     `json:"id"`
	InspectedAt  time.Time `json:"inspected_at"`
	QueenStatus  string    `json:"queen_status"`
	BroodPattern string    `json:"brood_pattern"`
	FramesBrood  *int      `json:"frames_brood"`
	FramesFeed   *int      `json:"frames_feed"`
	FramesPollen *int      `json:"frames_pollen"`
	Notes        string    `json:"notes"`
}

type TreatmentSummary struct {
	ID           int64     `json:"id"`
	TreatedAt    time.Time `json:"treated_at"`
	MedicineName string    `json:"medicine_name"`
	Dose         string    `json:"dose"`
	Notes        string    `json:"notes"`
}

type HarvestSummary struct {
	ID          int64     `json:"id"`
	HarvestedAt time.Time `json:"harvested_at"`
	Frames      int       `json:"frames"`
	HalfFrames  int       `json:"half_frames"`
	Kilograms   float64   `json:"kilograms"`
	Notes       string    `json:"notes"`
}

type FeedingSummary struct {
	ID       int64     `json:"id"`
	FedAt    time.Time `json:"fed_at"`
	FeedType string    `json:"feed_type"`
	Amount   string    `json:"amount"`
	Notes    string    `json:"notes"`
}

type HiveHistory struct {
	Hive        HiveSummary         `json:"hive"`
	Inspections []InspectionSummary `json:"inspections"`
	Treatments  []TreatmentSummary  `json:"treatments"`
	Harvests    []HarvestSummary    `json:"harvests"`
	Feedings    []FeedingSummary    `json:"feedings"`
}

// daysToRange turns an optional lookback window into a [from, to) pair for
// ListByHiveIDsAndRange; nil days means "since ever."
func daysToRange(days *int) (time.Time, time.Time) {
	to := time.Now().Add(time.Minute)
	if days == nil {
		return time.Time{}, to
	}
	return to.AddDate(0, 0, -*days), to
}

func (t *HiveTools) listInspections(ctx context.Context, hiveID int64, from, to time.Time) ([]InspectionSummary, error) {
	inspections, err := t.inspections.ListByHiveIDsAndRange(ctx, []int64{hiveID}, from, to)
	if err != nil {
		return nil, fmt.Errorf("list inspections: %w", err)
	}
	summaries := make([]InspectionSummary, len(inspections))
	for i, insp := range inspections {
		summaries[i] = InspectionSummary{
			ID:           insp.ID,
			InspectedAt:  insp.InspectedAt,
			QueenStatus:  insp.QueenStatus,
			BroodPattern: insp.BroodPattern,
			FramesBrood:  insp.FramesBrood,
			FramesFeed:   insp.FramesFeed,
			FramesPollen: insp.FramesPollen,
			Notes:        insp.Notes,
		}
	}
	return summaries, nil
}

func (t *HiveTools) listTreatments(ctx context.Context, hiveID int64, from, to time.Time) ([]TreatmentSummary, error) {
	treatments, err := t.treatments.ListByHiveIDsAndRange(ctx, []int64{hiveID}, from, to)
	if err != nil {
		return nil, fmt.Errorf("list treatments: %w", err)
	}
	summaries := make([]TreatmentSummary, len(treatments))
	for i, tr := range treatments {
		summaries[i] = TreatmentSummary{
			ID:           tr.ID,
			TreatedAt:    tr.TreatedAt,
			MedicineName: tr.MedicineName,
			Dose:         tr.Dose,
			Notes:        tr.Notes,
		}
	}
	return summaries, nil
}

func (t *HiveTools) listHarvests(ctx context.Context, hiveID int64, from, to time.Time) ([]HarvestSummary, error) {
	harvests, err := t.harvests.ListByHiveIDsAndRange(ctx, []int64{hiveID}, from, to)
	if err != nil {
		return nil, fmt.Errorf("list harvests: %w", err)
	}
	summaries := make([]HarvestSummary, len(harvests))
	for i, h := range harvests {
		summaries[i] = HarvestSummary{
			ID:          h.ID,
			HarvestedAt: h.HarvestedAt,
			Frames:      h.Frames,
			HalfFrames:  h.HalfFrames,
			Kilograms:   h.Kilograms,
			Notes:       h.Notes,
		}
	}
	return summaries, nil
}

func (t *HiveTools) listFeedings(ctx context.Context, hiveID int64, from, to time.Time) ([]FeedingSummary, error) {
	feedings, err := t.feedings.ListByHiveIDsAndRange(ctx, []int64{hiveID}, from, to)
	if err != nil {
		return nil, fmt.Errorf("list feedings: %w", err)
	}
	summaries := make([]FeedingSummary, len(feedings))
	for i, f := range feedings {
		summaries[i] = FeedingSummary{
			ID:       f.ID,
			FedAt:    f.FedAt,
			FeedType: f.FeedType,
			Amount:   f.Amount,
			Notes:    f.Notes,
		}
	}
	return summaries, nil
}

// GetHiveSummary aggregates one hive's inspections, treatments, harvests, and
// feedings within the given lookback window (or all of it, if days is nil).
func (t *HiveTools) GetHiveSummary(ctx context.Context, userID, hiveID int64, days *int) (*HiveHistory, error) {
	hive, err := authorizeHive(ctx, t.apiaries, t.hives, userID, hiveID)
	if err != nil {
		return nil, err
	}
	from, to := daysToRange(days)

	inspections, err := t.listInspections(ctx, hiveID, from, to)
	if err != nil {
		return nil, err
	}
	treatments, err := t.listTreatments(ctx, hiveID, from, to)
	if err != nil {
		return nil, err
	}
	harvests, err := t.listHarvests(ctx, hiveID, from, to)
	if err != nil {
		return nil, err
	}
	feedings, err := t.listFeedings(ctx, hiveID, from, to)
	if err != nil {
		return nil, err
	}

	diseases, err := t.hives.ListDiseasesByHiveIDs(ctx, []int64{hiveID})
	if err != nil {
		return nil, fmt.Errorf("list hive diseases: %w", err)
	}
	diseaseNames := make([]string, len(diseases))
	for i, d := range diseases {
		diseaseNames[i] = d.Disease
	}

	return &HiveHistory{
		Hive: HiveSummary{
			ID:              hive.ID,
			ApiaryID:        hive.ApiaryID,
			Name:            hive.Name,
			Type:            hive.Type,
			Active:          hive.Active,
			Queenless:       hive.Queenless,
			NeedsFood:       hive.NeedsFood,
			ReadyForHarvest: hive.ReadyForHarvest,
			Diseases:        diseaseNames,
		},
		Inspections: inspections,
		Treatments:  treatments,
		Harvests:    harvests,
		Feedings:    feedings,
	}, nil
}

type hiveDaysInput struct {
	HiveID int64 `json:"hive_id"`
	Days   *int  `json:"days,omitempty"`
}

var errHiveIDRequired = fmt.Errorf("hive_id is required")

func decodeHiveDaysInput(input json.RawMessage) (hiveDaysInput, error) {
	var in hiveDaysInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return in, fmt.Errorf("decode input: %w", err)
		}
	}
	if in.HiveID <= 0 {
		return in, errHiveIDRequired
	}
	return in, nil
}

func hiveDaysSchema(description string) InputSchema {
	return InputSchema{
		Properties: map[string]any{
			"hive_id": map[string]any{
				"type":        "integer",
				"description": "The hive to look up.",
			},
			"days": map[string]any{
				"type":        "integer",
				"description": description,
			},
		},
		Required: []string{"hive_id"},
	}
}

func (t *HiveTools) GetHiveSummaryTool() Tool {
	return Tool{
		Name:        "get_hive_summary",
		Description: "Get one hive's full picture: status flags, diseases, and its inspections/treatments/harvests/feedings, optionally limited to the last N days.",
		InputSchema: hiveDaysSchema("Only include history from the last N days; omit for all of it."),
		Handler: func(ctx context.Context, userID int64, input json.RawMessage) (any, error) {
			in, err := decodeHiveDaysInput(input)
			if err != nil {
				return nil, err
			}
			return t.GetHiveSummary(ctx, userID, in.HiveID, in.Days)
		},
	}
}
