package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// InspectionImageHandler handles HTTP requests for inspection image resources.
type InspectionImageHandler struct {
	images *service.InspectionImageService
}

// NewInspectionImageHandler creates an InspectionImageHandler backed by svc.
func NewInspectionImageHandler(images *service.InspectionImageService) *InspectionImageHandler {
	return &InspectionImageHandler{images: images}
}

func imageJSON(img *model.InspectionImage) map[string]any {
	return map[string]any{
		"id":            img.ID,
		"inspection_id": img.InspectionID,
		"mime_type":     img.MimeType,
		"created_at":    img.CreatedAt,
	}
}

func inspectionImageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrApiaryNotFound):
		respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
	case errors.Is(err, service.ErrHiveNotFound):
		respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
	case errors.Is(err, service.ErrInspectionNotFound):
		respond.Error(w, http.StatusNotFound, "INSPECTION_NOT_FOUND", "inspection not found")
	case errors.Is(err, service.ErrImageNotFound):
		respond.Error(w, http.StatusNotFound, "IMAGE_NOT_FOUND", "image not found")
	case errors.Is(err, service.ErrInvalidImageType):
		respond.Error(w, http.StatusBadRequest, "INVALID_IMAGE_TYPE", err.Error())
	case errors.Is(err, service.ErrImageTooLarge):
		respond.Error(w, http.StatusRequestEntityTooLarge, "IMAGE_TOO_LARGE", err.Error())
	case errors.Is(err, service.ErrMaxImagesReached):
		respond.Error(w, http.StatusUnprocessableEntity, "MAX_IMAGES_REACHED", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseImagePathIDs(r *http.Request) (apiaryID, hiveID, inspectionID int64, err error) {
	apiaryID, err = strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return
	}
	hiveID, err = strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	if err != nil {
		return
	}
	inspectionID, err = strconv.ParseInt(r.PathValue("inspectionId"), 10, 64)
	return
}

// Upload handles POST .../images — accepts a multipart upload and stores the image.
func (h *InspectionImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, inspectionID, err := parseImagePathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid path id")
		return
	}

	data, mimeType, ok := parseImageFile(w, r)
	if !ok {
		return
	}

	img, err := h.images.Upload(r.Context(), userID, apiaryID, hiveID, inspectionID, mimeType, data)
	if err != nil {
		inspectionImageError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, imageJSON(img))
}

// List handles GET .../images — returns all image metadata for an inspection.
func (h *InspectionImageHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, inspectionID, err := parseImagePathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid path id")
		return
	}

	imgs, err := h.images.List(r.Context(), userID, apiaryID, hiveID, inspectionID)
	if err != nil {
		inspectionImageError(w, err)
		return
	}

	items := make([]map[string]any, len(imgs))
	for i, img := range imgs {
		items[i] = imageJSON(img)
	}
	respond.JSON(w, http.StatusOK, items)
}

// Delete handles DELETE .../images/{imageId} — removes an image.
func (h *InspectionImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, inspectionID, err := parseImagePathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid path id")
		return
	}

	imageID, err := strconv.ParseInt(r.PathValue("imageId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid image id")
		return
	}

	if err := h.images.Delete(r.Context(), userID, apiaryID, hiveID, inspectionID, imageID); err != nil {
		inspectionImageError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ServeFile handles GET .../images/{imageId}/file — serves the image bytes.
func (h *InspectionImageHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, inspectionID, err := parseImagePathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid path id")
		return
	}

	imageID, err := strconv.ParseInt(r.PathValue("imageId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid image id")
		return
	}

	imgs, err := h.images.List(r.Context(), userID, apiaryID, hiveID, inspectionID)
	if err != nil {
		inspectionImageError(w, err)
		return
	}

	var target *model.InspectionImage
	for _, img := range imgs {
		if img.ID == imageID {
			target = img
			break
		}
	}
	if target == nil {
		respond.Error(w, http.StatusNotFound, "IMAGE_NOT_FOUND", "image not found")
		return
	}

	w.Header().Set("Content-Type", target.MimeType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	http.ServeFile(w, r, h.images.FilePath(target.Filename))
}
