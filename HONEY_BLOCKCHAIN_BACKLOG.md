# Honey Certification & Blockchain — Backlog

> Living document for Epic 9, broken out from the main [BACKLOG.md](BACKLOG.md) because the async/jobs-queue redesign roughly doubled its task count. Task breakdown mirrors [HONEY_BLOCKCHAIN_PLAN.md](HONEY_BLOCKCHAIN_PLAN.md) — see that file for full rationale and field-level detail behind each row.
>
> Treat each item like a Jira ticket — update status as work progresses.
>
> **Architecture in one line:** `CreateBatch` persists a batch + enqueues a `blockchain_jobs` row and returns immediately; a background `BlockchainWorker` owns all Polygon RPC interaction, writes results to an append-only `honey_batch_certifications` history, and drives a 7-state lifecycle (queued → submitting → submitted → pending_confirmation → confirmed / failed / reverted). Public verification uses an unguessable `verification_token`, never the numeric batch id. Honey amount is stored as integer grams, never `float64`.
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
| HC-DB-01 | `DB`  | `[x]`  | Create `honey_batches` table              | `backend/migrations/029_...`. No blockchain fields — those live in HC-DB-04. Includes `verification_token` (UUID, UNIQUE) and `amount_grams` (BIGINT, not float). `deleted_at` for soft delete. |
| HC-DB-02 | `DB`  | `[x]`  | Create `honey_batch_qr_codes` table       | `backend/migrations/030_...`. `qr_code_data` encodes `/verify/{verification_token}`, never the numeric id.                                                                                     |
| HC-DB-03 | `DB`  | `[x]`  | Create `blockchain_jobs` table            | `backend/migrations/031_...`. Durable queue: status, attempt_count, next_retry_at, last_error. Index (status, next_retry_at) for the worker's claim query.                                     |
| HC-DB-04 | `DB`  | `[x]`  | Create `honey_batch_certifications` table | `backend/migrations/032_...`. Append-only per-batch history (chain_id, contract_address, tx_hash, block_number, status, gas_used). Partial UNIQUE (batch_id) WHERE status is "live" — idempotency guard. |

---

## 9.2 Backend — Models & Persistence

| ID        | Layer | Status | Title                                             | Notes                                                                                                                       |
| --------- | ----- | ------ | --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| HC-BE-07  | `BE`  | `[x]`  | Model: `HoneyBatch` struct                        | Mirrors DB schema. `AmountGrams int64`, `VerificationToken string`. No blockchain fields on this struct.                    |
| HC-BE-08  | `BE`  | `[x]`  | Model: `ProcessingMethod` enum                    | raw, filtered, pasteurized + `IsValidProcessingMethod`                                                                      |
| HC-BE-07b | `BE`  | `[x]`  | Model: `HoneyBatchCertification` + status lifecycle | `CertificationStatus` type: queued/submitting/submitted/pending_confirmation/confirmed/failed/reverted. `IsTerminal()`/`IsLive()` helpers. Single source of truth mirrored in DB CHECK, API JSON, Dart enum (HC-FE-08b). |
| HC-BE-07c | `BE`  | `[x]`  | Model: `BlockchainJob` struct                     | Reuses `CertificationStatus` for its own status field.                                                                      |
| HC-BE-09  | `BE`  | `[x]`  | Repository: `HoneyBatchRepository` — Create        | Runs in a transaction together with the initial `blockchain_jobs` insert (HC-BE-13) — a batch is never persisted without a job. |
| HC-BE-10  | `BE`  | `[x]`  | Repository: Get by ID / by verification token      | `GetByID` (owner-scoped), `GetByVerificationToken` (public path)                                                            |
| HC-BE-11  | `BE`  | `[x]`  | Repository: List batches by user/apiary            | `ListByUserID`, `ListByApiaryID`, paginated                                                                                 |
| HC-BE-12  | `BE`  | `[x]`  | Repository: Update notes / soft delete             | `UpdateNotes`, `SoftDelete` — no status/blockchain mutation methods here anymore. `UpdateNotes` only touches `honey_type`; there's no `notes` column on `honey_batches` (plan text was stale on this). |
| HC-BE-12b | `BE`  | `[x]`  | Repository: `HoneyBatchCertificationRepository` + `BlockchainJobRepository` | Certification repo: Create, GetLatestByBatchID, ListByBatchID, UpdateStatus. Job repo: Create, `ClaimNext` (SELECT...FOR UPDATE SKIP LOCKED, atomically flips claimed job to `submitting` in the same tx), MarkSubmitting/Submitted/Failed, ListPendingConfirmation. |

---

## 9.3 Backend — Blockchain Integration

| ID        | Layer | Status | Title                                | Notes                                                                                                                                            |
| --------- | ----- | ------ | -------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| HC-BE-01  | `BE`  | `[x]`  | Blockchain config                    | RPC URL, contract address, private key, chain ID (default 80002 Amoy), plus `JobPollInterval` (5s), `ConfirmationPollInterval` (30s), `RequiredConfirmations` (12). |
| HC-BE-02  | `BE`  | `[x]`  | Smart contract (Solidity)            | `certify()` **reverts if batchID already certified** — contract-level idempotency. Event `CertificationCreated`. Owner-only caller.               |
| HC-BE-03  | `BE`  | `[ ]`  | Deploy contract to Polygon           | Amoy testnet (80002) — the only target needed. Mainnet (137) is **(optional)**, out of scope. Store ABI at `backend/internal/blockchain/contracts/HoneyCertification.abi`.                     |
| HC-BE-04  | `BE`  | `[ ]`  | Blockchain writer                    | `CertifyBatch(...)` — called **only** by the worker (HC-BE-15b), never from the HTTP path. Returns tx hash immediately, doesn't wait for confirmation. |
| HC-BE-05  | `BE`  | `[ ]`  | Blockchain reader                    | `VerifyBatch(...)` + `GetTransactionStatus(txHash)` (confirmed/reverted/blockNumber/gasUsed) — used by the confirmation-polling loop (HC-BE-15c).  |
| HC-BE-06  | `BE`  | `[ ]`  | Hash utilities — PDF                 | `SHA256File(filePath)` for PDF hashing.                                                                                                            |
| HC-BE-06b | `BE`  | `[ ]`  | Canonical metadata hashing           | `CanonicalMetadataHash(batch)` — exact 7-field order, `\x1f`-joined, UTF-8, SHA256. No `fmt.Sprintf`/struct-JSON. See plan §HC-BE-06b for the full field spec — this is the language-agnostic contract, not "read the Go source." |

---

## 9.4 Backend — Business Logic & Worker

| ID        | Layer | Status | Title                                    | Notes                                                                                                                                                          |
| --------- | ----- | ------ | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| HC-BE-13  | `BE`  | `[ ]`  | Service: Create honey batch              | Validate → hash PDF → generate verification token → compute metadata hash → **in one transaction**: insert batch + insert `blockchain_jobs` row (`status=queued`) → return. No blockchain call here. |
| HC-BE-14  | `BE`  | `[ ]`  | Service: Get batch + verify              | `GetBatchWithVerification(token)` — reads latest certification from DB (kept fresh by the worker), not a live RPC call per request.                          |
| HC-BE-15b | `BE`  | `[ ]`  | Worker: process certify jobs             | `ProcessNextJob` — claim job → idempotency check (skip if a "live" certification already exists) → submit tx → update certification/job status. Retry via exponential backoff (1s/2s/4s/8s, capped) on failure. |
| HC-BE-15c | `BE`  | `[ ]`  | Worker: poll for confirmations           | `PollSubmittedJobs` — for submitted/pending_confirmation jobs, check tx status; move to confirmed/reverted once mined & enough confirmations, or leave pending. |
| HC-BE-25  | `BE`  | `[ ]`  | Idempotency guarantees (docs + tests)    | Three layers: contract revert-on-duplicate, worker's pre-submit live-certification check, DB partial unique constraint. Add a test that kills the worker mid-broadcast and asserts exactly one live certification results. |
| HC-BE-16  | `BE`  | `[ ]`  | Service: Generate QR code data           | URL is `{appURL}/verify/{verificationToken}` — never the numeric id.                                                                                          |
| HC-BE-16b | `BE`  | `[ ]`  | Verification token generation            | UUID v4 via `crypto/rand` (`google/uuid`), generated once at batch creation, immutable.                                                                        |

---

## 9.5 Backend — API Handlers

| ID        | Layer | Status | Title                                                    | Notes                                                                                                                    |
| --------- | ----- | ------ | ----------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| HC-BE-17  | `BE`  | `[ ]`  | Handler: POST /api/v1/honey-batches                     | Auth required. Returns batch + `certification_status: "queued"` — no tx hash yet (async). Blockchain failure can never cause a 500 here. |
| HC-BE-18  | `BE`  | `[ ]`  | Handler: GET /api/v1/honey-batches/{id}                 | **Auth + ownership required** (no longer public — public access moved to HC-BE-19).                                     |
| HC-BE-19  | `BE`  | `[ ]`  | Handler: GET /api/v1/verify/{token}                      | Renamed from `/honey-batches/{id}/verify`. Public, token-scoped. Returns full lifecycle status, not a boolean.           |
| HC-BE-20  | `BE`  | `[ ]`  | Handler: GET /api/v1/honey-batches (list)                | Auth required. Each item includes its latest certification status for list-view badges.                                 |
| HC-BE-21  | `BE`  | `[ ]`  | Handler: GET /api/v1/verify/{token}/qr-code              | Moved under the token-scoped path. Public, cached (immutable).                                                          |
| HC-BE-22  | `BE`  | `[ ]`  | Handler: PATCH /api/v1/honey-batches/{id}                | Auth + ownership. notes/honey_type only.                                                                                 |
| HC-BE-23  | `BE`  | `[ ]`  | Handler: DELETE /api/v1/honey-batches/{id}               | Auth + ownership. Soft delete; on-chain record is untouched/immutable.                                                   |
| HC-BE-24  | `BE`  | `[ ]`  | Handler: GET /api/v1/honey-batches/{id}/pdf              | Auth + ownership — owner-scoped PDF access.                                                                              |
| HC-BE-24b | `BE`  | `[ ]`  | Handler: GET /api/v1/verify/{token}/pdf                  | New — public, token-scoped PDF access for QR scanners.                                                                   |
| HC-BE-24c | `BE`  | `[ ]`  | Handler: POST /api/v1/honey-batches/{id}/retry-certification | New — auth + ownership. Re-enqueues a `blockchain_jobs` row for a `failed` batch. Backs the FE "Retry" button (HC-10-06). |

---

## 9.6 Backend — Integration & Wiring

| ID       | Layer | Status | Title                            | Notes                                                                                                                          |
| -------- | ----- | ------ | ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| HC-BE-W1 | `BE`  | `[ ]`  | Wire blockchain components in main.go | Service no longer holds a direct `blockchain.Writer` dependency — only the worker does. Register `/verify/{token}` public route group. |
| HC-BE-26 | `BE`  | `[ ]`  | Background worker runner         | `BlockchainWorker.Run(ctx)` — two ticker loops (job processing @5s, confirmation polling @30s), graceful shutdown on context cancel. A `submitting`-stuck-past-timeout sweep handles crash-mid-step recovery. |

---

## 9.7 Frontend — Models & Repositories

| ID        | Layer | Status | Title                              | Notes                                                                                                                   |
| --------- | ----- | ------ | ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| HC-FE-08  | `FE`  | `[ ]`  | `HoneyBatchModel` (Dart)           | `amountGrams` (int) is canonical; `amountKg` getter for display only. No blockchain fields — those live on HC-FE-08b.  |
| HC-FE-08b | `FE`  | `[ ]`  | `HoneyBatchCertificationModel` (Dart) | New. `CertificationStatus` enum mirrors the Go/DB lifecycle 1:1 via explicit `fromJson`/`toJson`.                      |
| HC-FE-09  | `FE`  | `[ ]`  | `ProcessingMethodEnum` (Dart)      | raw, filtered, pasteurized + display labels                                                                             |
| HC-FE-10  | `FE`  | `[ ]`  | `HoneyBatchRepository` (Dart)      | Adds `retryCertification(id)` and `verifyByToken(token)` (public, no auth header) alongside standard CRUD.              |

---

## 9.8 Frontend — State Management

| ID       | Layer | Status | Title              | Notes                                                                                                    |
| -------- | ----- | ------ | -------------------- | ------------------------------------------------------------------------------------------------------------ |
| HC-FE-19 | `FE`  | `[ ]`  | Honey BLoC/Cubit    | `create()` result shows `certification_status: queued` immediately. Adds `retryCertification(id)` method. |

---

## 9.9 Frontend — Core Screens

| ID       | Layer | Status | Title                                     | Notes                                                                                                                             |
| -------- | ----- | ------ | -------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| HC-FE-01 | `FE`  | `[ ]`  | Honey batches home screen                 | List badges reflect full lifecycle, not just pending/confirmed/failed.                                                            |
| HC-FE-02 | `FE`  | `[ ]`  | Create honey batch screen                 | Amount entered in kg, converted to whole grams (`(kg*1000).round()`) before hitting the API. Success message clarifies async certification. |
| HC-FE-03 | `FE`  | `[ ]`  | Honey batch detail screen                 | Owner view (numeric id). Shows "Retry" action when latest certification is `failed`.                                             |
| HC-FE-04 | `FE`  | `[ ]`  | Honey batch verification screen           | Loaded via `verifyByToken` — public, no auth. Refresh re-fetches DB state, not a live RPC call.                                   |
| HC-FE-05 | `FE`  | `[ ]`  | QR code display screen                    | Unchanged from original plan.                                                                                                     |
| HC-FE-06 | `FE`  | `[ ]`  | QR code scanner screen                    | Extracts the **verification token** (UUID string) from the scanned URL, not a numeric id.                                        |
| HC-FE-18 | `FE`  | `[ ]`  | Add Honey Batches section to hive detail  | Unchanged from original plan.                                                                                                     |

---

## 9.10 Frontend — Utils & Widgets

| ID       | Layer | Status | Title                                        | Notes                                                                                                                    |
| -------- | ----- | ------ | ----------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| HC-FE-11 | `FE`  | `[ ]`  | QR code generation                           | `generateQRImageUrl(token)` → `/verify/{token}/qr-code`                                                                  |
| HC-FE-12 | `FE`  | `[ ]`  | QR code scanner util                         | `extractVerificationTokenFromQRData(qrData)` — parses token, not batch id                                                |
| HC-FE-13 | `FE`  | `[ ]`  | PDF preview / download                       | Unchanged from original plan.                                                                                             |
| HC-FE-14 | `FE`  | `[ ]`  | PDF upload UI                                | Unchanged from original plan.                                                                                             |
| HC-FE-15 | `FE`  | `[ ]`  | Certification status indicator (badge)       | Driven by the 7-state `CertificationStatus` enum — distinct color/icon per state, not a flattened 3-state badge.          |
| HC-FE-16 | `FE`  | `[ ]`  | Verification details modal                   | Shows full certification history (multiple rows) if a batch has more than one, most recent first.                        |
| HC-FE-17 | `FE`  | `[ ]`  | Hash comparison display                      | Unchanged from original plan.                                                                                             |

---

## 9.11 Polish & Edge Cases

| ID       | Layer | Status | Title                   | Notes                                                                                                                                    |
| -------- | ----- | ------ | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| HC-10-01 | `BE`  | `[ ]`  | PDF file validation     | MIME type + max size (10MB). ClamAV scan is **(optional)**, skip.                                                                          |
| HC-10-02 | `BE`  | `[ ]`  | Blockchain retry logic  | Moved to the worker (HC-BE-15b) — exponential backoff via `blockchain_jobs.next_retry_at`, 3 attempts then terminal `failed`.             |
| HC-10-03 | `BE`  | `[ ]`  | Storage for PDFs        | Local FS is enough for the thesis's testing env (matches existing photo-storage pattern). S3 + signed URLs is **(optional)**.             |
| HC-10-04 | `BE`  | `[ ]`  | Gas fee management      | `gas_used` persisted per-certification row (auditable per batch — useful for the thesis write-up, cheap to keep). Gas relay service + price-spike alerting is **(optional)**, unnecessary on free testnet gas. |
| HC-10-05 | `FE`  | `[ ]`  | Offline handling        | Simplified vs. original plan — no local blockchain-write queue needed, since the app never triggers chain writes directly.               |
| HC-10-06 | `FE`  | `[ ]`  | Loading states          | Distinct UI per lifecycle state; "Retry" button on `failed` batches (calls HC-BE-24c).                                                    |
| HC-10-07 | `FE`  | `[ ]`  | Error handling          | Error copy mapped from lifecycle status (queued/submitting/submitted/pending_confirmation = "in progress", not an error).                |
| HC-10-08 | `FE`  | `[ ]`  | Localization            | l10n keys for all 7 lifecycle states (not just pending/confirmed/failed), processing methods, verification text.                         |
| HC-10-09 | `FE`  | `[ ]`  | Empty states            | Unchanged from original plan.                                                                                                             |
| HC-10-10 | `BE`  | `[ ]`  | Database indexing       | Folded into migrations HC-DB-01–04 directly rather than a bolt-on later migration — see those rows for the actual index list.            |
