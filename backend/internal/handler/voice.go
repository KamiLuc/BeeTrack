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

func actionJSON(a *model.VoiceAction) map[string]any {
	return map[string]any{
		"id":               a.ID,
		"sequence":         a.Sequence,
		"hive_id":          a.HiveID,
		"tool_name":        a.ToolName,
		"status":           a.Status,
		"result_type":      a.ResultType,
		"result_record_id": a.ResultRecordID,
		"error_message":    a.ErrorMessage,
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
	case errors.Is(err, service.ErrRecordingNotFound):
		respond.Error(w, http.StatusNotFound, "RECORDING_NOT_FOUND", err.Error())
	case errors.Is(err, service.ErrRecordingNotCompleted):
		respond.Error(w, http.StatusConflict, "RECORDING_NOT_COMPLETED", err.Error())
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

func (h *VoiceHandler) Accept(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuth(w, r)
	if !ok {
		return
	}

	apiaryID, ok := parsePathID(w, r, "id", "invalid apiary id")
	if !ok {
		return
	}
	recordingID, ok := parsePathID(w, r, "recordingId", "invalid recording id")
	if !ok {
		return
	}

	rec, actions, err := h.voice.Accept(r.Context(), userID, apiaryID, recordingID)
	if err != nil {
		voiceError(w, err)
		return
	}

	items := make([]map[string]any, len(actions))
	for i, a := range actions {
		items[i] = actionJSON(a)
	}
	resp := recordingJSON(rec)
	resp["actions"] = items
	respond.JSON(w, http.StatusOK, resp)
}

func (h *VoiceHandler) Reject(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireAuth(w, r)
	if !ok {
		return
	}

	apiaryID, ok := parsePathID(w, r, "id", "invalid apiary id")
	if !ok {
		return
	}
	recordingID, ok := parsePathID(w, r, "recordingId", "invalid recording id")
	if !ok {
		return
	}

	rec, err := h.voice.Reject(r.Context(), userID, apiaryID, recordingID)
	if err != nil {
		voiceError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, recordingJSON(rec))
}
