package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	VoiceLLMCallPhaseResolveHive    = "resolve_hive"
	VoiceLLMCallPhaseProposeActions = "propose_actions"
)

type VoiceLLMCall struct {
	ID               int64
	VoiceRecordingID int64
	Phase            string
	Request          datatypes.JSON
	Response         datatypes.JSON
	IsError          bool
	CreatedAt        time.Time
}
