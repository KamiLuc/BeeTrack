package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
)

// AdminCertificationHandler handles HTTP requests for the admin honey batch
// certification review queue.
type AdminCertificationHandler struct {
	review      *service.CertificationReviewService
	batches     *service.HoneyBatchService
	gasEstimate *service.CertificationGasEstimateService
}

func NewAdminCertificationHandler(review *service.CertificationReviewService, batches *service.HoneyBatchService, gasEstimate *service.CertificationGasEstimateService) *AdminCertificationHandler {
	return &AdminCertificationHandler{review: review, batches: batches, gasEstimate: gasEstimate}
}

// gasEstimateJSON shapes a GasEstimate for the wire — big.Int fields become
// decimal strings since they can exceed JS's safe integer range.
func gasEstimateJSON(est *service.GasEstimate) map[string]any {
	out := map[string]any{
		"gas_units":             est.GasUnits,
		"gas_price_wei":         est.GasPriceWei.String(),
		"estimated_cost_wei":    est.EstimatedCostWei.String(),
		"estimated_cost_matic":  est.CostMatic(),
		"account_balance_wei":   est.AccountBalanceWei.String(),
		"account_balance_matic": est.BalanceMatic(),
		"matic_pln_rate":        est.MaticPLNRate,
		"estimated_cost_pln":    nil,
		"account_balance_pln":   nil,
	}
	if est.MaticPLNRate != nil {
		out["estimated_cost_pln"] = est.CostMatic() * *est.MaticPLNRate
		out["account_balance_pln"] = est.BalanceMatic() * *est.MaticPLNRate
	}
	return out
}

func certificationRequestJSON(req *model.HoneyBatchCertificationRequestDetail) map[string]any {
	return map[string]any{
		"id":                     req.ID,
		"batch_id":               req.BatchID,
		"requested_by":           req.RequestedBy,
		"requester_email":        req.RequesterEmail,
		"status":                 req.Status,
		"rejection_reason":       req.RejectionReason,
		"blockchain_job_id":      req.BlockchainJobID,
		"created_at":             req.CreatedAt,
		"gathering_date":         req.GatheringDate,
		"amount_grams":           req.AmountGrams,
		"honey_type":             req.HoneyType,
		"processing_method":      req.ProcessingMethod,
		"pdf_url":                fmt.Sprintf("/api/v1/admin/honey-batches/%d/pdf", req.BatchID),
		"job_status":             req.JobStatus,
		"job_last_error":         req.JobLastError,
		"transaction_hash":       req.TransactionHash,
		"block_number":           req.BlockNumber,
		"confirmation_timestamp": req.ConfirmationTimestamp,
	}
}

// allowedCertificationRequestStatuses are the values the admin panel's status
// filter accepts; empty means no filter (all statuses).
var allowedCertificationRequestStatuses = map[string]bool{
	"":                                       true,
	model.CertificationRequestStatusPending:  true,
	model.CertificationRequestStatusApproved: true,
	model.CertificationRequestStatusRejected: true,
}

// parseCertificationReviewQuery reads the "status", "q", and "sort" query params
// for the admin certifications queue, mirroring parseListingReviewQuery.
func parseCertificationReviewQuery(r *http.Request) (status, keyword, sortDir string, ok bool) {
	status = r.URL.Query().Get("status")
	if !allowedCertificationRequestStatuses[status] {
		return "", "", "", false
	}
	keyword = strings.TrimSpace(r.URL.Query().Get("q"))
	sortDir = r.URL.Query().Get("sort")
	if sortDir != "desc" {
		sortDir = "asc"
	}
	return status, keyword, sortDir, true
}

func adminCertificationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCertificationRequestNotFound):
		respond.Error(w, http.StatusNotFound, "CERTIFICATION_REQUEST_NOT_FOUND", "certification request not found")
	case errors.Is(err, service.ErrCertificationRequestNotPending):
		respond.Error(w, http.StatusConflict, "CERTIFICATION_REQUEST_NOT_PENDING", err.Error())
	case errors.Is(err, service.ErrCertificationRequestNotFailed):
		respond.Error(w, http.StatusConflict, "CERTIFICATION_REQUEST_NOT_FAILED", err.Error())
	case errors.Is(err, service.ErrRejectionReasonRequired):
		respond.Error(w, http.StatusBadRequest, "REJECTION_REASON_REQUIRED", err.Error())
	case errors.Is(err, service.ErrRejectionReasonTooShort):
		respond.Error(w, http.StatusBadRequest, "REJECTION_REASON_TOO_SHORT", err.Error())
	case errors.Is(err, service.ErrRejectionReasonTooLong):
		respond.Error(w, http.StatusBadRequest, "REJECTION_REASON_TOO_LONG", err.Error())
	case errors.Is(err, service.ErrBatchNotFound):
		respond.Error(w, http.StatusNotFound, "BATCH_NOT_FOUND", err.Error())
	case errors.Is(err, service.ErrBlockchainNotConfigured):
		respond.Error(w, http.StatusServiceUnavailable, "BLOCKCHAIN_NOT_CONFIGURED", err.Error())
	default:
		respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}

func parseCertificationRequestID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// List handles GET /api/v1/admin/certification-requests — returns certification
// requests for the review queue, optionally filtered by ?status=
// (pending/approved/rejected, default all), by ?q= (keyword matched against
// honey type and requester email), and ordered by ?sort= (asc/desc, default asc).
func (h *AdminCertificationHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	status, keyword, sortDir, ok := parseCertificationReviewQuery(r)
	if !ok {
		respond.Error(w, http.StatusBadRequest, "INVALID_STATUS", "invalid status filter")
		return
	}

	items, total, err := h.review.List(r.Context(), status, keyword, sortDir, limit, offset)
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
	adminID, ok := requireAuth(w, r)
	if !ok {
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
	adminID, ok := requireAuth(w, r)
	if !ok {
		return
	}

	id, err := parseCertificationRequestID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certification request id")
		return
	}

	var req rejectRequest
	if !decodeJSON(w, r, &req) {
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

// Delete handles DELETE /api/v1/admin/certification-requests/{id} — removes a
// certification request whose blockchain attempt failed or reverted, so the
// admin panel can clear retry noise. Requests still pending review, in
// flight, or confirmed can't be deleted this way.
func (h *AdminCertificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseCertificationRequestID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certification request id")
		return
	}

	if err := h.review.Delete(r.Context(), id); err != nil {
		adminCertificationError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"deleted": true})
}

// EstimateGas handles GET /api/v1/admin/certification-requests/{id}/estimate-gas
// — a dry-run gas cost preview (in wei, MATIC, and PLN) for the certify()
// transaction the worker would eventually submit once the request is
// approved. Nothing is signed or broadcast.
func (h *AdminCertificationHandler) EstimateGas(w http.ResponseWriter, r *http.Request) {
	id, err := parseCertificationRequestID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid certification request id")
		return
	}

	est, err := h.gasEstimate.Estimate(r.Context(), id)
	if err != nil {
		adminCertificationError(w, err)
		return
	}
	respond.JSON(w, http.StatusOK, gasEstimateJSON(est))
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
