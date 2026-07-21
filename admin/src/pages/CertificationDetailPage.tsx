import { useEffect, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { ApiError, getStoredToken, resourceUrl } from "../api/client";
import {
  approveCertificationRequest,
  getCertificationRequest,
  rejectCertificationRequest,
  type CertificationRequest,
} from "../api/certifications";
import { chainStatusBadgeClass, chainStatusGroup, chainStatusLabelKey } from "../certificationStatus";
import { ReasonPicker } from "../components/ReasonPicker";
import { formatKg } from "../format";
import { useI18n } from "../i18n/I18nContext";
import { certificationRejectReasons, processingMethodLabel, type TranslationKey } from "../i18n/translations";
import { isValidRejectionReason, MAX_REJECTION_REASON_LENGTH } from "../validation";

const STATUS_LABEL_KEY: Record<CertificationRequest["status"], TranslationKey> = {
  pending: "certificationsQueue.statusPending",
  approved: "certificationsQueue.statusApproved",
  rejected: "certificationsQueue.statusRejected",
};

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
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { lang, t } = useI18n();
  const [req, setReq] = useState<CertificationRequest | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const backToQueue = `/certifications?${params.toString()}`;

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
      navigate(backToQueue);
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
      navigate(backToQueue);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("certificationDetail.rejectError"));
    } finally {
      setSubmitting(false);
    }
  }

  if (error) return <div className="error">{error}</div>;
  if (!req) return <p>{t("common.loading")}</p>;

  const isPending = req.status === "pending";
  const group = chainStatusGroup(req.job_status);
  const hasChainInfo = req.job_status !== null;

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
        {t("certificationDetail.backToQueue")}
      </a>

      <div className="detail-header">
        <h1>{t("certificationDetail.batchTitle", { honeyType: req.honey_type, batchId: req.batch_id })}</h1>
        <span className={`badge badge-status-${req.status}`}>{t(STATUS_LABEL_KEY[req.status])}</span>
      </div>

      <div className="detail-layout">
        <div className="card">
          {pdfUrl ? (
            <embed src={pdfUrl} type="application/pdf" width="100%" height="600" />
          ) : (
            <p>{t("certificationDetail.loadingPdf")}</p>
          )}
        </div>

        <div className="card">
          <h2 className="card-title">{t("certificationDetail.detailsTitle")}</h2>
          <dl className="detail-list">
            <div className="detail-row">
              <dt>{t("certificationDetail.amount")}</dt>
              <dd>{formatKg(req.amount_grams)}</dd>
            </div>
            <div className="detail-row">
              <dt>{t("certificationDetail.processingMethod")}</dt>
              <dd>{processingMethodLabel(lang, req.processing_method)}</dd>
            </div>
            <div className="detail-row">
              <dt>{t("certificationDetail.gatheringDate")}</dt>
              <dd>{new Date(req.gathering_date).toLocaleDateString(lang)}</dd>
            </div>
            <div className="detail-row">
              <dt>{t("certificationDetail.requestedBy")}</dt>
              <dd>{req.requester_email}</dd>
            </div>
            <div className="detail-row">
              <dt>{t("certificationDetail.submitted")}</dt>
              <dd>{new Date(req.created_at).toLocaleString(lang)}</dd>
            </div>
            {req.rejection_reason && (
              <div className="detail-row rejection">
                <dt>{t("certificationDetail.rejectionReason")}</dt>
                <dd>{req.rejection_reason}</dd>
              </div>
            )}
          </dl>

          {hasChainInfo && (
            <>
              <h2 className="card-title">{t("certificationDetail.chainSectionTitle")}</h2>
              <dl className="detail-list">
                <div className="detail-row">
                  <dt>{t("certificationDetail.chainStatus")}</dt>
                  <dd>
                    <span className={`badge ${chainStatusBadgeClass(group)}`}>{t(chainStatusLabelKey(group))}</span>
                  </dd>
                </div>
                {req.transaction_hash && (
                  <div className="detail-row">
                    <dt>{t("certificationDetail.transactionHash")}</dt>
                    <dd className="mono">{req.transaction_hash}</dd>
                  </div>
                )}
                {req.block_number !== null && (
                  <div className="detail-row">
                    <dt>{t("certificationDetail.blockNumber")}</dt>
                    <dd>{req.block_number}</dd>
                  </div>
                )}
                {req.confirmation_timestamp && (
                  <div className="detail-row">
                    <dt>{t("certificationDetail.confirmedAt")}</dt>
                    <dd>{new Date(req.confirmation_timestamp).toLocaleString(lang)}</dd>
                  </div>
                )}
                {req.job_last_error && (
                  <div className="detail-row rejection">
                    <dt>{t("certificationDetail.lastError")}</dt>
                    <dd>{req.job_last_error}</dd>
                  </div>
                )}
              </dl>
            </>
          )}
        </div>
      </div>

      {isPending && (
        <div className="card" style={{ marginTop: "1.25rem" }}>
          <h2 className="card-title">{t("certificationDetail.rejectionReason")}</h2>
          <div className="field">
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
      )}

      {isPending && (
        <div className="actions">
          <button
            className="btn-approve"
            disabled={submitting || reason.trim().length > 0}
            onClick={handleApprove}
          >
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
      )}
    </div>
  );
}
