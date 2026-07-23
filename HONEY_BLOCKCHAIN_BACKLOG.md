# Honey Certification & Blockchain — Backlog

> Living document for Epic 9, broken out from the main [BACKLOG.md](BACKLOG.md) because the async/jobs-queue redesign roughly doubled its task count. Task breakdown mirrors [HONEY_BLOCKCHAIN_PLAN.md](HONEY_BLOCKCHAIN_PLAN.md) — see that file for full rationale and field-level detail behind each row.
>
> Treat each item like a Jira ticket — update status as work progresses.
>
> **Architecture in one line:** `CreateBatch` persists a batch and — only if the caller opts in via `request_certification` — enqueues a `blockchain_jobs` row, then returns immediately; a background `BlockchainWorker` owns all Polygon RPC interaction, writes results to an append-only `honey_batch_certifications` history, and drives a 7-state lifecycle (queued → submitting → submitted → pending_confirmation → confirmed / failed / reverted). Certification is opt-in, not automatic — a batch can have **no** certification row at all (nil/null, not a status value) indefinitely until the owner requests certification later. Public verification uses an unguessable `verification_token`, never the numeric batch id. Honey amount is stored as integer grams, never `float64`.
>
> **Thesis scope, not production:** this is a final CS thesis feature — target environment is **Polygon Amoy testnet only**, exercised in a testing environment. Rows tagged **(optional)** below are production-hardening that can be skipped without weakening the thesis; everything else is in scope because it's the engineering content being demonstrated (async jobs, idempotency, deterministic hashing, append-only history).

---

## Status Legend

| Symbol | Meaning     |
| ------ | ----------- |
| `[ ]`  | To do       |
| `[~]`  | In progress |
| `[x]`  | Done        |
| `[!]`  | Blocked     |

---

## Sections Overview

1. [Database Foundation](#91-database-foundation)
2. [Backend — Models & Persistence](#92-backend--models--persistence)
3. [Backend — Blockchain Integration](#93-backend--blockchain-integration)
4. [Backend — Business Logic & Worker](#94-backend--business-logic--worker)
5. [Backend — API Handlers](#95-backend--api-handlers)
6. [Backend — Integration & Wiring](#96-backend--integration--wiring)
7. [Frontend — Models & Repositories](#97-frontend--models--repositories)
8. [Frontend — State Management](#98-frontend--state-management)
9. [Frontend — Core Screens](#99-frontend--core-screens)
10. [Frontend — Utils & Widgets](#910-frontend--utils--widgets)
11. [Polish & Edge Cases](#911-polish--edge-cases)

---

## 9.1 Database Foundation

| ID       | Layer | Status | Title                                     | Notes                                                                                                                                                                                            |
| -------- | ----- | ------ | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| HC-DB-01 | `DB`  | `[x]`  | Create `honey_batches` table              | `backend/migrations/029_...`. No blockchain fields — those live in HC-DB-04. Includes `verification_token` (UUID, UNIQUE) and `amount_grams` (BIGINT, not float). `deleted_at` for soft delete. `apiary_id` was later dropped entirely (`034_remove_honey_batch_apiary.sql`) — a batch belongs only to its creating user, no apiary scoping (plan text was stale on this). A `pdf_filename TEXT NOT NULL DEFAULT ''` column was later added (`035_honey_batch_pdf_filename.sql`) to store the uploaded lab PDF's original filename alongside `pdf_file_hash`. |
| HC-DB-02 | `DB`  | `[x]`  | Create `honey_batch_qr_codes` table       | `backend/migrations/030_...`. `qr_code_data` encodes `/verify/{verification_token}`, never the numeric id.                                                                                     |
| HC-DB-03 | `DB`  | `[x]`  | Create `blockchain_jobs` table            | `backend/migrations/031_...`. Durable queue: status, attempt_count, next_retry_at, last_error. Index (status, next_retry_at) for the worker's claim query.                                     |
| HC-DB-04 | `DB`  | `[x]`  | Create `honey_batch_certifications` table | `backend/migrations/032_...`. Append-only per-batch history (chain_id, contract_address, tx_hash, block_number, status, gas_used). Partial UNIQUE (batch_id) WHERE status is "live" — idempotency guard. |

---

## 9.2 Backend — Models & Persistence

| ID        | Layer | Status | Title                                             | Notes                                                                                                                       |
| --------- | ----- | ------ | --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| HC-BE-07  | `BE`  | `[x]`  | Model: `HoneyBatch` struct                        | Mirrors DB schema. `AmountGrams int64`, `VerificationToken string`. No blockchain fields on this struct. Gained `PDFFilename string` alongside the existing PDF hash field (see HC-DB-01).                    |
| HC-BE-08  | `BE`  | `[x]`  | Model: `ProcessingMethod` enum                    | raw, filtered, pasteurized + `IsValidProcessingMethod`                                                                      |
| HC-BE-07b | `BE`  | `[x]`  | Model: `HoneyBatchCertification` + status lifecycle | `CertificationStatus` type: queued/submitting/submitted/pending_confirmation/confirmed/failed/reverted. No "not requested" enum value — a never-certified batch has a nil `*HoneyBatchCertification`, not a status. `IsTerminal()`/`IsLive()` helpers. Single source of truth mirrored in DB CHECK, API JSON, Dart enum (HC-FE-08b). |
| HC-BE-07c | `BE`  | `[x]`  | Model: `BlockchainJob` struct                     | Reuses `CertificationStatus` for its own status field.                                                                      |
| HC-BE-09  | `BE`  | `[x]`  | Repository: `HoneyBatchRepository` — Create        | Runs in a transaction together with the initial `blockchain_jobs` insert (HC-BE-13) — a batch is never persisted without a job. |
| HC-BE-10  | `BE`  | `[x]`  | Repository: Get by ID / by verification token      | `GetByID` (owner-scoped), `GetByVerificationToken` (public path)                                                            |
| HC-BE-11  | `BE`  | `[x]`  | Repository: List batches by user/apiary            | `ListByUserID`, paginated. `ListByApiaryID` was later removed as dead code once `apiary_id` was dropped from `honey_batches` (see HC-DB-01) — plan text was stale on this. |
| HC-BE-12  | `BE`  | `[x]`  | Repository: Update notes / soft delete             | `UpdateNotes`, `SoftDelete` — no status/blockchain mutation methods here anymore. `UpdateNotes` only touches `honey_type`; there's no `notes` column on `honey_batches` (plan text was stale on this). Later replaced by `UpdateFields(ctx, id, gatheringDate, amountGrams, processingMethod, honeyType, metadataHash)`, writing all four editable fields plus the recomputed `metadata_hash` in one statement (see HC-BE-22). Signature changed again to `UpdateFields(ctx, batch *model.HoneyBatch)`, which also persists `lab_pdf_url`/`pdf_filename`/`pdf_file_hash` so an edit can replace the batch's PDF. |
| HC-BE-12b | `BE`  | `[x]`  | Repository: `HoneyBatchCertificationRepository` + `BlockchainJobRepository` | Certification repo: Create, GetLatestByBatchID, ListByBatchID, UpdateStatus. Job repo: Create, `ClaimNext` (SELECT...FOR UPDATE SKIP LOCKED, atomically flips claimed job to `submitting` in the same tx), MarkSubmitting/Submitted/Failed, ListPendingConfirmation. |

---

## 9.3 Backend — Blockchain Integration

| ID        | Layer | Status | Title                                | Notes                                                                                                                                            |
| --------- | ----- | ------ | -------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| HC-BE-01  | `BE`  | `[x]`  | Blockchain config                    | RPC URL, contract address, private key, chain ID (default 80002 Amoy), plus `JobPollInterval` (10s), `ConfirmationPollInterval` (30s), `RequiredConfirmations` (12). |
| HC-BE-02  | `BE`  | `[x]`  | Smart contract (Solidity)            | `certify()` **reverts if batchID already certified** — contract-level idempotency. Event `CertificationCreated`. Owner-only caller.               |
| HC-BE-03  | `BE`  | `[x]`  | Deploy contract to Polygon           | Deployed to Amoy testnet (80002) at `0x5d92856257b2e0a8365c02aed826a857317f95ed` (tx `0x7d792e87289cbb4b613299c18c788b7207a6d2a2215289ac3c6ca4d1bcb2a6ff`). ABI stored at `backend/internal/blockchain/contracts/HoneyCertification.abi`. Set `CONTRACT_ADDRESS` env var to the address above. |
| HC-BE-04  | `BE`  | `[x]`  | Blockchain writer                    | `CertifyBatch(...)` — called **only** by the worker (HC-BE-15b), never from the HTTP path. Returns tx hash immediately, doesn't wait for confirmation. |
| HC-BE-05  | `BE`  | `[x]`  | Blockchain reader                    | `VerifyBatch(...)` + `GetTransactionStatus(txHash)` (confirmed/reverted/blockNumber/gasUsed) — used by the confirmation-polling loop (HC-BE-15c).  |
| HC-BE-06  | `BE`  | `[x]`  | Hash utilities — PDF                 | `SHA256File(filePath)` for PDF hashing.                                                                                                            |
| HC-BE-06b | `BE`  | `[x]`  | Canonical metadata hashing           | `CanonicalMetadataHash(batch)` — exact 6-field order, `\x1f`-joined, UTF-8, SHA256. No `fmt.Sprintf`/struct-JSON. `apiary_id` was dropped from the field list once `honey_batches.apiary_id` was removed (see HC-DB-01) — plan text (7 fields) was stale on this. See plan §HC-BE-06b for the full field spec — this is the language-agnostic contract, not "read the Go source." |

---

## 9.4 Backend — Business Logic & Worker

| ID        | Layer | Status | Title                                    | Notes                                                                                                                                                          |
| --------- | ----- | ------ | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| HC-BE-13  | `BE`  | `[x]`  | Service: Create honey batch              | `CreateBatch(ctx, userID, req)` — no `apiaryID` param and no apiary-membership check (plan text was stale on this; a batch belongs only to `user_id`, see HC-DB-01). Validate → hash PDF → generate verification token → compute metadata hash → **in one transaction**: insert batch, and only if `RequestCertification` is true, insert `blockchain_jobs` row (`status=queued`) → return. No blockchain call here. Default is **not** to certify. |
| HC-BE-14  | `BE`  | `[x]`  | Service: Get batch + verify              | `GetBatchWithVerification(token)` — reads latest certification from DB (kept fresh by the worker), not a live RPC call per request. No row → nil certification field, not an error. No DB-only hash-comparison field (schema doesn't store the on-chain hash to compare against). |
| HC-BE-15b | `BE`  | `[x]`  | Worker: process certify jobs             | `ProcessNextJob` — claim job → idempotency check (skip if a "live" certification already exists) → submit tx → update certification/job status. Retry via exponential backoff (1s/2s/4s/8s, capped) on failure. |
| HC-BE-15c | `BE`  | `[x]`  | Worker: poll for confirmations           | `PollSubmittedJobs` — for submitted/pending_confirmation jobs, check tx status; move to confirmed/reverted once mined & enough confirmations, or leave pending. |
| HC-BE-25  | `BE`  | `[x]`  | Idempotency guarantees (docs + tests)    | Three layers: contract revert-on-duplicate, worker's pre-submit live-certification check, DB partial unique constraint. Documented as a package doc comment on `backend/internal/worker/blockchain_worker.go`. `TestProcessNextJob_MidBroadcastCrashRecovery` simulates a worker dying right after broadcast (before the tx hash is recorded), then a retried job hitting the contract's duplicate-certify revert, and asserts exactly one live certification results. |
| HC-BE-16  | `BE`  | `[x]`  | Service: Generate QR code data           | `HoneyBatchService.GenerateQRCodeData` builds `{appURL}/verify/{verificationToken}` and persists it to `honey_batch_qr_codes` on first call; idempotent — later calls return the same row. Requires a confirmed certification (`ErrBatchNotCertified` otherwise) — a QR pointing at an uncertified batch would be misleading, and this matches HC-BE-21's "cached, immutable" assumption since confirmed is terminal. QR data itself stays just the URL, not batch fields — those are fetched live on scan, never baked in. |
| HC-BE-16b | `BE`  | `[x]`  | Verification token generation            | UUID v4 via `crypto/rand` (`google/uuid`), generated once at batch creation, immutable. Inline in `CreateBatch` (HC-BE-13) — small enough not to need its own helper. |

---

## 9.5 Backend — API Handlers

| ID        | Layer | Status | Title                                                    | Notes                                                                                                                    |
| --------- | ----- | ------ | ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| HC-BE-17  | `BE`  | `[x]`  | Handler: POST /api/v1/honey-batches                     | Auth required. Multipart form (`lab_pdf` file + fields), since PDF storage (HC-10-03) and validation (HC-10-01) landed as part of this ticket — `HoneyBatchService.CreateBatch` now owns writing/hashing the PDF instead of expecting a pre-saved path. Accepts `request_certification: bool` (default false). Returns batch + a synthetic `certification: {status: "queued"}` (no row exists yet — the worker creates one), or `null` if not requested. Blockchain failure can never cause a 500 here — CreateBatch never calls the chain. `CreateBatchRequest.PDFFilename` (the uploaded file's original `header.Filename`) is captured and stored without a new form field. |
| HC-BE-18  | `BE`  | `[x]`  | Handler: GET /api/v1/honey-batches/{id}                 | **Auth + ownership required** (ownership = `batch.UserID == callerID`; no longer public — public access moved to HC-BE-19). `HoneyBatchService.GetBatch` now synthesizes a `{status: "queued"}` placeholder when the batch has no certification row yet but a pending/unclaimed `blockchain_jobs` row exists, so the window before the worker claims the job shows "in progress" rather than looking uncertified. The token-based `GetBatchWithVerification` (HC-BE-14) is unaffected. |
| HC-BE-19  | `BE`  | `[x]`  | Handler: GET /api/v1/verify/{token}                      | Renamed from `/honey-batches/{id}/verify`. Public, token-scoped. Returns full lifecycle status, not a boolean.           |
| HC-BE-20  | `BE`  | `[x]`  | Handler: GET /api/v1/honey-batches (list)                | Auth required. Each item includes its latest certification status for list-view badges. Same pending-job placeholder as HC-BE-18 applies via `HoneyBatchService.ListBatches`.                                 |
| HC-BE-21  | `BE`  | `[x]`  | Handler: GET /api/v1/verify/{token}/qr-code              | Moved under the token-scoped path. Public, cached (immutable via `Cache-Control: max-age=31536000, immutable`). Serves a PNG (`github.com/skip2/go-qrcode`) encoding the URL from `GenerateQRCodeData` — requires a confirmed certification, same gate as HC-BE-16. Renders inline (no `Content-Disposition`), so it stays safe to embed in an `<img>` on Flutter Web. A sibling `GET /api/v1/verify/{token}/qr-code/download` serves the identical PNG with `Content-Disposition: attachment` so the FE can offer a one-click "download to print" action without forcing a download on the display screen. Download filename is derived from batch metadata (`qrCodeDownloadFilename`: `{gathering_date}_{sanitized honey_type}_{weight}kg.png`, e.g. `2024-05-01_wildflower_1.5kg.png`) rather than a fixed name — see `docs/api.md`. |
| HC-BE-22  | `BE`  | `[x]`  | Handler: PATCH /api/v1/honey-batches/{id}                | Auth + ownership. honey_type only (no `notes` column, see HC-BE-12). Later expanded to `gathering_date`, `amount_grams`, `processing_method`, and `honey_type` via `HoneyBatchService.UpdateBatch` — all four fields are now editable, and `metadata_hash` is recomputed on every edit. New `409 BATCH_LOCKED` once the batch has any certification attempt (row of any status, or a pending/unclaimed job). Endpoint changed from JSON to `multipart/form-data` with an added optional `lab_pdf` field, same semantics as Create: `UpdateBatch`/`UpdateBatchRequest` validates it (`ErrInvalidPDFType`/`ErrPDFTooLarge`), writes it to disk, and deletes the old PDF file after the DB write succeeds; if omitted, the existing PDF is left untouched. Added an optional `remove_pdf` form field: when `"true"` and no new `lab_pdf` is uploaded, `UpdateBatch` clears the existing PDF (`lab_pdf_url`, `pdf_filename`, `pdf_file_hash`) and deletes the stored file; a new upload always takes precedence over `remove_pdf`. |
| HC-BE-23  | `BE`  | `[x]`  | Handler: DELETE /api/v1/honey-batches/{id}               | Auth + ownership. Soft delete; on-chain record is untouched/immutable.                                                   |
| HC-BE-24  | `BE`  | `[x]`  | Handler: GET /api/v1/honey-batches/{id}/pdf              | Auth + ownership — owner-scoped PDF access. Not gated on certification status — the owner can always view their own upload. |
| HC-BE-24b | `BE`  | `[x]`  | Handler: GET /api/v1/verify/{token}/pdf                  | Public, token-scoped PDF access. Gated on a confirmed certification, same rationale as the QR gate — no public exposure of lab data for an uncertified batch. |
| HC-BE-24c | `BE`  | `[x]`  | Handler: POST /api/v1/honey-batches/{id}/retry-certification | Auth + ownership. `HoneyBatchService.RetryCertification` re-enqueues a `blockchain_jobs` row when no certification exists yet (first-time certify), or the latest is `failed`/`reverted`. Rejects (409, `BATCH_ALREADY_CERTIFIED`) if `IsLive()`. Backs both the FE "Certify" and "Retry" buttons (HC-10-06). |

---

## 9.6 Backend — Integration & Wiring

| ID       | Layer | Status | Title                            | Notes                                                                                                                          |
| -------- | ----- | ------ | ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| HC-BE-W1 | `BE`  | `[x]`  | Wire blockchain components in main.go | `startHoneyCertificationWorker(ctx, db)` builds the writer/reader/repos and starts the worker — optional, not fatal, if blockchain env vars aren't set (feature has no HTTP handlers yet). Adds the app's first graceful-shutdown context via `signal.NotifyContext` (SIGINT/SIGTERM), which didn't exist before. `/verify/{token}` route registration deferred to when the handlers (Phase 5) exist. |
| HC-BE-26 | `BE`  | `[x]`  | Background worker runner         | `BlockchainWorker.Run(ctx, jobInterval, confirmationInterval)` — two ticker loops, graceful shutdown on context cancel. `SweepStuckSubmitting` (5 min timeout) handles crash-mid-step recovery, run once per job-loop tick. Wired into `main.go` via HC-BE-W1. |

---

## 9.7 Frontend — Models & Repositories

| ID        | Layer | Status | Title                              | Notes                                                                                                                   |
| --------- | ----- | ------ | ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| HC-FE-08  | `FE`  | `[x]`  | `HoneyBatchModel` (Dart)           | `amountGrams` (int) is canonical; `amountKg` getter for display only. No blockchain fields — those live on HC-FE-08b.  |
| HC-FE-08b | `FE`  | `[x]`  | `HoneyBatchCertificationModel` (Dart) | New. `CertificationStatus` enum mirrors the Go/DB lifecycle 1:1 via explicit `fromJson`/`toJson`. `HoneyBatchModel.certification` is nullable — `null` means not yet certified, not an enum value. |
| HC-FE-09  | `FE`  | `[x]`  | `ProcessingMethodEnum` (Dart)      | raw, filtered, pasteurized + display labels                                                                             |
| HC-FE-10  | `FE`  | `[x]`  | `HoneyBatchRepository` (Dart)      | `createBatch` takes `requestCertification` (default false). Adds `requestCertification(id)` (certify-now or retry, same call) and `verifyByToken(token)` (public, no auth header) alongside standard CRUD. |

---

## 9.8 Frontend — State Management

| ID       | Layer | Status | Title              | Notes                                                                                                    |
| -------- | ----- | ------ | -------------------- | ------------------------------------------------------------------------------------------------------------ |
| HC-FE-19 | `FE`  | `[x]`  | Honey BLoC/Cubit    | `create()` takes `requestCertification` (default false); resulting batch's `certification` is `queued` or stays `null`. Adds `requestCertification(id)` method (certify-now or retry). |

---

## 9.9 Frontend — Core Screens

| ID       | Layer | Status | Title                                     | Notes                                                                                                                             |
| -------- | ----- | ------ | -------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| HC-FE-01 | `FE`  | `[x]`  | Honey batches home screen                 | List badges reflect full lifecycle, not just pending/confirmed/failed. New top-level `AppSection.honeyBatches` drawer entry (not hive/apiary-scoped — mirrors the repository, which lists all of the user's batches). |
| HC-FE-02 | `FE`  | `[x]`  | Create honey batch screen                 | Amount entered in kg, converted to whole grams (`(kg*1000).round()`) before hitting the API. Adds a "Certify on the blockchain" toggle, off by default. Success message depends on the toggle. PDF picked via `file_picker` (added as a new dependency; `withData: true` for web). Later made dual-purpose (create/edit) via an optional `existingBatch` param — the card's "Edit" action pushes this screen pre-filled instead of a honey-type-only dialog, letting the owner edit gathering date, amount, processing method, and honey type together. The certify toggle is hidden in edit mode; the PDF picker is now shown in edit mode too (previously hidden), letting the owner optionally replace the batch's PDF, with the current filename shown as a placeholder until a replacement is picked. **Later deviation:** the "Certify on the blockchain" toggle was removed entirely from create mode — a batch is now always created uncertified, and certification is requested afterward via the card's Certify/Retry button (HC-BE-24c). The PDF picker's clear/X button now also works in edit mode: if an existing PDF hasn't been replaced with a new file, clearing marks it for removal and passes `removePdf` through to `updateBatch` (HC-BE-22). |
| HC-FE-03 | `FE`  | `[x]`  | Honey batch detail screen                 | Owner view (numeric id). Shows "Certify" action when `certification` is `null`, "Retry" when `failed`/`reverted`. QR/share buttons shown once `confirmed`, wired to TODO stubs pending HC-FE-05. Later merged into `HoneyBatchCard` on the list screen instead of staying a standalone screen — there is no separate detail screen anymore, but the underlying requirement (certify/retry, view status) is still met, just via the card's overflow menu and action area (plan text was stale on this). |
| HC-FE-04 | `FE`  | `[ ]`  | Honey batch verification screen           | Loaded via `verifyByToken` — public, no auth. Refresh re-fetches DB state, not a live RPC call.                                   |
| HC-FE-05 | `FE`  | `[x]`  | QR code display screen                    | No separate screen — matches HC-FE-03's precedent of living on the card instead. `showQrPreviewDialog` (new `qr_preview_screen.dart`) shows the inline `/verify/{token}/qr-code` PNG via `Image.network` in a dialog, with a download action wired to `launchQrDownload` (`url_launcher`, opens `/verify/{token}/qr-code/download` in a new tab so the browser's `Content-Disposition` triggers the save). Wired into the confirmed-batch card's existing "View QR"/"Download QR" buttons (previously TODO stubs). New `HoneyBatchRepository.qrCodeImageUrl`/`qrCodeDownloadUrl` build these URLs from `ApiClient.baseUrl` + the batch's `verificationToken`. The download filename is derived server-side from batch metadata (see HC-BE-21's updated notes and `docs/api.md`) rather than a fixed name. |
| HC-FE-06 | `FE`  | `[ ]`  | QR code scanner screen                    | Extracts the **verification token** (UUID string) from the scanned URL, not a numeric id.                                        |
| HC-FE-18 | `FE`  | `[ ]`  | Add Honey Batches section to hive detail  | Plan text is stale: batches no longer have an `apiary_id`/hive link at all (see HC-DB-01), so "batches from this hive" no longer applies — batches are only scoped to their owning user. |

---

## 9.10 Frontend — Utils & Widgets

| ID       | Layer | Status | Title                                        | Notes                                                                                                                    |
| -------- | ----- | ------ | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| HC-FE-11 | `FE`  | `[x]`  | QR code generation                           | `HoneyBatchRepository.qrCodeImageUrl(token)`/`qrCodeDownloadUrl(token)` → `/verify/{token}/qr-code[/download]`, built from `ApiClient.baseUrl`. Landed as part of HC-FE-05. |
| HC-FE-12 | `FE`  | `[ ]`  | QR code scanner util                         | `extractVerificationTokenFromQRData(qrData)` — parses token, not batch id                                                |
| HC-FE-13 | `FE`  | `[x]`  | PDF preview / download                       | `showPdfPreviewDialog` (new `pdf_preview_screen.dart`) renders bytes via the `printing` package's `PdfPreview` in a dialog, with a `PdfShareAction` for download. Wired into both the create/edit screen (picked file or fetched via new `HoneyBatchRepository.getPdfBytes(id)`) and the batch card (existing PDF). Size-validated client-side against `maxPdfBytes` (10MB, mirrors backend) before upload. |
| HC-FE-14 | `FE`  | `[ ]`  | PDF upload UI                                | Unchanged from original plan.                                                                                             |
| HC-FE-15 | `FE`  | `[ ]`  | Certification status indicator (badge)       | Driven by the 7-state `CertificationStatus` enum, plus a distinct "not certified yet" rendering when `certification` is `null` — not a flattened 3-state badge. |
| HC-FE-16 | `FE`  | `[ ]`  | Verification details modal                   | Shows full certification history (multiple rows) if a batch has more than one, most recent first.                        |
| HC-FE-17 | `FE`  | `[ ]`  | Hash comparison display                      | Unchanged from original plan.                                                                                             |

---

## 9.11 Polish & Edge Cases

| ID       | Layer | Status | Title                   | Notes                                                                                                                                    |
| -------- | ----- | ------ | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| HC-10-01 | `BE`  | `[x]`  | PDF file validation     | MIME type + max size (10MB), enforced in `HoneyBatchService.validateCreateBatchRequest` (HC-BE-17). ClamAV scan is **(optional)**, skipped. |
| HC-10-02 | `BE`  | `[x]`  | Blockchain retry logic  | Done via the worker (HC-BE-15b) — exponential backoff via `blockchain_jobs.next_retry_at`, 3 attempts then terminal `failed`.             |
| HC-10-03 | `BE`  | `[x]`  | Storage for PDFs        | Local FS, written by `HoneyBatchService.CreateBatch` under `PDF_STORAGE_PATH` (HC-BE-17) — matches the existing photo-storage pattern. S3 + signed URLs is **(optional)**. |
| HC-10-04 | `BE`  | `[ ]`  | Gas fee management      | `gas_used` persisted per-certification row (auditable per batch — useful for the thesis write-up, cheap to keep). Gas relay service + price-spike alerting is **(optional)**, unnecessary on free testnet gas. |
| HC-10-05 | `FE`  | `[ ]`  | Offline handling        | Simplified vs. original plan — no local blockchain-write queue needed, since the app never triggers chain writes directly.               |
| HC-10-06 | `FE`  | `[ ]`  | Loading states          | Distinct UI per lifecycle state; "Certify" button when `certification` is `null`, "Retry" on `failed`/`reverted` (both call HC-BE-24c).   |
| HC-10-07 | `FE`  | `[ ]`  | Error handling          | Error copy mapped from lifecycle status (`null` = neutral, not an error; queued/submitting/submitted/pending_confirmation = "in progress", not an error). |
| HC-10-08 | `FE`  | `[ ]`  | Localization            | l10n keys for all 7 lifecycle states plus a separate non-enum key for the null-certification case, processing methods, verification text. |
| HC-10-09 | `FE`  | `[ ]`  | Empty states            | Unchanged from original plan.                                                                                                             |
| HC-10-10 | `BE`  | `[ ]`  | Database indexing       | Folded into migrations HC-DB-01–04 directly rather than a bolt-on later migration — see those rows for the actual index list.            |
