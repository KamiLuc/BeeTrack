import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ApiError, resourceUrl } from "../api/client";
import { approveListing, getListing, rejectListing, type AdminListing } from "../api/listings";
import { ReasonPicker } from "../components/ReasonPicker";
import { useI18n } from "../i18n/I18nContext";
import { listingRejectReasons } from "../i18n/translations";
import { isValidRejectionReason, MAX_REJECTION_REASON_LENGTH } from "../validation";

export function ListingDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { lang, t } = useI18n();
  const [listing, setListing] = useState<AdminListing | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);

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
      navigate("/listings");
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
      navigate("/listings");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("listingDetail.rejectError"));
    } finally {
      setSubmitting(false);
    }
  }

  if (error) return <div className="error">{error}</div>;
  if (!listing) return <p>{t("common.loading")}</p>;

  return (
    <div>
      <p>
        <Link to="/listings">{t("listingDetail.backToQueue")}</Link>
      </p>
      <h1>{listing.title}</h1>
      <span className={listing.is_edit ? "badge badge-edit" : "badge badge-new"}>
        {listing.is_edit ? t("listingsQueue.badgeEdited") : t("listingsQueue.badgeNew")}
      </span>

      <div className="card" style={{ marginTop: "1rem" }}>
        <p>{listing.description}</p>
        <div className="photo-gallery">
          {listing.images.map((img) => (
            <img key={img.id} src={resourceUrl(img.url)} alt={listing.title} />
          ))}
        </div>
        <dl>
          <dt>{t("listingDetail.colCategory")}</dt>
          <dd>{listing.category}</dd>
          <dt>{t("listingDetail.colPrice")}</dt>
          <dd>{listing.price}</dd>
          <dt>{t("listingDetail.colQuantity")}</dt>
          <dd>{listing.quantity}</dd>
          <dt>{t("listingDetail.colAddress")}</dt>
          <dd>{listing.address}</dd>
          <dt>{t("listingDetail.colContact")}</dt>
          <dd>
            {listing.contact_phone} &middot; {listing.contact_email}
          </dd>
        </dl>
      </div>

      <div className="card" style={{ marginTop: "1rem" }}>
        <div className="field">
          <label htmlFor="reason">{t("listingDetail.rejectionReason")}</label>
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

      <div className="actions">
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
      </div>
    </div>
  );
}
