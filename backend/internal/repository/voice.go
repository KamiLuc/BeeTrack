package repository

import (
	"context"
	"errors"

	"github.com/beetrack/backend/internal/model"
	"gorm.io/gorm"
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

func (r *VoiceRepository) ListRecordingsByApiaryID(ctx context.Context, apiaryID int64) ([]*model.VoiceRecording, error) {
	var recs []*model.VoiceRecording
	err := r.db.WithContext(ctx).
		Where("apiary_id = ?", apiaryID).
		Order("created_at DESC").
		Find(&recs).Error
	return recs, err
}

func (r *VoiceRepository) UpdateRecording(ctx context.Context, rec *model.VoiceRecording) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *VoiceRepository) DeleteRecording(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.VoiceRecording{}, id).Error
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

func (r *VoiceRepository) UpdateAction(ctx context.Context, action *model.VoiceAction) error {
	return r.db.WithContext(ctx).Save(action).Error
}
