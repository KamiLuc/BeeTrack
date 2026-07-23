package handler

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/beetrack/backend/internal/blockchain"
	"github.com/beetrack/backend/internal/model"
	"github.com/beetrack/backend/internal/service"
	"github.com/beetrack/backend/pkg/respond"
	qrcode "github.com/skip2/go-qrcode"
)

// qrCodeImageSize is the pixel width/height of generated QR code PNGs.
const qrCodeImageSize = 512

// filenameUnsafeChars matches runs of characters not safe to use unescaped in a Content-Disposition filename.
var filenameUnsafeChars = regexp.MustCompile(`[^a-z0-9]+`)

// sanitizeFilenamePart lowercases s and collapses runs of unsafe characters into a single hyphen.
func sanitizeFilenamePart(s string) string {
	return strings.Trim(filenameUnsafeChars.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

// qrCodeDownloadFilename builds a human-readable filename for a batch's downloaded QR code, e.g. "2024-05-01_wildflower_1.5kg.png".
func qrCodeDownloadFilename(batch *model.HoneyBatch) string {
	date := batch.GatheringDate.Format("2006-01-02")
	honeyType := sanitizeFilenamePart(batch.HoneyType)
	weightKg := strconv.FormatFloat(float64(batch.AmountGrams)/1000, 'f', -1, 64)
	return fmt.Sprintf("%s_%s_%skg.png", date, honeyType, weightKg)
}

// hexEncodeCertRecord hex-encodes an on-chain certification record's hashes
// (no "0x" prefix), matching the encoding convention used to store
// HoneyBatch.PDFFileHash/MetadataHash — see hex.EncodeToString calls in
// internal/service/honey_batch.go and decodeHash32 in
// internal/worker/blockchain_worker.go.
func hexEncodeCertRecord(record *blockchain.CertificationRecord) (pdfHashHex, metadataHashHex string) {
	return hex.EncodeToString(record.PDFHash[:]), hex.EncodeToString(record.MetadataHash[:])
}

// ChainCertReader reads a batch's on-chain certification record — satisfied
// by *blockchain.HoneyCertReader; an interface here so VerifyPage's live
// hash-comparison logic can be tested without a real RPC connection.
//
// Callers passing a possibly-nil *blockchain.HoneyCertReader (see
// cmd/api/main.go, where blockchain config is optional) must explicitly
// convert a nil pointer to a nil ChainCertReader rather than relying on
// Go's implicit conversion — a nil *HoneyCertReader wrapped directly in
// this interface is a non-nil interface value, and calling GetCertification
// on it would panic instead of hitting the certReader == nil check below.
type ChainCertReader interface {
	GetCertification(ctx context.Context, batchID int64) (*blockchain.CertificationRecord, error)
}

// HoneyBatchVerifyHandler handles public, token-scoped honey batch verification requests.
type HoneyBatchVerifyHandler struct {
	batches *service.HoneyBatchService
	// certReader is used by VerifyPage to live-check a confirmed batch's
	// hashes against the on-chain record. Optional — nil if blockchain isn't
	// configured (see cmd/api/main.go); VerifyPage degrades gracefully.
	certReader ChainCertReader
}

// NewHoneyBatchVerifyHandler creates a HoneyBatchVerifyHandler backed by svc. certReader may be nil.
func NewHoneyBatchVerifyHandler(batches *service.HoneyBatchService, certReader ChainCertReader) *HoneyBatchVerifyHandler {
	return &HoneyBatchVerifyHandler{batches: batches, certReader: certReader}
}

// Verify handles GET /api/v1/verify/{token} — public, returns the batch and its full certification lifecycle status.
func (h *HoneyBatchVerifyHandler) Verify(w http.ResponseWriter, r *http.Request) {
	result, err := h.batches.GetBatchWithVerification(r.Context(), r.PathValue("token"))
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	respond.JSON(w, http.StatusOK, publicHoneyBatchJSON(result.Batch, result.Certification, h.batches.VerificationURL(result.Batch.VerificationToken)))
}

// qrCodePNG generates the PNG bytes for a batch's QR code, resolved via its public verification token, together with the batch itself.
func (h *HoneyBatchVerifyHandler) qrCodePNG(r *http.Request) ([]byte, *model.HoneyBatch, error) {
	token := r.PathValue("token")
	result, err := h.batches.GetBatchWithVerification(r.Context(), token)
	if err != nil {
		return nil, nil, err
	}

	data, err := h.batches.GenerateQRCodeData(r.Context(), result.Batch.ID)
	if err != nil {
		return nil, nil, err
	}

	png, err := qrcode.Encode(data, qrcode.Medium, qrCodeImageSize)
	if err != nil {
		return nil, nil, err
	}
	return png, result.Batch, nil
}

// QRCode handles GET /api/v1/verify/{token}/qr-code — public, serves a PNG QR code encoding the batch's verification URL for inline display. Requires a confirmed certification; cached indefinitely once generated.
func (h *HoneyBatchVerifyHandler) QRCode(w http.ResponseWriter, r *http.Request) {
	png, _, err := h.qrCodePNG(r)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Write(png)
}

// QRCodeDownload handles GET /api/v1/verify/{token}/qr-code/download — public, same PNG as QRCode but forces a browser download instead of inline display (e.g. for printing), with a filename derived from the batch's gathering date, honey type, and weight.
func (h *HoneyBatchVerifyHandler) QRCodeDownload(w http.ResponseWriter, r *http.Request) {
	png, batch, err := h.qrCodePNG(r)
	if err != nil {
		honeyBatchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, qrCodeDownloadFilename(batch)))
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
