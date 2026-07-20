import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { ApiError } from "../api/client";
import { listPendingListings, type AdminListing } from "../api/listings";
import { useI18n } from "../i18n/I18nContext";

const PAGE_SIZE = 20;

export function ListingsQueuePage() {
  const { lang, t } = useI18n();
  const [items, setItems] = useState<AdminListing[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    listPendingListings(PAGE_SIZE, offset)
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
  }, [offset, t]);

  return (
    <div>
      <h1>{t("listingsQueue.title")}</h1>
      {error && <div className="error">{error}</div>}
      {loading ? (
        <p>{t("common.loading")}</p>
      ) : items.length === 0 ? (
        <p>{t("listingsQueue.noPending")}</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>{t("listingsQueue.colTitle")}</th>
              <th>{t("listingsQueue.colCategory")}</th>
              <th>{t("listingsQueue.colSubmitted")}</th>
              <th></th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {items.map((listing) => (
              <tr key={listing.id}>
                <td>{listing.title}</td>
                <td>{listing.category}</td>
                <td>{new Date(listing.created_at).toLocaleString(lang)}</td>
                <td>
                  <span className={listing.is_edit ? "badge badge-edit" : "badge badge-new"}>
                    {listing.is_edit ? t("listingsQueue.badgeEdited") : t("listingsQueue.badgeNew")}
                  </span>
                </td>
                <td>
                  <Link to={`/listings/${listing.id}`}>{t("listingsQueue.review")}</Link>
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
