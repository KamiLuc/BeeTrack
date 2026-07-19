package model

import "time"

// HoneyBatchQRCode caches the QR code data string generated for a honey
// batch's public verification URL, keyed by batch id. Generated once and
// reused — the URL is deterministic from the batch's (immutable)
// verification token, so there's never more than one row per batch.
type HoneyBatchQRCode struct {
	ID         int64
	BatchID    int64
	QRCodeData string
	CreatedAt  time.Time
}
