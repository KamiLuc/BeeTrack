package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/beetrack/backend/internal/middleware"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// AdminCertificationHandler handles HTTP requests for the admin honey batch
// certification review queue.
type AdminCertificationHandler struct {
	review  *service.CertificationReviewService
	batches *service.HoneyBatchService
}

func NewAdminCertificationHandler(review *service.CertificationReviewService, batches *service.HoneyBatchService) *AdminCertificationHandler {
	return &AdminCertificationHandler{review: review, batches: batches}
}

func certificationRequestJSON(req *model.HoneyBatchCertificationRequest) map[string]any {
	return map[string]any{
		"id":                req.ID,
		"batch_id":          req.BatchID,
		"requested_by":      req.RequestedBy,
		"status":            req.Status,
		"rejection_reason":  req.RejectionReason,
		"blockchain_job_id": req.BlockchainJobID,
		"created_at":        req.CreatedAt,
	}
}

func adminCertificationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCertificationRequestNotFound):
		respond.Error(w, http.StatusNotFound, "CERTIFICATION_REQUEST_NOT_FOUND", "certification request not found")
	case errors.Is(err, service.ErrCertificationRequestNotPending):
		respond.Error(w, http.StatusConflict, "CERTIFICATION_REQUEST_NOT_PENDING", err.Error())
	case errors.Is(err, service.ErrRejectionReasonRequired):
		respond.Error(w, http.StatusBadRequest, "REJECTION_REASON_REQUIRED", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseCertificationRequestID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// ListPending handles GET /api/v1/admin/certification-requests.
func (h *AdminCertificationHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)

	items, total, err := h.review.ListPending(r.Context(), limit, offset)
	if err != nil {
		adminCertificationError(w, err)
		return
	}

	out := make([]map[string]any, len(items))
	for i, req := range items {
		out[i] = certificationRequestJSON(req)
	}
	respond.JSON(w, http.StatusOK, map[string]any{"items": out, "total": total})
}

// Get handles GET /api/v1/admin/certification-requests/{id}.
func (h *AdminCertificationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseCertificationRequestID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certification request id")
		return
	}

	req, err := h.review.Get(r.Context(), id)
	if err != nil {
		adminCertificationError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, certificationRequestJSON(req))
}

// Approve handles POST /api/v1/admin/certification-requests/{id}/approve —
// enqueues the blockchain_jobs row the existing worker picks up.
func (h *AdminCertificationHandler) Approve(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseCertificationRequestID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certification request id")
		return
	}

	if err := h.review.Approve(r.Context(), adminID, id); err != nil {
		adminCertificationError(w, err)
		return
	}

	req, err := h.review.Get(r.Context(), id)
	if err != nil {
		adminCertificationError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, certificationRequestJSON(req))
}

// Reject handles POST /api/v1/admin/certification-requests/{id}/reject.
func (h *AdminCertificationHandler) Reject(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		respond.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "authorization token required")
		return
	}

	id, err := parseCertificationRequestID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certification request id")
		return
	}

	var req rejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := h.review.Reject(r.Context(), adminID, id, req.Reason); err != nil {
		adminCertificationError(w, err)
		return
	}

	certReq, err := h.review.Get(r.Context(), id)
	if err != nil {
		adminCertificationError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, certificationRequestJSON(certReq))
}

// PDF handles GET /api/v1/admin/honey-batches/{id}/pdf — serves a batch's lab
// PDF regardless of ownership, for admin review.
func (h *AdminCertificationHandler) PDF(w http.ResponseWriter, r *http.Request) {
	id, err := parseHoneyBatchID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid honey batch id")
		return
	}

	path, err := h.batches.GetBatchPDFForAdmin(r.Context(), id)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, path)
}
