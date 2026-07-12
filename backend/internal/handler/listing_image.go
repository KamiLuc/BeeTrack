package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// ListingImageHandler handles HTTP requests for marketplace listing images.
type ListingImageHandler struct {
	images *service.ListingImageService
}

// NewListingImageHandler creates a ListingImageHandler backed by svc.
func NewListingImageHandler(images *service.ListingImageService) *ListingImageHandler {
	return &ListingImageHandler{images: images}
}

func listingImageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrListingNotFound):
		respond.Error(w, http.StatusNotFound, "LISTING_NOT_FOUND", "listing not found")
	case errors.Is(err, service.ErrListingImageNotFound):
		respond.Error(w, http.StatusNotFound, "IMAGE_NOT_FOUND", "image not found")
	case errors.Is(err, service.ErrNotListingOwner):
		respond.Error(w, http.StatusForbidden, "NOT_OWNER", "not the listing owner")
	case errors.Is(err, service.ErrInvalidImageType):
		respond.Error(w, http.StatusBadRequest, "INVALID_IMAGE_TYPE", err.Error())
	case errors.Is(err, service.ErrImageTooLarge):
		respond.Error(w, http.StatusRequestEntityTooLarge, "IMAGE_TOO_LARGE", err.Error())
	case errors.Is(err, service.ErrListingImageMaxImages):
		respond.Error(w, http.StatusBadRequest, "TOO_MANY_IMAGES", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseListingImagePathIDs(r *http.Request) (listingID, imageID int64, err error) {
	listingID, err = strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return
	}
	imageID, err = strconv.ParseInt(r.PathValue("imageId"), 10, 64)
	return
}

// Upload handles POST /api/v1/listings/{id}/images — stores a multipart image upload.
func (h *ListingImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	listingID, err := parseListingID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid listing id")
		return
	}

	if err := r.ParseMultipartForm(11 * 1024 * 1024); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "could not parse multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "MISSING_FILE", "field 'image' is required")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	data, err := io.ReadAll(file)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not read file")
		return
	}

	img, err := h.images.Upload(r.Context(), userID, listingID, mimeType, data)
	if err != nil {
		listingImageError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, listingImageJSON(*img))
}

// ServeFile handles GET /api/v1/listings/{id}/images/{imageId}/file — serves the image bytes (public).
func (h *ListingImageHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	listingID, imageID, err := parseListingImagePathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid path id")
		return
	}

	img, err := h.images.GetFile(r.Context(), listingID, imageID)
	if err != nil {
		listingImageError(w, err)
		return
	}

	w.Header().Set("Content-Type", h.images.ContentType(img.ImageURL))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, h.images.FilePath(img.ImageURL))
}

// Delete handles DELETE /api/v1/listings/{id}/images/{imageId} — removes a listing image.
func (h *ListingImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	listingID, imageID, err := parseListingImagePathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid path id")
		return
	}

	if err := h.images.Delete(r.Context(), userID, listingID, imageID); err != nil {
		listingImageError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
