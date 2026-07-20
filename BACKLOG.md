# BeeTrack — Backlog

> Living document. Treat each item like a Jira ticket — update status as work progresses.
> **Stack:** Flutter (Android + Web) · Go (backend API) · PostgreSQL · Docker

---

## Status Legend

| Symbol | Meaning     |
| ------ | ----------- |
| `[ ]`  | To do       |
| `[~]`  | In progress |
| `[x]`  | Done        |
| `[!]`  | Blocked     |

---

## Epics Overview

1. [UX Polish](#1-ux-polish)
2. [Honey Harvest Tracking](#2-honey-harvest-tracking)
3. [Reports & Analytics](#3-reports--analytics)
4. [Bulk Operations](#4-bulk-operations)
5. [Voice Logging](#5-voice-logging)
6. [Queen Recognition (AI Feature)](#6-queen-recognition-ai-feature)
7. [MCP Server](#7-mcp-server)
8. [Infrastructure & DevOps](#8-infrastructure--devops)
9. [Honey Certification & Blockchain](#9-honey-certification--blockchain)
10. [Admin Panel & Moderation](#10-admin-panel--moderation)

---

## 3. Reports & Analytics

| ID       | Layer | Status | Title                   | Notes                                                |
| -------- | ----- | ------ | ----------------------- | ---------------------------------------------------- |
| RP-01-BE | `BE`  | `[ ]`  | Dashboard data endpoint | Hive count, last inspection dates, active treatments |
| RP-01-FE | `FE`  | `[ ]`  | Dashboard screen        | Overview of all hives at a glance                    |

---

## 5. Voice Logging

| ID       | Layer | Status | Title                                           | Notes                                                                                                                        |
| -------- | ----- | ------ | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| VC-01-BE | `BE`  | `[ ]`  | Voice endpoint                                  | POST /api/v1/hives/{hiveId}/voice — accepts audio file, calls Whisper → Claude, dispatches to correct service                |
| VC-02-BE | `BE`  | `[ ]`  | Claude intent parser                            | Given transcript + hive context, returns structured action (log_inspection / log_treatment / log_harvest) with fields filled |
| VC-03-FE | `FE`  | `[ ]`  | Hold-to-record mic button on hive detail screen | Uses `record` package; sends audio to VC-01 on release; saves immediately, no confirmation step                              |
| VC-04-FE | `FE`  | `[ ]`  | Result snackbar                                 | Show what was saved ("Inspection logged: queen added, brood good") so user knows what was recorded; tap to edit if wrong     |

---

## 6. Queen Recognition (AI Feature)

> **Deferred — implement after core app is stable.**

| ID       | Layer | Status | Title                                  | Notes                                               |
| -------- | ----- | ------ | --------------------------------------- | --------------------------------------------------- |
| QR-01-FE | `FE`  | `[ ]`  | Camera capture screen                  | Live preview + capture button                       |
| QR-02-BE | `BE`  | `[ ]`  | Upload photo endpoint                  | Multipart form POST                                 |
| QR-02-FE | `FE`  | `[ ]`  | Upload photo from app                  |                                                     |
| QR-03-BE | `BE`  | `[ ]`  | CV model inference endpoint            | Go service calls Python/ONNX/TFLite model           |
| QR-04-BE | `BE`  | `[ ]`  | Return bounding box + confidence score |                                                     |
| QR-05-FE | `FE`  | `[ ]`  | Overlay bounding box on image          |                                                     |
| QR-06-BE | `BE`  | `[ ]`  | Collect user feedback endpoint         | correct / incorrect for retraining                  |
| QR-06-FE | `FE`  | `[ ]`  | User feedback UI                       |                                                     |
| QR-07-BE | `BE`  | `[ ]`  | Model training pipeline                | Dataset, annotations, training script — thesis core |
| QR-08-BE | `BE`  | `[ ]`  | Model evaluation metrics               | mAP, precision, recall — document for thesis        |

---

## 7. MCP Server

> **Deferred — enables AI voice assistant integration.**

| ID        | Layer | Status | Title                      | Notes                                       |
| --------- | ----- | ------ | --------------------------- | ------------------------------------------- |
| MCP-01-BE | `BE`  | `[ ]`  | MCP server endpoint        | HTTP+SSE transport, runs alongside REST API |
| MCP-02-BE | `BE`  | `[ ]`  | Tool: `create_inspection`  |                                             |
| MCP-03-BE | `BE`  | `[ ]`  | Tool: `log_treatment`      |                                             |
| MCP-04-BE | `BE`  | `[ ]`  | Tool: `log_harvest`        |                                             |
| MCP-05-BE | `BE`  | `[ ]`  | Tool: `get_hive_summary`   | Latest inspection + active treatments       |
| MCP-06-BE | `BE`  | `[ ]`  | Tool: `list_hives`         |                                             |
| MCP-07-BE | `BE`  | `[ ]`  | Auth for MCP clients       | API key or OAuth                            |
| MCP-08-BE | `BE`  | `[ ]`  | Voice pipeline integration | Whisper → Claude/GPT with MCP tools         |

---

## 8. Infrastructure & DevOps

| ID        | Layer | Status | Title                                         | Notes |
| --------- | ----- | ------ | --------------------------------------------- | ----- |
| INF-05-BE | `BE`  | `[ ]`  | REST API — OpenAPI / Swagger spec             |       |
| INF-06-BE | `BE`  | `[ ]`  | Input validation & structured error responses |       |
| INF-07-BE | `BE`  | `[ ]`  | Structured JSON logging                       |       |

---

## 9. Honey Certification & Blockchain

> Immutable honey batch certification stored on Polygon blockchain. Each batch gets a QR code that verifies authenticity via blockchain hash of lab PDF.
>
> **Blockchain Strategy:** Store minimal data on-chain (hash, metadata hash, timestamp) for cost efficiency. PDF hash links to lab report; scanning verifies hash hasn't changed. Certification runs fully asynchronously via a durable jobs queue and background worker — see [HONEY_BLOCKCHAIN_PLAN.md](HONEY_BLOCKCHAIN_PLAN.md).
>
> **Full task breakdown moved to its own file:** [HONEY_BLOCKCHAIN_BACKLOG.md](HONEY_BLOCKCHAIN_BACKLOG.md) — the async/jobs-queue redesign roughly doubled the task count, so it no longer fits comfortably inline here. Update status there as work progresses; this entry stays as the pointer.

---

## 10. Admin Panel & Moderation

> New `ADMIN` user role (set manually in the DB only) plus a React admin panel for reviewing marketplace listings (new + edited) and honey batch certification requests before they go live/on-chain. Marketplace listings now require admin approval before appearing publicly; honey batch certification now requires admin approval before the existing blockchain job queue (Epic 9) is enqueued — the worker, idempotency, and retry logic from Epic 9 are unchanged. Full plan: [ADMIN_PANEL_PLAN.md](ADMIN_PANEL_PLAN.md).

| ID       | Layer   | Status | Title                                              | Notes                                                                 |
| -------- | ------- | ------ | --------------------------------------------------- | ---------------------------------------------------------------------- |
| AP-DB-01 | `BE`    | `[x]`  | Add `role` column to `users`                        | Manual SQL only to promote to admin, no API path                       |
| AP-DB-02 | `BE`    | `[x]`  | Add moderation fields to `listings`                  | `status`, `rejection_reason`, `first_approved_at`, `reviewed_by/at`; backfill existing rows to `approved` |
| AP-DB-03 | `BE`    | `[x]`  | Create `honey_batch_certification_requests` table    | Review queue sitting upstream of `blockchain_jobs`                     |
| AP-BE-01 | `BE`    | `[x]`  | Extend `User` model with `Role`/`IsAdmin()`          |                                                                        |
| AP-BE-02 | `BE`    | `[x]`  | Extend `Listing` model with moderation fields         |                                                                        |
| AP-BE-03 | `BE`    | `[x]`  | Create `HoneyBatchCertificationRequest` model         |                                                                        |
| AP-BE-04 | `BE`    | `[x]`  | Repository support for moderation + review queues     |                                                                        |
| AP-BE-04b| `BE`    | `[x]`  | Publish existing data + seed pending listings          | Backfill via AP-DB-02 migration; seed script approves its own listings and adds a few left `pending` for admin QA. Also surfaced (and fixed) a real bug: `User.Role`/`Listing.Status` need `gorm:"default:..."` tags so direct struct construction doesn't insert an empty string against the new CHECK constraints |
| AP-BE-05 | `BE`    | `[x]`  | `RequireAdmin` middleware                             | DB-checked role, not JWT-claimed — immediate revocation                |
| AP-BE-06 | `BE`    | `[x]`  | Listing create/edit defaults to `pending`             | Public reads filtered to `approved`; owner views unfiltered            |
| AP-BE-07 | `BE`    | `[x]`  | `ListingModerationService` (approve/reject)            | Reject requires a reason                                               |
| AP-BE-08 | `BE`    | `[x]`  | Certification review gate + `CertificationReviewService` | `RequestCertification` creates a review request, not a job directly |
| AP-BE-09 | `BE`    | `[x]`  | Admin PDF access bypass                                | Admin can view any batch's lab PDF regardless of ownership             |
| AP-BE-10 | `BE`    | `[x]`  | `GET /api/v1/admin/listings`                           | Includes computed `is_edit` flag                                       |
| AP-BE-11 | `BE`    | `[x]`  | `GET /api/v1/admin/listings/{id}`                      |                                                                        |
| AP-BE-12 | `BE`    | `[x]`  | `POST /api/v1/admin/listings/{id}/approve`             |                                                                        |
| AP-BE-13 | `BE`    | `[x]`  | `POST /api/v1/admin/listings/{id}/reject`              | Requires `reason`                                                      |
| AP-BE-14 | `BE`    | `[x]`  | `GET /api/v1/admin/certification-requests`             |                                                                        |
| AP-BE-15 | `BE`    | `[x]`  | `GET /api/v1/admin/certification-requests/{id}`        |                                                                        |
| AP-BE-16 | `BE`    | `[x]`  | `POST /api/v1/admin/certification-requests/{id}/approve` | Enqueues the `blockchain_jobs` row — Epic 9 worker takes over unchanged |
| AP-BE-17 | `BE`    | `[x]`  | `POST /api/v1/admin/certification-requests/{id}/reject` | Requires `reason`                                                      |
| AP-BE-18 | `BE`    | `[x]`  | `GET /api/v1/admin/honey-batches/{id}/pdf`             |                                                                        |
| AP-BE-19 | `BE`    | `[x]`  | Include `role` in `GET /api/v1/users/me`               | Client-side UX only, not a security boundary                           |
| AP-BE-20 | `BE`    | `[x]`  | Wire admin middleware + routes into `main.go`, CORS    |                                                                        |
| AP-FE-01 | `ADMIN` | `[x]`  | React admin panel project scaffold                      | New `admin/` directory, Vite + React + TS                              |
| AP-FE-02 | `ADMIN` | `[x]`  | API client + auth/listings/certifications modules       |                                                                        |
| AP-FE-03 | `ADMIN` | `[x]`  | Auth context + route guard                              |                                                                        |
| AP-FE-04 | `ADMIN` | `[x]`  | Login page                                              |                                                                        |
| AP-FE-05 | `ADMIN` | `[x]`  | App shell + nav                                         |                                                                        |
| AP-FE-06 | `ADMIN` | `[x]`  | Listings queue page                                      | "New" vs "Edited" badge                                                |
| AP-FE-07 | `ADMIN` | `[x]`  | Listing detail/review page                               | Photos + approve/reject (reason)                                       |
| AP-FE-08 | `ADMIN` | `[x]`  | Certification requests queue page                        |                                                                        |
| AP-FE-09 | `ADMIN` | `[x]`  | Certification detail/review page                         | Embedded lab PDF via fetch-then-blob-URL                               |
| AP-FE-10 | `ADMIN` | `[x]`  | Docker/dev wiring + README section                        |                                                                        |
| AP-10-01 | `FE`    | `[ ]`  | Listing status badge on My Listings screen (Flutter)      | Pending / Rejected (+reason) / Live                                     |
| AP-10-02 | `FE`    | `[ ]`  | Certification request status on honey batch card (Flutter) | "Pending admin review" / "Rejected by admin" states                    |
| AP-10-03 | `FE`    | `[ ]`  | Localization for new statuses                             | `app_en.arb` + `app_pl.arb`                                             |
