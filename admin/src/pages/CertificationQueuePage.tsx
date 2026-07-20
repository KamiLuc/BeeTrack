import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { ApiError } from "../api/client";
import { listPendingCertificationRequests, type CertificationRequest } from "../api/certifications";
import { useI18n } from "../i18n/I18nContext";

const PAGE_SIZE = 20;

export function CertificationQueuePage() {
  const { lang, t } = useI18n();
  const [items, setItems] = useState<CertificationRequest[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    listPendingCertificationRequests(PAGE_SIZE, offset)
      .then((page) => {
        if (cancelled) return;
        setItems(page.items);
        setTotal(page.total);
        setError(null);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof ApiError ? err.message : t("certificationsQueue.loadError"));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [offset, t]);

  return (
    <div>
      <h1>{t("certificationsQueue.title")}</h1>
      {error && <div className="error">{error}</div>}
      {loading ? (
        <p>{t("common.loading")}</p>
      ) : items.length === 0 ? (
        <p>{t("certificationsQueue.noPending")}</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>{t("certificationsQueue.colHoneyType")}</th>
              <th>{t("certificationsQueue.colAmount")}</th>
              <th>{t("certificationsQueue.colGatheringDate")}</th>
              <th>{t("certificationsQueue.colOwner")}</th>
              <th>{t("certificationsQueue.colSubmitted")}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {items.map((req) => (
              <tr key={req.id}>
                <td>{req.honey_type}</td>
                <td>{req.amount_grams}</td>
                <td>{new Date(req.gathering_date).toLocaleDateString(lang)}</td>
                <td>{req.requester_email}</td>
                <td>{new Date(req.created_at).toLocaleString(lang)}</td>
                <td>
                  <Link to={`/certifications/${req.id}`}>{t("certificationsQueue.review")}</Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      <div className="pagination">
        <button disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}>
          {t("common.previous")}
        </button>
        <span>
          {t("common.paginationRange", {
            from: total === 0 ? 0 : offset + 1,
            to: Math.min(offset + PAGE_SIZE, total),
            total,
          })}
        </span>
        <button disabled={offset + PAGE_SIZE >= total} onClick={() => setOffset(offset + PAGE_SIZE)}>
          {t("common.next")}
        </button>
      </div>
    </div>
  );
}
