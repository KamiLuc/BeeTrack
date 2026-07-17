package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// TreatmentHandler handles HTTP requests for treatment resources.
type TreatmentHandler struct {
	treatments *service.TreatmentService
}

// NewTreatmentHandler creates a TreatmentHandler backed by svc.
func NewTreatmentHandler(treatments *service.TreatmentService) *TreatmentHandler {
	return &TreatmentHandler{treatments: treatments}
}

type treatmentRequest struct {
	TreatedAt    *time.Time `json:"treated_at"`
	MedicineName string     `json:"medicine_name"`
	Dose         string     `json:"dose"`
	Notes        string     `json:"notes"`
}

type bulkTreatmentRequest struct {
	treatmentRequest
	HiveIDs []int64 `json:"hive_ids"`
}

func (req treatmentRequest) toParams() service.TreatmentParams {
	var at time.Time
	if req.TreatedAt != nil {
		at = *req.TreatedAt
	}
	return service.TreatmentParams{
		TreatedAt:    at,
		MedicineName: req.MedicineName,
		Dose:         req.Dose,
		Notes:        req.Notes,
	}
}

func treatmentJSON(t *model.Treatment) map[string]any {
	return map[string]any{
		"id":              t.ID,
		"hive_id":         t.HiveID,
		"treated_by":      t.TreatedBy,
		"treated_by_name": t.TreatedByName,
		"treated_at":      t.TreatedAt,
		"medicine_name":   t.MedicineName,
		"dose":            t.Dose,
		"notes":           t.Notes,
		"created_at":      t.CreatedAt,
		"updated_at":      t.UpdatedAt,
	}
}

func treatmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrApiaryNotFound):
		respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
	case errors.Is(err, service.ErrHiveNotFound):
		respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
	case errors.Is(err, service.ErrTreatmentNotFound):
		respond.Error(w, http.StatusNotFound, "TREATMENT_NOT_FOUND", "treatment not found")
	case errors.Is(err, service.ErrTreatedAtRequired):
		respond.Error(w, http.StatusBadRequest, "TREATED_AT_REQUIRED", err.Error())
	case errors.Is(err, service.ErrMedicineNameRequired):
		respond.Error(w, http.StatusBadRequest, "MEDICINE_NAME_REQUIRED", err.Error())
	case errors.Is(err, service.ErrMedicineNameTooLong):
		respond.Error(w, http.StatusBadRequest, "MEDICINE_NAME_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrDoseTooLong):
		respond.Error(w, http.StatusBadRequest, "DOSE_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrTreatmentNotesTooLong):
		respond.Error(w, http.StatusBadRequest, "NOTES_TOO_LONG", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseTreatmentPathIDs(r *http.Request) (apiaryID, hiveID int64, err error) {
	apiaryID, err = strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return
	}
	hiveID, err = strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	return
}

// Medicines handles GET /api/v1/medicines — returns medicine names this user has previously used.
func (h *TreatmentHandler) Medicines(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	names, err := h.treatments.MedicineSuggestions(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	respond.JSON(w, http.StatusOK, names)
}

// Doses handles GET /api/v1/doses — returns doses this user has previously used.
func (h *TreatmentHandler) Doses(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	doses, err := h.treatments.DoseSuggestions(r.Context(), userID)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	respond.JSON(w, http.StatusOK, doses)
}

// BulkCreate handles POST /api/v1/apiaries/{id}/treatments/bulk — creates one treatment per hive in the apiary.
func (h *TreatmentHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary id")
		return
	}

	var req bulkTreatmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	count, err := h.treatments.BulkTreat(r.Context(), userID, apiaryID, req.HiveIDs, req.treatmentRequest.toParams())
	if err != nil {
		treatmentError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, map[string]any{"count": count})
}

// Create handles POST /api/v1/apiaries/{id}/hives/{hiveId}/treatments — creates a new treatment.
func (h *TreatmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseTreatmentPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	var req treatmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	t, err := h.treatments.Create(r.Context(), userID, apiaryID, hiveID, req.toParams())
	if err != nil {
		treatmentError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, treatmentJSON(t))
}

// Get handles GET /api/v1/apiaries/{id}/hives/{hiveId}/treatments/{treatmentId} — returns a single treatment.
func (h *TreatmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseTreatmentPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	treatmentID, err := strconv.ParseInt(r.PathValue("treatmentId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid treatment id")
		return
	}

	t, err := h.treatments.Get(r.Context(), userID, apiaryID, hiveID, treatmentID)
	if err != nil {
		treatmentError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, treatmentJSON(t))
}

// List handles GET /api/v1/apiaries/{id}/hives/{hiveId}/treatments — returns paginated treatments.
func (h *TreatmentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseTreatmentPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	treatments, total, err := h.treatments.List(r.Context(), userID, apiaryID, hiveID, limit, offset)
	if err != nil {
		treatmentError(w, err)
		return
	}

	items := make([]map[string]any, len(treatments))
	for i, t := range treatments {
		items[i] = treatmentJSON(t)
	}

	respond.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// Update handles PATCH /api/v1/apiaries/{id}/hives/{hiveId}/treatments/{treatmentId} — updates a treatment.
func (h *TreatmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseTreatmentPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	treatmentID, err := strconv.ParseInt(r.PathValue("treatmentId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid treatment id")
		return
	}

	var req treatmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	t, err := h.treatments.Update(r.Context(), userID, apiaryID, hiveID, treatmentID, req.toParams())
	if err != nil {
		treatmentError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, treatmentJSON(t))
}

// Delete handles DELETE /api/v1/apiaries/{id}/hives/{hiveId}/treatments/{treatmentId} — deletes a treatment.
func (h *TreatmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseTreatmentPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	treatmentID, err := strconv.ParseInt(r.PathValue("treatmentId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid treatment id")
		return
	}

	if err := h.treatments.Delete(r.Context(), userID, apiaryID, hiveID, treatmentID); err != nil {
		treatmentError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
