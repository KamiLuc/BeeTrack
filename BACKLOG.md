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
4. [Bulk Operations](#4-bulk-operations)
5. [Queen Recognition (AI Feature)](#5-queen-recognition-ai-feature)
6. [Infrastructure & DevOps](#6-infrastructure--devops)
8. [Colony Strength & Queen Replacement](#8-colony-strength--queen-replacement)

> Voice logging and the MCP-based AI assistant moved out of this backlog — see
> [AI_ASSISTANT.md](AI_ASSISTANT.md) for the design and
> [BACKLOG_AI_ASSISTANT.md](BACKLOG_AI_ASSISTANT.md) for its tickets.

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

## 8. Colony Strength & Queen Replacement

> Three additions: a new `colony_strength` enum field on inspections (`very_weak`/`weak`/`medium`/`strong`/`very_strong` —
> Polish: Bardzo słaba/Słaba/Średnia/Silna/Bardzo silna), a new `box_added` boolean on inspections (Polish: "Dołożono
> korpus", mirrors the existing `queen_added`/"Poddano matkę"), and a new hive status flag `queen_needs_replacement`
> ("Matka do wymiany") — independent of `queenless` (a hive can have a queen that's failing and needs swapping,
> distinct from having no queen at all). Delivered in stages: backend core, then the AI layer (MCP tools + voice
> worker schemas), then Flutter UI + l10n.
>
> **Update:** `queenless` was later removed entirely and folded into `queen_needs_replacement`, since both cases
> call for the same beekeeper action (introduce a queen). The `queenless`/`Queenless` references below describe
> the state of the code as CS-01–CS-10 were built, not the current schema.

| ID       | Layer | Status | Title                                    | Notes |
| -------- | ----- | ------ | ------------------------------------------ | ----- |
| CS-01-BE | `BE`  | `[x]`  | Migrations + model fields                | `046_inspections_add_colony_strength.sql`, `047_hives_add_queen_needs_replacement.sql`, `048_inspections_add_box_added.sql` — `inspections.colony_strength VARCHAR(20)` (nullable, like `brood_pattern`), `inspections.box_added BOOLEAN NOT NULL DEFAULT FALSE` (like `queen_added`), `hives.queen_needs_replacement BOOLEAN NOT NULL DEFAULT FALSE` (like `queenless`) |
| CS-02-BE | `BE`  | `[x]`  | Service/repository/handler wiring        | Validation (`ValidColonyStrengths`, `ErrInvalidColonyStrength`), repository SQL update maps, handler request/response DTOs, apiary-clone transaction — same shape as `BroodPattern`/`QueenAdded`/`Queenless`. `HiveService.Add`/`.Update` gained a `queenNeedsReplacement bool` param |
| CS-03-BE | `BE`  | `[x]`  | PDF report + docs                        | `report_pdf.go`: `colonyStrengthLabel()` + a `box_added` → "Dołożono korpus" line; `docs/api.md` field docs + valid-values lists + new error codes |
| CS-04-BE | `BE`  | `[x]`  | AI layer: MCP tools + voice worker schema | `internal/mcp/` (hive_history, compare_hives, dashboard_summary, list_hives, list_hives_by_status) surface `colony_strength`/`queen_needs_replacement`, incl. a new dashboard counter and a `list_hives_by_status` filter case; `voice_worker.go`'s `hiveActionContext`/`createInspectionTool` schema gains `colony_strength`/`box_added`/`queen_needs_replacement` so Phase 2 proposals can use them (actually setting `queen_needs_replacement` via voice still needs `update_hive_status`, VC-11-BE — this only gets the field visible/proposable) |
| CS-05-FE | `FE`  | `[x]`  | Inspection form + detail screen          | New "Siła rodziny" dropdown in the Observations section (`_EnumDropdown`, mirrors Brood/Aggressiveness). "Poddano matkę"/"Dołożono korpus" ended up grouped with the frames-added/removed `_SignedFrameField`s (drawn/foundation/brood/feed — also "actions taken" during the inspection, not observed state) under a new "Akcje" section (new l10n key `inspectionSectionActions`), right after the absolute frame-count fields; `inspection_summary.dart` shows both in the detail view |
| CS-06-FE | `FE`  | `[x]`  | Hive status: toggle + badge + filter     | New `HiveQueenNeedsReplacementToggle` (mirrors `HiveQueenlessToggle`) on add/edit hive screens; status chip on hive detail; filter chip + grid icon on the apiary grid. On the inspection form it lives in the "Do zrobienia" section (see CS-08-FE) rather than Hive State |
| CS-07-FE | `FE`  | `[x]`  | l10n strings (PL/EN)                     | `hiveQueenNeedsReplacement`, `inspectionColonyStrength` + 5 values, `inspectionBoxAdded`, `inspectionSectionActions`, `inspectionSectionTodo` added to both `app_en.arb`/`app_pl.arb` |
| CS-08-FE | `FE`  | `[x]`  | To-do section (ready for harvest / needs food / queen needs replacement / box needs adding) | A real section on the inspection form, titled "Do zrobienia" (new l10n key `inspectionSectionTodo`), containing the actual `_BoolRow` toggles for `hiveReadyForHarvest`/`hiveNeedsFood`/`hiveQueenNeedsReplacement`/`hiveBoxNeedsAdding` — not a derived read-only text summary. Final form order (top to bottom): Observations → Frames → Akcje (queen/box added + frames-added deltas) → Notes → **Do zrobienia** → Stan ula (now just `hiveActive`, since `queenless` was removed — see CS-11) → Choroby. Not on the hive detail screen (redundant with its existing status chips) |
| CS-09-BE | `BE`  | `[x]`  | New hive flag: `box_needs_adding`        | "Wymaga dodania korpusu" — independent hive status flag mirroring `queen_needs_replacement`, all layers (migration, model, service `Add`/`Update` param, repository update map, handler DTOs + `hiveJSON`, apiary-clone transaction, MCP tools incl. dashboard counter + `list_hives_by_status` case, `voice_worker.go`'s `hiveActionContext`), plus docs/api.md |
| CS-10-FE | `FE`  | `[x]`  | New hive flag: FE + interaction          | `HiveBoxNeedsAddingToggle` on add/edit hive screens, status chip (with since-date) on hive detail, filter chip + grid icon on apiary grid + dashboard hive-status icons, toggle in the inspection form's "Do zrobienia" section. Toggling "Dołożono korpus" (Akcje section) true resets `box_needs_adding` to false client-side, mirroring "Poddano matkę" resetting `queen_needs_replacement`; the toggle is `enabled: !_boxAdded` the same way the queen one is `enabled: !_queenAdded` |
| CS-11-BE/FE | `BE`+`FE` | `[x]` | Remove `queenless`, fold into `queen_needs_replacement` | Both flags meant "hive needs a queen" (no queen vs. a failing one) — consolidated to one. Backend: `051_hives_drop_queenless.sql`, removed from model/service/repository/handler/MCP/voice-worker/seed/docs. Frontend: removed from `Hive` model, repository, `HiveQueenlessToggle` deleted, all screens/filters/icons/tests updated. Wording for the surviving flag changed PL "Matka do wymiany" → "Wymaga poddania matki" |
| CS-12-BE/FE | `BE`+`FE` | `[x]` | Status/disease "since" timestamps | New nullable `*_since` columns (`ready_for_harvest_since`, `queen_needs_replacement_since`, `needs_food_since`, `box_needs_adding_since`) on `hives`, computed server-side in `HiveService` via a `statusSince` helper: newly-true gets `now()`, still-true keeps its existing timestamp, false clears it. Exposed in `hiveJSON`. Flutter: added to `Hive`/`HiveDisease` (disease already had `created_at`, just wasn't consumed), status/disease chips on the hive detail screen now show `d.MM` alongside the label. Two bugs found/fixed post-review: (1) `cmd/seed/main.go` inserted hives directly via the repository, bypassing `HiveService` entirely, so every seeded status flag had a `null` since-date — added a `statusSinceIfTrue` seed helper mirroring the service's logic; (2) `EditHiveScreen._submit` discarded `HiveRepository.updateHive`'s response and reconstructed the popped `Hive` from pre-edit local state, so a status just toggled via that screen (not through an inspection) kept showing no date until the next full reload — `updateHive` now returns the server's fresh `Hive` (parses the PATCH response body) and `EditHiveScreen` pops `updated.copyWith(diseases: ...)` instead of hand-building one. Also added a "Do zrobienia" section (new `HiveSectionTitle` widget, reused `inspectionSectionTodo` l10n key) to the add/edit hive screens, grouping ready-for-harvest/needs-food/queen-needs-replacement/box-needs-adding in the same order as the inspection form, separate from "Aktywny" |
