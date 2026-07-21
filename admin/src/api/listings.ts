import { request } from "./client";

export type ListingImage = {
  id: number;
  listing_id: number;
  url: string;
  display_order: number;
  created_at: string;
};

export type AdminListing = {
  id: number;
  user_id: number;
  owner_email: string;
  title: string;
  description: string;
  category: string;
  price: number;
  quantity: number;
  address: string;
  lat: number | null;
  lng: number | null;
  contact_phone: string;
  contact_email: string;
  status: "pending" | "approved" | "rejected" | "removed";
  rejection_reason: string | null;
  is_edit: boolean;
  created_at: string;
  updated_at: string;
  images: ListingImage[];
};

export type Page<T> = { items: T[]; total: number };

export type ListingStatusFilter = "" | "pending" | "approved" | "rejected" | "removed";
export type SortDir = "asc" | "desc";

export function listListings(
  status: ListingStatusFilter,
  query: string,
  sort: SortDir,
  limit: number,
  offset: number,
): Promise<Page<AdminListing>> {
  return request<Page<AdminListing>>("/admin/listings", {
    query: { status: status || undefined, q: query || undefined, sort, limit, offset },
  });
}

export function getListing(id: number): Promise<AdminListing> {
  return request<AdminListing>(`/admin/listings/${id}`);
}

export function approveListing(id: number): Promise<AdminListing> {
  return request<AdminListing>(`/admin/listings/${id}/approve`, { method: "POST" });
}

export function rejectListing(id: number, reason: string): Promise<AdminListing> {
  return request<AdminListing>(`/admin/listings/${id}/reject`, {
    method: "POST",
    body: { reason },
  });
}

export function removeListing(id: number, reason: string): Promise<AdminListing> {
  return request<AdminListing>(`/admin/listings/${id}/remove`, {
    method: "POST",
    body: { reason },
  });
}

export function restoreListing(id: number): Promise<AdminListing> {
  return request<AdminListing>(`/admin/listings/${id}/restore`, { method: "POST" });
}
