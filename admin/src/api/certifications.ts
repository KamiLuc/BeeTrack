import { request } from "./client";
import type { Page } from "./listings";

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
  pdf_url: string;
};

export function listPendingCertificationRequests(
  limit: number,
  offset: number,
): Promise<Page<CertificationRequest>> {
  return request<Page<CertificationRequest>>("/admin/certification-requests", {
    query: { status: "pending", limit, offset },
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
