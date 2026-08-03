package repository

import (
	"context"
	"errors"
	"time"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VoiceRepository struct {
	db *gorm.DB
}

func NewVoiceRepository(db *gorm.DB) *VoiceRepository {
	return &VoiceRepository{db: db}
}

func (r *VoiceRepository) CreateRecording(ctx context.Context, rec *model.VoiceRecording) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *VoiceRepository) CountRecordingsByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.VoiceRecording{}).
		Where("user_id = ? AND status IN ?", userID, []string{
			model.VoiceRecordingStatusPending,
			model.VoiceRecordingStatusProcessing,
			model.VoiceRecordingStatusCompleted,
		}).
		Count(&count).Error
	return count, err
}

// Returns nil, nil if not found — callers must still check UserID themselves before treating this as an ownership check.
func (r *VoiceRepository) GetRecordingByID(ctx context.Context, id int64) (*model.VoiceRecording, error) {
	var rec model.VoiceRecording
	err := r.db.WithContext(ctx).First(&rec, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *VoiceRepository) ListRecordingsByApiaryID(ctx context.Context, apiaryID int64, limit, offset int) ([]*model.VoiceRecording, error) {
	var recs []*model.VoiceRecording
	err := r.db.WithContext(ctx).
		Where("apiary_id = ?", apiaryID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&recs).Error
	return recs, err
}

func (r *VoiceRepository) CountRecordingsByApiaryID(ctx context.Context, apiaryID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.VoiceRecording{}).
		Where("apiary_id = ?", apiaryID).
		Count(&count).Error
	return count, err
}

func (r *VoiceRepository) UpdateRecording(ctx context.Context, rec *model.VoiceRecording) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *VoiceRepository) DeleteRecording(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.VoiceRecording{}, id).Error
}

func (r *VoiceRepository) ClaimNext(ctx context.Context) (*model.VoiceRecording, error) {
	var rec model.VoiceRecording
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND (next_attempt_at IS NULL OR next_attempt_at <= NOW())",
				model.VoiceRecordingStatusPending).
			Order("created_at ASC").
			Limit(1).
			First(&rec).Error
		if err != nil {
			return err
		}
		return tx.Model(&rec).Updates(map[string]any{
			"status":     model.VoiceRecordingStatusProcessing,
			"claimed_at": gorm.Expr("NOW()"),
		}).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rec.Status = model.VoiceRecordingStatusProcessing
	return &rec, nil
}

func (r *VoiceRepository) MarkCompleted(ctx context.Context, id int64, transcript, language string) error {
	updates := map[string]any{
		"status":       model.VoiceRecordingStatusCompleted,
		"transcript":   transcript,
		"audio_path":   nil,
		"claimed_at":   nil,
		"processed_at": gorm.Expr("NOW()"),
	}
	if language != "" {
		updates["detected_language"] = language
	}
	return r.db.WithContext(ctx).Model(&model.VoiceRecording{}).Where("id = ?", id).Updates(updates).Error
}

func (r *VoiceRepository) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	return r.db.WithContext(ctx).Model(&model.VoiceRecording{}).Where("id = ?", id).Updates(map[string]any{
		"status":        model.VoiceRecordingStatusFailed,
		"error_message": errMsg,
		"audio_path":    nil,
		"claimed_at":    nil,
		"processed_at":  gorm.Expr("NOW()"),
	}).Error
}

func (r *VoiceRepository) MarkRetry(ctx context.Context, id int64, errMsg string, nextAttemptAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.VoiceRecording{}).Where("id = ?", id).Updates(map[string]any{
		"status":          model.VoiceRecordingStatusPending,
		"retry_count":     gorm.Expr("retry_count + 1"),
		"error_message":   errMsg,
		"next_attempt_at": nextAttemptAt,
		"claimed_at":      nil,
	}).Error
}

// Does not touch retry_count — distinct from the Whisper-call retry above.
func (r *VoiceRepository) SweepStuckProcessing(ctx context.Context, olderThan time.Duration) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&model.VoiceRecording{}).
		Where("status = ? AND claimed_at < ?", model.VoiceRecordingStatusProcessing, time.Now().Add(-olderThan)).
		Updates(map[string]any{
			"status":     model.VoiceRecordingStatusPending,
			"claimed_at": nil,
		})
	return result.RowsAffected, result.Error
}

func (r *VoiceRepository) CreateAction(ctx context.Context, action *model.VoiceAction) error {
	return r.db.WithContext(ctx).Create(action).Error
}

func (r *VoiceRepository) ListActionsByRecordingID(ctx context.Context, recordingID int64) ([]*model.VoiceAction, error) {
	var actions []*model.VoiceAction
	err := r.db.WithContext(ctx).
		Where("voice_recording_id = ?", recordingID).
		Order("sequence ASC").
		Find(&actions).Error
	return actions, err
}

func (r *VoiceRepository) ListActionsByRecordingIDs(ctx context.Context, recordingIDs []int64) ([]*model.VoiceAction, error) {
	if len(recordingIDs) == 0 {
		return []*model.VoiceAction{}, nil
	}
	var actions []*model.VoiceAction
	err := r.db.WithContext(ctx).
		Where("voice_recording_id IN ?", recordingIDs).
		Order("voice_recording_id ASC, sequence ASC").
		Find(&actions).Error
	return actions, err
}

func (r *VoiceRepository) UpdateAction(ctx context.Context, action *model.VoiceAction) error {
	return r.db.WithContext(ctx).Save(action).Error
}

func (r *VoiceRepository) DeleteActionsByRecordingID(ctx context.Context, recordingID int64) error {
	return r.db.WithContext(ctx).Where("voice_recording_id = ?", recordingID).Delete(&model.VoiceAction{}).Error
}

func (r *VoiceRepository) CreateLLMCall(ctx context.Context, call *model.VoiceLLMCall) error {
	return r.db.WithContext(ctx).Create(call).Error
}
