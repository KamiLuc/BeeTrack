package handler

import (
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
		"chain_id":               c.ChainID,
		"contract_address":       c.ContractAddress,
		"transaction_hash":       c.TransactionHash,
		"block_number":           c.BlockNumber,
		"gas_used":               c.GasUsed,
		"confirmation_timestamp": c.ConfirmationTimestamp,
		"created_at":             c.CreatedAt,
	}
}

func certificationRequestSummaryJSON(req *model.HoneyBatchCertificationRequest) any {
	if req == nil {
		return nil
	}
	return map[string]any{
		"status":           req.Status,
		"rejection_reason": req.RejectionReason,
		"created_at":       req.CreatedAt,
	}
}

// honeyBatchJSON builds the owner-facing batch representation, including the internal numeric id.
func honeyBatchJSON(b *model.HoneyBatch, cert *model.HoneyBatchCertification, certRequest *model.HoneyBatchCertificationRequest, verificationURL string) map[string]any {
	out := publicHoneyBatchJSON(b, cert, verificationURL)
	out["id"] = b.ID
	out["certification_request"] = certificationRequestSummaryJSON(certRequest)
	return out
}

// publicHoneyBatchJSON builds the batch representation safe to expose on the public, token-scoped verification page — no internal numeric id, no certification request state.
func publicHoneyBatchJSON(b *model.HoneyBatch, cert *model.HoneyBatchCertification, verificationURL string) map[string]any {
	return map[string]any{
		"verification_token": b.VerificationToken,
		"verification_url":   verificationURL,
		"gathering_date":     b.GatheringDate,
		"amount_grams":       b.AmountGrams,
		"processing_method":  b.ProcessingMethod,
		"honey_type":         b.HoneyType,
		"pdf_filename":       b.PDFFilename,
		"pdf_file_hash":      b.PDFFileHash,
		"metadata_hash":      b.MetadataHash,
		"created_at":         b.CreatedAt,
		"updated_at":         b.UpdatedAt,
		"certification":      certificationJSON(cert),
	}
}

func honeyBatchError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrBatchNotFound):
		respond.Error(w, http.StatusNotFound, "BATCH_NOT_FOUND", "honey batch not found")
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
	case errors.Is(err, service.ErrBatchAlreadyCertified):
		respond.Error(w, http.StatusConflict, "BATCH_ALREADY_CERTIFIED", err.Error())
	case errors.Is(err, service.ErrCertificationRequestPending):
		respond.Error(w, http.StatusConflict, "CERTIFICATION_REQUEST_PENDING", err.Error())
	case errors.Is(err, service.ErrBatchHasNoPDF):
		respond.Error(w, http.StatusConflict, "BATCH_HAS_NO_PDF", err.Error())
	case errors.Is(err, service.ErrBatchLocked):
		respond.Error(w, http.StatusConflict, "BATCH_LOCKED", err.Error())
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

	requestCertification := r.FormValue("request_certification") == "true"

	// The lab PDF is only required up front when certification is requested
	// immediately; otherwise it can be attached later via retry-certification.
	var pdfMimeType, pdfFilename string
	var pdfData []byte
	file, header, ferr := r.FormFile("lab_pdf")
	if ferr == nil {
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not read file")
			return
		}
		pdfMimeType = header.Header.Get("Content-Type")
		pdfFilename = header.Filename
		pdfData = data
	} else if requestCertification {
		respond.Error(w, http.StatusBadRequest, "MISSING_FILE", "field 'lab_pdf' is required when requesting certification")
		return
	}

	req := service.CreateBatchRequest{
		GatheringDate:        gatheringDate,
		AmountGrams:          amountGrams,
		ProcessingMethod:     r.FormValue("processing_method"),
		HoneyType:            r.FormValue("honey_type"),
		PDFMimeType:          pdfMimeType,
		PDFData:              pdfData,
		PDFFilename:          pdfFilename,
		RequestCertification: requestCertification,
	}

	batch, err := h.batches.CreateBatch(r.Context(), userID, req)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	// No blockchain job exists yet — admin approval of the certification
	// request is what creates one (see CertificationReviewService.Approve).
	var certRequest *model.HoneyBatchCertificationRequest
	if req.RequestCertification {
		certRequest = &model.HoneyBatchCertificationRequest{Status: model.CertificationRequestStatusPending}
	}
	respond.JSON(w, http.StatusCreated, honeyBatchJSON(batch, nil, certRequest, h.batches.VerificationURL(batch.VerificationToken)))
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

	respond.JSON(w, http.StatusOK, honeyBatchJSON(result.Batch, result.Certification, result.CertificationRequest, h.batches.VerificationURL(result.Batch.VerificationToken)))
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
		out[i] = honeyBatchJSON(it.Batch, it.Certification, it.CertificationRequest, h.batches.VerificationURL(it.Batch.VerificationToken))
	}

	respond.JSON(w, http.StatusOK, map[string]any{"items": out, "total": total})
}

// Update handles PATCH /api/v1/honey-batches/{id} — updates a batch's gathering
// date, amount, processing method, honey type, and optionally replaces its lab
// PDF (multipart form, same field name as Create — "lab_pdf" — but optional
// here). Locked (409) once the batch has any certification attempt, since its
// metadata hash may already be live.
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

	if err := r.ParseMultipartForm(maxCreateHoneyBatchBytes); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "could not parse multipart form")
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

	var pdfMimeType, pdfFilename string
	var pdfData []byte
	if file, header, ferr := r.FormFile("lab_pdf"); ferr == nil {
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "could not read file")
			return
		}
		pdfMimeType = header.Header.Get("Content-Type")
		pdfFilename = header.Filename
		pdfData = data
	}

	if _, err := h.batches.UpdateBatch(r.Context(), userID, id, service.UpdateBatchRequest{
		GatheringDate:    gatheringDate,
		AmountGrams:      amountGrams,
		ProcessingMethod: r.FormValue("processing_method"),
		HoneyType:        r.FormValue("honey_type"),
		PDFData:          pdfData,
		PDFMimeType:      pdfMimeType,
		PDFFilename:      pdfFilename,
		RemovePDF:        r.FormValue("remove_pdf") == "true",
	}); err != nil {
		honeyBatchError(w, err)
		return
	}

	result, err := h.batches.GetBatch(r.Context(), userID, id)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, honeyBatchJSON(result.Batch, result.Certification, result.CertificationRequest, h.batches.VerificationURL(result.Batch.VerificationToken)))
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

// RetryCertification handles POST /api/v1/honey-batches/{id}/retry-certification
// — submits a batch for admin certification review.
func (h *HoneyBatchHandler) RetryCertification(w http.ResponseWriter, r *http.Request) {
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

	if err := h.batches.RetryCertification(r.Context(), userID, id); err != nil {
		honeyBatchError(w, err)
		return
	}

	result, err := h.batches.GetBatch(r.Context(), userID, id)
	if err != nil {
		honeyBatchError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, honeyBatchJSON(result.Batch, result.Certification, result.CertificationRequest, h.batches.VerificationURL(result.Batch.VerificationToken)))
}
