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
9. [Marketplace — Sale Announcements](#9-marketplace--sale-announcements)
10. [Honey Certification & Blockchain](#10-honey-certification--blockchain)

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
| -------- | ----- | ------ | -------------------------------------- | --------------------------------------------------- |
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
| --------- | ----- | ------ | -------------------------- | ------------------------------------------- |
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

## 9. Marketplace — Sale Announcements

> Public marketplace for beekeeping products/services. Listings require auth to create; viewing is public. Includes search, filters, favorites, and map display.

### 9.1 Database Schema

| ID        | Layer | Status | Title                            | Notes                                                                                                                                                                      |
| --------- | ----- | ------ | -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| MKT-DB-01 | `DB`  | `[x]`  | Create `listings` table          | id, user_id, title, description, category, price, quantity, address, apiary_id, contact_phone, contact_email, is_hidden, created_at, updated_at                            |
| MKT-DB-02 | `DB`  | `[x]`  | Create `listing_images` table    | id, listing_id, image_url, display_order, created_at                                                                                                                       |
| MKT-DB-03 | `DB`  | `[x]`  | Create `listing_favorites` table | id, user_id, listing_id, created_at                                                                                                                                        |
| MKT-DB-04 | `DB`  | `[x]`  | Create `listing_categories` enum | TEXT + CHECK constraint (codebase convention); HONEY, POLLEN, BEE_COLONIES, QUEEN_BEES, BEEHIVES, POPULATED_BEEHIVES, EQUIPMENT, EXTRACTION_EQUIPMENT, FEED, SUPPLIES, WAX_FOUNDATION, BEESWAX, PROPOLIS, SERVICES, OTHER |

### 9.2 Backend — Models & Persistence

| ID        | Layer | Status | Title                             | Notes                                                                                                                                           |
| --------- | ----- | ------ | --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| MKT-BE-01 | `BE`  | `[ ]`  | Model: `Listing` struct           | id, user_id, title, description, category, price, quantity, address, apiary_id, contact_phone, contact_email, is_hidden, created_at, updated_at |
| MKT-BE-02 | `BE`  | `[ ]`  | Model: `ListingImage` struct      | id, listing_id, image_url, display_order, created_at                                                                                            |
| MKT-BE-03 | `BE`  | `[ ]`  | Repository: Create listing        | Insert + associated images                                                                                                                      |
| MKT-BE-04 | `BE`  | `[ ]`  | Repository: Get listing by ID     | With images and apiary details                                                                                                                  |
| MKT-BE-05 | `BE`  | `[ ]`  | Repository: List/search listings  | Filters: category, price_min/max, keyword, date_range, distance (if location provided), hidden status (only own)                                |
| MKT-BE-06 | `BE`  | `[ ]`  | Repository: Update listing        | Title, description, category, price, quantity, address, contact_phone, contact_email, is_hidden                                                 |
| MKT-BE-07 | `BE`  | `[ ]`  | Repository: Hide/show listing     | Toggle is_hidden (soft delete, not permanent)                                                                                                   |
| MKT-BE-08 | `BE`  | `[ ]`  | Repository: Delete listing images | Remove old images before update                                                                                                                 |

### 9.3 Backend — Business Logic

| ID        | Layer | Status | Title                           | Notes                                                                        |
| --------- | ----- | ------ | ------------------------------- | ---------------------------------------------------------------------------- |
| MKT-BE-09 | `BE`  | `[ ]`  | Service: Create listing         | Validate auth, images (max 3), required fields; return listing with images   |
| MKT-BE-10 | `BE`  | `[ ]`  | Service: Update listing         | Validate ownership, handle image updates                                     |
| MKT-BE-11 | `BE`  | `[ ]`  | Service: Get listing            | Check if hidden; allow owner or public view; include apiary info if attached |
| MKT-BE-12 | `BE`  | `[ ]`  | Service: Search/filter listings | Build dynamic query based on filters; exclude hidden for non-owners          |

### 9.4 Backend — API Handlers

| ID        | Layer | Status | Title                                     | Notes                                                             |
| --------- | ----- | ------ | ----------------------------------------- | ----------------------------------------------------------------- |
| MKT-BE-13 | `BE`  | `[ ]`  | Handler: POST /api/v1/listings            | Create listing (auth required)                                    |
| MKT-BE-14 | `BE`  | `[ ]`  | Handler: GET /api/v1/listings             | Search/filter (public, excludes hidden)                           |
| MKT-BE-15 | `BE`  | `[ ]`  | Handler: GET /api/v1/listings/{id}        | Get single listing (public)                                       |
| MKT-BE-16 | `BE`  | `[ ]`  | Handler: PATCH /api/v1/listings/{id}      | Update listing (auth + ownership required)                        |
| MKT-BE-17 | `BE`  | `[ ]`  | Handler: PATCH /api/v1/listings/{id}/hide | Hide listing (toggle is_hidden; auth + ownership)                 |
| MKT-BE-18 | `BE`  | `[ ]`  | Handler: DELETE /api/v1/listings/{id}     | Delete listing (auth + ownership required)                        |
| MKT-BE-19 | `BE`  | `[ ]`  | Handler: Image upload endpoint            | Multipart POST, validate MIME type, store in S3/local, return URL |

### 9.5 Frontend — Core Screens

| ID        | Layer | Status | Title                   | Notes                                                                                                                |
| --------- | ----- | ------ | ----------------------- | -------------------------------------------------------------------------------------------------------------------- |
| MKT-FE-01 | `FE`  | `[ ]`  | Marketplace home screen | Feed of listings with search/filter UI, category chips, map button                                                   |
| MKT-FE-02 | `FE`  | `[ ]`  | Listing detail screen   | Full details, images carousel, contact info, apiary summary (if attached), add to favorites button                   |
| MKT-FE-03 | `FE`  | `[ ]`  | Create listing screen   | Form: title, description, category, price, quantity, address, contact_phone, contact_email, attach apiary (optional) |
| MKT-FE-04 | `FE`  | `[ ]`  | Edit listing screen     | Reuse create form, pre-filled with existing data, image management                                                   |
| MKT-FE-05 | `FE`  | `[ ]`  | My listings screen      | Show all user's listings (including hidden), edit/delete/hide actions                                                |
| MKT-FE-06 | `FE`  | `[ ]`  | Favorites screen        | Saved listings, add to favorites from detail view                                                                    |

### 9.6 Frontend — Search & Filters

| ID        | Layer | Status | Title                    | Notes                                                        |
| --------- | ----- | ------ | ------------------------ | ------------------------------------------------------------ |
| MKT-FE-07 | `FE`  | `[ ]`  | Category filter          | Chips for each category, multi-select UI                     |
| MKT-FE-08 | `FE`  | `[ ]`  | Price range slider       | Min/max inputs or slider widget                              |
| MKT-FE-09 | `FE`  | `[ ]`  | Keyword search           | Text field, real-time or search button                       |
| MKT-FE-10 | `FE`  | `[ ]`  | Date range filter        | Posted within last X days (or date picker)                   |
| MKT-FE-11 | `FE`  | `[ ]`  | Distance/location filter | If user location available, show radius filter; else disable |

### 9.7 Frontend — Map Integration

| ID        | Layer | Status | Title                | Notes                                                                      |
| --------- | ----- | ------ | -------------------- | -------------------------------------------------------------------------- |
| MKT-FE-12 | `FE`  | `[ ]`  | Map screen           | Show listings with apiary attached as pins; tap pin to view listing        |
| MKT-FE-13 | `FE`  | `[ ]`  | Map filters          | Filter map pins by category, price range, same filters as feed             |
| MKT-FE-14 | `FE`  | `[ ]`  | Distance calculation | Show distance from user to listing apiary (if location permission granted) |

### 9.8 Frontend — Image Management

| ID        | Layer | Status | Title                      | Notes                                                               |
| --------- | ----- | ------ | -------------------------- | ------------------------------------------------------------------- |
| MKT-FE-15 | `FE`  | `[ ]`  | Image picker (create/edit) | Select up to 3 images from gallery or camera, preview before upload |
| MKT-FE-16 | `FE`  | `[ ]`  | Image carousel on detail   | Swipeable carousel for multiple images                              |
| MKT-FE-17 | `FE`  | `[ ]`  | Image upload progress      | Show upload progress/loading state                                  |

### 9.9 Frontend — Data Models & Repositories

| ID        | Layer | Status | Title               | Notes                                        |
| --------- | ----- | ------ | ------------------- | -------------------------------------------- |
| MKT-FE-18 | `FE`  | `[ ]`  | ListingModel (Dart) | Mirrors Listing struct from backend          |
| MKT-FE-19 | `FE`  | `[ ]`  | ListingRepository   | CRUD + search methods, call backend handlers |
| MKT-FE-20 | `FE`  | `[ ]`  | FavoritesRepository | Add/remove favorite, list favorites          |

### 9.10 Frontend — Navigation & State

| ID        | Layer | Status | Title                             | Notes                                                 |
| --------- | ----- | ------ | --------------------------------- | ----------------------------------------------------- |
| MKT-FE-21 | `FE`  | `[ ]`  | Add Marketplace tab to bottom nav | Between Hives and Account (or similar placement)      |
| MKT-FE-22 | `FE`  | `[ ]`  | Marketplace BLoC/Cubit            | Manage listings, search state, filters                |
| MKT-FE-23 | `FE`  | `[ ]`  | Handle public vs. auth views      | Show "Create Listing" button for logged-in users only |

### 9.11 Polish & Edge Cases

| ID        | Layer   | Status | Title                        | Notes                                                                                   |
| --------- | ------- | ------ | ---------------------------- | --------------------------------------------------------------------------------------- |
| MKT-09-01 | `BE/FE` | `[ ]`  | Soft delete validation       | Ensure deleted listings don't appear in search; verify ownership on hide                |
| MKT-09-02 | `FE`    | `[ ]`  | Empty states                 | "No listings yet", "No favorites yet" screens                                           |
| MKT-09-03 | `FE`    | `[ ]`  | Pagination / infinite scroll | For search results (backend limit, frontend pagination)                                 |
| MKT-09-04 | `BE`    | `[ ]`  | Add index on listings table  | category, user_id, created_at, is_hidden for query performance                          |
| MKT-09-05 | `FE`    | `[ ]`  | Confirmation dialogs         | Before delete/hide listing                                                              |
| MKT-09-06 | `BE/FE` | `[ ]`  | Prevent self-contact         | If listing has apiary, show apiary owner info (not user contact) if viewing own listing |
| MKT-09-07 | `FE`    | `[ ]`  | Localization                 | Add l10n strings for all UI text (categories, labels, filters)                          |

---

## 10. Honey Certification & Blockchain

> Immutable honey batch certification stored on Polygon blockchain. Each batch gets a QR code that verifies authenticity via blockchain hash of lab PDF.
>
> **Blockchain Strategy:** Store minimal data on-chain (hash, metadata hash, timestamp) for cost efficiency. PDF hash links to lab report; scanning verifies hash hasn't changed.

### 10.1 Database Schema

| ID       | Layer | Status | Title                               | Notes                                                                                                                                                                                                                                                                                                                        |
| -------- | ----- | ------ | ----------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HC-DB-01 | `DB`  | `[ ]`  | Create `honey_batches` table        | id, user_id, apiary_id, gathering_date, amount (kg), processing_method (enum: raw/filtered/pasteurized), honey_type (text: wildflower/acacia/etc), lab_pdf_url, pdf_file_hash (SHA256), metadata_hash, blockchain_tx_hash, blockchain_contract_address, blockchain_status (pending/confirmed/failed), created_at, updated_at |
| HC-DB-02 | `DB`  | `[ ]`  | Create `honey_batch_qr_codes` table | id, batch_id, qr_code_data (URL), created_at                                                                                                                                                                                                                                                                                 |

### 10.2 Backend — Blockchain Integration

| ID       | Layer | Status | Title                      | Notes                                                                                                                 |
| -------- | ----- | ------ | -------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| HC-BE-01 | `BE`  | `[ ]`  | Blockchain config          | Polygon RPC endpoint, contract address, private key management (env vars)                                             |
| HC-BE-02 | `BE`  | `[ ]`  | Smart contract (Solidity)  | Simple registry: store (batch_id, pdf_hash, metadata_hash, timestamp, owner_address); emit CertificationCreated event |
| HC-BE-03 | `BE`  | `[ ]`  | Deploy contract to Polygon | Testnet first (Mumbai), then mainnet                                                                                  |
| HC-BE-04 | `BE`  | `[ ]`  | Blockchain writer          | Function to hash batch metadata, call contract.certify() with PDF hash, wait for tx confirmation                      |
| HC-BE-05 | `BE`  | `[ ]`  | Blockchain reader          | Function to query contract state; check if tx_hash exists and is confirmed; return stored hash                        |
| HC-BE-06 | `BE`  | `[ ]`  | Hash utilities             | SHA256 for PDF file + batch metadata; ensure consistent hashing                                                       |

### 10.3 Backend — Models & Persistence

| ID       | Layer | Status | Title                                   | Notes                                                                       |
| -------- | ----- | ------ | --------------------------------------- | --------------------------------------------------------------------------- |
| HC-BE-07 | `BE`  | `[ ]`  | Model: `HoneyBatch` struct              | Mirrors DB schema                                                           |
| HC-BE-08 | `BE`  | `[ ]`  | Model: `ProcessingMethod` enum          | raw, filtered, pasteurized                                                  |
| HC-BE-09 | `BE`  | `[ ]`  | Repository: Create batch                | Insert batch + store blockchain metadata (before tx)                        |
| HC-BE-10 | `BE`  | `[ ]`  | Repository: Get batch by ID             | Fetch from DB + fetch blockchain verification status                        |
| HC-BE-11 | `BE`  | `[ ]`  | Repository: List batches by user/apiary | Filter by user_id or apiary_id                                              |
| HC-BE-12 | `BE`  | `[ ]`  | Repository: Update batch status         | When blockchain tx confirmed, update blockchain_status + blockchain_tx_hash |

### 10.4 Backend — Business Logic

| ID       | Layer | Status | Title                           | Notes                                                                                                    |
| -------- | ----- | ------ | ------------------------------- | -------------------------------------------------------------------------------------------------------- |
| HC-BE-13 | `BE`  | `[ ]`  | Service: Create honey batch     | Validate auth, apiary ownership, calculate PDF hash, trigger blockchain write, return batch with tx_hash |
| HC-BE-14 | `BE`  | `[ ]`  | Service: Get batch + verify     | Fetch batch from DB, query blockchain for verification status, compare PDF hash                          |
| HC-BE-15 | `BE`  | `[ ]`  | Service: Poll blockchain status | Background job: check pending txs, update status when confirmed                                          |
| HC-BE-16 | `BE`  | `[ ]`  | Service: Generate QR code data  | Create URL (e.g., https://beetrack.app/verify/batch/{id}); encode as QR                                  |

### 10.5 Backend — API Handlers

| ID       | Layer | Status | Title                                           | Notes                                                                                 |
| -------- | ----- | ------ | ----------------------------------------------- | ------------------------------------------------------------------------------------- |
| HC-BE-17 | `BE`  | `[ ]`  | Handler: POST /api/v1/honey-batches             | Create batch (auth required; multipart form: data + PDF file)                         |
| HC-BE-18 | `BE`  | `[ ]`  | Handler: GET /api/v1/honey-batches/{id}         | Get batch + blockchain verification (public endpoint)                                 |
| HC-BE-19 | `BE`  | `[ ]`  | Handler: GET /api/v1/honey-batches/{id}/verify  | Detailed verification: on-chain hash vs PDF hash, tx status                           |
| HC-BE-20 | `BE`  | `[ ]`  | Handler: GET /api/v1/honey-batches              | List batches (filters: user_id, apiary_id, honey_type, auth-protected)                |
| HC-BE-21 | `BE`  | `[ ]`  | Handler: GET /api/v1/honey-batches/{id}/qr-code | Generate/return QR code image (PNG/SVG)                                               |
| HC-BE-22 | `BE`  | `[ ]`  | Handler: PATCH /api/v1/honey-batches/{id}       | Update batch (auth + ownership; title, notes only; cannot change PDF/blockchain data) |
| HC-BE-23 | `BE`  | `[ ]`  | Handler: DELETE /api/v1/honey-batches/{id}      | Soft delete batch (auth + ownership required)                                         |
| HC-BE-24 | `BE`  | `[ ]`  | Handler: POST /api/v1/honey-batches/{id}/pdf    | Retrieve lab PDF URL (public)                                                         |

### 10.6 Frontend — Core Screens

| ID       | Layer | Status | Title                           | Notes                                                                                                             |
| -------- | ----- | ------ | ------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| HC-FE-01 | `FE`  | `[ ]`  | Honey batches home screen       | Tab in hive detail or separate "Honey" section; list user's batches + create button                               |
| HC-FE-02 | `FE`  | `[ ]`  | Create honey batch screen       | Form: gathering_date, amount (kg), processing_method (dropdown), honey_type (text), attach apiary, upload lab PDF |
| HC-FE-03 | `FE`  | `[ ]`  | Honey batch detail screen       | Show: all batch info, PDF preview/download, QR code display, blockchain verification status                       |
| HC-FE-04 | `FE`  | `[ ]`  | Honey batch verification screen | Display: "✓ Verified on Polygon", "PDF hash matches", tx link, timestamp                                          |
| HC-FE-05 | `FE`  | `[ ]`  | QR code display screen          | Show QR code + "Share" button; long-press/right-click to save as image                                            |
| HC-FE-06 | `FE`  | `[ ]`  | QR code scanner screen          | Use camera, scan QR, navigate to verification screen                                                              |
| HC-FE-07 | `FE`  | `[ ]`  | My honey batches screen         | List all user's certified batches, edit/delete/view details, show blockchain status (pending/confirmed)           |

### 10.7 Frontend — Data Models & Repositories

| ID       | Layer | Status | Title                       | Notes                                          |
| -------- | ----- | ------ | --------------------------- | ---------------------------------------------- |
| HC-FE-08 | `FE`  | `[ ]`  | HoneyBatchModel (Dart)      | Mirrors HoneyBatch struct from backend         |
| HC-FE-09 | `FE`  | `[ ]`  | ProcessingMethodEnum (Dart) | raw, filtered, pasteurized with display labels |
| HC-FE-10 | `FE`  | `[ ]`  | HoneyBatchRepository        | CRUD + get verification status, download PDF   |

### 10.8 Frontend — QR Code & PDF Handling

| ID       | Layer | Status | Title                  | Notes                                                                        |
| -------- | ----- | ------ | ---------------------- | ---------------------------------------------------------------------------- |
| HC-FE-11 | `FE`  | `[ ]`  | QR code generation     | Generate QR from batch ID URL; display as image in UI                        |
| HC-FE-12 | `FE`  | `[ ]`  | QR code scanner        | Use `qr_code_scanner` or `mobile_scanner` package; extract batch ID from URL |
| HC-FE-13 | `FE`  | `[ ]`  | PDF preview / download | Embed PDF viewer or link to download; show in detail screen                  |
| HC-FE-14 | `FE`  | `[ ]`  | PDF upload UI          | File picker; show file name + size before upload; upload progress indicator  |

### 10.9 Frontend — Verification UI

| ID       | Layer | Status | Title                       | Notes                                                                                       |
| -------- | ----- | ------ | --------------------------- | ------------------------------------------------------------------------------------------- |
| HC-FE-15 | `FE`  | `[ ]`  | Blockchain status indicator | Show "Pending", "✓ Verified", or "✗ Failed" with visual icon                                |
| HC-FE-16 | `FE`  | `[ ]`  | Verification details modal  | Show: tx hash (clickable link to Polygonscan), confirmed timestamp, PDF hash, metadata hash |
| HC-FE-17 | `FE`  | `[ ]`  | Hash comparison display     | Show "On-chain hash: xyz..." and "Current PDF hash: xyz..." side-by-side                    |

### 10.10 Frontend — Navigation & State

| ID       | Layer | Status | Title                                    | Notes                                                              |
| -------- | ----- | ------ | ---------------------------------------- | ------------------------------------------------------------------ |
| HC-FE-18 | `FE`  | `[ ]`  | Add Honey Batches section to hive detail | Tab or collapsible section with "Create Batch" button              |
| HC-FE-19 | `FE`  | `[ ]`  | Honey BLoC/Cubit                         | Manage batches, verification state, QR scanner state               |
| HC-FE-20 | `FE`  | `[ ]`  | Deep link for QR verification            | Support scanning external QR codes that link to /verify/batch/{id} |

### 10.11 Polish & Edge Cases

| ID       | Layer | Status | Title                  | Notes                                                                                  |
| -------- | ----- | ------ | ---------------------- | -------------------------------------------------------------------------------------- |
| HC-10-01 | `BE`  | `[ ]`  | PDF file validation    | Check MIME type, max size (e.g., 10MB), scan for malware (optional)                    |
| HC-10-02 | `BE`  | `[ ]`  | Blockchain retry logic | If tx fails, retry with exponential backoff; notify user if permanently failed         |
| HC-10-03 | `BE`  | `[ ]`  | Storage for PDFs       | Use S3/cloud storage or local file system; secure access control                       |
| HC-10-04 | `BE`  | `[ ]`  | Gas fee management     | Log tx costs; consider implementing gas relay for user (pay on behalf)                 |
| HC-10-05 | `FE`  | `[ ]`  | Offline handling       | Cache batch info locally; show "Verification pending" if network offline               |
| HC-10-06 | `FE`  | `[ ]`  | Loading states         | Show spinner while blockchain tx pending; allow user to check status later             |
| HC-10-07 | `FE`  | `[ ]`  | Error handling         | Display user-friendly blockchain errors (tx failed, contract error, network issue)     |
| HC-10-08 | `FE`  | `[ ]`  | Localization           | Add l10n for processing methods, blockchain status labels, verification text           |
| HC-10-09 | `FE`  | `[ ]`  | Empty states           | "No honey batches yet", "No verified batches" screens                                  |
| HC-10-10 | `BE`  | `[ ]`  | Database indexing      | Index on (user_id, created_at), (apiary_id), (blockchain_status) for query performance |
