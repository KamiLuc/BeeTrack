package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/beetrack/backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrInvalidAudioType       = errors.New("unsupported audio type; allowed: audio/webm, audio/mp4, audio/wav")
	ErrRecordingTooLong       = errors.New("recording exceeds 15 MB limit")
	ErrMaxRecordingsReached   = errors.New("you already have the maximum of 10 recordings awaiting review")
	ErrRecordingNotFound      = errors.New("voice recording not found")
	ErrRecordingNotCompleted  = errors.New("voice recording is not awaiting review")
	ErrRecordingNotCancelable = errors.New("voice recording can only be cancelled while pending")
	ErrActionNotFound         = errors.New("voice action not found")
	ErrActionNotProposed      = errors.New("voice action is not proposed")
)

// MaxAudioBytes bounds an uploaded recording's size as a proxy for the client's 3-minute duration
// cap (VC-18-FE), sized to the worst case: a 3-minute mono 16-bit 44.1kHz WAV is ~15MB, the largest
// of the accepted formats — avoids needing to parse webm/m4a/wav headers for an exact duration.
const MaxAudioBytes = 15 * 1024 * 1024

const maxRecordingsPerUser = 10

// voiceClaimDelay holds a freshly uploaded recording back from the worker for a
// window after upload, giving the beekeeper a real chance to Cancel it (VC-14-BE)
// before it flips to processing — otherwise a fast worker poll cycle could claim
// it before the client even finishes rendering the pending row.
const voiceClaimDelay = 10 * time.Second

var allowedAudioMIME = map[string]string{
	"audio/webm":  ".webm",
	"audio/mp4":   ".m4a",
	"audio/wav":   ".wav",
	"audio/x-wav": ".wav",
}

type VoiceRepository interface {
	CreateRecording(ctx context.Context, rec *model.VoiceRecording) error
	CountRecordingsByUserID(ctx context.Context, userID int64) (int64, error)
	GetRecordingByID(ctx context.Context, id int64) (*model.VoiceRecording, error)
	UpdateRecording(ctx context.Context, rec *model.VoiceRecording) error
	ListRecordingsByApiaryID(ctx context.Context, apiaryID int64, limit, offset int) ([]*model.VoiceRecording, error)
	CountRecordingsByApiaryID(ctx context.Context, apiaryID int64) (int64, error)
	ListActionsByRecordingID(ctx context.Context, recordingID int64) ([]*model.VoiceAction, error)
	ListActionsByRecordingIDs(ctx context.Context, recordingIDs []int64) ([]*model.VoiceAction, error)
	GetActionByID(ctx context.Context, id int64) (*model.VoiceAction, error)
	UpdateAction(ctx context.Context, action *model.VoiceAction) error
	DeleteActionsByRecordingID(ctx context.Context, recordingID int64) error
}

type InspectionCreator interface {
	Create(ctx context.Context, userID, apiaryID, hiveID int64, params InspectionParams) (*model.Inspection, error)
}

type TreatmentCreator interface {
	Create(ctx context.Context, userID, apiaryID, hiveID int64, params TreatmentParams) (*model.Treatment, error)
}

type HarvestCreator interface {
	Create(ctx context.Context, userID, apiaryID, hiveID int64, params HarvestParams) (*model.Harvest, error)
}

type FeedingCreator interface {
	Create(ctx context.Context, userID, apiaryID, hiveID int64, params FeedingParams) (*model.Feeding, error)
}

type VoiceHiveService interface {
	Get(ctx context.Context, userID, apiaryID, hiveID int64) (*model.Hive, error)
	Update(ctx context.Context, userID, apiaryID, hiveID int64, name, hiveType string, active, readyForHarvest, queenNeedsReplacement, needsFood, boxNeedsAdding bool) (*model.Hive, error)
	DiseasesByHive(ctx context.Context, hiveID int64) ([]*model.HiveDisease, error)
	AddDisease(ctx context.Context, userID, apiaryID, hiveID int64, disease string) (*model.HiveDisease, error)
	RemoveDisease(ctx context.Context, userID, apiaryID, hiveID, diseaseID int64) error
}

type VoiceService struct {
	apiaries    ApiaryMembershipReader
	recordings  VoiceRepository
	inspections InspectionCreator
	treatments  TreatmentCreator
	harvests    HarvestCreator
	feedings    FeedingCreator
	hives       VoiceHiveService
	storagePath string
}

func NewVoiceService(
	apiaries ApiaryMembershipReader,
	recordings VoiceRepository,
	inspections InspectionCreator,
	treatments TreatmentCreator,
	harvests HarvestCreator,
	feedings FeedingCreator,
	hives VoiceHiveService,
	storagePath string,
) *VoiceService {
	return &VoiceService{
		apiaries:    apiaries,
		recordings:  recordings,
		inspections: inspections,
		treatments:  treatments,
		harvests:    harvests,
		feedings:    feedings,
		hives:       hives,
		storagePath: storagePath,
	}
}

func (s *VoiceService) Upload(ctx context.Context, userID, apiaryID int64, mimeType string, data []byte) (*model.VoiceRecording, error) {
	return s.upload(ctx, userID, apiaryID, nil, mimeType, data)
}

func (s *VoiceService) UploadForHive(ctx context.Context, userID, apiaryID, hiveID int64, mimeType string, data []byte) (*model.VoiceRecording, error) {
	return s.upload(ctx, userID, apiaryID, &hiveID, mimeType, data)
}

func (s *VoiceService) upload(ctx context.Context, userID, apiaryID int64, hiveID *int64, mimeType string, data []byte) (*model.VoiceRecording, error) {
	if len(data) > MaxAudioBytes {
		return nil, ErrRecordingTooLong
	}
	ext, ok := allowedAudioMIME[mimeType]
	if !ok {
		return nil, ErrInvalidAudioType
	}
	if hiveID != nil {
		if _, err := s.hives.Get(ctx, userID, apiaryID, *hiveID); err != nil {
			return nil, err
		}
	} else if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}
	count, err := s.recordings.CountRecordingsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count recordings: %w", err)
	}
	if count >= maxRecordingsPerUser {
		return nil, ErrMaxRecordingsReached
	}
	if err := os.MkdirAll(s.storagePath, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	filename := uuid.New().String() + ext
	if err := os.WriteFile(filepath.Join(s.storagePath, filename), data, 0o644); err != nil {
		return nil, fmt.Errorf("write audio: %w", err)
	}
	claimAfter := time.Now().Add(voiceClaimDelay)
	rec := &model.VoiceRecording{
		UserID:        userID,
		ApiaryID:      apiaryID,
		HiveID:        hiveID,
		Status:        model.VoiceRecordingStatusPending,
		AudioPath:     &filename,
		NextAttemptAt: &claimAfter,
	}
	if err := s.recordings.CreateRecording(ctx, rec); err != nil {
		_ = os.Remove(filepath.Join(s.storagePath, filename))
		return nil, fmt.Errorf("create recording: %w", err)
	}
	return rec, nil
}

func (s *VoiceService) List(ctx context.Context, userID, apiaryID int64, limit, offset int) ([]*model.VoiceRecording, map[int64][]*model.VoiceAction, int64, error) {
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, 0, ErrApiaryNotFound
		}
		return nil, nil, 0, fmt.Errorf("get apiary: %w", err)
	}
	total, err := s.recordings.CountRecordingsByApiaryID(ctx, apiaryID)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("count recordings: %w", err)
	}
	recs, err := s.recordings.ListRecordingsByApiaryID(ctx, apiaryID, limit, offset)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("list recordings: %w", err)
	}

	ids := make([]int64, len(recs))
	for i, rec := range recs {
		ids[i] = rec.ID
	}
	actions, err := s.recordings.ListActionsByRecordingIDs(ctx, ids)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("list actions: %w", err)
	}
	grouped := make(map[int64][]*model.VoiceAction, len(ids))
	for _, id := range ids {
		grouped[id] = []*model.VoiceAction{}
	}
	for _, a := range actions {
		grouped[a.VoiceRecordingID] = append(grouped[a.VoiceRecordingID], a)
	}

	return recs, grouped, total, nil
}

func (s *VoiceService) Cancel(ctx context.Context, userID, apiaryID, recordingID int64) (*model.VoiceRecording, error) {
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}
	rec, err := s.recordings.GetRecordingByID(ctx, recordingID)
	if err != nil {
		return nil, fmt.Errorf("get recording: %w", err)
	}
	if rec == nil || rec.ApiaryID != apiaryID {
		return nil, ErrRecordingNotFound
	}
	if rec.Status != model.VoiceRecordingStatusPending {
		return nil, ErrRecordingNotCancelable
	}

	if rec.AudioPath != nil {
		_ = os.Remove(filepath.Join(s.storagePath, *rec.AudioPath))
	}
	rec.Status = model.VoiceRecordingStatusCancelled
	rec.AudioPath = nil
	if err := s.recordings.UpdateRecording(ctx, rec); err != nil {
		return nil, fmt.Errorf("update recording: %w", err)
	}
	return rec, nil
}

func (s *VoiceService) getOwnedCompletedRecording(ctx context.Context, userID, apiaryID, recordingID int64) (*model.VoiceRecording, error) {
	if _, _, err := s.apiaries.GetMembership(ctx, apiaryID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApiaryNotFound
		}
		return nil, fmt.Errorf("get apiary: %w", err)
	}
	rec, err := s.recordings.GetRecordingByID(ctx, recordingID)
	if err != nil {
		return nil, fmt.Errorf("get recording: %w", err)
	}
	if rec == nil || rec.ApiaryID != apiaryID {
		return nil, ErrRecordingNotFound
	}
	if rec.Status != model.VoiceRecordingStatusCompleted {
		return nil, ErrRecordingNotCompleted
	}
	return rec, nil
}

func (s *VoiceService) Accept(ctx context.Context, userID, apiaryID, recordingID int64) (*model.VoiceRecording, []*model.VoiceAction, error) {
	rec, err := s.getOwnedCompletedRecording(ctx, userID, apiaryID, recordingID)
	if err != nil {
		return nil, nil, err
	}

	actions, err := s.recordings.ListActionsByRecordingID(ctx, recordingID)
	if err != nil {
		return nil, nil, fmt.Errorf("list actions: %w", err)
	}

	for _, action := range actions {
		if action.Status != model.VoiceActionStatusProposed {
			continue
		}
		s.applyAction(ctx, userID, apiaryID, action)
		if err := s.recordings.UpdateAction(ctx, action); err != nil {
			return nil, nil, fmt.Errorf("update action: %w", err)
		}
	}

	rec.Status = model.VoiceRecordingStatusAccepted
	if err := s.recordings.UpdateRecording(ctx, rec); err != nil {
		return nil, nil, fmt.Errorf("update recording: %w", err)
	}
	return rec, actions, nil
}

func (s *VoiceService) Reject(ctx context.Context, userID, apiaryID, recordingID int64) (*model.VoiceRecording, error) {
	rec, err := s.getOwnedCompletedRecording(ctx, userID, apiaryID, recordingID)
	if err != nil {
		return nil, err
	}
	if err := s.recordings.DeleteActionsByRecordingID(ctx, recordingID); err != nil {
		return nil, fmt.Errorf("delete actions: %w", err)
	}
	rec.Status = model.VoiceRecordingStatusRejected
	if err := s.recordings.UpdateRecording(ctx, rec); err != nil {
		return nil, fmt.Errorf("update recording: %w", err)
	}
	return rec, nil
}

func (s *VoiceService) UpdateActionArguments(ctx context.Context, userID, apiaryID, recordingID, actionID int64, toolArguments json.RawMessage) (*model.VoiceAction, error) {
	rec, err := s.getOwnedCompletedRecording(ctx, userID, apiaryID, recordingID)
	if err != nil {
		return nil, err
	}
	action, err := s.recordings.GetActionByID(ctx, actionID)
	if err != nil {
		return nil, fmt.Errorf("get action: %w", err)
	}
	if action == nil || action.VoiceRecordingID != rec.ID {
		return nil, ErrActionNotFound
	}
	if action.Status != model.VoiceActionStatusProposed {
		return nil, ErrActionNotProposed
	}
	action.ToolArguments = datatypes.JSON(toolArguments)
	if err := s.recordings.UpdateAction(ctx, action); err != nil {
		return nil, fmt.Errorf("update action: %w", err)
	}
	return action, nil
}

func (s *VoiceService) applyAction(ctx context.Context, userID, apiaryID int64, action *model.VoiceAction) {
	if action.HiveID == nil || action.ToolName == nil {
		s.failAction(action, errors.New("action missing hive_id or tool_name"))
		return
	}
	hiveID := *action.HiveID

	if snapshot, err := s.hiveStateSnapshot(ctx, userID, apiaryID, hiveID); err == nil {
		action.PreviousHiveState = snapshot
	}

	var (
		resultType string
		resultID   int64
		err        error
	)
	switch *action.ToolName {
	case model.VoiceActionToolCreateInspection:
		resultType, resultID, err = s.applyCreateInspection(ctx, userID, apiaryID, hiveID, action.ToolArguments)
	case model.VoiceActionToolCreateTreatment:
		resultType, resultID, err = s.applyCreateTreatment(ctx, userID, apiaryID, hiveID, action.ToolArguments)
	case model.VoiceActionToolCreateHarvest:
		resultType, resultID, err = s.applyCreateHarvest(ctx, userID, apiaryID, hiveID, action.ToolArguments)
	case model.VoiceActionToolCreateFeeding:
		resultType, resultID, err = s.applyCreateFeeding(ctx, userID, apiaryID, hiveID, action.ToolArguments)
	case model.VoiceActionToolUpdateHiveStatus:
		resultType, resultID, err = s.applyUpdateHiveStatus(ctx, userID, apiaryID, hiveID, action.ToolArguments)
	default:
		err = fmt.Errorf("unsupported tool: %s", *action.ToolName)
	}

	if err != nil {
		s.failAction(action, err)
		return
	}
	action.Status = model.VoiceActionStatusApplied
	action.ResultType = &resultType
	action.ResultRecordID = &resultID
	action.ErrorMessage = nil
}

func (s *VoiceService) failAction(action *model.VoiceAction, err error) {
	msg := err.Error()
	action.Status = model.VoiceActionStatusError
	action.ErrorMessage = &msg
}

type hiveStateSnapshotJSON struct {
	Type                  string   `json:"type"`
	Active                bool     `json:"active"`
	ReadyForHarvest       bool     `json:"ready_for_harvest"`
	QueenNeedsReplacement bool     `json:"queen_needs_replacement"`
	NeedsFood             bool     `json:"needs_food"`
	BoxNeedsAdding        bool     `json:"box_needs_adding"`
	Diseases              []string `json:"diseases"`
}

func (s *VoiceService) hiveStateSnapshot(ctx context.Context, userID, apiaryID, hiveID int64) (datatypes.JSON, error) {
	hive, err := s.hives.Get(ctx, userID, apiaryID, hiveID)
	if err != nil {
		return nil, err
	}
	diseases, err := s.hives.DiseasesByHive(ctx, hiveID)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(diseases))
	for i, d := range diseases {
		names[i] = d.Disease
	}
	snap := hiveStateSnapshotJSON{
		Type:                  hive.Type,
		Active:                hive.Active,
		ReadyForHarvest:       hive.ReadyForHarvest,
		QueenNeedsReplacement: hive.QueenNeedsReplacement,
		NeedsFood:             hive.NeedsFood,
		BoxNeedsAdding:        hive.BoxNeedsAdding,
		Diseases:              names,
	}
	b, err := json.Marshal(snap)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}

type createInspectionArgs struct {
	QueenStatus           string `json:"queen_status"`
	BroodPattern          string `json:"brood_pattern"`
	Aggressiveness        string `json:"aggressiveness"`
	ColonyStrength        string `json:"colony_strength"`
	FramesBrood           *int   `json:"frames_brood"`
	FramesFeed            *int   `json:"frames_feed"`
	FramesPollen          *int   `json:"frames_pollen"`
	QueenCellsCount       *int   `json:"queen_cells_count"`
	FramesAddedFoundation *int   `json:"frames_added_foundation"`
	FramesAddedDrawn      *int   `json:"frames_added_drawn"`
	FramesAddedBrood      *int   `json:"frames_added_brood"`
	FramesAddedFeed       *int   `json:"frames_added_feed"`
	QueenAdded            bool   `json:"queen_added"`
	BoxAdded              bool   `json:"box_added"`
	Notes                 string `json:"notes"`
}

func (s *VoiceService) applyCreateInspection(ctx context.Context, userID, apiaryID, hiveID int64, raw datatypes.JSON) (string, int64, error) {
	var args createInspectionArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", 0, fmt.Errorf("parse tool arguments: %w", err)
	}
	insp, err := s.inspections.Create(ctx, userID, apiaryID, hiveID, InspectionParams{
		InspectedAt:           time.Now(),
		QueenStatus:           args.QueenStatus,
		BroodPattern:          args.BroodPattern,
		FramesBrood:           args.FramesBrood,
		FramesFeed:            args.FramesFeed,
		FramesPollen:          args.FramesPollen,
		QueenCellsCount:       args.QueenCellsCount,
		Aggressiveness:        args.Aggressiveness,
		ColonyStrength:        args.ColonyStrength,
		FramesAddedFoundation: args.FramesAddedFoundation,
		FramesAddedDrawn:      args.FramesAddedDrawn,
		FramesAddedBrood:      args.FramesAddedBrood,
		FramesAddedFeed:       args.FramesAddedFeed,
		QueenAdded:            args.QueenAdded,
		BoxAdded:              args.BoxAdded,
		Notes:                 args.Notes,
	})
	if err != nil {
		return "", 0, err
	}
	return model.VoiceActionResultTypeInspection, insp.ID, nil
}

type createTreatmentArgs struct {
	MedicineName string `json:"medicine_name"`
	Dose         string `json:"dose"`
	Notes        string `json:"notes"`
}

func (s *VoiceService) applyCreateTreatment(ctx context.Context, userID, apiaryID, hiveID int64, raw datatypes.JSON) (string, int64, error) {
	var args createTreatmentArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", 0, fmt.Errorf("parse tool arguments: %w", err)
	}
	t, err := s.treatments.Create(ctx, userID, apiaryID, hiveID, TreatmentParams{
		TreatedAt:    time.Now(),
		MedicineName: args.MedicineName,
		Dose:         args.Dose,
		Notes:        args.Notes,
	})
	if err != nil {
		return "", 0, err
	}
	return model.VoiceActionResultTypeTreatment, t.ID, nil
}

type createHarvestArgs struct {
	Frames     int     `json:"frames"`
	HalfFrames int     `json:"half_frames"`
	Kilograms  float64 `json:"kilograms"`
	Notes      string  `json:"notes"`
}

func (s *VoiceService) applyCreateHarvest(ctx context.Context, userID, apiaryID, hiveID int64, raw datatypes.JSON) (string, int64, error) {
	var args createHarvestArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", 0, fmt.Errorf("parse tool arguments: %w", err)
	}
	h, err := s.harvests.Create(ctx, userID, apiaryID, hiveID, HarvestParams{
		HarvestedAt: time.Now(),
		Frames:      args.Frames,
		HalfFrames:  args.HalfFrames,
		Kilograms:   args.Kilograms,
		Notes:       args.Notes,
	})
	if err != nil {
		return "", 0, err
	}
	return model.VoiceActionResultTypeHarvest, h.ID, nil
}

type createFeedingArgs struct {
	FeedType string `json:"feed_type"`
	Amount   string `json:"amount"`
	Notes    string `json:"notes"`
}

func (s *VoiceService) applyCreateFeeding(ctx context.Context, userID, apiaryID, hiveID int64, raw datatypes.JSON) (string, int64, error) {
	var args createFeedingArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", 0, fmt.Errorf("parse tool arguments: %w", err)
	}
	f, err := s.feedings.Create(ctx, userID, apiaryID, hiveID, FeedingParams{
		FedAt:    time.Now(),
		FeedType: args.FeedType,
		Amount:   args.Amount,
		Notes:    args.Notes,
	})
	if err != nil {
		return "", 0, err
	}
	return model.VoiceActionResultTypeFeeding, f.ID, nil
}

type updateHiveStatusArgs struct {
	ReadyForHarvest       *bool     `json:"ready_for_harvest"`
	QueenNeedsReplacement *bool     `json:"queen_needs_replacement"`
	NeedsFood             *bool     `json:"needs_food"`
	BoxNeedsAdding        *bool     `json:"box_needs_adding"`
	Diseases              *[]string `json:"diseases"`
}

func (s *VoiceService) applyUpdateHiveStatus(ctx context.Context, userID, apiaryID, hiveID int64, raw datatypes.JSON) (string, int64, error) {
	var args updateHiveStatusArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "", 0, fmt.Errorf("parse tool arguments: %w", err)
	}
	hive, err := s.hives.Get(ctx, userID, apiaryID, hiveID)
	if err != nil {
		return "", 0, err
	}

	readyForHarvest := hive.ReadyForHarvest
	if args.ReadyForHarvest != nil {
		readyForHarvest = *args.ReadyForHarvest
	}
	queenNeedsReplacement := hive.QueenNeedsReplacement
	if args.QueenNeedsReplacement != nil {
		queenNeedsReplacement = *args.QueenNeedsReplacement
	}
	needsFood := hive.NeedsFood
	if args.NeedsFood != nil {
		needsFood = *args.NeedsFood
	}
	boxNeedsAdding := hive.BoxNeedsAdding
	if args.BoxNeedsAdding != nil {
		boxNeedsAdding = *args.BoxNeedsAdding
	}

	updated, err := s.hives.Update(ctx, userID, apiaryID, hiveID, hive.Name, hive.Type, hive.Active, readyForHarvest, queenNeedsReplacement, needsFood, boxNeedsAdding)
	if err != nil {
		return "", 0, err
	}

	if args.Diseases != nil {
		if err := s.syncHiveDiseases(ctx, userID, apiaryID, hiveID, *args.Diseases); err != nil {
			return "", 0, err
		}
	}

	return model.VoiceActionResultTypeHiveStatus, updated.ID, nil
}

func (s *VoiceService) syncHiveDiseases(ctx context.Context, userID, apiaryID, hiveID int64, desired []string) error {
	current, err := s.hives.DiseasesByHive(ctx, hiveID)
	if err != nil {
		return err
	}
	want := toSet(desired)
	have := make(map[string]int64, len(current))
	for _, d := range current {
		have[d.Disease] = d.ID
	}
	for name := range want {
		if _, ok := have[name]; !ok {
			if _, err := s.hives.AddDisease(ctx, userID, apiaryID, hiveID, name); err != nil {
				return fmt.Errorf("add disease %q: %w", name, err)
			}
		}
	}
	for name, id := range have {
		if !want[name] {
			if err := s.hives.RemoveDisease(ctx, userID, apiaryID, hiveID, id); err != nil {
				return fmt.Errorf("remove disease %q: %w", name, err)
			}
		}
	}
	return nil
}
