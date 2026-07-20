package handler

import (
	"net/http"

	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
	qrcode "github.com/skip2/go-qrcode"
)

// qrCodeImageSize is the pixel width/height of generated QR code PNGs.
const qrCodeImageSize = 512

// HoneyBatchVerifyHandler handles public, token-scoped honey batch verification requests.
type HoneyBatchVerifyHandler struct {
	batches *service.HoneyBatchService
}

// NewHoneyBatchVerifyHandler creates a HoneyBatchVerifyHandler backed by svc.
func NewHoneyBatchVerifyHandler(batches *service.HoneyBatchService) *HoneyBatchVerifyHandler {
	return &HoneyBatchVerifyHandler{batches: batches}
}

// Verify handles GET /api/v1/verify/{token} — public, returns the batch and its full certification lifecycle status.
func (h *HoneyBatchVerifyHandler) Verify(w http.ResponseWriter, r *http.Request) {
	result, err := h.batches.GetBatchWithVerification(r.Context(), r.PathValue("token"))
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, honeyBatchJSON(result.Batch, result.Certification))
}

// qrCodePNG generates the PNG bytes for a batch's QR code, resolved via its public verification token.
func (h *HoneyBatchVerifyHandler) qrCodePNG(r *http.Request) ([]byte, error) {
	data, err := h.batches.GenerateQRCodeDataByToken(r.Context(), r.PathValue("token"))
	if err != nil {
		return nil, err
	}
	return qrcode.Encode(data, qrcode.Medium, qrCodeImageSize)
}

// QRCode handles GET /api/v1/verify/{token}/qr-code — public, serves a PNG QR code encoding the batch's verification URL for inline display. Requires a confirmed certification; cached indefinitely once generated.
func (h *HoneyBatchVerifyHandler) QRCode(w http.ResponseWriter, r *http.Request) {
	png, err := h.qrCodePNG(r)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(png)
}

// QRCodeDownload handles GET /api/v1/verify/{token}/qr-code/download — public, same PNG as QRCode but forces a browser download instead of inline display (e.g. for printing).
func (h *HoneyBatchVerifyHandler) QRCodeDownload(w http.ResponseWriter, r *http.Request) {
	png, err := h.qrCodePNG(r)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `attachment; filename="honey-batch-qr-code.png"`)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(png)
}

// PDF handles GET /api/v1/verify/{token}/pdf — public, serves the lab PDF for a confirmed batch.
func (h *HoneyBatchVerifyHandler) PDF(w http.ResponseWriter, r *http.Request) {
	path, err := h.batches.GetBatchPDFByToken(r.Context(), r.PathValue("token"))
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, path)
}
