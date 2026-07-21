import type { CertificationJobStatus } from "./api/certifications";
import type { TranslationKey } from "./i18n/translations";

// Collapses the blockchain job's fine-grained lifecycle into the coarse
// groups the admin panel displays — mirrors the Flutter app's honey batch
// card, which shows the same "in progress" collapse instead of per-step
// jargon (queued/submitting/submitted/pending_confirmation).
export type ChainStatusGroup = "none" | "in_progress" | "confirmed" | "failed";

export function chainStatusGroup(status: CertificationJobStatus | null): ChainStatusGroup {
  switch (status) {
    case null:
      return "none";
    case "confirmed":
      return "confirmed";
    case "failed":
    case "reverted":
      return "failed";
    default:
      return "in_progress";
  }
}

// Reuses the existing badge-status-* palette (pending=amber, approved=green,
// rejected=red, removed=grey) so no new CSS is needed.
const BADGE_CLASS: Record<ChainStatusGroup, string> = {
  none: "badge-status-removed",
  in_progress: "badge-status-pending",
  confirmed: "badge-status-approved",
  failed: "badge-status-rejected",
};

export function chainStatusBadgeClass(group: ChainStatusGroup): string {
  return BADGE_CLASS[group];
}

const LABEL_KEY: Record<ChainStatusGroup, TranslationKey> = {
  none: "certificationsQueue.chainStatusNone",
  in_progress: "certificationsQueue.chainStatusInProgress",
  confirmed: "certificationsQueue.chainStatusConfirmed",
  failed: "certificationsQueue.chainStatusFailed",
};

export function chainStatusLabelKey(group: ChainStatusGroup): TranslationKey {
  return LABEL_KEY[group];
}
