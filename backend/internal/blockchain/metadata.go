package blockchain

import (
	"crypto/sha256"
	"strconv"
	"strings"

	"github.com/beetrack/backend/internal/model"
)

// fieldSeparator is the ASCII unit separator (0x1F), which cannot legally
// appear in any joined field, so it can't shift a field boundary.
const fieldSeparator = "\x1f"

// CanonicalMetadataHash hashes a fixed, ordered set of a HoneyBatch's fields
// so the same batch always produces the same hash, in any language — this
// exact spec is the cross-language contract, changing it breaks every
// already-certified batch's hash:
//  1. batch_id          — decimal string, no leading zeros
//  2. apiary_id         — decimal string
//  3. gathering_date    — UTC, RFC 3339 date-only ("2006-01-02")
//  4. amount_grams      — decimal string (integer, never a float)
//  5. processing_method — exact enum string ("raw"/"filtered"/"pasteurized")
//  6. honey_type        — UTF-8, NFC-normalized as stored (normalized once
//     at write time in the service layer; this function does not re-normalize)
//  7. pdf_file_hash     — lowercase hex string of the PDF's SHA256
func CanonicalMetadataHash(batch *model.HoneyBatch) [32]byte {
	fields := []string{
		strconv.FormatInt(batch.ID, 10),
		strconv.FormatInt(batch.ApiaryID, 10),
		batch.GatheringDate.UTC().Format("2006-01-02"),
		strconv.FormatInt(batch.AmountGrams, 10),
		batch.ProcessingMethod,
		batch.HoneyType,
		batch.PDFFileHash,
	}
	return sha256.Sum256([]byte(strings.Join(fields, fieldSeparator)))
}
