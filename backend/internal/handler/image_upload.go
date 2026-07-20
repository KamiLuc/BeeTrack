package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// maxUploadBodyBytes bounds the raw multipart request body.
const maxUploadBodyBytes = service.MaxImageBytes + 1<<20

// parseImageFile enforces the upload size limit, parses the multipart body, and reads the "image" field.
func parseImageFile(w http.ResponseWriter, r *http.Request) (data []byte, mimeType string, ok bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBodyBytes)

	if err := r.ParseMultipartForm(maxUploadBodyBytes); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			respond.Error(w, http.StatusRequestEntityTooLarge, "IMAGE_TOO_LARGE", service.ErrImageTooLarge.Error())
		} else {
			respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "could not parse multipart form")
		}
		return nil, "", false
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "MISSING_FILE", "field 'image' is required")
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
