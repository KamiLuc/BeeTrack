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
5. [Queen Recognition (AI Feature)](#5-queen-recognition-ai-feature)
6. [Infrastructure & DevOps](#6-infrastructure--devops)
7. [Honey Certification & Blockchain](#7-honey-certification--blockchain)

> Voice logging and the MCP-based AI assistant moved out of this backlog — see
> [AI_ASSISTANT.md](AI_ASSISTANT.md) for the design and
> [BACKLOG_AI_ASSISTANT.md](BACKLOG_AI_ASSISTANT.md) for its tickets.

---

## 3. Reports & Analytics

| ID       | Layer | Status | Title                   | Notes                                                |
| -------- | ----- | ------ | ----------------------- | ---------------------------------------------------- |
| RP-01-BE | `BE`  | `[x]`  | Dashboard data endpoint | No new endpoint — composed client-side from existing per-hive endpoints (hive list, inspections, treatments, feedings, harvests) |
| RP-01-FE | `FE`  | `[x]`  | Dashboard screen        | Overview of all hives at a glance                    |
| RP-02-BE | `BE`  | `[x]`  | PDF report export       | `POST /apiaries/{id}/report/pdf` — generates a formal PDF (per-hive, per-category sections) from the same hive/category/date-range filters as the dashboard, using an embedded DejaVu Sans font for Polish diacritics |
| RP-02-FE | `FE`  | `[x]`  | PDF report download button | "Pobierz PDF"/"Download PDF" button on the Dashboard screen; posts current filters to RP-02-BE and shares/saves the result via the `printing` package |

---

## 5. Queen Recognition (AI Feature)

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

## 6. Infrastructure & DevOps

| ID        | Layer | Status | Title                                         | Notes |
| --------- | ----- | ------ | --------------------------------------------- | ----- |
| INF-05-BE | `BE`  | `[x]`  | REST API — OpenAPI / Swagger spec             | `docs/openapi.yaml`, generated from `docs/api.md` by `docs/generate_openapi.py` (re-run after editing api.md). Writing it surfaced real drift in api.md itself: the entire Feedings feature (8 endpoints) and admin listing remove/restore were undocumented, `GET /medicines` was documented as unauthenticated/static when it's actually an authed per-user history endpoint, and a `PATCH .../frames` endpoint was documented but no longer exists in code — all fixed. `backend/cmd/api/openapi_sync_test.go` now fails CI if main.go's registered routes and openapi.yaml's paths ever diverge again. |
| INF-06-BE | `BE`  | `[x]`  | Input validation & structured error responses | First pass: shared `requireAuth`/`decodeJSON`/`parsePathID` helpers in `internal/handler/helpers.go` collapse ~280 duplicated boilerplate call sites (auth-context check, JSON body decode, path-ID parsing) across ~17 handler files. Zero change to wire format/error codes — same `{code, message}` shape and same per-resource messages as before. Consolidating the 87 duplicated "required"/"too long"/"invalid" service-layer validation sentinels is a separate, larger follow-up, deliberately out of scope here. |
| INF-07-BE | `BE`  | `[x]`  | Structured JSON logging                       | `pkg/logging` sets `slog`'s global default to a JSON handler (`LOG_LEVEL` env var, default `info`); `internal/middleware/logging.go` wraps the whole app, generating/reusing an `X-Request-Id` header (echoed on the response) and attaching a request-scoped logger to context that emits one line per request (method, path, status, duration_ms, remote_addr, request_id). `Auth`/`OptionalAuth` enrich that context logger with `user_id` once a token validates, so downstream logs (e.g. `auth.go`'s email-send failures) carry both IDs for correlation without extra plumbing. Converted the remaining ad-hoc `log.Printf`/`Fatal` call sites in `main.go` and the blockchain worker to `slog`. `cmd/seed`/`cmd/resetdb` deliberately left on plain stdout — human-run dev scripts, not server logs. Bug fix (found via the AI assistant's SSE streaming, AST-09-BE): the `statusWriter` response wrapper this middleware installs to capture the status code didn't implement `http.Flusher`, so it silently swallowed every streaming response's `Flush()` calls once it passed through this middleware — not just the assistant's, any future streaming endpoint would have hit the same bug. Fixed by having `statusWriter` forward `Flush()` to the underlying `ResponseWriter`. |
| INF-08-BE | `BE`  | `[ ]`  | Server-side re-compression of uploaded inspection/listing images | Client already caps uploads at 5 MB (`generalPhotoTooLarge` guard); re-encode on the backend after upload to shrink stored file size regardless of client behavior (gallery picks, PNG/WebP, web). Needs a Go image-decode/re-encode dependency, a target quality/resolution, and EXIF orientation handling. |

---

## 7. Honey Certification & Blockchain

> Immutable honey batch certification stored on Polygon blockchain. Each batch gets a QR code that verifies authenticity via blockchain hash of lab PDF.
>
> **Blockchain Strategy:** Store minimal data on-chain (hash, metadata hash, timestamp) for cost efficiency. PDF hash links to lab report; scanning verifies hash hasn't changed. Certification runs fully asynchronously via a durable jobs queue and background worker.
>
> Core flow (DB, models, blockchain integration, worker, API handlers, PDF upload, status badge) is done. Remaining items below; production-hardening items are optional — this is thesis/testnet scope.

| ID       | Layer | Status | Title                        | Notes                                                                                                     |
| -------- | ----- | ------ | ----------------------------- | ----------------------------------------------------------------------------------------------------------- |
| HC-FE-16 | `FE`  | `[x]`  | ~~Verification details modal~~ | **Descoped** — owners don't need in-app certification history; already fully visible in the admin panel. Endpoint (`GET /api/v1/honey-batches/{id}/certifications`) stays for the admin panel's use. |
| HC-FE-17 | `FE`  | `[x]`  | Hash comparison display      | Show stored hash vs. on-chain hash                                                                        |
| HC-10-04 | `BE`  | `[x]`  | Gas fee management            | Scoped down to a pre-approval gas cost preview — `GET /admin/certification-requests/{id}/estimate-gas` dry-runs `certify()` and returns cost in wei/POL/PLN. Gas relay retry-with-bumped-gas and price-spike alerting stay out of scope on free testnet gas |
| HC-10-05 | `FE`  | `[ ]`  | Offline handling              | *(optional)* No local blockchain-write queue needed — app never triggers chain writes directly            |
| HC-10-06 | `FE`  | `[ ]`  | Loading states                | Distinct UI per lifecycle state; "Certify" when `certification` is `null`, "Retry" on `failed`/`reverted`  |
| HC-10-07 | `FE`  | `[ ]`  | Error handling                | Error copy mapped from lifecycle status (`null` = neutral, not an error; in-progress states aren't errors) |
| HC-10-08 | `FE`  | `[ ]`  | Localization                  | l10n keys for all 7 lifecycle states + separate non-enum key for null-certification case                  |
| HC-10-09 | `FE`  | `[ ]`  | Empty states                  |                                                                                                             |
| HC-10-10 | `BE`  | `[ ]`  | Database indexing             | *(optional)* Already folded into migrations HC-DB-01–04 — this row is for any additional indexing found later |
