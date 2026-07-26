# BeeTrack — AI Assistant Backlog

> Living document. Treat each item like a Jira ticket — update status as work progresses.
> Design doc: [AI_ASSISTANT.md](AI_ASSISTANT.md) — read that first; these tickets implement it
> section by section. Replaces the old Epic 5 (Voice Logging) / Epic 7 (MCP Server)
> placeholder rows removed from [BACKLOG.md](BACKLOG.md).

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

1. [Voice-Logged Inspections](#1-voice-logged-inspections)
2. [AI Apiary Assistant (Chat + MCP)](#2-ai-apiary-assistant-chat--mcp)
3. [Shared Infrastructure](#3-shared-infrastructure)

---

## 1. Voice-Logged Inspections

> Ref: [AI_ASSISTANT.md](AI_ASSISTANT.md) §2. Recording is queued and processed by a
> background worker (§2.1), not handled inline in the HTTP request.

| ID       | Layer | Status | Title                                            | Notes                                                                                                                                    |
| -------- | ----- | ------ | ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| VC-01-BE | `BE`  | `[ ]`  | Migration: `voice_recordings` + `voice_actions`  | §2.4 schema — status enum incl. `cancelled`, `audio_path`, `previous_hive_state` for hive-status undo                                     |
| VC-02-BE | `BE`  | `[ ]`  | `AUDIO_STORAGE_PATH` config + Docker volume       | Same UUID-filename convention as `IMAGE_STORAGE_PATH`; transient, worker deletes after transcription (§2.4)                                |
| VC-03-BE | `BE`  | `[ ]`  | `internal/llm/` — Anthropic API client wrapper    | Messages API + tool-use support; shared by voice worker and the chat assistant (§4)                                                        |
| VC-04-BE | `BE`  | `[ ]`  | Whisper client wrapper                            | Hosted STT call; returns transcript + per-segment confidence for the quality gate (§2.2)                                                   |
| VC-05-BE | `BE`  | `[ ]`  | `POST /apiaries/{id}/voice` — enqueue endpoint    | Multipart upload; validates `RECORDING_TOO_LONG` server-side; saves audio, inserts `pending` row, returns `202 Accepted` (§2.1, §2.3)      |
| VC-06-BE | `BE`  | `[ ]`  | Voice worker: poll loop + status lifecycle        | `pending → processing → completed/failed`, same shape as the blockchain worker; stuck-`processing` recovery/retry (§2.1, §2.4)             |
| VC-07-BE | `BE`  | `[ ]`  | Worker: transcription + audio-quality gate        | `POOR_AUDIO_QUALITY` short-circuit before Claude is ever called (§2.2)                                                                     |
| VC-08-BE | `BE`  | `[ ]`  | Worker Phase 1: hive-name resolution              | Calls `list_hives` (shared with AST-02-BE) filtered to the current apiary; `HIVE_NOT_IDENTIFIED` on no/ambiguous match (§2.1, §2.2)         |
| VC-09-BE | `BE`  | `[ ]`  | Worker Phase 2: write tools                       | `create_inspection` / `create_treatment` / `create_harvest` / `create_feeding` / `update_hive_status`, each calling the existing service layer (§2.1, §2.2) |
| VC-10-BE | `BE`  | `[ ]`  | Reuse autocomplete history as Claude context      | `GET /medicines`, `/feed-types`, `/feed-amounts` fed into Phase 2 so spoken values match existing entries instead of forking duplicates (§2.2) |
| VC-11-BE | `BE`  | `[ ]`  | `update_hive_status` tool + undo support           | Wraps `PATCH .../hives/{hiveId}` (flags + diseases); records `previous_hive_state` so Undo can restore instead of delete (§2.2, §2.6)       |
| VC-12-BE | `BE`  | `[ ]`  | Multi-action splitting in one recording           | Claude may emit 1..N write-tool calls per transcript; each executed/persisted independently (§2.1, §2.2)                                    |
| VC-13-BE | `BE`  | `[ ]`  | `GET /apiaries/{id}/voice-recordings`             | Paginated, newest first; nested `voice_actions` once `completed` (§2.6)                                                                     |
| VC-14-BE | `BE`  | `[ ]`  | `DELETE /apiaries/{id}/voice-recordings/{id}`     | Cancel — only while `status = 'pending'`; marks `cancelled`, deletes server-side audio file (§2.5)                                          |
| VC-15-BE | `BE`  | `[ ]`  | `DELETE /apiaries/{id}/voice-actions/{id}`        | Undo — deletes/reverts the underlying record via existing service calls, sets `reverted_at` (§2.6)                                          |
| VC-16-FE | `FE`  | `[ ]`  | Mic icon + recording dialog on `ApiaryGridScreen` | Bottom amber banner icon opens a modal (same pattern as `_HiveListDialog`); round record button inside (§2.1, §2.5)                          |
| VC-17-FE | `FE`  | `[ ]`  | Recording controls: toggle, hard cap, auto-stop   | Tap to start/stop; ~2.5s silence auto-stop with start grace period; hard cap (e.g. 3 min) with warning (§2.2)                               |
| VC-18-FE | `FE`  | `[ ]`  | Local audio storage + Play in pending list        | Save recording via `path_provider` before upload; Play button plays the local file (§2.5)                                                   |
| VC-19-FE | `FE`  | `[ ]`  | Cancel a pending recording (FE)                   | Calls VC-14-BE while `pending`; hidden once `processing` (§2.5)                                                                             |
| VC-20-FE | `FE`  | `[ ]`  | Voice Activity screen                             | History-icon button next to the mic icon; paginated recording/action list with status, edit-tap-through, undo (§2.6)                        |
| VC-21-FE | `FE`  | `[ ]`  | l10n strings for voice logging UI                 | Recording dialog, statuses, error copy (`NO_ACTION_RECOGNIZED`, `HIVE_NOT_IDENTIFIED`, `POOR_AUDIO_QUALITY`, `RECORDING_TOO_LONG`)           |

---

## 2. AI Apiary Assistant (Chat + MCP)

> Ref: [AI_ASSISTANT.md](AI_ASSISTANT.md) §3. Read-only in v1 — no write tools here;
> voice logging (Epic 1 above) is the write path.

| ID        | Layer | Status | Title                                    | Notes                                                                                                          |
| --------- | ----- | ------ | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------- |
| AST-01-BE | `BE`  | `[x]`  | `internal/mcp/` package + tool registry  | In-process, no separate deployment; tools call straight into existing repositories (§3.2)                       |
| AST-02-BE | `BE`  | `[x]`  | Tool: `list_hives`                       | Optional `apiary_id` filter, status flags, and active diseases — also called directly by VC-08-BE for hive-name resolution (§3.3). Excludes inactive hives (added alongside AST-05-BE, applies to every multi-hive tool via the shared `resolveHives` helper) |
| AST-03-BE | `BE`  | `[x]`  | Tools: `list_hive_records` + `get_hive_summary` | Deviates from the original single-tool sketch: `Treatment` has no active/inactive concept in the schema, so `list_hive_records(hive_id, record_types?, days?)` replaces a fixed "latest"/"active" shape with a recency window, and a `record_types` array lets Claude fetch several categories (e.g. inspections + feedings) in one call instead of one per type. `get_hive_summary` aggregates all four plus status flags/diseases over the same optional day window. (Originally shipped as 4 separate per-type tools, then collapsed into one — see AST-04-BE's note for why.) |
| AST-04-BE | `BE`  | `[x]`  | Tool: `list_hives_missing_records`       | Originally shipped as 3 separate tools (`list_untreated_hives`/`list_uninspected_hives`/`list_unfed_hives`), then collapsed into one `list_hives_missing_records(apiary_id?, record_types?, days?)` for the same reason as AST-03-BE: near-identical tools differing only by record type waste a tool call per category and clutter the tool list Claude has to choose from. Each result lists which requested types it's actually missing (§3.3) |
| AST-05-BE | `BE`  | `[x]`  | Tool: `list_hives_by_status`             | Widened from just `needs_food` to `(apiary_id?, statuses?)` covering queenless/needs_food/sick/ready_for_harvest — same collapse rationale as AST-03/04-BE. `sick` means the hive has an active disease (from `list_hives`' existing disease lookup); omitting `statuses` matches any of the four. Also the ticket where inactive-hive exclusion was added to the shared `resolveHives` helper, so `list_hives`/`list_hives_missing_records`/this tool all stopped surfacing inactive hives — single-hive lookups (`get_hive_summary`, `list_hive_records`) are unaffected since they authorize by hive_id directly, not through `resolveHives` (§3.3) |
| AST-06-BE | `BE`  | `[x]`  | Tool: `compare_hives`                    | `compare_hives(hive_ids)`: status flags, diseases, and each hive's latest inspection, treatment, feeding, and harvest — widened beyond the original inspection-only sketch since a "compare hive 3 and hive 7" question benefits from the full recent-activity picture, not just brood pattern. Authorized per hive_id — fails entirely if any hive_id isn't the caller's, same as get_hive_summary/list_hive_records (§3.3) |
| AST-07-BE | `BE`  | `[ ]`  | Tool: `search_listings`                  | Wraps existing public listing search/filter (§3.3, §3.6)                                                        |
| AST-08-BE | `BE`  | `[ ]`  | Tool: `get_listing`                      | Single listing detail for follow-up questions (§3.3)                                                            |
| AST-09-BE | `BE`  | `[ ]`  | `POST /assistant/messages` — agent loop  | Claude Messages API + tool-use loop scoped to caller's `userID`; streamed response (§3.2, §3.4)                 |
| AST-10-BE | `BE`  | `[ ]`  | HTTP+SSE MCP transport at `/mcp`         | For future external MCP clients; in-app chat talks to tools directly, not over this transport (§3.2)            |
| AST-11-BE | `BE`  | `[ ]`  | Personal access token auth for external MCP clients | Deferred — no concrete external-client use case yet (§3.5)                                          |
| AST-12-FE | `FE`  | `[ ]`  | `AssistantScreen` — chat UI              | Cubit pattern; message list + input field (§3.4)                                                                |
| AST-13-FE | `FE`  | `[ ]`  | Nav drawer entry for Assistant           | New `AppSection`, signed-in only (§3.4)                                                                         |
| AST-14-FE | `FE`  | `[ ]`  | Streaming response rendering             | Incremental display while the agent loop runs multiple tool calls (§3.4)                                        |
| AST-15-FE | `FE`  | `[ ]`  | l10n strings for assistant chat UI       |                                                                                                                  |

---

## 3. Shared Infrastructure

| ID     | Layer | Status | Title                          | Notes                                                                                          |
| ------ | ----- | ------ | ------------------------------- | ------------------------------------------------------------------------------------------------ |
| AI-01  | `BE`  | `[ ]`  | New env vars                   | `ANTHROPIC_API_KEY`, `WHISPER_API_KEY`, `AUDIO_STORAGE_PATH` — `getEnv`-with-validation pattern (§4) |
| AI-02  | `BE`  | `[ ]`  | `docs/api.md` updates          | All new endpoints across Epics 1 and 2, regenerate `docs/openapi.yaml` after                     |
