import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { ApiError } from "../api/client";
import {
  listCertificationRequests,
  type CertificationRequest,
  type CertificationStatusFilter,
  type SortDir,
} from "../api/certifications";
import { chainStatusBadgeClass, chainStatusGroup, chainStatusLabelKey } from "../certificationStatus";
import { formatKg } from "../format";
import { useI18n } from "../i18n/I18nContext";
import type { TranslationKey } from "../i18n/translations";

const PAGE_SIZE = 20;
const SEARCH_DEBOUNCE_MS = 600;
const FILTERS_STORAGE_KEY = "beetrack_admin_certifications_filters";

const STATUS_LABEL_KEY: Record<CertificationRequest["status"], TranslationKey> = {
  pending: "certificationsQueue.statusPending",
  approved: "certificationsQueue.statusApproved",
  rejected: "certificationsQueue.statusRejected",
};

function parseOffset(v: string | null): number {
  const n = Number(v);
  return Number.isFinite(n) && n >= 0 ? n : 0;
}

// persist saves next's query string so filters survive navigating away and back
// (e.g. to the Listings tab and back), then returns next unchanged.
function persist(next: URLSearchParams): URLSearchParams {
  sessionStorage.setItem(FILTERS_STORAGE_KEY, next.toString());
  return next;
}

export function CertificationQueuePage() {
  const { lang, t } = useI18n();
  const navigate = useNavigate();
  const [params, setParams] = useSearchParams();

  // Defaults to "all" statuses when nothing is set (fresh visit with no remembered filters).
  const status = (params.get("status") ?? "") as CertificationStatusFilter;
  const sort = (params.get("sort") === "desc" ? "desc" : "asc") as SortDir;
  const offset = parseOffset(params.get("offset"));
  const query = params.get("q") ?? "";

  const [searchInput, setSearchInput] = useState(query);
  const [items, setItems] = useState<CertificationRequest[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  // On a fresh navigation to this page with no query string, restore the last-used
  // filters from sessionStorage instead of resetting to defaults.
  useEffect(() => {
    if (params.toString() === "") {
      const saved = sessionStorage.getItem(FILTERS_STORAGE_KEY);
      if (saved) setParams(new URLSearchParams(saved), { replace: true });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Keep the visible search box in sync if the URL changes from elsewhere (e.g. back/forward).
  useEffect(() => {
    setSearchInput(query);
  }, [query]);

  // Debounce typing into a "q" URL param update, resetting to the first page.
  useEffect(() => {
    if (searchInput === query) return;
    const id = setTimeout(() => {
      setParams((prev) => {
        const next = new URLSearchParams(prev);
        if (searchInput) next.set("q", searchInput);
        else next.delete("q");
        next.delete("offset");
        return persist(next);
      });
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(id);
  }, [searchInput, query, setParams]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    listCertificationRequests(status, query, sort, PAGE_SIZE, offset)
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
  }, [status, query, sort, offset, t]);

  function handleStatusChange(value: CertificationStatusFilter) {
    setParams((prev) => {
      const next = new URLSearchParams(prev);
      next.set("status", value);
      next.delete("offset");
      return persist(next);
    });
  }

  function handleSortChange(value: SortDir) {
    setParams((prev) => {
      const next = new URLSearchParams(prev);
      next.set("sort", value);
      next.delete("offset");
      return persist(next);
    });
  }

  function handleOffsetChange(value: number) {
    setParams((prev) => {
      const next = new URLSearchParams(prev);
      if (value > 0) next.set("offset", String(value));
      else next.delete("offset");
      return persist(next);
    });
  }

  return (
    <div>
      <h1>{t("certificationsQueue.title")}</h1>
      <div className="filters">
        <label>
          {t("certificationsQueue.search")}
          <input
            type="text"
            value={searchInput}
            placeholder={t("certificationsQueue.searchPlaceholder")}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </label>
        <label>
          {t("certificationsQueue.filterStatus")}
          <select
            value={status}
            onChange={(e) => handleStatusChange(e.target.value as CertificationStatusFilter)}
          >
            <option value="">{t("certificationsQueue.statusAll")}</option>
            <option value="pending">{t("certificationsQueue.statusPending")}</option>
            <option value="approved">{t("certificationsQueue.statusApproved")}</option>
            <option value="rejected">{t("certificationsQueue.statusRejected")}</option>
          </select>
        </label>
        <label>
          {t("certificationsQueue.sortBy")}
          <select value={sort} onChange={(e) => handleSortChange(e.target.value as SortDir)}>
            <option value="asc">{t("certificationsQueue.sortOldest")}</option>
            <option value="desc">{t("certificationsQueue.sortNewest")}</option>
          </select>
        </label>
      </div>
      {error && <div className="error">{error}</div>}
      {loading ? (
        <p>{t("common.loading")}</p>
      ) : items.length === 0 ? (
        <p>{t("certificationsQueue.noResults")}</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>{t("certificationsQueue.colHoneyType")}</th>
              <th>{t("certificationsQueue.colAmount")}</th>
              <th>{t("certificationsQueue.colOwner")}</th>
              <th>{t("certificationsQueue.colStatus")}</th>
              <th>{t("certificationsQueue.colChainStatus")}</th>
              <th>{t("certificationsQueue.colSubmitted")}</th>
            </tr>
          </thead>
          <tbody>
            {items.map((req) => {
              const group = chainStatusGroup(req.job_status);
              return (
                <tr
                  key={req.id}
                  className="clickable-row"
                  onClick={() => navigate(`/certifications/${req.id}?${params.toString()}`)}
                >
                  <td>{req.honey_type}</td>
                  <td>{formatKg(req.amount_grams)}</td>
                  <td>{req.requester_email}</td>
                  <td>
                    <span className={`badge badge-status-${req.status}`}>{t(STATUS_LABEL_KEY[req.status])}</span>
                  </td>
                  <td>
                    <span className={`badge ${chainStatusBadgeClass(group)}`}>{t(chainStatusLabelKey(group))}</span>
                  </td>
                  <td>{new Date(req.created_at).toLocaleString(lang)}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
      {total > 0 && (
        <div className="pagination">
          <button disabled={offset === 0} onClick={() => handleOffsetChange(Math.max(0, offset - PAGE_SIZE))}>
            {t("common.previous")}
          </button>
          <span>
            {t("common.paginationRange", {
              from: offset + 1,
              to: Math.min(offset + PAGE_SIZE, total),
              total,
            })}
          </span>
          <button disabled={offset + PAGE_SIZE >= total} onClick={() => handleOffsetChange(offset + PAGE_SIZE)}>
            {t("common.next")}
          </button>
        </div>
      )}
    </div>
  );
}
