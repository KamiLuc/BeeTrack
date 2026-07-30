package handler

import (
	"errors"
	"net/http"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

type VoiceHandler struct {
	voice *service.VoiceService
}

func NewVoiceHandler(voice *service.VoiceService) *VoiceHandler {
	return &VoiceHandler{voice: voice}
}

func recordingJSON(rec *model.VoiceRecording) map[string]any {
	return map[string]any{
		"recording_id": rec.ID,
		"status":       rec.Status,
	}
}

func voiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrApiaryNotFound):
		respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
	case errors.Is(err, service.ErrInvalidAudioType):
		respond.Error(w, http.StatusBadRequest, "INVALID_AUDIO_TYPE", err.Error())
	case errors.Is(err, service.ErrRecordingTooLong):
		respond.Error(w, http.StatusRequestEntityTooLarge, "RECORDING_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrMaxRecordingsReached):
		respond.Error(w, http.StatusUnprocessableEntity, "MAX_RECORDINGS_REACHED", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func (h *VoiceHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuth(w, r)
	if !ok {
		return
	}

	apiaryID, ok := parsePathID(w, r, "id", "invalid apiary id")
	if !ok {
		return
	}

	data, mimeType, ok := parseAudioFile(w, r)
	if !ok {
		return
	}

	rec, err := h.voice.Upload(r.Context(), userID, apiaryID, mimeType, data)
	if err != nil {
		voiceError(w, err)
		return
	}

	respond.JSON(w, http.StatusAccepted, recordingJSON(rec))
}
