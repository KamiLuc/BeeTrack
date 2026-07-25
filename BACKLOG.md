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

---

## 3. Reports & Analytics

| ID       | Layer | Status | Title                   | Notes                                                |
| -------- | ----- | ------ | ----------------------- | ---------------------------------------------------- |
| RP-01-BE | `BE`  | `[x]`  | Dashboard data endpoint | No new endpoint — composed client-side from existing per-hive endpoints (hive list, inspections, treatments, feedings, harvests) |
| RP-01-FE | `FE`  | `[x]`  | Dashboard screen        | Overview of all hives at a glance                    |
| RP-02-BE | `BE`  | `[x]`  | PDF report export       | `POST /apiaries/{id}/report/pdf` — generates a formal PDF (per-hive, per-category sections) from the same hive/category/date-range filters as the dashboard, using an embedded DejaVu Sans font for Polish diacritics |
| RP-02-FE | `FE`  | `[x]`  | PDF report download button | "Pobierz PDF"/"Download PDF" button on the Dashboard screen; posts current filters to RP-02-BE and shares/saves the result via the `printing` package |

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
| INF-06-BE | `BE`  | `[x]`  | Input validation & structured error responses | First pass: shared `requireAuth`/`decodeJSON`/`parsePathID` helpers in `internal/handler/helpers.go` collapse ~280 duplicated boilerplate call sites (auth-context check, JSON body decode, path-ID parsing) across ~17 handler files. Zero change to wire format/error codes — same `{code, message}` shape and same per-resource messages as before. Consolidating the 87 duplicated "required"/"too long"/"invalid" service-layer validation sentinels is a separate, larger follow-up, deliberately out of scope here. |
| INF-07-BE | `BE`  | `[ ]`  | Structured JSON logging                       |       |
| INF-08-BE | `BE`  | `[ ]`  | Server-side re-compression of uploaded inspection/listing images | Client already caps uploads at 5 MB (`generalPhotoTooLarge` guard); re-encode on the backend after upload to shrink stored file size regardless of client behavior (gallery picks, PNG/WebP, web). Needs a Go image-decode/re-encode dependency, a target quality/resolution, and EXIF orientation handling. |

---

## 9. Honey Certification & Blockchain

> Immutable honey batch certification stored on Polygon blockchain. Each batch gets a QR code that verifies authenticity via blockchain hash of lab PDF.
>
> **Blockchain Strategy:** Store minimal data on-chain (hash, metadata hash, timestamp) for cost efficiency. PDF hash links to lab report; scanning verifies hash hasn't changed. Certification runs fully asynchronously via a durable jobs queue and background worker.
>
> Core flow (DB, models, blockchain integration, worker, API handlers, PDF upload, status badge) is done. Remaining items below; production-hardening items are optional — this is thesis/testnet scope.

| ID       | Layer | Status | Title                        | Notes                                                                                                     |
| -------- | ----- | ------ | ----------------------------- | ----------------------------------------------------------------------------------------------------------- |
| HC-FE-16 | `FE`  | `[x]`  | ~~Verification details modal~~ | **Descoped** — owners don't need in-app certification history; already fully visible in the admin panel. Endpoint (`GET /api/v1/honey-batches/{id}/certifications`) stays for the admin panel's use. |
| HC-FE-17 | `FE`  | `[x]`  | Hash comparison display      | Show stored hash vs. on-chain hash                                                                        |
| HC-10-04 | `BE`  | `[ ]`  | Gas fee management            | *(optional)* `gas_used` already persisted per-certification row; gas relay/price-spike alerting unnecessary on free testnet gas |
| HC-10-05 | `FE`  | `[ ]`  | Offline handling              | *(optional)* No local blockchain-write queue needed — app never triggers chain writes directly            |
| HC-10-06 | `FE`  | `[ ]`  | Loading states                | Distinct UI per lifecycle state; "Certify" when `certification` is `null`, "Retry" on `failed`/`reverted`  |
| HC-10-07 | `FE`  | `[ ]`  | Error handling                | Error copy mapped from lifecycle status (`null` = neutral, not an error; in-progress states aren't errors) |
| HC-10-08 | `FE`  | `[ ]`  | Localization                  | l10n keys for all 7 lifecycle states + separate non-enum key for null-certification case                  |
| HC-10-09 | `FE`  | `[ ]`  | Empty states                  |                                                                                                             |
| HC-10-10 | `BE`  | `[ ]`  | Database indexing             | *(optional)* Already folded into migrations HC-DB-01–04 — this row is for any additional indexing found later |
