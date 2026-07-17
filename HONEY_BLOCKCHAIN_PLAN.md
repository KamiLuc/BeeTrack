# Honey Certification & Blockchain Implementation Plan

## Context

Epic 9 ("Honey Certification & Blockchain") from BACKLOG.md aims to create an immutable honey batch certification system using the Polygon blockchain. Each honey batch gets a QR code that verifies authenticity by checking a blockchain-stored hash against the lab PDF. The feature spans database schema, Go backend API, smart contracts, and Flutter frontend screens.

**Blockchain Strategy:** Store minimal data on-chain (hash, metadata hash, timestamp) for cost efficiency. Lab PDF is stored off-chain; blockchain only stores its SHA256 hash.

---

## Implementation Phases

### Phase 1: Database Foundation (4 tasks)
**Goals:** Create schema for honey batches and blockchain state tracking.

1. **HC-DB-01: Create `honey_batches` table migration**
   - File: `backend/migrations/029_create_honey_batches.sql`
   - Columns: id, user_id (created_by), apiary_id, gathering_date, amount (NUMERIC), processing_method (VARCHAR + CHECK), honey_type (TEXT), lab_pdf_url, pdf_file_hash (VARCHAR SHA256), metadata_hash, blockchain_tx_hash, blockchain_contract_address, blockchain_status (enum: pending/confirmed/failed), created_at, updated_at
   - Indexes: (user_id, created_at DESC), (apiary_id), (blockchain_status)
   - FK constraints: user_id → users(id), apiary_id → apiaries(id) ON DELETE CASCADE
   - Timestamps: TIMESTAMPTZ with DEFAULT NOW()

2. **HC-DB-02: Create `honey_batch_qr_codes` table migration**
   - File: `backend/migrations/030_create_honey_batch_qr_codes.sql`
   - Columns: id, batch_id, qr_code_data (VARCHAR), created_at
   - FK: batch_id → honey_batches(id) ON DELETE CASCADE
   - Index: (batch_id)

### Phase 2: Backend Models & Persistence (4 tasks)
**Goals:** Define Go domain models and database layer.

3. **HC-BE-07: Create HoneyBatch model**
   - File: `backend/internal/model/honey_batch.go`
   - Struct fields match DB schema (id, UserID, ApiaryID, GatheringDate, Amount, ProcessingMethod, HoneyType, LabPDFURL, PDFFileHash, MetadataHash, BlockchainTxHash, BlockchainContractAddress, BlockchainStatus, CreatedAt, UpdatedAt)
   - ProcessingMethod as string type (raw, filtered, pasteurized)
   - BlockchainStatus as string type (pending, confirmed, failed)
   - Use `*float64` for nullable Amount if needed; otherwise NUMERIC(8,2) maps to `int64` (cents) or `float64`

4. **HC-BE-08: Create ProcessingMethod enum**
   - File: `backend/internal/model/honey_batch.go` (same file as model)
   - Constants: ProcessingMethodRaw, ProcessingMethodFiltered, ProcessingMethodPasteurized
   - Validation function: IsValidProcessingMethod(method string) bool

5. **HC-BE-09 to HC-BE-12: Create HoneyBatchRepository**
   - File: `backend/internal/repository/honey_batch.go`
   - Methods: Create(ctx, batch) error, GetByID(ctx, id) (*HoneyBatch, error), ListByUserID(ctx, userID, limit, offset), ListByApiaryID(ctx, apiaryID, limit, offset), UpdateStatus(ctx, id, status) error
   - Use GORM pattern (receiver *HoneyBatchRepository(db *gorm.DB))
   - Follow error handling: return nil, nil for not found; return error otherwise
   - Transaction pattern for Create (insert batch + related data)

### Phase 3: Blockchain Integration (6 tasks)
**Goals:** Smart contract deployment and on-chain interaction layer.

6. **HC-BE-01: Blockchain configuration**
   - File: `backend/internal/config/blockchain_config.go`
   - Fields: PolygonRPCURL, ContractAddress, PrivateKey, ChainID (for testnet/mainnet detection)
   - Validation: Ensure private key is 64 hex chars, RPC URL is valid, contract address is 42 chars (0x...)
   - Environment variables: POLYGON_RPC_URL, CONTRACT_ADDRESS, BLOCKCHAIN_PRIVATE_KEY, CHAIN_ID (default 80002 for Amoy testnet)

7. **HC-BE-02: Smart contract (Solidity)**
   - File: `backend/contracts/HoneyCertification.sol` (or upload to external repo)
   - Simple registry: stores (batchID uint256, pdfHash bytes32, metadataHash bytes32, timestamp uint256, ownerAddress address)
   - Function: certify(batchID uint256, pdfHash bytes32, metadataHash bytes32) returns (tx hash)
   - Event: CertificationCreated(indexed batchID, pdfHash, metadataHash, timestamp, ownerAddress)
   - Read function: getCertification(batchID) returns stored data (if exists)
   - Owner validation: Only minter address (BeeTrack backend) can call certify()

8. **HC-BE-03: Deploy contract to Polygon**
   - Deploy to Amoy testnet (chain ID 80002) first; later to Polygon mainnet (137)
   - Use Remix or hardhat for deployment; store contract address in config
   - Document: Store ABI in `backend/internal/blockchain/contracts/HoneyCertification.abi` for Go bindings

9. **HC-BE-04: Blockchain writer (on-chain transaction builder)**
   - File: `backend/internal/blockchain/writer.go`
   - Function: CertifyBatch(ctx context.Context, batchID int64, pdfHash, metadataHash [32]byte) (txHash string, err error)
   - Steps: 
     1. Connect to Polygon RPC (via ethclient.Dial)
     2. Get nonce from pending transactions
     3. Build transaction call to contract.certify()
     4. Sign with private key (ECDSA)
     5. Broadcast to RPC
     6. Return transaction hash immediately (do not wait for confirmation)
   - Error handling: Return descriptive errors (network failure, gas limit, signing error)

10. **HC-BE-05: Blockchain reader (verification)**
    - File: `backend/internal/blockchain/reader.go`
    - Function: VerifyBatch(ctx context.Context, batchID int64) (txConfirmed bool, storedPdfHash [32]byte, err error)
    - Steps:
      1. Connect to Polygon RPC
      2. Call contract.getCertification(batchID) via eth_call
      3. Return stored hash + confirmations count
      4. If not found, return ErrBatchNotCertified
    - Also provide: GetTransactionStatus(txHash string) (confirmed bool, blockNumber uint64, err error)

11. **HC-BE-06: Hash utilities**
    - File: `backend/internal/blockchain/hash.go`
    - Function: SHA256BatchMetadata(batch *model.HoneyBatch) [32]byte
    - Function: SHA256File(filePath string) ([32]byte, error) — for PDF hashing
    - Use standard `crypto/sha256` package; ensure consistent hashing order

### Phase 4: Backend Business Logic (4 tasks)
**Goals:** Service layer with validation and blockchain orchestration.

12. **HC-BE-13: Service — Create honey batch**
    - File: `backend/internal/service/honey_batch.go`
    - Function: CreateBatch(ctx context.Context, userID, apiaryID int64, req CreateBatchRequest) (*model.HoneyBatch, error)
    - Validation:
      - User owns apiary (via apiary repo)
      - Apiary exists
      - Amount > 0 and <= 10000 kg (reasonable upper bound)
      - HoneyType not empty, <= 100 chars
      - Processing method is valid
      - PDF file provided and accessible
    - Steps:
      1. Validate inputs
      2. Hash PDF file
      3. Create batch in DB with blockchain_status = "pending"
      4. Trigger blockchain write (async or fire-and-forget)
      5. Return batch with tx_hash (or empty if async)
    - Error responses: Use domain errors (ErrApiaryNotFound, ErrInvalidAmount, etc.)

13. **HC-BE-14: Service — Get batch + verify**
    - File: `backend/internal/service/honey_batch.go`
    - Function: GetBatchWithVerification(ctx context.Context, batchID int64) (*BatchVerification, error)
    - Returns struct: batch data + on-chain status + hash comparison + confirmation count
    - If tx_hash exists, query blockchain for status
    - Compare PDF hash with stored on-chain hash
    - Return verification details for UI display

14. **HC-BE-15: Service — Poll blockchain status**
    - File: `backend/internal/service/honey_batch.go`
    - Function: PollPendingBatchStatuses(ctx context.Context) error (background job)
    - Queries DB for batches with blockchain_status = "pending"
    - For each, check blockchain reader for confirmation
    - Update DB with new status (confirmed/failed)
    - Use transaction to atomic update
    - Schedule this to run periodically (every 30 seconds via background worker)

15. **HC-BE-16: Service — Generate QR code data**
    - File: `backend/internal/service/honey_batch.go`
    - Function: GenerateQRCodeData(batchID int64, appURL string) (qrData string, error)
    - Returns URL: `{appURL}/verify/batch/{batchID}`
    - Encoded as data:image/svg+xml or PNG (use qrcode library)
    - Store QR code data in honey_batch_qr_codes table

### Phase 5: Backend API Handlers (8 tasks)
**Goals:** HTTP endpoints for batch CRUD and verification.

16. **HC-BE-17: Handler — POST /api/v1/honey-batches**
    - File: `backend/internal/handler/honey_batch.go`
    - Auth required (extract userID from context)
    - Parse multipart form: batch data (JSON) + PDF file
    - Validate request (see service HC-BE-13)
    - Call service.CreateBatch()
    - Return: batch object + blockchain_status + tx_hash
    - Error mapping: 400 for validation, 403 for ownership, 500 for blockchain failure

17. **HC-BE-18: Handler — GET /api/v1/honey-batches/{id}**
    - File: `backend/internal/handler/honey_batch.go`
    - Public endpoint (no auth required)
    - Parse ID from URL
    - Call service.GetBatchWithVerification()
    - Return batch + verification status
    - 404 if not found

18. **HC-BE-19: Handler — GET /api/v1/honey-batches/{id}/verify**
    - File: `backend/internal/handler/honey_batch.go`
    - Public endpoint
    - Detailed verification: on-chain hash vs PDF hash, tx status, confirmation count
    - Return JSON: {onChainHash, pdfHash, txHash, confirmed, blockNumber, timestamp}

19. **HC-BE-20: Handler — GET /api/v1/honey-batches** (list)
    - File: `backend/internal/handler/honey_batch.go`
    - Auth required
    - Query params: apiary_id, honey_type, limit, offset
    - Return paginated list: {items: [], total: int}

20. **HC-BE-21: Handler — GET /api/v1/honey-batches/{id}/qr-code**
    - File: `backend/internal/handler/honey_batch.go`
    - Public endpoint
    - Returns QR code as PNG or SVG image
    - Content-Type: image/png or image/svg+xml
    - Cache headers: Cache-Control: public, max-age=31536000 (immutable QR)

21. **HC-BE-22: Handler — PATCH /api/v1/honey-batches/{id}**
    - File: `backend/internal/handler/honey_batch.go`
    - Auth + ownership required
    - Allow updating only: notes, honey_type (not PDF or blockchain data)
    - Call service.UpdateBatch()
    - Return updated batch

22. **HC-BE-23: Handler — DELETE /api/v1/honey-batches/{id}**
    - File: `backend/internal/handler/honey_batch.go`
    - Auth + ownership required
    - Soft delete (set deleted_at timestamp in DB)
    - Return 204 No Content or {success: true}

23. **HC-BE-24: Handler — POST /api/v1/honey-batches/{id}/pdf**
    - File: `backend/internal/handler/honey_batch.go`
    - Public endpoint
    - Retrieve lab PDF URL (redirect or proxy)
    - Return 302 redirect to S3/cloud storage URL or stream PDF

### Phase 6: Backend Integration & Wiring (2 tasks)
**Goals:** Wire handlers, services, repos into main app.

24. **Wire blockchain components in main.go**
    - File: `backend/cmd/api/main.go`
    - Create BlockchainConfig from env vars
    - Create blockchain.Writer and blockchain.Reader (or singleton wrapped service)
    - Create HoneyBatchRepository
    - Create HoneyBatchService with repo + blockchain deps
    - Create HoneyBatchHandler with service
    - Register routes: POST/GET /api/v1/honey-batches, etc.
    - Start background job: PollPendingBatchStatuses every 30 seconds

25. **Add background job scheduler**
    - File: `backend/internal/jobs/jobs.go` (or pkg/jobs/)
    - Function: StartPollBlockchainJob(ctx context.Context, service *service.HoneyBatchService, interval time.Duration)
    - Ticker loop: run PollPendingBatchStatuses every interval
    - Graceful shutdown: cancel context

### Phase 7: Frontend Models & Repositories (2 tasks)
**Goals:** Dart layer for API communication.

26. **HC-FE-08: HoneyBatchModel (Dart)**
    - File: `app/lib/features/honey/data/honey_batch_model.dart`
    - Mirrors Go HoneyBatch struct
    - Immutable const constructor
    - fromJson factory: parses API response
    - Fields: id, userId, apiaryId, gatheringDate, amount, processingMethod, honeyType, labPdfUrl, pdfFileHash, blockchainTxHash, blockchainStatus, createdAt, updatedAt

27. **HC-FE-09: ProcessingMethodEnum (Dart)**
    - File: `app/lib/features/honey/data/honey_batch_model.dart` (same file)
    - Enum: raw, filtered, pasteurized
    - Display labels: "Raw", "Filtered", "Pasteurized"
    - Validation function

28. **HC-FE-10: HoneyBatchRepository (Dart)**
    - File: `app/lib/features/honey/data/honey_batch_repository.dart`
    - ApiClient dependency injection
    - Methods:
      - listBatches(apiaryId, {limit, offset}) → Future<(items: List<HoneyBatch>, total: int)>
      - getBatch(id) → Future<HoneyBatch>
      - createBatch(data, pdfFile) → Future<HoneyBatch> (multipart upload)
      - updateBatch(id, data) → Future<HoneyBatch>
      - deleteBatch(id) → Future<void>
      - verifyBatch(id) → Future<VerificationDetails>
      - getQRCode(id) → Future<String> (URL to QR image)
    - Error handling: Convert DioException to ApiException

### Phase 8: Frontend State Management (1 task)
**Goals:** Cubit for honey batch state.

29. **HC-FE-19: Honey BLoC/Cubit**
    - File: `app/lib/features/honey/cubit/honey_batches_cubit.dart` + `honey_batches_state.dart`
    - States: HoneyBatchesInitial, HoneyBatchesLoading, HoneyBatchesLoaded(batches, total, isLoadingMore), HoneyBatchesError(code)
    - Methods:
      - load(apiaryId) → emit Loading then Loaded
      - loadMore() → append to existing list
      - create(data, file) → emit Loading, update state
      - delete(id) → remove from list, refresh if error
      - setApiaryFilter(apiaryId) → reload
    - Pagination: track offset, hasMore flag

### Phase 9: Frontend Core Screens (7 tasks)
**Goals:** User-facing UI for creating, viewing, and verifying batches.

30. **HC-FE-01 & HC-FE-07: Honey batches home screen + My honey batches**
    - File: `app/lib/features/honey/view/honey_batches_screen.dart`
    - Main screen with Cubit setup
    - Displays list of user's batches
    - "Create Batch" button → navigate to create screen
    - List items show: batch name/type, date, blockchain status (badge)
    - Tap item → detail screen
    - Pull-to-refresh support

31. **HC-FE-02: Create honey batch screen**
    - File: `app/lib/features/honey/view/create_honey_batch_screen.dart`
    - Form fields:
      - Gathering date (date picker)
      - Amount in kg (text input, numeric)
      - Processing method (dropdown: raw/filtered/pasteurized)
      - Honey type (text input, autocomplete suggestions)
      - Apiary selector (dropdown, pre-filled if from apiary detail)
      - PDF upload (file picker, show file name + size)
    - Submit button: calls cubit.create(), shows progress
    - Error toast if validation fails
    - Success → pop screen + reload parent

32. **HC-FE-03: Honey batch detail screen**
    - File: `app/lib/features/honey/view/honey_batch_detail_screen.dart`
    - Displays: all batch info, gathering date, amount, processing method, honey type, apiary name
    - PDF section: preview/download link
    - QR code section: display QR image, "Share" button
    - Blockchain status section: "Pending"/"✓ Verified"/"✗ Failed" badge + details button
    - Edit button (if user owns batch) → edit screen
    - Delete button (if user owns batch)

33. **HC-FE-04: Honey batch verification screen**
    - File: `app/lib/features/honey/view/honey_batch_verification_screen.dart`
    - Displays verification details from verifyBatch() endpoint:
      - "✓ Verified on Polygon" (with timestamp)
      - "PDF hash matches"
      - Transaction hash (clickable link to Polygonscan)
      - Block number + confirmation count
      - Metadata hash displayed
    - Refresh button: re-query blockchain

34. **HC-FE-05: QR code display screen**
    - File: `app/lib/features/honey/view/qr_code_display_screen.dart`
    - Full-screen QR code image
    - "Share" button → share QR image via OS share sheet
    - Long-press context menu: "Save as image" (to gallery)

35. **HC-FE-06: QR code scanner screen**
    - File: `app/lib/features/honey/view/qr_code_scanner_screen.dart`
    - Uses `mobile_scanner` or `qr_code_scanner` package
    - Live camera preview
    - Detect QR code, extract batch ID from URL
    - Navigate to verification screen
    - Handle errors: invalid QR, network error, batch not found

36. **HC-FE-18: Add Honey Batches section to hive detail**
    - File: `app/lib/features/hive/view/hive_detail_screen.dart` (modify existing)
    - Add tab or collapsible section "Honey"
    - Display batches from this hive
    - "Create Batch" button opens create screen with hive pre-selected

### Phase 10: Frontend Utils & Widgets (4 tasks)
**Goals:** Shared components and helper functions.

37. **HC-FE-11 & HC-FE-12: QR code generation + scanner**
    - File: `app/lib/features/honey/utils/qr_utils.dart`
    - generateQRImageUrl(batchID) → returns URL to QR endpoint
    - scanQRCodeFromCamera() → uses mobile_scanner
    - extractBatchIDFromQRData(qrData) → parses URL to get ID

38. **HC-FE-13 & HC-FE-14: PDF preview + upload UI**
    - File: `app/lib/features/honey/widgets/pdf_upload_widget.dart` + `pdf_preview_widget.dart`
    - PdfUploadWidget: file picker, show name + size, upload progress
    - PdfPreviewWidget: embed PDF.js or use pdf_viewer_plugin, link to download
    - Both handle errors gracefully

39. **HC-FE-15 & HC-FE-16: Blockchain status indicator + verification modal**
    - File: `app/lib/features/honey/widgets/blockchain_status_widget.dart` + `verification_modal.dart`
    - BlockchainStatusWidget: badge with "Pending"/"✓ Verified"/"✗ Failed" + icon
    - VerificationModal: detailed view of on-chain state

40. **HC-FE-17: Hash comparison display**
    - File: `app/lib/features/honey/widgets/hash_comparison_widget.dart`
    - Side-by-side display: "On-chain hash: xyz..." and "Current PDF hash: xyz..."
    - Monospace font for hashes
    - Highlight match/mismatch

### Phase 11: Polish & Edge Cases (10 tasks)
**Goals:** Production-ready error handling, offline support, localization.

41. **HC-10-01: PDF file validation (backend)**
    - File: `backend/internal/service/honey_batch.go` (add to HC-BE-13)
    - Check MIME type: application/pdf only
    - Max size: 10MB
    - Optional: scan with ClamAV if security required

42. **HC-10-02: Blockchain retry logic (backend)**
    - File: `backend/internal/service/honey_batch.go`
    - If blockchain write fails: retry with exponential backoff (1s, 2s, 4s, 8s max)
    - After 3 failed retries, set blockchain_status = "failed", notify user
    - PollPendingBatchStatuses can attempt recovery for failed txs

43. **HC-10-03: PDF storage (backend)**
    - File: `backend/internal/storage/pdf_storage.go`
    - Use S3 or local file system
    - Generate signed URLs for secure access (public batches: no signature)
    - Cleanup: soft-delete PDFs when batch deleted
    - Security: validate PDF MIME before storing

44. **HC-10-04: Gas fee management (backend)**
    - File: `backend/internal/blockchain/writer.go`
    - Log transaction costs (gas used × gas price)
    - Consider gas relay service (OpenZeppelin Defender) for user-pays-nothing model (optional phase 2)
    - Alert if gas price spikes

45. **HC-10-05: Offline handling (frontend)**
    - File: `app/lib/features/honey/cubit/honey_batches_cubit.dart`
    - Cache batch list locally (Hive/SharedPreferences)
    - Show "Verification pending" if offline and status not confirmed
    - Retry verification when network returns

46. **HC-10-06: Loading states (frontend)**
    - File: `app/lib/features/honey/view/honey_batch_*.dart`
    - Show spinner while blockchain tx pending
    - Allow user to check status later without blocking UI
    - "Check status" button on pending batches

47. **HC-10-07: Error handling (frontend)**
    - File: `app/lib/core/widgets/error_dialog.dart` (modify existing)
    - Display user-friendly blockchain errors:
      - "Verification pending — try again in 30 seconds"
      - "Contract error — contact support"
      - "Network error — check your connection"
    - Provide retry buttons

48. **HC-10-08: Localization (frontend)**
    - File: `app/lib/l10n/app_en.arb` + `app_pl.arb`
    - Add keys: processingMethod_raw, processingMethod_filtered, processingMethod_pasteurized
    - Add keys: blockchainStatus_pending, blockchainStatus_confirmed, blockchainStatus_failed
    - Add keys: verification_verified, verification_details, verification_pending
    - Run `flutter gen-l10n` to regenerate

49. **HC-10-09: Empty states (frontend)**
    - File: `app/lib/features/honey/view/honey_batches_screen.dart`
    - "No honey batches yet" screen with "Create Batch" CTA
    - "No verified batches" filter view

50. **HC-10-10: Database indexing (backend)**
    - File: `backend/migrations/030_create_honey_batch_qr_codes.sql` (or new migration)
    - Add indexes:
      - honey_batches: (user_id, created_at DESC)
      - honey_batches: (apiary_id, created_at DESC)
      - honey_batches: (blockchain_status, created_at DESC)
    - Improves query performance for list/filter operations

---

## Critical Files Summary

### New Files to Create
- **Backend:**
  - `backend/migrations/029_create_honey_batches.sql`
  - `backend/migrations/030_create_honey_batch_qr_codes.sql`
  - `backend/internal/model/honey_batch.go`
  - `backend/internal/repository/honey_batch.go`
  - `backend/internal/handler/honey_batch.go`
  - `backend/internal/service/honey_batch.go`
  - `backend/internal/blockchain/writer.go`
  - `backend/internal/blockchain/reader.go`
  - `backend/internal/blockchain/hash.go`
  - `backend/internal/config/blockchain_config.go`
  - `backend/internal/jobs/blockchain_poller.go`
  - `backend/contracts/HoneyCertification.sol` (or external repo)
  - `backend/internal/blockchain/contracts/HoneyCertification.abi` (generated)

- **Frontend:**
  - `app/lib/features/honey/` (new feature folder)
  - `app/lib/features/honey/data/honey_batch_model.dart`
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
  - `backend/cmd/api/main.go` — wire blockchain components, start poller job
  - `backend/internal/handler/handler.go` — register HoneyBatchHandler

- **Frontend:**
  - `app/lib/main.dart` — add honey feature to navigation (drawer)
  - `app/lib/features/hive/view/hive_detail_screen.dart` — add Honey tab
  - `app/lib/l10n/app_en.arb` + `app_pl.arb` — add localizations
  - `app/lib/l10n/app_localizations.dart` + `app_localizations_en.dart` + `app_localizations_pl.dart` — regenerated by flutter gen-l10n

---

## Implementation Sequence (Recommended Order)

### Week 1: Foundation
1. Create migrations (HC-DB-01, HC-DB-02)
2. Create Go models (HC-BE-07, HC-BE-08)
3. Create repository (HC-BE-09 to HC-BE-12)
4. Wire into main.go and test DB operations

### Week 2: Blockchain
5. Blockchain config (HC-BE-01)
6. Smart contract (HC-BE-02, HC-BE-03) — deploy to Amoy
7. Writer & reader (HC-BE-04, HC-BE-05, HC-BE-06)
8. Integration tests for blockchain operations

### Week 3: Backend Service & API
9. Service layer (HC-BE-13 to HC-BE-16)
10. Handler layer (HC-BE-17 to HC-BE-24)
11. Background job (HC-BE-15, wire into main.go)
12. End-to-end API tests

### Week 4: Frontend Foundation
13. Dart models & repository (HC-FE-08, HC-FE-09, HC-FE-10)
14. State management (HC-FE-19)
15. Core screens (HC-FE-01, HC-FE-02, HC-FE-03)

### Week 5: Frontend Completion
16. Verification & QR screens (HC-FE-04, HC-FE-05, HC-FE-06)
17. Widgets (HC-FE-11 to HC-FE-17)
18. Integration with hive detail (HC-FE-18)
19. Localization (HC-10-08)

### Week 6: Polish & Edge Cases
20. Error handling, offline support, empty states
21. Testing on Android + Web
22. Performance & UX polish

---

## Testing Strategy

### Backend Tests
- **Unit:** honey_batch_service_test.go, blockchain_writer_test.go (mock RPC calls)
- **Integration:** Test DB migrations, repository CRUD, blockchain poller
- **End-to-end:** POST /api/v1/honey-batches with mock file → verify DB + blockchain state

### Frontend Tests
- **Unit:** Cubit logic, model serialization (golden tests for JSON)
- **Widget:** Create screen form validation, QR display
- **Integration:** Create batch → list batches → verify → delete

### Blockchain Tests
- Deploy contract to Amoy testnet
- Test certify() function, event emission
- Test hash verification (on-chain vs stored)
- Test failed tx retry logic

---

## Verification

### Manual Testing Checklist
1. ✅ Create honey batch with PDF upload
2. ✅ Blockchain status transitions: pending → confirmed
3. ✅ Verify batch: hash matches, tx confirmed
4. ✅ Generate & scan QR code
5. ✅ List batches by apiary + pagination
6. ✅ Edit batch notes (no blockchain changes)
7. ✅ Delete batch (soft delete)
8. ✅ Offline: batch listed, verification pending
9. ✅ Error cases: invalid PDF, network failure, gas exhaustion

### Performance Baselines
- Create batch: < 5s (API response, async blockchain)
- List batches (20 items): < 2s
- Blockchain poller: runs every 30s, finishes in < 5s
- QR generation: < 1s

---

## Known Constraints & Decisions

1. **Blockchain costs:** Using Polygon (cheap gas) instead of Ethereum mainnet
2. **Off-chain storage:** PDF stored in S3 or local FS; only hash on-chain for cost efficiency
3. **Async blockchain:** Create batch returns immediately; blockchain write happens in background
4. **No smart contract UI:** Users don't interact directly with contract; BeeTrack backend is sole caller
5. **Soft deletes:** Batches marked deleted, not hard-deleted (audit trail)
6. **QR code immutability:** URLs permanent; scanning always verifies current blockchain state

---

## Decisions Made

1. **PDF Storage:** Match photo storage pattern (currently URL-based). PDFs stored as URLs in DB, can be backed by S3, cloud storage, or local file system per deployment
2. **Blockchain Network:** **Polygon Amoy testnet** (development, chain ID 80002) — costs fake MATIC, safe to test. Can migrate to Polygon mainnet later once validated
3. **Gas Fees:** **App/backend pays** — simpler UX, no wallet required from users
4. **Scope:** **Full feature (50 tasks, 6 weeks)** — include polish, error handling, offline support, localization
5. **Verification UI:** Public endpoint for verification; QR scanning via app
