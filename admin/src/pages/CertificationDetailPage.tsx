import { useEffect, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { ApiError, getStoredToken, resourceUrl } from "../api/client";
import {
  approveCertificationRequest,
  deleteCertificationRequest,
  estimateCertificationGas,
  getCertificationRequest,
  rejectCertificationRequest,
  type CertificationRequest,
  type GasEstimate,
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
  const [gasEstimate, setGasEstimate] = useState<GasEstimate | null>(null);
  const [gasEstimateError, setGasEstimateError] = useState<string | null>(null);
  const [estimatingGas, setEstimatingGas] = useState(false);

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

  async function handleCheckGasEstimate() {
    if (!req) return;
    setEstimatingGas(true);
    setGasEstimateError(null);
    try {
      setGasEstimate(await estimateCertificationGas(req.id));
    } catch (err) {
      setGasEstimate(null);
      setGasEstimateError(err instanceof ApiError ? err.message : t("certificationDetail.gasEstimateError"));
    } finally {
      setEstimatingGas(false);
    }
  }

  async function handleDelete() {
    if (!req || !window.confirm(t("certificationDetail.deleteConfirm"))) return;
    setSubmitting(true);
    try {
      await deleteCertificationRequest(req.id);
      navigate(backToQueue);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("certificationDetail.deleteError"));
    } finally {
      setSubmitting(false);
    }
  }

  if (error) return <div className="error">{error}</div>;
  if (!req) return <p>{t("common.loading")}</p>;

  const isPending = req.status === "pending";
  const group = chainStatusGroup(req.job_status);
  const hasChainInfo = req.job_status !== null;
  const insufficientBalance =
    gasEstimate !== null && BigInt(gasEstimate.account_balance_wei) < BigInt(gasEstimate.estimated_cost_wei);
  const balanceAfterMatic =
    gasEstimate !== null
      ? Number(BigInt(gasEstimate.account_balance_wei) - BigInt(gasEstimate.estimated_cost_wei)) / 1e18
      : null;
  const balanceAfterPln =
    balanceAfterMatic !== null && gasEstimate?.matic_pln_rate != null ? balanceAfterMatic * gasEstimate.matic_pln_rate : null;

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

          {isPending && (
            <>
              <h2 className="card-title">{t("certificationDetail.gasEstimateTitle")}</h2>
              <button className="btn-approve" disabled={estimatingGas} onClick={handleCheckGasEstimate}>
                {estimatingGas && <span className="spinner" aria-hidden="true" />}
                {estimatingGas
                  ? t("certificationDetail.checkingGasEstimate")
                  : gasEstimate
                    ? t("certificationDetail.recheckGasEstimate")
                    : t("certificationDetail.checkGasEstimate")}
              </button>
              {gasEstimateError && (
                <p className="error" style={{ marginTop: "0.75rem" }}>
                  {gasEstimateError}
                </p>
              )}
              {gasEstimate && (
                <dl className="detail-list" style={{ marginTop: "0.75rem" }}>
                  <div className="detail-row">
                    <dt>{t("certificationDetail.gasUnits")}</dt>
                    <dd>{gasEstimate.gas_units.toLocaleString(lang)}</dd>
                  </div>
                  <div className="detail-row">
                    <dt>{t("certificationDetail.gasPrice")}</dt>
                    <dd>{(Number(gasEstimate.gas_price_wei) / 1e9).toLocaleString(lang, { maximumFractionDigits: 4 })} Gwei</dd>
                  </div>
                  <div className="detail-row">
                    <dt>{t("certificationDetail.estimatedCost")}</dt>
                    <dd>
                      {gasEstimate.estimated_cost_matic.toLocaleString(lang, { maximumFractionDigits: 6 })} POL
                      {gasEstimate.estimated_cost_pln !== null && (
                        <> ({gasEstimate.estimated_cost_pln.toLocaleString(lang, { maximumFractionDigits: 2 })} PLN)</>
                      )}
                    </dd>
                  </div>
                  <div className="detail-row">
                    <dt>{t("certificationDetail.polPlnRate")}</dt>
                    <dd>
                      {gasEstimate.matic_pln_rate !== null
                        ? `1 POL = ${gasEstimate.matic_pln_rate.toLocaleString(lang, { maximumFractionDigits: 2 })} PLN`
                        : t("certificationDetail.pricePlnUnavailable")}
                    </dd>
                  </div>
                  <div className="detail-row">
                    <dt>{t("certificationDetail.accountBalance")}</dt>
                    <dd>
                      {gasEstimate.account_balance_matic.toLocaleString(lang, { maximumFractionDigits: 4 })} POL
                      {gasEstimate.account_balance_pln !== null && (
                        <> ({gasEstimate.account_balance_pln.toLocaleString(lang, { maximumFractionDigits: 2 })} PLN)</>
                      )}
                    </dd>
                  </div>
                  <div className="detail-row">
                    <dt>{t("certificationDetail.accountBalanceAfter")}</dt>
                    <dd>
                      {balanceAfterMatic!.toLocaleString(lang, { maximumFractionDigits: 4 })} POL
                      {balanceAfterPln !== null && (
                        <> ({balanceAfterPln.toLocaleString(lang, { maximumFractionDigits: 2 })} PLN)</>
                      )}
                    </dd>
                  </div>
                </dl>
              )}
              {insufficientBalance && (
                <p className="error" style={{ marginTop: "0.75rem" }}>
                  {t("certificationDetail.insufficientBalanceWarning")}
                </p>
              )}
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
            disabled={submitting || reason.trim().length > 0 || insufficientBalance}
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

      {group === "failed" && (
        <div className="actions">
          <button className="btn-remove" disabled={submitting} onClick={handleDelete}>
            {t("certificationDetail.delete")}
          </button>
        </div>
      )}
    </div>
  );
}
