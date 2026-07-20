import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ApiError, getStoredToken, resourceUrl } from "../api/client";
import {
  approveCertificationRequest,
  getCertificationRequest,
  rejectCertificationRequest,
  type CertificationRequest,
} from "../api/certifications";
import { ReasonPicker } from "../components/ReasonPicker";
import { useI18n } from "../i18n/I18nContext";
import { certificationRejectReasons } from "../i18n/translations";
import { isValidRejectionReason, MAX_REJECTION_REASON_LENGTH } from "../validation";

// <embed src=...> can't send an auth header, so fetch the PDF and use a blob: URL instead.
function usePdfBlobUrl(pdfPath: string | undefined) {
  const [url, setUrl] = useState<string | null>(null);

  useEffect(() => {
    if (!pdfPath) return;
    let objectUrl: string | null = null;
    let cancelled = false;

    fetch(resourceUrl(pdfPath), {
      headers: { Authorization: `Bearer ${getStoredToken()}` },
    })
      .then((res) => res.blob())
      .then((blob) => {
        if (cancelled) return;
        objectUrl = URL.createObjectURL(blob);
        setUrl(objectUrl);
      });

    return () => {
      cancelled = true;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [pdfPath]);

  return url;
}

export function CertificationDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { lang, t } = useI18n();
  const [req, setReq] = useState<CertificationRequest | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    getCertificationRequest(Number(id))
      .then(setReq)
      .catch((err) => setError(err instanceof ApiError ? err.message : t("certificationDetail.loadError")));
  }, [id, t]);

  const pdfUrl = usePdfBlobUrl(req?.pdf_url);

  async function handleApprove() {
    if (!req) return;
    setSubmitting(true);
    try {
      await approveCertificationRequest(req.id);
      navigate("/certifications");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("certificationDetail.approveError"));
    } finally {
      setSubmitting(false);
    }
  }

  async function handleReject() {
    if (!req || !isValidRejectionReason(reason)) return;
    setSubmitting(true);
    try {
      await rejectCertificationRequest(req.id, reason.trim());
      navigate("/certifications");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("certificationDetail.rejectError"));
    } finally {
      setSubmitting(false);
    }
  }

  if (error) return <div className="error">{error}</div>;
  if (!req) return <p>{t("common.loading")}</p>;

  return (
    <div>
      <p>
        <Link to="/certifications">{t("certificationDetail.backToQueue")}</Link>
      </p>
      <h1>{t("certificationDetail.batchTitle", { honeyType: req.honey_type, batchId: req.batch_id })}</h1>

      <div className="card" style={{ marginTop: "1rem" }}>
        <dl>
          <dt>{t("certificationDetail.amount")}</dt>
          <dd>{req.amount_grams} g</dd>
          <dt>{t("certificationDetail.gatheringDate")}</dt>
          <dd>{new Date(req.gathering_date).toLocaleDateString(lang)}</dd>
          <dt>{t("certificationDetail.requestedBy")}</dt>
          <dd>{req.requester_email}</dd>
          <dt>{t("certificationDetail.submitted")}</dt>
          <dd>{new Date(req.created_at).toLocaleString(lang)}</dd>
        </dl>
        {pdfUrl ? (
          <embed src={pdfUrl} type="application/pdf" width="100%" height="600" />
        ) : (
          <p>{t("certificationDetail.loadingPdf")}</p>
        )}
      </div>

      <div className="card" style={{ marginTop: "1rem" }}>
        <div className="field">
          <label htmlFor="reason">{t("certificationDetail.rejectionReason")}</label>
          <ReasonPicker options={certificationRejectReasons[lang]} onSelect={setReason} />
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
          {t("certificationDetail.approve")}
        </button>
        <button
          className="btn-reject"
          disabled={submitting || !isValidRejectionReason(reason)}
          onClick={handleReject}
        >
          {t("certificationDetail.reject")}
        </button>
      </div>
    </div>
  );
}
