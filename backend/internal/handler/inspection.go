package handler

import (
	"context"
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

// InspectionHandler handles HTTP requests for inspection resources.
type InspectionHandler struct {
	inspections *service.InspectionService
}

// NewInspectionHandler creates an InspectionHandler backed by svc.
func NewInspectionHandler(inspections *service.InspectionService) *InspectionHandler {
	return &InspectionHandler{inspections: inspections}
}

type inspectionRequest struct {
	Aggressiveness        string     `json:"aggressiveness"`
	BroodPattern          string     `json:"brood_pattern"`
	FramesAddedDrawn      *int       `json:"frames_added_drawn"`
	FramesAddedFoundation *int       `json:"frames_added_foundation"`
	FramesAddedHoney      *int       `json:"frames_added_honey"`
	FramesBrood           *int       `json:"frames_brood"`
	FramesHoney           *int       `json:"frames_honey"`
	FramesPollen          *int       `json:"frames_pollen"`
	InspectedAt           *time.Time `json:"inspected_at"`
	Notes                 string     `json:"notes"`
	QueenAdded            bool       `json:"queen_added"`
	QueenCellsCount       *int       `json:"queen_cells_count"`
	QueenStatus           string     `json:"queen_status"`
}

func (req inspectionRequest) toParams() service.InspectionParams {
	var at time.Time
	if req.InspectedAt != nil {
		at = *req.InspectedAt
	}
	return service.InspectionParams{
		Aggressiveness:        req.Aggressiveness,
		BroodPattern:          req.BroodPattern,
		FramesAddedDrawn:      req.FramesAddedDrawn,
		FramesAddedFoundation: req.FramesAddedFoundation,
		FramesAddedHoney:      req.FramesAddedHoney,
		FramesBrood:           req.FramesBrood,
		FramesHoney:           req.FramesHoney,
		FramesPollen:          req.FramesPollen,
		InspectedAt:           at,
		Notes:                 req.Notes,
		QueenAdded:            req.QueenAdded,
		QueenCellsCount:       req.QueenCellsCount,
		QueenStatus:           req.QueenStatus,
	}
}

func diseaseJSON(d *model.InspectionDisease) map[string]any {
	return map[string]any{
		"id":         d.ID,
		"disease":    d.Disease,
		"notes":      d.Notes,
		"created_at": d.CreatedAt,
	}
}

func inspectionJSON(insp *model.Inspection, diseases []*model.InspectionDisease) map[string]any {
	dd := make([]map[string]any, len(diseases))
	for i, d := range diseases {
		dd[i] = diseaseJSON(d)
	}
	return map[string]any{
		"id":                      insp.ID,
		"hive_id":                 insp.HiveID,
		"inspected_by":            insp.InspectedBy,
		"inspected_at":            insp.InspectedAt,
		"queen_status":            insp.QueenStatus,
		"brood_pattern":           insp.BroodPattern,
		"frames_brood":            insp.FramesBrood,
		"frames_honey":            insp.FramesHoney,
		"frames_pollen":           insp.FramesPollen,
		"queen_cells_count":       insp.QueenCellsCount,
		"aggressiveness":          insp.Aggressiveness,
		"frames_added_foundation": insp.FramesAddedFoundation,
		"frames_added_drawn":      insp.FramesAddedDrawn,
		"frames_added_honey":      insp.FramesAddedHoney,
		"queen_added":             insp.QueenAdded,
		"notes":                   insp.Notes,
		"diseases":                dd,
		"created_at":              insp.CreatedAt,
		"updated_at":              insp.UpdatedAt,
	}
}

func (h *InspectionHandler) withDiseases(ctx context.Context, insp *model.Inspection) (map[string]any, error) {
	diseases, err := h.inspections.DiseasesByInspection(ctx, insp.ID)
	if err != nil {
		return nil, err
	}
	return inspectionJSON(insp, diseases), nil
}

func inspectionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrApiaryNotFound):
		respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
	case errors.Is(err, service.ErrHiveNotFound):
		respond.Error(w, http.StatusNotFound, "HIVE_NOT_FOUND", "hive not found")
	case errors.Is(err, service.ErrInspectionNotFound):
		respond.Error(w, http.StatusNotFound, "INSPECTION_NOT_FOUND", "inspection not found")
	case errors.Is(err, service.ErrInspectedAtRequired):
		respond.Error(w, http.StatusBadRequest, "INSPECTED_AT_REQUIRED", err.Error())
	case errors.Is(err, service.ErrInvalidAggressiveness):
		respond.Error(w, http.StatusBadRequest, "INVALID_AGGRESSIVENESS", err.Error())
	case errors.Is(err, service.ErrInvalidBroodPattern):
		respond.Error(w, http.StatusBadRequest, "INVALID_BROOD_PATTERN", err.Error())
	case errors.Is(err, service.ErrInvalidQueenStatus):
		respond.Error(w, http.StatusBadRequest, "INVALID_QUEEN_STATUS", err.Error())
	case errors.Is(err, service.ErrInvalidDisease):
		respond.Error(w, http.StatusBadRequest, "INVALID_DISEASE", err.Error())
	case errors.Is(err, service.ErrDiseaseNotFound):
		respond.Error(w, http.StatusNotFound, "DISEASE_NOT_FOUND", "disease not found")
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseInspectionPathIDs(r *http.Request) (apiaryID, hiveID int64, err error) {
	apiaryID, err = strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return
	}
	hiveID, err = strconv.ParseInt(r.PathValue("hiveId"), 10, 64)
	return
}

// Create handles POST /api/v1/apiaries/{id}/hives/{hiveId}/inspections — creates a new inspection.
func (h *InspectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseInspectionPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	var req inspectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	insp, err := h.inspections.Create(r.Context(), userID, apiaryID, hiveID, req.toParams())
	if err != nil {
		inspectionError(w, err)
		return
	}

	body, err := h.withDiseases(r.Context(), insp)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	respond.JSON(w, http.StatusCreated, body)
}

// Get handles GET /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId} — returns a single inspection.
func (h *InspectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseInspectionPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	inspectionID, err := strconv.ParseInt(r.PathValue("inspectionId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid inspection id")
		return
	}

	insp, err := h.inspections.Get(r.Context(), userID, apiaryID, hiveID, inspectionID)
	if err != nil {
		inspectionError(w, err)
		return
	}

	body, err := h.withDiseases(r.Context(), insp)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	respond.JSON(w, http.StatusOK, body)
}

// List handles GET /api/v1/apiaries/{id}/hives/{hiveId}/inspections — returns paginated inspections.
func (h *InspectionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseInspectionPathIDs(r)
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

	inspections, err := h.inspections.List(r.Context(), userID, apiaryID, hiveID, limit, offset)
	if err != nil {
		inspectionError(w, err)
		return
	}

	ids := make([]int64, len(inspections))
	for i, insp := range inspections {
		ids[i] = insp.ID
	}
	diseaseMap, err := h.inspections.DiseasesForInspections(r.Context(), ids)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	items := make([]map[string]any, len(inspections))
	for i, insp := range inspections {
		items[i] = inspectionJSON(insp, diseaseMap[insp.ID])
	}

	respond.JSON(w, http.StatusOK, items)
}

// Update handles PATCH /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId} — updates an inspection.
func (h *InspectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseInspectionPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	inspectionID, err := strconv.ParseInt(r.PathValue("inspectionId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid inspection id")
		return
	}

	var req inspectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	insp, err := h.inspections.Update(r.Context(), userID, apiaryID, hiveID, inspectionID, req.toParams())
	if err != nil {
		inspectionError(w, err)
		return
	}

	body, err := h.withDiseases(r.Context(), insp)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	respond.JSON(w, http.StatusOK, body)
}

// AddDisease handles POST /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/diseases — adds a disease to an inspection.
func (h *InspectionHandler) AddDisease(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseInspectionPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	inspectionID, err := strconv.ParseInt(r.PathValue("inspectionId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid inspection id")
		return
	}

	var req struct {
		Disease string `json:"disease"`
		Notes   string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	d, err := h.inspections.AddDisease(r.Context(), userID, apiaryID, hiveID, inspectionID, req.Disease, req.Notes)
	if err != nil {
		inspectionError(w, err)
		return
	}

	respond.JSON(w, http.StatusCreated, diseaseJSON(d))
}

// RemoveDisease handles DELETE /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/diseases/{diseaseId} — removes a disease from an inspection.
func (h *InspectionHandler) RemoveDisease(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseInspectionPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	inspectionID, err := strconv.ParseInt(r.PathValue("inspectionId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid inspection id")
		return
	}

	diseaseID, err := strconv.ParseInt(r.PathValue("diseaseId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid disease id")
		return
	}

	if err := h.inspections.RemoveDisease(r.Context(), userID, apiaryID, hiveID, inspectionID, diseaseID); err != nil {
		inspectionError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId} — deletes an inspection.
func (h *InspectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	apiaryID, hiveID, err := parseInspectionPathIDs(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary or hive id")
		return
	}

	inspectionID, err := strconv.ParseInt(r.PathValue("inspectionId"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid inspection id")
		return
	}

	if err := h.inspections.Delete(r.Context(), userID, apiaryID, hiveID, inspectionID); err != nil {
		inspectionError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
