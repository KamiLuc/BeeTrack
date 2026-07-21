import { useEffect, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { ApiError, resourceUrl } from "../api/client";
import {
  approveListing,
  getListing,
  rejectListing,
  removeListing,
  restoreListing,
  type AdminListing,
} from "../api/listings";
import { ReasonPicker } from "../components/ReasonPicker";
import { useI18n } from "../i18n/I18nContext";
import { listingCategoryLabel, listingRejectReasons, type TranslationKey } from "../i18n/translations";
import { isValidRejectionReason, MAX_REJECTION_REASON_LENGTH } from "../validation";

const STATUS_LABEL_KEY: Record<AdminListing["status"], TranslationKey> = {
  pending: "listingsQueue.statusPending",
  approved: "listingsQueue.statusApproved",
  rejected: "listingsQueue.statusRejected",
  removed: "listingsQueue.statusRemoved",
};

export function ListingDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { lang, t } = useI18n();
  const [listing, setListing] = useState<AdminListing | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [lightboxUrl, setLightboxUrl] = useState<string | null>(null);

  useEffect(() => {
    if (!lightboxUrl) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") setLightboxUrl(null);
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [lightboxUrl]);

  const backToQueue = `/listings?${params.toString()}`;

  useEffect(() => {
    getListing(Number(id))
      .then(setListing)
      .catch((err) => setError(err instanceof ApiError ? err.message : t("listingDetail.loadError")));
  }, [id, t]);

  async function handleApprove() {
    if (!listing) return;
    setSubmitting(true);
    try {
      await approveListing(listing.id);
      navigate(backToQueue);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("listingDetail.approveError"));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleReject() {
    if (!listing || !isValidRejectionReason(reason)) return;
    setSubmitting(true);
    try {
      await rejectListing(listing.id, reason.trim());
      navigate(backToQueue);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("listingDetail.rejectError"));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleRemove() {
    if (!listing || !isValidRejectionReason(reason)) return;
    setSubmitting(true);
    try {
      await removeListing(listing.id, reason.trim());
      navigate(backToQueue);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("listingDetail.removeError"));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleRestore() {
    if (!listing) return;
    setSubmitting(true);
    try {
      await restoreListing(listing.id);
      navigate(backToQueue);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("listingDetail.restoreError"));
    } finally {
      setSubmitting(false);
    }
  }

  if (error) return <div className="error">{error}</div>;
  if (!listing) return <p>{t("common.loading")}</p>;

  const isPending = listing.status === "pending";
  const isApproved = listing.status === "approved";
  const isRemoved = listing.status === "removed";

  return (
    <div>
      <a
        className="back-link"
        href={backToQueue}
        onClick={(e) => {
          e.preventDefault();
          navigate(backToQueue);
        }}
      >
        {t("listingDetail.backToQueue")}
      </a>

      <div className="detail-header">
        <h1>{listing.title}</h1>
        <span className={`badge badge-status-${listing.status}`}>{t(STATUS_LABEL_KEY[listing.status])}</span>
        {listing.status === "pending" && (
          <span className={listing.is_edit ? "badge badge-edit" : "badge badge-new"}>
            {listing.is_edit ? t("listingsQueue.badgeEdited") : t("listingsQueue.badgeNew")}
          </span>
        )}
      </div>

      <div className="detail-layout">
        <div className="card">
          <p className="description-text">{listing.description}</p>
          <div className="photo-gallery">
            {listing.images.map((img) => (
              <img
                key={img.id}
                src={resourceUrl(img.url)}
                alt={listing.title}
                onClick={() => setLightboxUrl(resourceUrl(img.url))}
              />
            ))}
          </div>
        </div>

        <div className="card">
          <h2 className="card-title">{t("listingDetail.detailsTitle")}</h2>
          <dl className="detail-list">
            <div className="detail-row">
              <dt>{t("listingDetail.colCategory")}</dt>
              <dd>{listingCategoryLabel(lang, listing.category)}</dd>
            </div>
            <div className="detail-row">
              <dt>{t("listingDetail.colPrice")}</dt>
              <dd>{listing.price}</dd>
            </div>
            <div className="detail-row">
              <dt>{t("listingDetail.colQuantity")}</dt>
              <dd>{listing.quantity}</dd>
            </div>
            <div className="detail-row">
              <dt>{t("listingDetail.colAddress")}</dt>
              <dd>{listing.address}</dd>
            </div>
            <div className="detail-row">
              <dt>{t("listingDetail.colOwner")}</dt>
              <dd>{listing.owner_email}</dd>
            </div>
            <div className="detail-row">
              <dt>{t("listingDetail.colContact")}</dt>
              <dd>
                {listing.contact_phone}
                <br />
                {listing.contact_email}
              </dd>
            </div>
            {listing.rejection_reason && (
              <div className="detail-row rejection">
                <dt>{t("listingDetail.rejectionReason")}</dt>
                <dd>{listing.rejection_reason}</dd>
              </div>
            )}
          </dl>
        </div>
      </div>

      {(isPending || isApproved) && (
        <div className="card" style={{ marginTop: "1.25rem" }}>
          <h2 className="card-title">
            {isPending ? t("listingDetail.rejectionReason") : t("listingDetail.removalReason")}
          </h2>
          <div className="field">
            <ReasonPicker options={listingRejectReasons[lang]} onSelect={setReason} />
            <textarea
              id="reason"
              rows={3}
              maxLength={MAX_REJECTION_REASON_LENGTH}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </div>
        </div>
      )}

      <div className="actions">
        {isPending && (
          <>
            <button className="btn-approve" disabled={submitting} onClick={handleApprove}>
              {t("listingDetail.approve")}
            </button>
            <button
              className="btn-reject"
              disabled={submitting || !isValidRejectionReason(reason)}
              onClick={handleReject}
            >
              {t("listingDetail.reject")}
            </button>
          </>
        )}
        {isApproved && (
          <button
            className="btn-remove"
            disabled={submitting || !isValidRejectionReason(reason)}
            onClick={handleRemove}
          >
            {t("listingDetail.remove")}
          </button>
        )}
        {isRemoved && (
          <button className="btn-restore" disabled={submitting} onClick={handleRestore}>
            {t("listingDetail.restore")}
          </button>
        )}
      </div>

      {lightboxUrl && (
        <div className="lightbox-overlay" onClick={() => setLightboxUrl(null)}>
          <img className="lightbox-image" src={lightboxUrl} alt={listing.title} />
        </div>
      )}
    </div>
  );
}
