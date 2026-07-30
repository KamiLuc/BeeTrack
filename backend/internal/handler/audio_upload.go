package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

const maxAudioUploadBodyBytes = service.MaxAudioBytes + 1<<20

func parseAudioFile(w http.ResponseWriter, r *http.Request) (data []byte, mimeType string, ok bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAudioUploadBodyBytes)

	if err := r.ParseMultipartForm(maxAudioUploadBodyBytes); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respond.Error(w, http.StatusRequestEntityTooLarge, "RECORDING_TOO_LONG", service.ErrRecordingTooLong.Error())
		} else {
			respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "could not parse multipart form")
		}
		return nil, "", false
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "MISSING_FILE", "field 'audio' is required")
		return nil, "", false
	}
	defer file.Close()

	data, err = io.ReadAll(file)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not read file")
		return nil, "", false
	}
	return data, header.Header.Get("Content-Type"), true
}
