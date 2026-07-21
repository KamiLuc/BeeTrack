import { request } from "./client";
import type { Page } from "./listings";

export type CertificationJobStatus =
  | "queued"
  | "submitting"
  | "submitted"
  | "pending_confirmation"
  | "confirmed"
  | "failed"
  | "reverted";

export type CertificationRequest = {
  id: number;
  batch_id: number;
  requested_by: number;
  requester_email: string;
  status: "pending" | "approved" | "rejected";
  rejection_reason: string | null;
  blockchain_job_id: number | null;
  created_at: string;
  gathering_date: string;
  amount_grams: number;
  honey_type: string;
  processing_method: string;
  pdf_url: string;
  job_status: CertificationJobStatus | null;
  job_last_error: string | null;
  transaction_hash: string | null;
  block_number: number | null;
  confirmation_timestamp: string | null;
};

export type CertificationStatusFilter = "" | "pending" | "approved" | "rejected";
export type SortDir = "asc" | "desc";

export function listCertificationRequests(
  status: CertificationStatusFilter,
  query: string,
  sort: SortDir,
  limit: number,
  offset: number,
): Promise<Page<CertificationRequest>> {
  return request<Page<CertificationRequest>>("/admin/certification-requests", {
    query: { status: status || undefined, q: query || undefined, sort, limit, offset },
  });
}

export function getCertificationRequest(id: number): Promise<CertificationRequest> {
  return request<CertificationRequest>(`/admin/certification-requests/${id}`);
}

export function approveCertificationRequest(id: number): Promise<CertificationRequest> {
  return request<CertificationRequest>(`/admin/certification-requests/${id}/approve`, {
    method: "POST",
  });
}

export function rejectCertificationRequest(id: number, reason: string): Promise<CertificationRequest> {
  return request<CertificationRequest>(`/admin/certification-requests/${id}/reject`, {
    method: "POST",
    body: { reason },
  });
}
