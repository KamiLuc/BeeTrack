// Mirrors backend/internal/service/listing_moderation.go's rejection-reason bounds.
export const MIN_REJECTION_REASON_LENGTH = 3;
export const MAX_REJECTION_REASON_LENGTH = 500;

export function isValidRejectionReason(reason: string): boolean {
  const length = reason.trim().length;
  return length >= MIN_REJECTION_REASON_LENGTH && length <= MAX_REJECTION_REASON_LENGTH;
}
