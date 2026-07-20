package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// HoneyBatchHandler handles HTTP requests for honey batch resources.
type HoneyBatchHandler struct {
	batches *service.HoneyBatchService
}

// NewHoneyBatchHandler creates a HoneyBatchHandler backed by svc.
func NewHoneyBatchHandler(batches *service.HoneyBatchService) *HoneyBatchHandler {
	return &HoneyBatchHandler{batches: batches}
}

func certificationJSON(c *model.HoneyBatchCertification) any {
	if c == nil {
		return nil
	}
	return map[string]any{
		"status":                 c.Status,
		"transaction_hash":       c.TransactionHash,
		"block_number":           c.BlockNumber,
		"gas_used":               c.GasUsed,
		"confirmation_timestamp": c.ConfirmationTimestamp,
		"created_at":             c.CreatedAt,
	}
}

func honeyBatchJSON(b *model.HoneyBatch, cert *model.HoneyBatchCertification) map[string]any {
	return map[string]any{
		"id":                 b.ID,
		"apiary_id":          b.ApiaryID,
		"verification_token": b.VerificationToken,
		"gathering_date":     b.GatheringDate,
		"amount_grams":       b.AmountGrams,
		"processing_method":  b.ProcessingMethod,
		"honey_type":         b.HoneyType,
		"pdf_file_hash":      b.PDFFileHash,
		"created_at":         b.CreatedAt,
		"updated_at":         b.UpdatedAt,
		"certification":      certificationJSON(cert),
	}
}

func honeyBatchError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrBatchNotFound):
		respond.Error(w, http.StatusNotFound, "BATCH_NOT_FOUND", "honey batch not found")
	case errors.Is(err, service.ErrApiaryNotFound):
		respond.Error(w, http.StatusNotFound, "APIARY_NOT_FOUND", "apiary not found")
	case errors.Is(err, service.ErrInvalidAmount):
		respond.Error(w, http.StatusBadRequest, "INVALID_AMOUNT", err.Error())
	case errors.Is(err, service.ErrHoneyTypeRequired):
		respond.Error(w, http.StatusBadRequest, "HONEY_TYPE_REQUIRED", err.Error())
	case errors.Is(err, service.ErrHoneyTypeTooLong):
		respond.Error(w, http.StatusBadRequest, "HONEY_TYPE_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrInvalidProcessingMethod):
		respond.Error(w, http.StatusBadRequest, "INVALID_PROCESSING_METHOD", err.Error())
	case errors.Is(err, service.ErrPDFRequired):
		respond.Error(w, http.StatusBadRequest, "PDF_REQUIRED", err.Error())
	case errors.Is(err, service.ErrInvalidPDFType):
		respond.Error(w, http.StatusBadRequest, "INVALID_PDF_TYPE", err.Error())
	case errors.Is(err, service.ErrPDFTooLarge):
		respond.Error(w, http.StatusRequestEntityTooLarge, "PDF_TOO_LARGE", err.Error())
	case errors.Is(err, service.ErrBatchNotCertified):
		respond.Error(w, http.StatusConflict, "BATCH_NOT_CERTIFIED", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseHoneyBatchID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

const maxCreateHoneyBatchBytes = 11 * 1024 * 1024

// Create handles POST /api/v1/honey-batches — creates a new honey batch from a multipart form with a "lab_pdf" file. Never fails due to blockchain state: certification, if requested, is only enqueued.
func (h *HoneyBatchHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	if err := r.ParseMultipartForm(maxCreateHoneyBatchBytes); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "could not parse multipart form")
		return
	}

	apiaryID, err := strconv.ParseInt(r.FormValue("apiary_id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid apiary_id")
		return
	}

	gatheringDate, err := time.Parse("2006-01-02", r.FormValue("gathering_date"))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_DATE", "gathering_date must be YYYY-MM-DD")
		return
	}

	amountGrams, err := strconv.ParseInt(r.FormValue("amount_grams"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_AMOUNT", "invalid amount_grams")
		return
	}

	file, header, err := r.FormFile("lab_pdf")
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "MISSING_FILE", "field 'lab_pdf' is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not read file")
		return
	}

	req := service.CreateBatchRequest{
		GatheringDate:        gatheringDate,
		AmountGrams:          amountGrams,
		ProcessingMethod:     r.FormValue("processing_method"),
		HoneyType:            r.FormValue("honey_type"),
		PDFMimeType:          header.Header.Get("Content-Type"),
		PDFData:              data,
		RequestCertification: r.FormValue("request_certification") == "true",
	}

	batch, err := h.batches.CreateBatch(r.Context(), userID, apiaryID, req)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	// No certification row exists yet — the worker creates one once it claims the job — so a synthetic "queued" placeholder reflects it here.
	var cert *model.HoneyBatchCertification
	if req.RequestCertification {
		cert = &model.HoneyBatchCertification{Status: model.CertificationStatusQueued}
	}
	respond.JSON(w, http.StatusCreated, honeyBatchJSON(batch, cert))
}

// Get handles GET /api/v1/honey-batches/{id} — returns a single batch owned by the caller.
func (h *HoneyBatchHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseHoneyBatchID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid honey batch id")
		return
	}

	result, err := h.batches.GetBatch(r.Context(), userID, id)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, honeyBatchJSON(result.Batch, result.Certification))
}

// List handles GET /api/v1/honey-batches — returns the caller's paginated batches, each with its latest certification status.
func (h *HoneyBatchHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
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

	items, total, err := h.batches.ListBatches(r.Context(), userID, limit, offset)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	out := make([]map[string]any, len(items))
	for i, it := range items {
		out[i] = honeyBatchJSON(it.Batch, it.Certification)
	}

	respond.JSON(w, http.StatusOK, map[string]any{"items": out, "total": total})
}

type updateHoneyBatchRequest struct {
	HoneyType string `json:"honey_type"`
}

// Update handles PATCH /api/v1/honey-batches/{id} — updates a batch's honey_type, the only mutable field.
func (h *HoneyBatchHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseHoneyBatchID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid honey batch id")
		return
	}

	var req updateHoneyBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if _, err := h.batches.UpdateHoneyType(r.Context(), userID, id, req.HoneyType); err != nil {
		honeyBatchError(w, err)
		return
	}

	result, err := h.batches.GetBatch(r.Context(), userID, id)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, honeyBatchJSON(result.Batch, result.Certification))
}

// Delete handles DELETE /api/v1/honey-batches/{id} — soft-deletes a batch owned by the caller.
func (h *HoneyBatchHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseHoneyBatchID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid honey batch id")
		return
	}

	if err := h.batches.DeleteBatch(r.Context(), userID, id); err != nil {
		honeyBatchError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PDF handles GET /api/v1/honey-batches/{id}/pdf — serves the lab PDF for a batch owned by the caller.
func (h *HoneyBatchHandler) PDF(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseHoneyBatchID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid honey batch id")
		return
	}

	path, err := h.batches.GetBatchPDF(r.Context(), userID, id)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, path)
}
