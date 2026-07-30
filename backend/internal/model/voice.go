package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	VoiceRecordingStatusPending    = "pending"
	VoiceRecordingStatusProcessing = "processing"
	VoiceRecordingStatusCompleted  = "completed"
	VoiceRecordingStatusAccepted   = "accepted"
	VoiceRecordingStatusRejected   = "rejected"
	VoiceRecordingStatusFailed     = "failed"
	VoiceRecordingStatusCancelled  = "cancelled"

	MaxWhisperRetries = 3
)

const (
	VoiceActionStatusProposed = "proposed"
	VoiceActionStatusApplied  = "applied"
	VoiceActionStatusError    = "error"
)

const (
	VoiceActionToolCreateInspection = "create_inspection"
	VoiceActionToolCreateTreatment  = "create_treatment"
	VoiceActionToolCreateHarvest    = "create_harvest"
	VoiceActionToolCreateFeeding    = "create_feeding"
	VoiceActionToolUpdateHiveStatus = "update_hive_status"
)

const (
	VoiceActionResultTypeInspection = "inspection"
	VoiceActionResultTypeTreatment  = "treatment"
	VoiceActionResultTypeHarvest    = "harvest"
	VoiceActionResultTypeFeeding    = "feeding"
	VoiceActionResultTypeHiveStatus = "hive_status"
	VoiceActionResultTypeError      = "error"
)

// VoiceRecording is one uploaded voice memo and its worker lifecycle. No hive_id here — a single
// recording can name zero, one, or several hives within the apiary; which hive(s) it resolved to
// lives per-action on VoiceAction instead.
type VoiceRecording struct {
	ID               int64
	UserID           int64
	ApiaryID         int64
	Status           string
	AudioPath        *string
	Transcript       *string
	DetectedLanguage *string
	ErrorMessage     *string
	RetryCount       int
	NextAttemptAt    *time.Time
	CreatedAt        time.Time
	ProcessedAt      *time.Time
}

// VoiceAction is one proposed (and later applied/errored) tool call Claude produced against a
// VoiceRecording's transcript. ToolArguments IS the proposal — nothing is executed against
// inspections/treatments/harvests/feedings/hives until Accept runs the real service call.
type VoiceAction struct {
	ID                int64
	VoiceRecordingID  int64
	Sequence          int
	SpokenHiveName    *string
	HiveID            *int64
	ToolName          *string
	ToolArguments     datatypes.JSON
	Status            string
	PreviousHiveState datatypes.JSON
	ResultType        *string
	ResultRecordID    *int64
	ErrorMessage      *string
	CreatedAt         time.Time
}
