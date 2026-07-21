import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { ApiError } from "../api/client";
import {
  listListings,
  type AdminListing,
  type ListingStatusFilter,
  type SortDir,
} from "../api/listings";
import { useI18n } from "../i18n/I18nContext";
import { listingCategoryLabel, type TranslationKey } from "../i18n/translations";

const PAGE_SIZE = 20;
const SEARCH_DEBOUNCE_MS = 600;
const FILTERS_STORAGE_KEY = "beetrack_admin_listings_filters";

const STATUS_LABEL_KEY: Record<AdminListing["status"] | "removed", TranslationKey> = {
  pending: "listingsQueue.statusPending",
  approved: "listingsQueue.statusApproved",
  rejected: "listingsQueue.statusRejected",
  removed: "listingsQueue.statusRemoved",
};

function parseOffset(v: string | null): number {
  const n = Number(v);
  return Number.isFinite(n) && n >= 0 ? n : 0;
}

// persist saves next's query string so filters survive navigating away and back
// (e.g. to the Certifications tab and back), then returns next unchanged.
function persist(next: URLSearchParams): URLSearchParams {
  sessionStorage.setItem(FILTERS_STORAGE_KEY, next.toString());
  return next;
}

export function ListingsQueuePage() {
  const { lang, t } = useI18n();
  const [params, setParams] = useSearchParams();

  // Defaults to "all" statuses when nothing is set (fresh visit with no remembered filters).
  const status = (params.get("status") ?? "") as ListingStatusFilter;
  const sort = (params.get("sort") === "desc" ? "desc" : "asc") as SortDir;
  const offset = parseOffset(params.get("offset"));
  const query = params.get("q") ?? "";

  const [searchInput, setSearchInput] = useState(query);
  const [items, setItems] = useState<AdminListing[]>([]);
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
    listListings(status, query, sort, PAGE_SIZE, offset)
      .then((page) => {
        if (cancelled) return;
        setItems(page.items);
        setTotal(page.total);
        setError(null);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof ApiError ? err.message : t("listingsQueue.loadError"));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [status, query, sort, offset, t]);

  function handleStatusChange(value: ListingStatusFilter) {
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
      <h1>{t("listingsQueue.title")}</h1>
      <div className="filters">
        <label>
          {t("listingsQueue.search")}
          <input
            type="text"
            value={searchInput}
            placeholder={t("listingsQueue.searchPlaceholder")}
            onChange={(e) => setSearchInput(e.target.value)}
          />
        </label>
        <label>
          {t("listingsQueue.filterStatus")}
          <select value={status} onChange={(e) => handleStatusChange(e.target.value as ListingStatusFilter)}>
            <option value="">{t("listingsQueue.statusAll")}</option>
            <option value="pending">{t("listingsQueue.statusPending")}</option>
            <option value="approved">{t("listingsQueue.statusApproved")}</option>
            <option value="rejected">{t("listingsQueue.statusRejected")}</option>
            <option value="removed">{t("listingsQueue.statusRemoved")}</option>
          </select>
        </label>
        <label>
          {t("listingsQueue.sortBy")}
          <select value={sort} onChange={(e) => handleSortChange(e.target.value as SortDir)}>
            <option value="asc">{t("listingsQueue.sortOldest")}</option>
            <option value="desc">{t("listingsQueue.sortNewest")}</option>
          </select>
        </label>
      </div>
      {error && <div className="error">{error}</div>}
      {loading ? (
        <p>{t("common.loading")}</p>
      ) : items.length === 0 ? (
        <p>{t("listingsQueue.noResults")}</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>{t("listingsQueue.colTitle")}</th>
              <th>{t("listingsQueue.colCategory")}</th>
              <th>{t("listingsQueue.colOwner")}</th>
              <th>{t("listingsQueue.colStatus")}</th>
              <th>{t("listingsQueue.colSubmitted")}</th>
              <th></th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {items.map((listing) => (
              <tr key={listing.id}>
                <td>{listing.title}</td>
                <td>{listingCategoryLabel(lang, listing.category)}</td>
                <td>{listing.owner_email}</td>
                <td>
                  <span className={`badge badge-status-${listing.status}`}>
                    {t(STATUS_LABEL_KEY[listing.status])}
                  </span>
                </td>
                <td>{new Date(listing.created_at).toLocaleString(lang)}</td>
                <td>
                  {listing.status === "pending" && (
                    <span className={listing.is_edit ? "badge badge-edit" : "badge badge-new"}>
                      {listing.is_edit ? t("listingsQueue.badgeEdited") : t("listingsQueue.badgeNew")}
                    </span>
                  )}
                </td>
                <td>
                  <Link to={`/listings/${listing.id}?${params.toString()}`}>{t("listingsQueue.review")}</Link>
                </td>
              </tr>
            ))}
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
