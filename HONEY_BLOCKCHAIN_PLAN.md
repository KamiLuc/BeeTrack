# Honey Certification & Blockchain Implementation Plan

## Context

Epic 9 ("Honey Certification & Blockchain") from BACKLOG.md aims to create an immutable honey batch certification system using the Polygon blockchain. Each honey batch gets a QR code that verifies authenticity by checking a blockchain-stored hash against the lab PDF. The feature spans database schema, Go backend API, smart contracts, and Flutter frontend screens.

**Blockchain Strategy:** Store minimal data on-chain (hash, metadata hash, timestamp) for cost efficiency. Lab PDF is stored off-chain; blockchain only stores its SHA256 hash.

**Architecture Strategy (revised):** Blockchain certification is fully asynchronous. The HTTP API never blocks on Polygon RPC calls — creating a batch persists the batch and enqueues a durable job; a dedicated background worker owns all chain interaction, retries, and status transitions. Certification results are recorded as an append-only history (`honey_batch_certifications`) rather than mutable fields on the batch itself, and batches are exposed publicly through an unguessable verification token rather than their sequential database ID.

**Thesis Scope:** This feature is being built for a final computer science thesis, not for a production deployment. The target environment is **Polygon Amoy testnet only** — there is no mainnet migration, no real-money gas spend, and no external users depending on uptime. The architectural rigor above (async jobs, idempotency, append-only certification history, deterministic hashing) is kept because it's the interesting engineering content of the thesis and needs to demonstrably work correctly under test, not because the deployment needs to survive production traffic. Anything below tagged **(production-hardening — optional for thesis)** can be skipped, stubbed, or left as documented future work without weakening the thesis; it's noted for completeness rather than as a requirement.

---

## Implementation Phases

### Phase 1: Database Foundation (4 tasks)
**Goals:** Create schema for honey batches, verification, and blockchain job/certification tracking.

1. **HC-DB-01: Create `honey_batches` table migration**
   - File: `backend/migrations/029_create_honey_batches.sql`
   - Columns: id, user_id (created_by), apiary_id, verification_token (UUID, UNIQUE, NOT NULL — public identifier, see HC-BE-06b), gathering_date, amount_grams (BIGINT — precise integer grams, see "Amount Representation" below), processing_method (VARCHAR + CHECK), honey_type (TEXT), lab_pdf_url, pdf_file_hash (CHAR(64), hex-encoded SHA256), metadata_hash (CHAR(64), hex-encoded SHA256), deleted_at (TIMESTAMPTZ, nullable — soft delete), created_at, updated_at
   - **No blockchain status/tx fields live here** — certification state is tracked entirely in `honey_batch_certifications` (HC-DB-04), keeping the batch row immutable aside from user-editable fields (notes, honey_type).
   - Indexes: (user_id, created_at DESC), (apiary_id, created_at DESC), UNIQUE (verification_token)
   - FK constraints: user_id → users(id), apiary_id → apiaries(id) ON DELETE CASCADE
   - Timestamps: TIMESTAMPTZ with DEFAULT NOW()

2. **HC-DB-02: Create `honey_batch_qr_codes` table migration**
   - File: `backend/migrations/030_create_honey_batch_qr_codes.sql`
   - Columns: id, batch_id, qr_code_data (VARCHAR — encodes `{appURL}/verify/{verification_token}`, never the numeric id), created_at
   - FK: batch_id → honey_batches(id) ON DELETE CASCADE
   - Index: (batch_id)

3. **HC-DB-03: Create `blockchain_jobs` table migration**
   - File: `backend/migrations/031_create_blockchain_jobs.sql`
   - Columns:
     - id BIGSERIAL PRIMARY KEY
     - batch_id BIGINT NOT NULL REFERENCES honey_batches(id) ON DELETE CASCADE
     - job_type TEXT NOT NULL (e.g. `certify`; extensible for future job types)
     - status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','submitting','submitted','pending_confirmation','confirmed','failed','reverted'))
     - attempt_count INT NOT NULL DEFAULT 0
     - next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW() — worker polls `WHERE status IN ('queued','failed') AND next_retry_at <= NOW()`
     - last_error TEXT
     - certification_id BIGINT REFERENCES honey_batch_certifications(id) — set once a certification row is created for this job's attempt
     - created_at, updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   - Indexes: (status, next_retry_at) — the worker's primary claim query; (batch_id)
   - This table is the durable queue described in "Async Certification Architecture" below: it survives server restarts, drives retries, and gives us observability into every certification attempt.

4. **HC-DB-04: Create `honey_batch_certifications` table migration**
   - File: `backend/migrations/032_create_honey_batch_certifications.sql`
   - Columns:
     - id BIGSERIAL PRIMARY KEY
     - batch_id BIGINT NOT NULL REFERENCES honey_batches(id) ON DELETE CASCADE
     - chain_id INT NOT NULL (e.g. 80002 for Amoy, 137 for mainnet)
     - contract_address CHAR(42) NOT NULL
     - transaction_hash CHAR(66) — null until submitted
     - block_number BIGINT — null until mined
     - status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','submitting','submitted','pending_confirmation','confirmed','failed','reverted'))
     - gas_used BIGINT
     - confirmation_timestamp TIMESTAMPTZ — set when status transitions to 'confirmed'
     - created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   - Indexes: (batch_id, created_at DESC) — a batch may accumulate multiple rows across retries, contract migrations, or additional chains; (batch_id) partial UNIQUE WHERE status IN ('submitted','pending_confirmation','confirmed') to prevent two "live" certifications existing for the same batch at once (see "Idempotency" below).
   - A batch's *current* certification state for display purposes is simply the most recent row by `created_at`; older rows are kept for audit history.

**Amount Representation:** `amount_grams` is a `BIGINT` (Go `int64`), not `float64`. Honey weight is entered/displayed in kilograms in the UI but stored and transmitted as whole grams internally, eliminating floating-point rounding entirely and keeping the metadata hash (HC-BE-06b) deterministic across languages/runtimes. Conversion (`kg ↔ g`) happens only at the UI edge (Dart) and API request/response boundary (Go), never inside hashing or persistence logic.

### Phase 2: Backend Models & Persistence (6 tasks)
**Goals:** Define Go domain models and database layer, including the jobs queue and certification history.

5. **HC-BE-07: Create HoneyBatch model**
   - File: `backend/internal/model/honey_batch.go`
   - Struct fields match DB schema: ID, UserID, ApiaryID, VerificationToken (string, UUID), GatheringDate, AmountGrams (int64), ProcessingMethod, HoneyType, LabPDFURL, PDFFileHash, MetadataHash, DeletedAt (*time.Time), CreatedAt, UpdatedAt
   - No blockchain fields on this struct — see HoneyBatchCertification (HC-BE-07b)
   - ProcessingMethod as string type (raw, filtered, pasteurized)

6. **HC-BE-08: Create ProcessingMethod enum**
   - File: `backend/internal/model/honey_batch.go` (same file as model)
   - Constants: ProcessingMethodRaw, ProcessingMethodFiltered, ProcessingMethodPasteurized
   - Validation function: IsValidProcessingMethod(method string) bool

7. **HC-BE-07b: Create HoneyBatchCertification model + status lifecycle**
   - File: `backend/internal/model/honey_batch_certification.go`
   - Struct fields: ID, BatchID, ChainID (int), ContractAddress, TransactionHash (*string), BlockNumber (*int64), Status (CertificationStatus), GasUsed (*int64), ConfirmationTimestamp (*time.Time), CreatedAt
   - Type `CertificationStatus string` with constants matching the DB CHECK constraint:
     - `CertificationStatusQueued` — job created, not yet picked up by worker
     - `CertificationStatusSubmitting` — worker is building/signing the transaction
     - `CertificationStatusSubmitted` — transaction broadcast, tx hash known, awaiting inclusion
     - `CertificationStatusPendingConfirmation` — included in a block, waiting for enough confirmations
     - `CertificationStatusConfirmed` — terminal success state
     - `CertificationStatusFailed` — terminal failure (RPC error, reverted before inclusion, retries exhausted)
     - `CertificationStatusReverted` — mined but the transaction reverted on-chain (terminal)
   - Helper: `IsTerminal() bool` (confirmed/failed/reverted) and `IsLive() bool` (submitted/pending_confirmation/confirmed) used by the idempotency check in HC-BE-25.
   - This same `CertificationStatus` type (and its string values) is the single source of truth reflected in the DB CHECK constraints, `blockchain_jobs.status`, API JSON, and — mirrored 1:1 — the Dart enum in HC-FE-09b, so the lifecycle never drifts between layers. A batch that was never certified simply has **no** `HoneyBatchCertification` at all (a nil pointer in Go, `null` in JSON, a nullable field in Dart) — this is not a value of the enum, it's the absence of one.

8. **HC-BE-07c: Create BlockchainJob model**
   - File: `backend/internal/model/blockchain_job.go`
   - Struct fields: ID, BatchID, JobType, Status (CertificationStatus, reusing the type from HC-BE-07b), AttemptCount (int), NextRetryAt (time.Time), LastError (*string), CertificationID (*int64), CreatedAt, UpdatedAt

9. **HC-BE-09 to HC-BE-12: Create HoneyBatchRepository**
   - File: `backend/internal/repository/honey_batch.go`
   - Methods: Create(ctx, batch) error, GetByID(ctx, id) (*HoneyBatch, error), GetByVerificationToken(ctx, token) (*HoneyBatch, error), ListByUserID(ctx, userID, limit, offset), ListByApiaryID(ctx, apiaryID, limit, offset), UpdateNotes(ctx, id, notes, honeyType) error, SoftDelete(ctx, id) error
   - Use GORM pattern (receiver *HoneyBatchRepository(db *gorm.DB))
   - Follow error handling: return nil, nil for not found; return error otherwise
   - `Create` runs in a transaction together with the initial `blockchain_jobs` row insert (see HC-BE-13), so a batch is never persisted without a corresponding certification job.

10. **HC-BE-12b: Create HoneyBatchCertificationRepository + BlockchainJobRepository**
    - File: `backend/internal/repository/honey_batch_certification.go` and `backend/internal/repository/blockchain_job.go`
    - `HoneyBatchCertificationRepository`: Create(ctx, cert) error, GetLatestByBatchID(ctx, batchID) (*HoneyBatchCertification, error), ListByBatchID(ctx, batchID) ([]*HoneyBatchCertification, error) — full audit history, UpdateStatus(ctx, id, status, fields...) error
    - `BlockchainJobRepository`: Create(ctx, job) error, ClaimNext(ctx) (*BlockchainJob, error) — atomically selects and locks one runnable job (`SELECT ... FOR UPDATE SKIP LOCKED WHERE status IN ('queued','failed') AND next_retry_at <= NOW() ORDER BY created_at LIMIT 1`) so multiple worker instances can run safely, MarkSubmitting/MarkSubmitted/MarkFailed(ctx, id, err, nextRetryAt) error, ListPendingConfirmation(ctx) ([]*BlockchainJob, error) — jobs whose certification is `submitted`/`pending_confirmation`, for the confirmation-polling loop

### Phase 3: Blockchain Integration (7 tasks)
**Goals:** Smart contract deployment and on-chain interaction layer.

11. **HC-BE-01: Blockchain configuration**
    - File: `backend/internal/config/blockchain_config.go`
    - Fields: PolygonRPCURL, ContractAddress, PrivateKey, ChainID (for testnet/mainnet detection), JobPollInterval (default 5s — how often the worker claims new jobs), ConfirmationPollInterval (default 30s — how often it checks pending transactions), RequiredConfirmations (default 12)
    - Validation: Ensure private key is 64 hex chars, RPC URL is valid, contract address is 42 chars (0x...)
    - Environment variables: POLYGON_RPC_URL, CONTRACT_ADDRESS, BLOCKCHAIN_PRIVATE_KEY, CHAIN_ID (default 80002 for Amoy testnet)

12. **HC-BE-02: Smart contract (Solidity)**
    - File: `backend/contracts/HoneyCertification.sol` (or upload to external repo)
    - Simple registry: stores (batchID uint256, pdfHash bytes32, metadataHash bytes32, timestamp uint256, ownerAddress address)
    - Function: `certify(batchID uint256, pdfHash bytes32, metadataHash bytes32) returns (tx hash)`
    - **Idempotency at the contract level (see improvement #6):** `certify()` `require`s that no record exists yet for `batchID` (i.e. reverts if `certifications[batchID].timestamp != 0`). This makes accidental duplicate submission from the backend a safe no-op-that-reverts rather than a silent double-certification, and gives the worker a deterministic way to detect "already certified" via a revert reason.
    - Event: CertificationCreated(indexed batchID, pdfHash, metadataHash, timestamp, ownerAddress)
    - Read function: getCertification(batchID) returns stored data (if exists)
    - Owner validation: Only minter address (BeeTrack backend) can call certify()

13. **HC-BE-03: Deploy contract to Polygon**
    - Deploy to Amoy testnet (chain ID 80002) — this is the only deployment target needed for the thesis. Mainnet (137) deployment is **(production-hardening — optional for thesis)**; the code stays chain-agnostic (`chain_id` is a config value, never hardcoded) so it's possible later, but is not part of this scope.
    - Use Remix or hardhat for deployment; store contract address in config
    - Document: Store ABI in `backend/internal/blockchain/contracts/HoneyCertification.abi` for Go bindings

14. **HC-BE-04: Blockchain writer (on-chain transaction builder)**
    - File: `backend/internal/blockchain/writer.go`
    - Function: `CertifyBatch(ctx context.Context, batchID int64, pdfHash, metadataHash [32]byte) (txHash string, err error)`
    - Called exclusively by the background worker (HC-BE-15b), never from the HTTP request path.
    - Steps:
      1. Connect to Polygon RPC (via ethclient.Dial)
      2. Get nonce from pending transactions
      3. Build transaction call to contract.certify()
      4. Sign with private key (ECDSA)
      5. Broadcast to RPC
      6. Return transaction hash immediately (do not wait for confirmation)
    - Error handling: Return descriptive, classifiable errors (network failure, gas limit, signing error, "already certified" contract revert) so the worker can decide whether to retry or treat as terminal.

15. **HC-BE-05: Blockchain reader (verification)**
    - File: `backend/internal/blockchain/reader.go`
    - Function: `VerifyBatch(ctx context.Context, batchID int64) (txConfirmed bool, storedPdfHash [32]byte, err error)`
    - Steps:
      1. Connect to Polygon RPC
      2. Call contract.getCertification(batchID) via eth_call
      3. Return stored hash + confirmations count
      4. If not found, return ErrBatchNotCertified
    - Also provide: `GetTransactionStatus(txHash string) (confirmed bool, reverted bool, blockNumber uint64, gasUsed uint64, err error)` — used by the worker's confirmation-polling loop (HC-BE-15b) to move jobs from `submitted`/`pending_confirmation` to `confirmed`/`reverted`.

16. **HC-BE-06: Hash utilities**
    - File: `backend/internal/blockchain/hash.go`
    - Function: `SHA256File(filePath string) ([32]byte, error)` — for PDF hashing (unchanged)
    - Metadata hashing is specified separately in HC-BE-06b below, since it needs a precise, versionable contract rather than "whatever a struct happens to serialize to".

17. **HC-BE-06b: Canonical metadata hashing** *(new — addresses improvement #5)*
    - File: `backend/internal/blockchain/metadata.go`
    - Problem: `metadata_hash` was previously referenced but never specified. `fmt.Sprintf("%v", batch)` or `json.Marshal(struct)` are **not** used, because Go struct-to-JSON field order is an implementation detail (not a stable cross-language contract) and float formatting is locale/precision-sensitive.
    - Function: `CanonicalMetadataHash(batch *model.HoneyBatch) [32]byte`
    - **Exact field list and order** (only these fields participate; nothing else, so adding an unrelated column later never changes existing hashes):
      1. `batch_id` — decimal string, no leading zeros (e.g. `"42"`)
      2. `apiary_id` — decimal string
      3. `gathering_date` — UTC, RFC 3339 date-only, e.g. `"2026-07-18"`
      4. `amount_grams` — decimal string (integer, never a float — this is exactly why HC-DB-01 stores grams as BIGINT rather than kg as NUMERIC/float64)
      5. `processing_method` — exact enum string (`"raw"` / `"filtered"` / `"pasteurized"`)
      6. `honey_type` — UTF-8, NFC-normalized, as stored (no trimming/casing changes at hash time — normalization happens once at write time in the service layer so the hash always matches what's persisted)
      7. `pdf_file_hash` — lowercase hex string of the PDF's SHA256 (from `SHA256File`)
    - **Construction:** join the seven fields with the ASCII unit separator `\x1f` (0x1F) in the exact order above, encode the resulting string as UTF-8 bytes, then `sha256.Sum256(...)`. The unit separator (rather than e.g. `,` or `|`) is chosen because it cannot legally appear in any of the field values, so there is no ambiguity/injection risk between fields (e.g. a honey_type containing a comma can't shift field boundaries).
    - Output is stored as a lowercase hex string in `honey_batches.metadata_hash` and passed on-chain as `bytes32`.
    - This function is pure and deterministic: same `HoneyBatch` row → same hash, on any machine, any time, in Go or (if ever reimplemented) any other language — the spec above is the language-agnostic contract, not "read the Go source".

### Phase 4: Backend Business Logic (7 tasks)
**Goals:** Service layer with validation, async job orchestration, and a dedicated worker for all blockchain interaction.

**Async Certification Architecture** *(addresses improvement #1)*: `CreateBatch` no longer talks to Polygon at all. It validates the request, persists the batch, and — only if the caller opted in — enqueues a `blockchain_jobs` row with `status='queued'` in the same DB transaction, then returns. A separate long-running worker process (started once at app boot, HC-BE-15b) is the *only* code path that ever calls `blockchain.Writer` or `blockchain.Reader`. This means the API's response time is bounded purely by Postgres, never by Polygon RPC latency, and a Polygon outage degrades to "certifications queue up" rather than "batch creation fails".

**Certification is opt-in, not automatic** *(revised)*: not every honey batch needs an on-chain certification — a beekeeper may want to log a batch's data without paying for/waiting on certification, and decide later whether it's worth certifying. `CreateBatchRequest` carries a `RequestCertification bool` field (default `false` if omitted). A batch created with it `false` has **no** `blockchain_jobs` or `honey_batch_certifications` rows at all — not a `queued` job, just nothing. `GetLatestByBatchID` simply returns no row for such a batch, and every layer (service, API JSON, Dart model) represents that as an absent/nil certification rather than inventing a status value for it — see HC-BE-07b. The owner can request certification at any later time via the broadened retry/certify endpoint (see HC-BE-24c below).

18. **HC-BE-13: Service — Create honey batch**
    - File: `backend/internal/service/honey_batch.go`
    - Function: `CreateBatch(ctx context.Context, userID, apiaryID int64, req CreateBatchRequest) (*model.HoneyBatch, error)`
    - `CreateBatchRequest` includes `RequestCertification bool` (default `false`) alongside the batch fields
    - Validation:
      - User owns apiary (via apiary repo)
      - Apiary exists
      - AmountGrams > 0 and <= 100,000,000 (100,000 kg / 100 tonnes reasonable upper bound, expressed in grams — a sanity guard against typos/garbage input, not a realistic per-batch limit)
      - HoneyType not empty, <= 100 chars
      - Processing method is valid
      - PDF file provided and accessible
    - Steps:
      1. Validate inputs
      2. Hash PDF file (`SHA256File`)
      3. Generate `verification_token` (crypto-random UUID v4, see HC-BE-16 rename below)
      4. Compute `metadata_hash` (`CanonicalMetadataHash`, HC-BE-06b)
      5. In a single DB transaction: insert the batch row, and — **only if `RequestCertification` is true** — insert a `blockchain_jobs` row (`job_type='certify'`, `status='queued'`, `next_retry_at=NOW()`). If false, the transaction is just the batch insert.
      6. Return the created batch immediately — **no blockchain call happens here**
    - Error responses: Use domain errors (ErrApiaryNotFound, ErrInvalidAmount, etc.)

19. **HC-BE-14: Service — Get batch + verify**
    - File: `backend/internal/service/honey_batch.go`
    - Function: `GetBatchWithVerification(ctx context.Context, token string) (*BatchVerification, error)` — looked up by verification token (HC-BE-16b), not numeric ID, for the public-facing path
    - Returns struct: batch data + latest certification (from `HoneyBatchCertificationRepository.GetLatestByBatchID`) + hash comparison + confirmation count
    - If `GetLatestByBatchID` returns no row, the returned certification field is nil/absent (not a status value) — this is a normal, expected state, not an error
    - If the latest certification has a `transaction_hash`, this only reads from the DB (already kept fresh by the worker's confirmation loop) — it does **not** make a live RPC call on every request, keeping the endpoint fast under load. A "Refresh" action (HC-FE-04) can trigger an on-demand re-check via a lightweight endpoint that nudges the worker or does a bounded live read.
    - Compare stored `pdf_file_hash` with the batch's current `pdf_file_hash` (detects if the PDF was ever swapped post-certification — should always match since PDFs are immutable once uploaded)

20. **HC-BE-15b: Worker — process certify jobs** *(new — addresses improvements #1, #2, #6)*
    - File: `backend/internal/worker/blockchain_worker.go`
    - Function: `(w *BlockchainWorker) ProcessNextJob(ctx context.Context) (processed bool, err error)`
    - Steps:
      1. `BlockchainJobRepository.ClaimNext(ctx)` — atomically claim one runnable job (or return `false, nil` if none)
      2. **Idempotency check (improvement #6):** query `HoneyBatchCertificationRepository.GetLatestByBatchID` for this batch; if a "live" certification already exists (`IsLive()` — submitted/pending_confirmation/confirmed), mark the job `confirmed`/skip resubmission instead of calling the contract again. This covers the case where a previous worker crashed *after* broadcasting but *before* updating job status.
      3. Mark job `submitting`; create a new `honey_batch_certifications` row (`status='submitting'`, chain_id/contract_address from config)
      4. Call `blockchain.Writer.CertifyBatch(...)`. If the contract reverts with "already certified" (see HC-BE-02), treat as success (another safety net for improvement #6) and mark the certification `confirmed`-pending-verification via the reader instead of erroring.
      5. On successful broadcast: update the certification row (`status='submitted'`, `transaction_hash=...`) and the job (`status='submitted'`)
      6. On failure: increment `attempt_count`; if under the retry limit, set `status='failed'`... `next_retry_at = now + backoff(attempt_count)` (exponential: 1s, 2s, 4s, 8s, capped) so `ClaimNext` picks it up again later; if attempts exhausted, leave `status='failed'` permanently and record `last_error`
    - Run in a loop on a ticker (`JobPollInterval`, default 5s), started from `main.go` (HC-BE-24).

21. **HC-BE-15c: Worker — poll for confirmations** *(replaces the old HC-BE-15 "PollPendingBatchStatuses")*
    - File: `backend/internal/worker/blockchain_worker.go`
    - Function: `(w *BlockchainWorker) PollSubmittedJobs(ctx context.Context) error`
    - `BlockchainJobRepository.ListPendingConfirmation(ctx)` → for each, call `blockchain.Reader.GetTransactionStatus(txHash)`
    - Not yet mined: leave as `submitted`. Mined but under `RequiredConfirmations`: move to `pending_confirmation`. Mined with enough confirmations and not reverted: move certification + job to `confirmed`, set `block_number`, `gas_used`, `confirmation_timestamp`. Reverted: move both to `reverted` (terminal — does not retry, since a revert is a semantic failure like "already certified" or a contract-level rejection, not a transient one)
    - Runs on a separate ticker (`ConfirmationPollInterval`, default 30s), started alongside `ProcessNextJob`'s loop.

22. **HC-BE-25: Idempotency guarantees** *(new — addresses improvement #6, consolidates the mechanisms introduced above)*
    - Three layers, so no single failure mode causes double-certification:
      1. **Contract-level:** `certify()` reverts if the batch already has a stored certification (HC-BE-02).
      2. **Application-level:** the worker checks `HoneyBatchCertificationRepository.GetLatestByBatchID` before every submission attempt and skips resubmission if a live certification already exists (HC-BE-15b step 2).
      3. **Database-level:** a partial unique index on `honey_batch_certifications (batch_id) WHERE status IN ('submitted','pending_confirmation','confirmed')` (HC-DB-04) makes it a constraint violation — not just a logic bug — for two "live" certification rows to exist for the same batch simultaneously.
    - This is documented explicitly here (rather than left implicit) because RPC timeouts are the expected failure mode after a transaction is broadcast but before the response is read — the worker cannot assume "no response" means "not submitted".

23. **HC-BE-16: Service — Generate QR code data**
    - File: `backend/internal/service/honey_batch.go`
    - Function: `GenerateQRCodeData(verificationToken string, appURL string) (qrData string, error)`
    - Returns URL: `{appURL}/verify/{verificationToken}` (see improvement #7 — never the numeric batch id)
    - Encoded as data:image/svg+xml or PNG (use qrcode library)
    - Store QR code data in honey_batch_qr_codes table

24. **HC-BE-16b: Verification token generation** *(new — addresses improvement #7)*
    - File: `backend/internal/service/honey_batch.go` (helper used by `CreateBatch`, HC-BE-13)
    - `verification_token` is a UUID v4 generated via `crypto/rand` (Go's `google/uuid` package with the default crypto-random generator — never `math/rand`) at batch-creation time and stored once, immutably, in `honey_batches.verification_token`.
    - Sequential/enumerable numeric IDs (`honey_batches.id`) are used only for authenticated, owner-scoped routes (`GET/PATCH/DELETE /api/v1/honey-batches/{id}`) and never appear in QR codes, the public verification API, or any response served to an unauthenticated caller.

### Phase 5: Backend API Handlers (10 tasks)
**Goals:** HTTP endpoints for batch CRUD (owner-scoped, numeric ID) and verification (public, token-scoped).

25. **HC-BE-17: Handler — POST /api/v1/honey-batches**
    - File: `backend/internal/handler/honey_batch.go`
    - Auth required (extract userID from context)
    - Parse multipart form: batch data (JSON, including optional `request_certification: bool`, default `false`) + PDF file
    - Validate request (see service HC-BE-13)
    - Call service.CreateBatch()
    - Return: batch object including `verification_token`, plus `certification_status: "queued"` if `request_certification` was true, or no `certification` field at all (null) if false/omitted — **no tx_hash yet either way**, since certification (when requested) hasn't been submitted at request time (this is the API-visible consequence of the async architecture in improvement #1)
    - Error mapping: 400 for validation, 403 for ownership, 500 for unexpected DB failure. A blockchain-side failure can never produce a 500 here anymore, since no blockchain call happens synchronously.

26. **HC-BE-18: Handler — GET /api/v1/honey-batches/{id}**
    - File: `backend/internal/handler/honey_batch.go`
    - **Auth + ownership required** (changed from public — public access now goes exclusively through the token-based `/verify/{token}` route below, per improvement #7)
    - Returns the batch plus its current certification status/history for the owner's own management UI
    - 404 if not found or not owned by the caller

27. **HC-BE-19: Handler — GET /api/v1/verify/{token}** *(renamed from `/honey-batches/{id}/verify`)*
    - File: `backend/internal/handler/honey_batch.go`
    - Public endpoint, keyed by `verification_token` — the numeric ID is never exposed here (improvement #7)
    - Detailed verification: on-chain hash vs PDF hash, latest certification status, tx status, block number, confirmation count
    - Return JSON: `{onChainHash, pdfHash, txHash, status, blockNumber, confirmationTimestamp}` — `status` uses the full lifecycle enum (queued/submitting/submitted/pending_confirmation/confirmed/failed/reverted), not just a boolean
    - 404 if no batch matches the token (indistinguishable, by design, from "token never existed" vs "batch deleted" — avoids leaking which)

28. **HC-BE-20: Handler — GET /api/v1/honey-batches** (list)
    - File: `backend/internal/handler/honey_batch.go`
    - Auth required
    - Query params: apiary_id, honey_type, limit, offset
    - Return paginated list: `{items: [], total: int}` — each item includes its latest certification status for the list-view badge

29. **HC-BE-21: Handler — GET /api/v1/verify/{token}/qr-code** *(moved under the token-scoped verify path)*
    - File: `backend/internal/handler/honey_batch.go`
    - Public endpoint
    - Returns QR code as PNG or SVG image, encoding `/verify/{token}`
    - Content-Type: image/png or image/svg+xml
    - Cache headers: Cache-Control: public, max-age=31536000 (immutable QR)

30. **HC-BE-22: Handler — PATCH /api/v1/honey-batches/{id}**
    - File: `backend/internal/handler/honey_batch.go`
    - Auth + ownership required
    - Allow updating only: notes, honey_type (not PDF, amount, or any certification data)
    - Call service.UpdateBatch()
    - Return updated batch

31. **HC-BE-23: Handler — DELETE /api/v1/honey-batches/{id}**
    - File: `backend/internal/handler/honey_batch.go`
    - Auth + ownership required
    - Soft delete (set deleted_at timestamp in DB)
    - Return 204 No Content or {success: true}
    - Note: the on-chain certification is immutable and intentionally *not* affected by a soft delete — this only hides the batch from the app's own listings.

32. **HC-BE-24: Handler — GET /api/v1/honey-batches/{id}/pdf** *(owner-scoped)*
    - File: `backend/internal/handler/honey_batch.go`
    - Auth + ownership required
    - Retrieve lab PDF URL (redirect or proxy) for the batch owner's own management screens

33. **HC-BE-24b: Handler — GET /api/v1/verify/{token}/pdf** *(new — public, token-scoped)*
    - File: `backend/internal/handler/honey_batch.go`
    - Public endpoint, mirrors HC-BE-24 but keyed by verification token so anyone scanning the QR code can view the certifying lab PDF without ever learning the batch's numeric ID
    - Return 302 redirect to S3/cloud storage URL or stream PDF

34. **HC-BE-24c: Handler — POST /api/v1/honey-batches/{id}/retry-certification** *(broadened — now doubles as "certify this batch now")*
    - File: `backend/internal/handler/honey_batch.go`
    - Auth + ownership required
    - Looks up the batch's latest certification (if any) and enqueues a fresh `blockchain_jobs` row (`status='queued'`) in these cases:
      - **No certification exists yet** (`GetLatestByBatchID` returns nil) — this is now the primary "certify this batch" action for batches created with `request_certification: false`, not just a retry path
      - Latest certification is `failed` or `reverted` — re-enqueue, same as the original retry behavior
    - Rejects (409) if the latest certification is already `queued`/`submitting`/`submitted`/`pending_confirmation`/`confirmed` — there's already a live or in-flight attempt, nothing to do
    - Idempotency guarantees from HC-BE-25 apply identically to manually-triggered jobs, whether this is a first-time request or a retry
    - Surfaced by HC-FE-03/HC-FE-06 as both a "Certify" button (when no certification exists yet) and a "Retry" button (on `failed`/`reverted` batches) — same endpoint, different button copy depending on current state

### Phase 6: Backend Integration & Wiring (2 tasks)
**Goals:** Wire handlers, services, repos, and the worker into main app.

35. **Wire blockchain components in main.go**
    - File: `backend/cmd/api/main.go`
    - Create BlockchainConfig from env vars
    - Create blockchain.Writer and blockchain.Reader
    - Create HoneyBatchRepository, HoneyBatchCertificationRepository, BlockchainJobRepository
    - Create HoneyBatchService with repo deps (the service no longer holds a direct blockchain.Writer dependency — only the worker does, per improvement #1)
    - Create HoneyBatchHandler with service
    - Register routes: POST/GET /api/v1/honey-batches, GET /api/v1/verify/{token}(/qr-code|/pdf), etc.
    - Start the `BlockchainWorker` (HC-BE-26 below) as a background goroutine

36. **HC-BE-26: Background worker runner** *(replaces the old "Add background job scheduler")*
    - File: `backend/internal/worker/blockchain_worker.go`
    - Function: `(w *BlockchainWorker) Run(ctx context.Context)` — runs two independent ticker loops concurrently:
      - `ProcessNextJob` every `JobPollInterval` (default 5s) — claims and submits queued/retryable jobs
      - `PollSubmittedJobs` every `ConfirmationPollInterval` (default 30s) — checks on submitted transactions
    - Graceful shutdown: both loops select on `ctx.Done()` and exit cleanly; `main.go` cancels the context on SIGTERM/SIGINT the same way other background work in the app already shuts down.
    - On startup, the worker doesn't need any special "recovery" logic beyond its normal claim query — jobs left `queued` or `submitting`-then-crashed are picked up naturally because `ClaimNext` only looks at DB state, not in-memory state. (A job stuck in `submitting` past a timeout is treated as `failed`+retryable by a periodic sweep, to handle a crash mid-step 3 of HC-BE-15b.)

### Phase 7: Frontend Models & Repositories (3 tasks)
**Goals:** Dart layer for API communication.

37. **HC-FE-08: HoneyBatchModel (Dart)**
    - File: `app/lib/features/honey/data/honey_batch_model.dart`
    - Mirrors Go HoneyBatch struct
    - Immutable const constructor
    - fromJson factory: parses API response
    - Fields: id, userId, apiaryId, verificationToken, gatheringDate, amountGrams (int), processingMethod, honeyType, labPdfUrl, pdfFileHash, createdAt, updatedAt — **no blockchainStatus/blockchainTxHash fields here**; those live on `HoneyBatchCertification` (HC-FE-08b)
    - `amountKg` getter (`amountGrams / 1000.0`) purely for display formatting — the canonical value passed around the app remains the integer `amountGrams`

38. **HC-FE-08b: HoneyBatchCertificationModel (Dart)** *(new)*
    - File: `app/lib/features/honey/data/honey_batch_certification_model.dart`
    - Fields: id, batchId, chainId, contractAddress, transactionHash, blockNumber, status (CertificationStatus enum), gasUsed, confirmationTimestamp, createdAt
    - `CertificationStatus` enum: `queued, submitting, submitted, pendingConfirmation, confirmed, failed, reverted` — string values match the Go/DB constants exactly (`pending_confirmation` ↔ `pendingConfirmation`, etc.) via explicit `fromJson`/`toJson` mapping, so the lifecycle defined in HC-BE-07b is the single cross-stack source of truth (addresses improvement #4). `HoneyBatchModel.certification` is nullable (`HoneyBatchCertificationModel?`); `null` means "not certified yet" — the UI checks for null rather than switching on an extra enum value.

39. **HC-FE-09: ProcessingMethodEnum (Dart)**
    - File: `app/lib/features/honey/data/honey_batch_model.dart` (same file)
    - Enum: raw, filtered, pasteurized
    - Display labels: "Raw", "Filtered", "Pasteurized"
    - Validation function

40. **HC-FE-10: HoneyBatchRepository (Dart)**
    - File: `app/lib/features/honey/data/honey_batch_repository.dart`
    - ApiClient dependency injection
    - Methods:
      - listBatches(apiaryId, {limit, offset}) → Future<(items: List<HoneyBatch>, total: int)>
      - getBatch(id) → Future<HoneyBatch> (owner-scoped, numeric id)
      - createBatch(data, pdfFile, {requestCertification = false}) → Future<HoneyBatch> (multipart upload)
      - updateBatch(id, data) → Future<HoneyBatch>
      - deleteBatch(id) → Future<void>
      - requestCertification(id) → Future<void> (calls HC-BE-24c; also used to retry a failed/reverted certification, same endpoint)
      - verifyByToken(token) → Future<VerificationDetails> (calls the public `/verify/{token}` route — no auth header needed, works for scanned QR codes from any user, including logged-out)
      - getQRCodeUrl(token) → String (URL to QR image at `/verify/{token}/qr-code`)
    - Error handling: Convert DioException to ApiException

### Phase 8: Frontend State Management (1 task)
**Goals:** Cubit for honey batch state.

41. **HC-FE-19: Honey BLoC/Cubit**
    - File: `app/lib/features/honey/cubit/honey_batches_cubit.dart` + `honey_batches_state.dart`
    - States: HoneyBatchesInitial, HoneyBatchesLoading, HoneyBatchesLoaded(batches, total, isLoadingMore), HoneyBatchesError(code)
    - Methods:
      - load(apiaryId) → emit Loading then Loaded
      - loadMore() → append to existing list
      - create(data, file, {requestCertification = false}) → emit Loading, update state (resulting batch's `certification` is populated with status `queued` if requested, else stays `null`)
      - delete(id) → remove from list, refresh if error
      - requestCertification(id) → re-enqueue (works whether `certification` is currently `null` or `failed`/`reverted`) and optimistically update the item's badge to "queued"
      - setApiaryFilter(apiaryId) → reload
    - Pagination: track offset, hasMore flag

### Phase 9: Frontend Core Screens (7 tasks)
**Goals:** User-facing UI for creating, viewing, and verifying batches.

42. **HC-FE-01 & HC-FE-07: Honey batches home screen + My honey batches**
    - File: `app/lib/features/honey/view/honey_batches_screen.dart`
    - Main screen with Cubit setup
    - Displays list of user's batches
    - "Create Batch" button → navigate to create screen
    - List items show: batch name/type, date, certification status (badge reflecting the full lifecycle — see HC-FE-15)
    - Tap item → detail screen (owner view, numeric id)
    - Pull-to-refresh support

43. **HC-FE-02: Create honey batch screen**
    - File: `app/lib/features/honey/view/create_honey_batch_screen.dart`
    - Form fields:
      - Gathering date (date picker)
      - Amount in kg (text input, numeric, decimal allowed for entry — converted to whole grams via `(kg * 1000).round()` before hitting the repository/API; the form validates the input still resolves to a positive integer gram amount)
      - Processing method (dropdown: raw/filtered/pasteurized)
      - Honey type (text input, autocomplete suggestions)
      - Apiary selector (dropdown, pre-filled if from apiary detail)
      - PDF upload (file picker, show file name + size)
      - "Certify on the blockchain" toggle/checkbox — **off by default**, since certification is opt-in; when off, the batch is created with no certification attempt at all (`certification` stays `null`, certifiable later from the detail screen)
    - Submit button: calls cubit.create(), shows progress
    - Error toast if validation fails
    - Success → pop screen + reload parent; success message depends on the toggle: "Batch created — blockchain certification is in progress" if requested, or "Batch created" if not

44. **HC-FE-03: Honey batch detail screen**
    - File: `app/lib/features/honey/view/honey_batch_detail_screen.dart`
    - Displays: all batch info, gathering date, amount (formatted from `amountGrams` as kg), processing method, honey type, apiary name
    - PDF section: preview/download link (owner-scoped endpoint)
    - QR code section: display QR image (encodes the verification token URL), "Share" button — hidden/disabled while `certification` is `null`, since there's nothing on-chain yet to verify
    - Certification status section: badge shows a distinct "Not certified yet" indicator when `certification` is `null`, otherwise reflects the full lifecycle (queued/submitting/submitted/pending confirmation/confirmed/failed/reverted) + details button; when `null`, shows a "Certify" action, and if `failed`/`reverted`, shows a "Retry" action — both wired to the same `requestCertification` call
    - Edit button (if user owns batch) → edit screen
    - Delete button (if user owns batch)

45. **HC-FE-04: Honey batch verification screen**
    - File: `app/lib/features/honey/view/honey_batch_verification_screen.dart`
    - Loaded via `verifyByToken(token)` — works for any visitor, no auth required
    - Displays verification details from the `/verify/{token}` endpoint:
      - Status badge showing the exact lifecycle state ("Queued" / "Submitted" / "Pending confirmation" / "✓ Confirmed on Polygon" (with timestamp) / "✗ Failed" / "✗ Reverted")
      - "PDF hash matches" indicator
      - Transaction hash (clickable link to Polygonscan, once available)
      - Block number + confirmation count (once mined)
      - Metadata hash displayed
    - Refresh button: re-fetches from the `/verify/{token}` endpoint (does not itself trigger a live RPC call — reflects the worker's latest DB state, per HC-BE-14)

46. **HC-FE-05: QR code display screen**
    - File: `app/lib/features/honey/view/qr_code_display_screen.dart`
    - Full-screen QR code image
    - "Share" button → share QR image via OS share sheet
    - Long-press context menu: "Save as image" (to gallery)

47. **HC-FE-06: QR code scanner screen**
    - File: `app/lib/features/honey/view/qr_code_scanner_screen.dart`
    - Uses `mobile_scanner` or `qr_code_scanner` package
    - Live camera preview
    - Detect QR code, extract the **verification token** from the scanned URL (`/verify/{token}` — a UUID string, not a numeric id, per improvement #7)
    - Navigate to verification screen
    - Handle errors: invalid QR, network error, batch not found

48. **HC-FE-18: Add Honey Batches section to hive detail**
    - File: `app/lib/features/hive/view/hive_detail_screen.dart` (modify existing)
    - Add tab or collapsible section "Honey"
    - Display batches from this hive
    - "Create Batch" button opens create screen with hive pre-selected

### Phase 10: Frontend Utils & Widgets (4 tasks)
**Goals:** Shared components and helper functions.

49. **HC-FE-11 & HC-FE-12: QR code generation + scanner**
    - File: `app/lib/features/honey/utils/qr_utils.dart`
    - generateQRImageUrl(token) → returns URL to QR endpoint (`/verify/{token}/qr-code`)
    - scanQRCodeFromCamera() → uses mobile_scanner
    - extractVerificationTokenFromQRData(qrData) → parses URL to get the token string

50. **HC-FE-13 & HC-FE-14: PDF preview + upload UI**
    - File: `app/lib/features/honey/widgets/pdf_upload_widget.dart` + `pdf_preview_widget.dart`
    - PdfUploadWidget: file picker, show name + size, upload progress
    - PdfPreviewWidget: embed PDF.js or use pdf_viewer_plugin, link to download
    - Both handle errors gracefully

51. **HC-FE-15 & HC-FE-16: Certification status indicator + verification modal**
    - File: `app/lib/features/honey/widgets/blockchain_status_widget.dart` + `verification_modal.dart`
    - `BlockchainStatusWidget`: badge driven by the `CertificationStatus` enum (HC-FE-08b) — distinct color/icon per state (e.g. neutral outline "Not requested", grey "Queued", blue spinner "Submitting"/"Submitted", amber "Pending confirmation", green check "Confirmed", red "Failed"/"Reverted") rather than a flattened three-state pending/confirmed/failed badge
    - `VerificationModal`: detailed view of on-chain state, including certification history (all rows from `ListByBatchID`, most recent first) if more than one exists — surfaces retries/migrations transparently

52. **HC-FE-17: Hash comparison display**
    - File: `app/lib/features/honey/widgets/hash_comparison_widget.dart`
    - Side-by-side display: "On-chain hash: xyz..." and "Current PDF hash: xyz..."
    - Monospace font for hashes
    - Highlight match/mismatch

### Phase 11: Polish & Edge Cases (10 tasks)
**Goals:** Solid error handling, offline support, localization for a testing/demo environment. Items marked **(production-hardening — optional for thesis)** below can be skipped without weakening the thesis's engineering content.

53. **HC-10-01: PDF file validation (backend)**
    - File: `backend/internal/service/honey_batch.go` (add to HC-BE-13)
    - Check MIME type: application/pdf only
    - Max size: 10MB
    - ClamAV scanning: **(production-hardening — optional for thesis)**, skip

54. **HC-10-02: Blockchain retry logic (backend)**
    - File: `backend/internal/worker/blockchain_worker.go` (moved from the service layer — retries are now entirely the worker's responsibility, per improvement #1/#2)
    - Exponential backoff via `blockchain_jobs.next_retry_at` (1s, 2s, 4s, 8s max), tracked per-job via `attempt_count`
    - After 3 failed retries, job `status='failed'` permanently (until the owner explicitly retries via HC-BE-24c); the associated `honey_batch_certifications` row is also left `failed` for audit history
    - The confirmation-polling loop (HC-BE-15c) handles the separate case of "submitted but not yet confirmed" — that's not a retry scenario, just patience

55. **HC-10-03: PDF storage (backend)**
    - File: `backend/internal/storage/pdf_storage.go`
    - Local file system is sufficient for the thesis's testing environment, matching the existing photo-storage pattern in the app; S3 + signed URLs are **(production-hardening — optional for thesis)**
    - Cleanup: soft-delete PDFs when batch deleted
    - Security: validate PDF MIME before storing (cheap to keep, not skipped)

56. **HC-10-04: Gas fee management (backend)**
    - File: `backend/internal/blockchain/writer.go`
    - `gas_used` is persisted per-certification (HC-DB-04), so cost is auditable per batch — useful for the thesis write-up (e.g. reporting real testnet gas costs), and effectively free to keep
    - Gas relay service (OpenZeppelin Defender) and gas-price-spike alerting: **(production-hardening — optional for thesis)**, not needed since testnet gas is free (fake MATIC)

57. **HC-10-05: Offline handling (frontend)**
    - File: `app/lib/features/honey/cubit/honey_batches_cubit.dart`
    - Cache batch list locally (Hive/SharedPreferences)
    - Because certification is already async server-side, "offline" for the frontend just means "can't refresh status yet" — there's no local queuing of blockchain writes to reconcile, since the app never triggers them directly
    - Retry status refresh when network returns

58. **HC-10-06: Loading states (frontend)**
    - File: `app/lib/features/honey/view/honey_batch_*.dart`
    - Show the appropriate badge for every lifecycle state (never a generic spinner covering "queued" through "pending confirmation" — that's exactly what the richer status model in improvement #4 is for)
    - Allow user to check status later without blocking UI
    - "Certify" button when `certification` is `null`, "Check status" / "Retry" button on `failed`/`reverted` batches (both call HC-BE-24c)

59. **HC-10-07: Error handling (frontend)**
    - File: `app/lib/core/widgets/error_dialog.dart` (modify existing)
    - Display user-friendly certification errors mapped from the lifecycle status:
      - `certification == null` → not an error state; shows a neutral "Not certified yet" indicator with a "Certify" action, not an error dialog
      - `queued`/`submitting`/`submitted`/`pending_confirmation` → "Certification in progress — check back shortly" (not an error state)
      - `failed` → "Certification failed — tap Retry to try again"
      - `reverted` → "Certification was rejected on-chain — contact support"
      - Generic network error while fetching status → "Network error — check your connection"
    - Provide retry buttons where applicable

60. **HC-10-08: Localization (frontend)**
    - File: `app/lib/l10n/app_en.arb` + `app_pl.arb`
    - Add keys: processingMethod_raw, processingMethod_filtered, processingMethod_pasteurized
    - Add keys for every lifecycle state (addresses improvement #4 needing consistent UI badges): certificationStatus_queued, certificationStatus_submitting, certificationStatus_submitted, certificationStatus_pendingConfirmation, certificationStatus_confirmed, certificationStatus_failed, certificationStatus_reverted
    - Add a separate key for the null-certification case (not part of the enum): certificationNotRequested
    - Add keys: verification_verified, verification_details, verification_pending
    - Run `flutter gen-l10n` to regenerate

61. **HC-10-09: Empty states (frontend)**
    - File: `app/lib/features/honey/view/honey_batches_screen.dart`
    - "No honey batches yet" screen with "Create Batch" CTA
    - "No verified batches" filter view (filters on `status == confirmed`)

62. **HC-10-10: Database indexing (backend)**
    - File: covered by the migrations in Phase 1 directly (HC-DB-01 through HC-DB-04) rather than a bolt-on later migration
    - Indexes in place:
      - honey_batches: (user_id, created_at DESC), (apiary_id, created_at DESC), UNIQUE (verification_token)
      - blockchain_jobs: (status, next_retry_at) — the worker's claim query
      - honey_batch_certifications: (batch_id, created_at DESC), partial UNIQUE (batch_id) WHERE status IN ('submitted','pending_confirmation','confirmed')

---

## Critical Files Summary

### New Files to Create
- **Backend:**
  - `backend/migrations/029_create_honey_batches.sql`
  - `backend/migrations/030_create_honey_batch_qr_codes.sql`
  - `backend/migrations/031_create_blockchain_jobs.sql`
  - `backend/migrations/032_create_honey_batch_certifications.sql`
  - `backend/internal/model/honey_batch.go`
  - `backend/internal/model/honey_batch_certification.go`
  - `backend/internal/model/blockchain_job.go`
  - `backend/internal/repository/honey_batch.go`
  - `backend/internal/repository/honey_batch_certification.go`
  - `backend/internal/repository/blockchain_job.go`
  - `backend/internal/handler/honey_batch.go`
  - `backend/internal/service/honey_batch.go`
  - `backend/internal/blockchain/writer.go`
  - `backend/internal/blockchain/reader.go`
  - `backend/internal/blockchain/hash.go`
  - `backend/internal/blockchain/metadata.go`
  - `backend/internal/config/blockchain_config.go`
  - `backend/internal/worker/blockchain_worker.go`
  - `backend/contracts/HoneyCertification.sol` (or external repo)
  - `backend/internal/blockchain/contracts/HoneyCertification.abi` (generated)

- **Frontend:**
  - `app/lib/features/honey/` (new feature folder)
  - `app/lib/features/honey/data/honey_batch_model.dart`
  - `app/lib/features/honey/data/honey_batch_certification_model.dart`
  - `app/lib/features/honey/data/honey_batch_repository.dart`
  - `app/lib/features/honey/cubit/honey_batches_cubit.dart`
  - `app/lib/features/honey/cubit/honey_batches_state.dart`
  - `app/lib/features/honey/view/honey_batches_screen.dart`
  - `app/lib/features/honey/view/create_honey_batch_screen.dart`
  - `app/lib/features/honey/view/honey_batch_detail_screen.dart`
  - `app/lib/features/honey/view/honey_batch_verification_screen.dart`
  - `app/lib/features/honey/view/qr_code_display_screen.dart`
  - `app/lib/features/honey/view/qr_code_scanner_screen.dart`
  - `app/lib/features/honey/widgets/blockchain_status_widget.dart`
  - `app/lib/features/honey/widgets/verification_modal.dart`
  - `app/lib/features/honey/widgets/pdf_upload_widget.dart`
  - `app/lib/features/honey/widgets/pdf_preview_widget.dart`
  - `app/lib/features/honey/widgets/hash_comparison_widget.dart`
  - `app/lib/features/honey/utils/qr_utils.dart`

### Files to Modify
- **Backend:**
  - `backend/cmd/api/main.go` — wire blockchain components, start the `BlockchainWorker`
  - `backend/internal/handler/handler.go` — register HoneyBatchHandler (including the `/verify/{token}` public route group)

- **Frontend:**
  - `app/lib/main.dart` — add honey feature to navigation (drawer)
  - `app/lib/features/hive/view/hive_detail_screen.dart` — add Honey tab
  - `app/lib/l10n/app_en.arb` + `app_pl.arb` — add localizations
  - `app/lib/l10n/app_localizations.dart` + `app_localizations_en.dart` + `app_localizations_pl.dart` — regenerated by flutter gen-l10n

---

## Implementation Sequence (Recommended Order)

### Week 1: Foundation
1. Create migrations (HC-DB-01 through HC-DB-04, including `blockchain_jobs` and `honey_batch_certifications` up front, since later phases depend on them)
2. Create Go models (HC-BE-07, HC-BE-08, HC-BE-07b, HC-BE-07c)
3. Create repositories (HC-BE-09 to HC-BE-12, HC-BE-12b)
4. Wire into main.go and test DB operations

### Week 2: Blockchain
5. Blockchain config (HC-BE-01)
6. Smart contract with duplicate-rejection (HC-BE-02), deploy to Amoy (HC-BE-03)
7. Writer & reader (HC-BE-04, HC-BE-05)
8. Canonical metadata hashing (HC-BE-06, HC-BE-06b) — nail this down early since both the contract call and the DB row depend on it
9. Integration tests for blockchain operations, including idempotency (double-submit a job, assert only one live certification results)

### Week 3: Backend Service, Worker & API
10. Service layer — synchronous half (HC-BE-13, HC-BE-14, HC-BE-16, HC-BE-16b)
11. Background worker (HC-BE-15b, HC-BE-15c, HC-BE-25, HC-BE-26) — this is the new critical-path component; get `CreateBatch` returning instantly and the worker independently draining the queue before building handlers on top
12. Handler layer (HC-BE-17 to HC-BE-24c), including the public `/verify/{token}` route group
13. End-to-end API tests: create batch → assert immediate 201 response with no tx hash → poll `/verify/{token}` until worker moves it to `confirmed`

### Week 4: Frontend Foundation
14. Dart models & repository (HC-FE-08, HC-FE-08b, HC-FE-09, HC-FE-10)
15. State management (HC-FE-19)
16. Core screens (HC-FE-01, HC-FE-02, HC-FE-03)

### Week 5: Frontend Completion
17. Verification & QR screens (HC-FE-04, HC-FE-05, HC-FE-06) — using verification tokens throughout
18. Widgets, including the full-lifecycle status badge (HC-FE-11 to HC-FE-17)
19. Integration with hive detail (HC-FE-18)
20. Localization (HC-10-08)

### Week 6: Polish & Edge Cases
21. Error handling, offline support, empty states
22. Testing on Android + Web
23. Performance & UX polish

---

## Testing Strategy

### Backend Tests
- **Unit:** honey_batch_service_test.go (CreateBatch never calls the blockchain package — assert via a mock/spy that it's untouched), blockchain_worker_test.go (mock RPC calls; cover retry/backoff, idempotency skip-on-live-certification, revert handling), metadata_test.go (fixed input → fixed expected hash, golden-value test to lock the canonical hashing spec in place)
- **Integration:** Test DB migrations, repository CRUD, `BlockchainJobRepository.ClaimNext` under concurrent callers (assert `SKIP LOCKED` prevents double-claim), worker end-to-end against a local/mock chain
- **End-to-end:** POST /api/v1/honey-batches with mock file → assert immediate response with `certification_status=queued` and no tx hash → run worker tick → assert DB reaches `confirmed`

### Frontend Tests
- **Unit:** Cubit logic, model serialization (golden tests for JSON), `CertificationStatus` enum round-trip against every backend status string
- **Widget:** Create screen form validation (including kg→grams conversion), QR display, status badge rendering for every lifecycle state
- **Integration:** Create batch → list batches (shows "queued") → verify via token → delete

### Blockchain Tests
- Deploy contract to Amoy testnet
- Test certify() function, event emission, and the duplicate-batchID revert path
- Test hash verification (on-chain vs stored)
- Test failed tx retry logic via the worker, not the old inline retry

---

## Verification

### Manual Testing Checklist
1. ✅ Create honey batch with PDF upload — API responds immediately, before any chain interaction
2. ✅ Certification status transitions, observed via `/verify/{token}`: queued → submitting → submitted → pending_confirmation → confirmed
3. ✅ Verify batch: hash matches, tx confirmed
4. ✅ Generate & scan QR code — scanned URL contains a verification token, not a numeric id
5. ✅ List batches by apiary + pagination
6. ✅ Edit batch notes (no certification changes)
7. ✅ Delete batch (soft delete; on-chain record untouched)
8. ✅ Offline: batch listed, certification status shows last-known state
9. ✅ Error cases: invalid PDF, RPC network failure mid-worker-run (job retries with backoff), gas exhaustion, contract revert (duplicate batchID)
10. ✅ Idempotency: kill the worker process immediately after it broadcasts a transaction (before marking the job submitted) and restart — assert exactly one live certification exists for the batch, not two
11. ✅ Attempting to fetch `/api/v1/honey-batches/{id}` (numeric, owner-scoped) without auth is rejected; attempting to guess sequential ids does not expose other users' batches via `/verify/{token}`

### Performance Baselines
- Create batch: < 500ms (API response is now purely DB-bound — no Polygon RPC in the critical path, a significant improvement over the original "< 5s, async blockchain" baseline)
- List batches (20 items): < 2s
- Worker job claim → submit: < 5s after enqueue (bounded by `JobPollInterval`)
- Worker confirmation detection: < 30s after on-chain inclusion (bounded by `ConfirmationPollInterval`)
- QR generation: < 1s

---

## Known Constraints & Decisions

1. **Blockchain costs:** Using Polygon (cheap gas) instead of Ethereum mainnet
2. **Off-chain storage:** PDF stored in S3 or local FS; only hash on-chain for cost efficiency
3. **Fully async blockchain:** `CreateBatch` returns immediately after persisting the batch and enqueuing a `blockchain_jobs` row; a dedicated `BlockchainWorker` owns every chain interaction. The HTTP layer has zero dependency on Polygon RPC latency or availability.
4. **Durable job queue:** `blockchain_jobs` (not an in-memory goroutine/channel) backs all certification work, so restarts, crashes, and retries are all handled by re-reading DB state rather than requiring in-process recovery logic.
5. **Certification history is append-only:** `honey_batch_certifications` allows multiple rows per batch across retries, contract migrations, or additional chains — `honey_batches` itself carries no mutable blockchain fields.
6. **Precise amounts:** Honey weight is stored as `amount_grams` (`int64`), never `float64`, eliminating floating-point drift anywhere it matters (persistence, API payloads, and especially the deterministic metadata hash).
7. **Public verification identifiers:** QR codes and the public verification API use a random `verification_token` (UUID v4), never the sequential database id — batches cannot be enumerated by walking `/verify/{n}`.
8. **No smart contract UI:** Users don't interact directly with contract; BeeTrack backend is sole caller
9. **Soft deletes:** Batches marked deleted, not hard-deleted (audit trail); on-chain certification records are immutable regardless
10. **QR code immutability:** URLs permanent; scanning always reflects the worker's latest known on-chain state (not a live RPC call per scan, for both latency and rate-limit reasons)
11. **Idempotency is defense-in-depth:** enforced at three independent layers — smart contract revert-on-duplicate, application-level pre-submission check, and a DB partial unique constraint — so no single missed edge case causes a double-certification.
12. **Thesis scope, not production:** target environment is Polygon Amoy testnet only. Mainnet deployment, gas relay services, malware scanning, and signed-URL cloud storage are documented as future work but are **(production-hardening — optional for thesis)** and not required to consider the feature complete.
13. **Certification is opt-in, not automatic:** `CreateBatch` no longer always enqueues a certification job. A batch can exist indefinitely with zero rows in `blockchain_jobs`/`honey_batch_certifications` — represented everywhere as a nil/null certification, not a synthesized status value — until the owner explicitly requests certification via the broadened HC-BE-24c endpoint. This reflects that not every batch needs on-chain proof — some beekeepers just want the record-keeping.

---

## Decisions Made

1. **PDF Storage:** Match photo storage pattern (currently URL-based). PDFs stored as URLs in DB, can be backed by S3, cloud storage, or local file system per deployment
2. **Blockchain Network:** **Polygon Amoy testnet** (chain ID 80002) — costs fake MATIC, safe to test, and is the only target this thesis needs. `chain_id` is still stored per-certification (not globally assumed), which costs nothing extra but keeps a mainnet migration possible later without reinterpreting historical records — that migration itself is out of scope.
3. **Gas Fees:** **App/backend pays** — simpler UX, no wallet required from users
4. **Scope:** **Full feature (revised task breakdown, 6 weeks)** — include polish, error handling, offline support, localization; the async/jobs/certifications-table architecture adds a handful of tasks but does not extend the timeline, since it replaces (rather than adds alongside) the original inline-blockchain-call approach
5. **Verification UI:** Public endpoint for verification, keyed by an unguessable token; QR scanning via app
6. **Amount precision:** Integer grams (`int64`) chosen over `NUMERIC` + a Go decimal library — avoids adding a new dependency, and honey batch weights have no need for sub-gram precision, so the simpler representation fully covers the domain
7. **Async architecture:** A durable `blockchain_jobs` queue processed by a dedicated worker, rather than a fire-and-forget goroutine or an inline synchronous call — chosen specifically for restart-safety and observability, both of which a goroutine-based approach cannot provide
