# BeeTrack — AI Assistant Project

> Design doc for the AI feature set: voice-logged inspections and a conversational
> apiary/marketplace assistant. This describes **how it's going to work**; ticket
> breakdown lives in a separate backlog once this is agreed.
>
> Relates to existing backlog placeholders: Epic 5 (Voice Logging), Epic 7 (MCP Server).
> This doc supersedes those two epics' scope — the separate backlog to be created next
> will replace their rows.

---

## 1. Overview

Two related but independently shippable features:

1. **Voice-logged inspections** — the mic button lives on a single apiary's grid
   screen (`ApiaryGridScreen`), not on an individual hive's screen, so the beekeeper
   names the hive out loud as part of the recording ("Hive 3, brood looks good...")
   while walking the whole apiary rather than opening each hive in turn. Tap to
   start, tap again to stop; each recording is **queued and processed by a background
   worker** rather than blocking on the spot, so the beekeeper can immediately record
   hive 4, then hive 5, without waiting for hive 3's transcription/reasoning to
   finish — the matching inspection (or treatment/harvest/feeding) record on the
   *correct* hive appears once the worker gets to it.
2. **AI apiary assistant** — a chat screen in the app where the beekeeper asks natural
   language questions about their own data ("which hive wasn't treated this month?",
   "which colonies need feeding?", "compare hive 3 and hive 7") and about the
   marketplace ("find me a seller of Dadant frames near me"). The assistant answers by
   calling read-only tools exposed over the user's own data via an MCP server — it
   never writes data.

Both features share the same underlying Anthropic API client and its tool-use
(function-calling) mechanism, but they use it differently:

- **Voice logging** needs Claude to (a) figure out *which hive, within the current
  apiary,* the beekeeper named and (b) turn the transcript into *one or more
  structured writes* against it — a small, fixed write-tool schema
  (`create_inspection` / `create_treatment` / `create_harvest` / `create_feeding` /
  `update_hive_status`),
  possibly more than one call per recording if it covers more than one topic (e.g.
  inspection notes followed by "also gave them a litre of syrup"). Step (a) needs to
  know that apiary's hive names, which is exactly the data `list_hives` already
  exposes for the chat assistant (§3.3), filtered to one apiary — so the voice
  pipeline calls that *same tool implementation* as a plain Go function to resolve a
  spoken name to a hive ID, without needing the MCP transport layer (no external
  client ever needs to reach the voice pipeline's internals). Each write-tool
  invocation is then handled by a plain Go `switch` that calls straight into the
  existing service layer.
- **The chat assistant** needs Claude to run a multi-step, open-ended *read* loop
  (look up hives, maybe compare two, maybe follow up with a listing search) where the
  same tool definitions are useful outside the app too. That reusability is what
  justifies formalizing the read tools as an **MCP server** (`internal/mcp/`) — a
  single place both features (and, later, external clients) call into, instead of
  duplicating hive-lookup logic in two places.

---

## 2. Voice-logged inspections

### 2.1 Flow

The flow has two halves that don't run in the same request: the mobile app's upload
returns almost immediately, and a background worker does the actual transcription and
reasoning afterward — the same producer/consumer shape the blockchain certification
worker already uses in this codebase (`internal/worker`, poll interval, job status
column), just for a different job type.

**Half 1 — enqueue (synchronous, fast):**

```
[Mic icon in the ApiaryGridScreen's bottom amber banner]
        │  tap → opens a modal dialog (same pattern as the existing
        │  Hive list dialog: × close button in the header)
        ▼
[Recording dialog]
  - A round record button with a mic icon, front and center: tap to start
    → auto-stops after ~2.5s of silence following speech, OR tap again to
    stop early, OR hits the hard cap (e.g. 3 min) and auto-submits
  - Below it: if there's at least one recording still pending/processing,
    a list of them — see §2.5 for Play/Cancel
        │  stop → the audio is saved to local device storage first (so it
        │  can be played back regardless of upload outcome — §2.5), then
        │  uploaded (multipart); the apiary is known from the screen
        │  (it's in the URL), the hive is not
        ▼
POST /api/v1/apiaries/{id}/voice
        │  saves the audio file server-side, inserts a `voice_recordings`
        │  row with status = 'pending', and returns immediately — no
        │  Whisper or Claude call happens in this request at all
        ▼
Response `202 Accepted`: { recording_id, status: "pending" }
        │
        ▼
[The new recording appears in the dialog's pending list (§2.5); the round record
 button is immediately available again, so the beekeeper can record hive 4, then
 hive 5, back-to-back, without waiting on hive 3's processing]
```

**Half 2 — background worker (asynchronous, one recording at a time or with limited
concurrency):**

```
Worker polls voice_recordings WHERE status = 'pending' (same JobPollInterval-style
loop as the blockchain worker), claims one, sets status = 'processing'
        │
        ▼
  Speech-to-text (Whisper)  →  transcript + per-segment confidence            ─┐
        │                                                                      │
        ▼                                                                      │
  QUALITY GATE: low-confidence / high no-speech-probability / garbled           │
  transcript → short-circuits here as `POOR_AUDIO_QUALITY`, never reaches       │
  Claude (§2.2); recording marked status = 'failed'                            │
        │  (passes)                                                            │
        ▼                                                                      │
  PHASE 1 — resolve targets: Claude is given the transcript + *this             │  every step's
  apiary's* hives (id, name) via the same `list_hives` read tool the            │  input/output
  chat assistant uses (§3.3, called directly as a Go function here, not         │  gets persisted
  over MCP transport, filtered to this apiary). It matches each named           │  — see §2.4
  hive mentioned in the transcript to a hive_id.                                │
        │                                                                       │
        ▼                                                                       │
  PHASE 2 — write: for each resolved (hive_id, topic) pair, Claude calls        │
  one write tool with that hive_id fixed in the arguments:                     │
     - create_inspection / create_treatment / create_harvest / create_feeding / update_hive_status   │
        │  picks 1..N tools + fills each one's arguments from the transcript    │
        ▼                                                                       │
  Backend executes each tool call in turn against the corresponding             │
  existing service (service.InspectionService.Create /                        │
  TreatmentService.Create / HarvestService.Create / FeedingService.Create),     │
  writing one voice_actions row per call                                      │
        │                                                                    │
        ▼                                                                   ─┘
Recording marked status = 'completed'; the audio file is deleted (§2.4) — the
transcript, already persisted, is all that's kept from here on
        │
        ▼
[The recording drops off the recording dialog's pending list (§2.5) once it leaves
 `processing`; next time the beekeeper opens the separate Voice Activity screen (§2.6)
 or pulls to refresh there, they see its actions: "Hive 3 — Inspection logged: brood
 good" + "Hive 3 — Feeding logged: 1L syrup", each tap-to-edit/undo]
```

The mic is a toggle, not a hold — tap to start, tap again to stop and submit. This
matters on Android as much as web: a held gesture is easy to lose (accidental release,
a notification stealing focus) mid-sentence, and a beekeeper's hands are often not
free enough for a sustained hold while also handling a frame. It also fits a longer
recording better — a beekeeper narrating a full inspection plus a feeding note, for a
named hive, is naturally 30-60+ seconds, not a quick press-and-release. On top of the
manual toggle, the recording also auto-stops on its own — see the time limits and
silence-detection design in §2.2.

The banner-icon-opens-a-dialog shape mirrors the existing hive list dialog on the same
screen (`_HiveListDialog`) rather than introducing a new interaction pattern: a bottom
banner icon is consistent with how the grid screen already surfaces secondary actions
(filter, hive list, dashboard), and a modal dialog is a natural home for both starting
a new recording and seeing what's still in flight, without leaving the grid screen the
beekeeper is actively working from.

Queuing is what makes rapid-fire, multi-hive walk-throughs practical: transcription
and a multi-tool-call Claude turn together can easily take several seconds, and
forcing the beekeeper to stand and wait after every single hive defeats the point of a
hands-free tool. Queuing decouples "how long the AI takes" from "how fast the
beekeeper can talk" — three recordings taken 10 seconds apart just become three
`pending` rows, processed in the order they were queued.

Putting the button on the apiary grid (rather than a single hive's screen) means the
endpoint knows *which apiary* from the URL, same as every other `/apiaries/{id}/...`
route, but still can't infer *which hive* — hive identification has to happen from
speech, which is why Phase 1 exists. This also means one recording can name more than
one hive within that apiary ("Hive 3 looked good... Hive 4 needs feeding") and produce
actions against each — each `voice_actions` row's `hive` is what lets the Voice
Activity screen say which hive each result belongs to. It deliberately does not
resolve across apiaries: a beekeeper working one apiary's hives at a time matches how
the grid screen itself is scoped, and keeping resolution to a smaller, unambiguous
hive set is both cheaper for Claude and safer (no risk of matching a same-named hive
in a different apiary the beekeeper isn't even standing in).

Nothing about existing inspection/treatment/harvest/feeding validation changes — the
worker's Phase 2 is a new entry point in front of the *same* service-layer `Create`
calls used by the manual forms, so a bad transcript-derived value fails the same
validation a bad form submission would (e.g. `ErrInvalidBroodPattern`). Each action is
validated and saved independently: if the feeding fields are bad but the inspection
is fine, the inspection is still saved and only the feeding action comes back as an
error (§2.4) — one bad action doesn't roll back a good one, or the sibling recordings
still queued behind it.

### 2.2 Design decisions

- **Tap to start, tap to stop** (not hold-to-record) — see the rationale in §2.1.
- **Recording is bounded, and stops itself when the beekeeper's done talking.** Two
  independent limits, both client-side (Flutter, via the `record` package's amplitude
  stream — no separate VAD library needed for a simple silence threshold):
  - **Hard cap** — a fixed maximum length (e.g. 3 minutes) after which recording
    auto-stops and submits whatever was captured, regardless of speech/silence. This
    bounds worst-case Whisper/Claude request size and cost, and catches a beekeeper
    forgetting the mic is still running (pocket, distraction).
  - **Auto-stop on silence** — once speech has been detected, ~2.5s of continuous
    silence auto-stops and submits, so a normal "I'm done talking" pause ends the
    recording without a second tap. A short grace period at the very start (no
    auto-stop before any speech is detected at all) avoids cutting off a beekeeper who
    pauses to gather their thoughts before speaking. The manual stop tap always still
    works, for a beekeeper who wants to end it before either limit triggers.
  Both give visible feedback as they approach (e.g. a brief warning at the last few
  seconds of the hard cap) so the cutoff isn't a surprise.
- **The recording is saved locally on the device, and a pending/cancel queue is
  visible in the same recording dialog** — see §2.5 for the full design (local
  playback via Play, and Cancel while still `pending`). This is what makes it
  reasonable to fire off several recordings back-to-back: the beekeeper isn't just
  trusting an upload happened somewhere in the background, they can see and hear
  exactly what's still waiting to be processed.
- **One recording can produce multiple records.** A single note that covers both an
  inspection and a feeding ("brood looks good, queen seen... also topped them up with
  a litre of syrup") should become one inspection row and one feeding row, not force
  the beekeeper into two separate recordings or lose one of the two topics. Claude is
  free to call more than one write tool per request; the worker executes and persists
  each call independently (§2.1, §2.4) and the Voice Activity screen (§2.6) shows one
  entry per action produced, once processing finishes.
- **Hive state flags can be set by voice too, not just logged as notes.** The manual
  hive edit screen lets a beekeeper flip `queenless` / `needs_food` /
  `ready_for_harvest` / `active` and add/remove diseases (`PATCH
  /apiaries/{id}/hives/{hiveId}`, plus the disease add/remove endpoints) — a beekeeper
  saying "this one's queenless, mark it for feeding" should update the same flags,
  not just end up as free-text notes on an inspection Claude also happened to log. A
  fifth write tool, `update_hive_status`, wraps exactly those existing endpoints; it's
  independent of `create_inspection` (a recording can trigger one, the other, or
  both) since the beekeeper might mention a status change without describing a full
  inspection at all.
- **Splitting is topic-based, not sentence-based.** Claude decides how many distinct
  actions the transcript actually describes — a long recording that's all about one
  inspection still produces exactly one `create_inspection` call; it doesn't fragment
  just because it's long.
- **No confirmation step before saving.** Since results now only ever show up later
  (§2.1's queuing means there's nothing to confirm at record time even if we wanted
  to), the Voice Activity screen's edit/undo (§2.6) *is* the correct/undo path, not a
  pre-save review screen. This keeps the flow hands-free, which is the whole point in
  the field: a real undo mechanism after the fact is what makes it safe to skip a
  confirmation step before it.
- **Hive naming is required, not optional.** Since the button no longer implies a
  hive, a recording that never names one has nothing to resolve against. The app's
  recording UI should hint this upfront (e.g. a placeholder tip "say the hive name
  first") rather than the beekeeper discovering it only after an empty result.
- **Ambiguous or unmatched hive names don't guess.** If the spoken name matches more
  than one hive in the current apiary (duplicate hive names are allowed by the data
  model even though most beekeepers won't create them) or matches nothing closely
  enough, that portion of the transcript becomes an `error` action
  (`HIVE_NOT_IDENTIFIED`, §2.4) rather than picking a hive at random — the app surfaces
  it as "Couldn't tell which hive — please open it and log manually" instead of
  silently writing to the wrong colony's history. Matching only needs to be forgiving
  of minor mishearing (Whisper transcription noise), not of genuinely ambiguous input.
- **Per-hive context is fetched after resolution, not upfront.** Once Phase 1 resolves
  a hive_id, the backend loads *that* hive's context before Phase 2 — type and last
  inspection's frame counts (so "added two frames" resolves to a delta relative to
  last known state, consistent with how `FramesAddedBrood` etc. already work as signed
  deltas). This is a straightforward consequence of not knowing the hive until Phase 1
  runs.
- **Reuse the manual forms' own autocomplete suggestions, don't invent a separate
  source of truth.** The Treatment and Feeding forms already suggest previously-used
  values as the beekeeper types (medicine name, food type, food amount) specifically
  so a spoken/typed variant doesn't fork into a near-duplicate entry — Claude should
  be given the caller's own existing values for the field it's filling, from the same
  endpoints those forms already call: `GET /api/v1/medicines` (past medicine names,
  most recent first) for `create_treatment`, and `GET /api/v1/feed-types` /
  `GET /api/v1/feed-amounts` for `create_feeding`. So "treated with Apiwarol" maps to
  however the beekeeper already spelled Apiwarol before, and "gave them sugar syrup"
  maps to an existing "sugar syrup" food type rather than creating "Sugar Syrup" as a
  second, slightly different entry. These are per-user history lists, not a fixed
  vocabulary — an entirely new medicine or food type the beekeeper has never logged
  before still passes through as free text, same as typing it manually would. Dose
  values stay free text regardless (already allowed, recent fix), so a spoken dose
  like "two strips" passes through as-is without needing a suggestion match.
- **Language:** Polish and English, matching the app's l10n. Whisper auto-detects;
  Claude is prompted in the detected language so field values (notes) preserve the
  beekeeper's own words — the same detected language is used for hive-name matching,
  so Polish diacritics in a hive's name aren't a mismatch source.
- **Bad audio is rejected before it ever reaches Claude, not guessed at.** Whisper
  returns per-segment confidence alongside the transcript (avg log-probability,
  no-speech probability). If a meaningful portion of the recording comes back
  low-confidence — overlapping voices, heavy wind/machinery noise, mostly
  unintelligible — the backend short-circuits with `POOR_AUDIO_QUALITY` right after
  transcription and never invokes Claude at all: there's no reliable transcript to
  reason over, so letting the model try would mean it's effectively inventing an
  action from noise. The app shows "Recording wasn't clear enough — try again
  somewhere quieter" — distinct from the "understood the words, nothing actionable in
  them" case below, because the advice to the beekeeper is different (re-record vs.
  say more).
- **Failure mode (transcript was fine, nothing to act on):** if Claude can't
  confidently map a clean transcript to a named hive plus any of the five write
  actions at all (e.g. silence, unrelated speech, small talk), the endpoint returns a
  distinct `NO_ACTION_RECOGNIZED` error and the app shows "Couldn't understand that —
  try again" rather than guessing. If it maps *some* but not all of what was said
  (e.g. it caught the feeding but not the inspection, or resolved one hive but not a
  second one mentioned), that's a partial result, not a failure — the app shows what
  was saved and the beekeeper can fill in the rest manually.

### 2.3 New backend surface

| Endpoint | Notes |
|---|---|
| `POST /api/v1/apiaries/{id}/voice` | Apiary-scoped, not hive-scoped — multipart, field `audio` (webm/m4a/wav); rejects synchronously with `RECORDING_TOO_LONG` if the file's duration exceeds the same hard cap the client enforces (§2.2) — a server-side safety net, not just a client-side limit; otherwise stores the file, inserts a `pending` `voice_recordings` row, and returns `202 Accepted` with `{ recording_id, status: "pending" }` immediately — no Whisper/Claude work happens in this request (§2.1) |

### 2.4 Persistence, review & undo

Because one recording can now produce several actions, the log needs a
one-recording-to-many-actions shape: one row for the recording/transcript itself, and
one child row per tool call Claude made against it. This log is no longer just an
engineering trail — it's also what backs a user-facing **Voice Activity** list (§2.6)
so the beekeeper can see what the assistant did and undo it if something went wrong,
not only what a developer can `SELECT` from psql:

```sql
-- migrations/NNN_create_voice_logs.sql
CREATE TABLE voice_recordings (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL REFERENCES users(id),
    apiary_id         BIGINT NOT NULL REFERENCES apiaries(id), -- known from the URL
    status            TEXT NOT NULL DEFAULT 'pending', -- 'pending' | 'processing' |
                                                         -- 'completed' | 'failed' | 'cancelled'
                                                         -- ('cancelled' set by the beekeeper
                                                         -- via §2.5, only while still 'pending')
    audio_path        TEXT,                      -- UUID-named file under AUDIO_STORAGE_PATH
                                                    -- (same convention as inspection images);
                                                    -- cleared once transcription succeeds
    transcript        TEXT,                       -- Whisper output; NULL until the worker
                                                     -- has processed this recording
    detected_language TEXT,                      -- e.g. "pl", "en"
    error_message     TEXT,                       -- set on status = 'failed' for a
                                                     -- recording-level failure (Whisper API
                                                     -- error, corrupt file) — distinct from a
                                                     -- per-action error, which lives on the
                                                     -- voice_actions row instead
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at      TIMESTAMPTZ                 -- when the worker finished with it
                                                    -- (either outcome)
);
-- No hive_id here: a single recording can name zero, one, or several hives within
-- the apiary — which hive(s) it resolved to lives per-action below, not on the
-- recording itself.

CREATE TABLE voice_actions (
    id                  BIGSERIAL PRIMARY KEY,
    voice_recording_id  BIGINT NOT NULL REFERENCES voice_recordings(id),
    sequence            SMALLINT NOT NULL,        -- order Claude emitted this tool call in
    spoken_hive_name     TEXT,                     -- Claude's best-effort extraction of the
                                                     -- name as said, kept even on a failed match
    hive_id             BIGINT REFERENCES hives(id), -- resolved target; NULL if unresolved
    tool_name           TEXT,                      -- 'create_inspection' | 'create_treatment' |
                                                     -- 'create_harvest' | 'create_feeding' |
                                                     -- 'update_hive_status';
                                                     -- NULL when hive resolution itself failed
    tool_arguments      JSONB,                      -- exact structured args Claude filled in
    previous_hive_state JSONB,                      -- only for 'update_hive_status': the flags'
                                                      -- values before this change, so Undo can
                                                      -- restore them instead of deleting a row
    result_type         TEXT NOT NULL,              -- 'inspection' | 'treatment' | 'harvest' |
                                                      -- 'feeding' | 'hive_status' | 'error'
    result_record_id    BIGINT,                     -- FK-less pointer to the created/updated
                                                       -- row, if any (the hive itself, for
                                                       -- 'hive_status')
    error_message       TEXT,                       -- set when result_type = 'error'
                                                      -- (includes HIVE_NOT_IDENTIFIED cases)
    reverted_at         TIMESTAMPTZ,                 -- set when the beekeeper undoes this action
                                                      -- (§2.6); NULL means still in effect
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

A recording with no recognizable action at all (`NO_ACTION_RECOGNIZED`, §2.2) still
gets its `voice_recordings` row — just zero `voice_actions` children — so "the
assistant heard nothing useful" is as visible in the log as a successful split. A
recording that named a hive but couldn't resolve it (`HIVE_NOT_IDENTIFIED`) gets one
`voice_actions` row with `spoken_hive_name` set, `hive_id` NULL, and `result_type =
'error'` — so "the beekeeper said a name we couldn't match" is distinguishable in the
log from "the beekeeper didn't name a hive at all." A `POOR_AUDIO_QUALITY` rejection
still gets its `voice_recordings` row (transcript included, however unreliable it is)
plus one `voice_actions` error row with `tool_name` NULL — Claude was never called for
that recording at all, which the log makes visible the same way a failed hive match
does.

The transcript and per-action arguments stay even after an action is reverted (the row
is marked, not deleted) — that's what makes "why did the assistant only log the
feeding and not the inspection I also described?" answerable after the fact, whether
the beekeeper is debugging it themselves from the Voice Activity screen or it's being
looked at from psql during development.

Audio itself is only ever *transiently* stored — it has to exist on disk long enough
for the worker to pick it up and run Whisper on it (the whole reason it's queued
instead of transcribed inline), but there's no reason to keep it a moment longer. The
worker deletes the file (and clears `audio_path`) immediately after a successful
transcription, and also on a terminal failure (`POOR_AUDIO_QUALITY`,
`NO_ACTION_RECOGNIZED`, or a hard worker error) — so the only trace of a recording
that survives long-term is its transcript, never the raw audio. This is the same
UUID-filename-under-a-storage-path convention inspection images already use
(`IMAGE_STORAGE_PATH`), just a new `AUDIO_STORAGE_PATH` / Docker volume for a
short-lived file instead of a long-lived one.

Each `voice_actions` row is written *before* its service `Create` call (or before the
hive-resolution attempt, if that's what it's recording) and updated with the outcome
right after, so a crash mid-job, a failed hive match, or a validation failure (e.g.
`ErrInvalidBroodPattern`) on one action still leaves a diagnosable record for that
action specifically, without blocking or rolling back the sibling actions from the
same recording — or the sibling *recordings* still sitting `pending` behind it in the
queue. If the worker itself dies mid-recording (not mid-action — e.g. a deploy
restarts the process), the recording is left `processing`; the worker's polling loop
picks up anything stuck in `processing` past a timeout and retries it from the top,
the same recovery approach the blockchain worker already uses for stuck jobs.

### 2.5 Recording popup: local playback & cancelling a pending recording

This is the mic-icon dialog itself (§2.1) — a short-term, in-flight view distinct from
the Voice Activity screen (§2.6), which is the longer-term history of *completed*
actions. This one only ever shows recordings that haven't finished processing yet.

- **The audio is saved to the device before it's uploaded.** As soon as recording
  stops, the raw audio file is written to the app's own local storage (e.g. via
  `path_provider`, the same place other local, non-synced app state would live) —
  independent of whether the upload succeeds, retries, or the recording is later
  processed. This is what makes "Play" (below) possible without depending on the
  server ever serving audio back: audio is only ever transiently stored server-side
  (§2.4) and is never re-downloadable, so local-first is the only way to hear it
  again.
- **Pending list** — whenever at least one recording for this apiary hasn't reached a
  terminal state yet (`pending` or `processing`, whether still uploading, queued, or
  actively being worked on), the dialog shows them as a list below the round record
  button, each with two actions:
  - **Play** — plays back the *locally saved* audio file, so the beekeeper can
    double-check what they actually said while it's still in flight, without waiting
    for a transcript.
  - **Cancel** — calls `DELETE /api/v1/apiaries/{id}/voice-recordings/{recordingId}`
    (§2.3), which only succeeds while the recording is still `pending` (not yet
    claimed by the worker); it marks the row `cancelled` so the worker's `WHERE status
    = 'pending'` poll simply never picks it up, and no Whisper/Claude call happens for
    it at all. Once a recording flips to `processing`, Cancel is no longer offered —
    it's too late to stop cleanly, and the (probably small) remaining wait is shorter
    than the complexity of interrupting a job partway through and rolling back
    whatever it had already done. If the result turns out wrong once it completes,
    that's what Undo (§2.6) is for.
- Once a recording leaves `pending`/`processing` (whichever way), it drops out of this
  list — there's nothing left to play-before-processing or cancel. The locally saved
  audio file is deleted at that point too, since the transcript (kept server-side,
  §2.4) is what matters from then on, and the Voice Activity screen (§2.6) is where
  its outcome lives.

New backend surface for this:

| Endpoint | Notes |
|---|---|
| `DELETE /api/v1/apiaries/{id}/voice-recordings/{recordingId}` | Cancel: sets `status = 'cancelled'` and deletes the server-side audio file; only valid while `status = 'pending'` — returns a conflict error if the worker has already claimed it (`processing`) or it's already terminal |

### 2.6 Reviewing & undoing voice actions

Because processing is now asynchronous, this screen is not just a review/undo
convenience — it's *the* place the beekeeper finds out what a queued recording
actually did, since the upload response (§2.1) never carries a result.

- **Voice Activity screen** — a new history-icon button in the apiary grid's bottom
  amber banner (next to the mic button) opens a paginated list of past voice
  recordings for *this apiary*, newest first (same page-based pagination pattern as
  `InspectionHistoryScreen`). Each recording shows its status — "Queued" /
  "Processing…" / its resulting action rows once `completed` (hive, action type, a
  one-line summary, tap-to-edit) / an error reason once `failed`. Reverted actions
  show as struck-through/"undone" rather than disappearing, so the history stays
  honest about what actually happened. The screen refreshes on open and via
  pull-to-refresh — same lightweight polling-on-demand approach already used for the
  honey batch certification badge's in-progress state, no push notifications or
  websockets needed for v1.
- **Editing** a successful action doesn't need a new editing UI at all: a voice-created
  inspection/treatment/harvest/feeding is stored as exactly the same row a manually
  created one would be, so tapping a row just opens the existing edit screen for that
  record type (`InspectionFormScreen`, `TreatmentFormScreen`, etc., already in edit
  mode) — the same screen reached from the hive detail history lists.
- **Undo** removes the underlying record via that record type's *existing* delete
  service call (e.g. `InspectionService.Delete`) and sets `reverted_at` on the
  `voice_actions` row — it does not delete the log row itself, so an undone action is
  still visible (as reverted) rather than vanishing from the history. `update_hive_status`
  actions undo differently, since there's no row to delete: it calls the existing
  `PATCH .../hives/{hiveId}` with the flags from `previous_hive_state`, restoring
  exactly what they were before, then sets `reverted_at` the same way.
- **Error and cancelled rows** (`NO_ACTION_RECOGNIZED` / `HIVE_NOT_IDENTIFIED` /
  `POOR_AUDIO_QUALITY` / a recording-level `failed` status / `cancelled` via §2.5)
  have no record to edit or undo — they're listed for transparency only ("couldn't
  understand this recording" / "couldn't tell which hive" / "recording wasn't clear
  enough" / "cancelled").

New backend surface for this:

| Endpoint | Notes |
|---|---|
| `GET /api/v1/apiaries/{id}/voice-recordings` | Paginated (`limit`/`offset`), newest first; each item is a recording with its `status` and, once `completed`, its nested `voice_actions` (resolved hive, action type, a short summary, `reverted_at`, and enough of `result_record_id` to deep-link into the corresponding edit screen) — this is what the Voice Activity screen polls/refreshes against |
| `DELETE /api/v1/apiaries/{id}/voice-actions/{actionId}` | Undo: deletes the underlying record via the existing per-type delete service call and sets `reverted_at`; a no-op error if the action has no `result_record_id` (nothing was created) or is already reverted |

---

## 3. AI apiary assistant (chat + MCP)

### 3.1 Why an MCP server specifically

The in-app chat feature, the voice-logging pipeline's hive-name resolution (§2.1), and
(optionally, later) external AI clients all need the same read access to a user's
apiary data. An MCP server is the one abstraction that serves all three: the backend
defines each tool once, and the *same* tool implementations back the chat agent loop,
the voice pipeline's Phase 1, and any external MCP client — voice logging just calls
them as plain Go functions rather than over the MCP transport (§1). It also keeps a
hard boundary between "things the assistant is allowed to see" and "things it's
allowed to do" — v1 exposes read-only tools only; nothing in the assistant path can
mutate apiary data (voice logging's write tools, §2.1, are a separate, unrelated tool
schema, not part of this MCP catalog).

### 3.2 Architecture

```
Flutter app (new Assistant chat screen)
        │  POST /api/v1/assistant/messages  { conversationId?, message }
        ▼
Go backend — AssistantHandler → AssistantService
        │
        │  runs a Claude agent loop (Anthropic Messages API, tool use),
        │  scoped to the authenticated user's own userID
        ▼
   MCP server (in-process, same Go binary — no separate deployment)
        │  tools declare a Go interface backed by the *existing* repositories
        │  (ApiaryRepository, HiveRepository, InspectionRepository, ...)
        │  every tool call is implicitly filtered to the caller's own apiaries/hives —
        │  same ownership check pattern as REST handlers (`GetMembership`)
        ▼
   PostgreSQL (read-only queries; nothing new to migrate for v1 tools)
        │
        ▼
   Claude synthesizes a natural-language answer from tool results
        ▼
Response streamed back to the chat screen
```

The MCP server runs **in-process** inside the existing `cmd/api` binary rather than as
a separate service — it's a Go package (`internal/mcp/`) whose tools call straight into
the service layer, the same way handlers do. This avoids a second deployable, a second
auth boundary, and network hops for what's fundamentally "read the same Postgres the
REST API already reads." An HTTP+SSE transport is exposed at a dedicated path
(`/mcp`) purely so the same tool set is *also* reachable by external MCP clients later
(see §3.5) — the in-app chat path talks to the tools directly as Go function calls,
not over that transport.

### 3.3 Tool catalog (v1, read-only)

| Tool | Purpose |
|---|---|
| `list_hives` | All **active** hives across the caller's apiaries, with status flags (queenless, needs_food, ready_for_harvest) and active diseases; accepts an optional `apiary_id` filter — also the tool voice logging's Phase 1 calls directly, filtered to the current apiary, to resolve a spoken hive name to a hive_id (§2.1). Inactive hives are excluded here and from every other multi-hive tool below, matching the app's own convention of excluding inactive hives from anything actionable — a single hive looked up directly by ID (`get_hive_summary`, `list_hive_records`) is unaffected, since that's an explicit request for a known hive, not a "what needs attention" scan |
| `list_hive_records` | One hive's inspections/treatments/harvests/feedings, filtered to a `(hive_id, record_types?, days?)` window — `record_types` selects a subset (e.g. `["inspection","feeding"]`) in one call instead of one tool call per type; omitting it returns all four. `Treatment` has no active/inactive concept in the schema, so this replaces a single "active treatments" notion with a recency window each caller controls |
| `get_hive_summary` | Aggregates all four record types for one hive over the same optional day window, plus its status flags and diseases |
| `list_hives_missing_records` | Hives across the caller's apiaries (or one, via `apiary_id`) missing at least one of `record_types` (inspection/treatment/feeding; all three if omitted) in the last `days` days, or that never had one at all if `days` is omitted. Each result lists which of the requested types it's actually missing and, if any, when it last happened |
| `list_hives_by_status` | Hives across the caller's apiaries (or one, via `apiary_id`) matching any of `statuses` (queenless/needs_food/sick/ready_for_harvest; matches any of them if omitted) — `sick` means the hive has at least one active disease. Generalizes the original "just needs_food" idea into one filter tool for every status flag |
| `compare_hives` | Side-by-side of key metrics (brood pattern, frame counts, disease flags, last inspection date) for a set of hive IDs |
| `search_listings` | Wraps the existing public `GET /api/v1/listings` search/filter — lets the assistant answer "find me X" marketplace questions |
| `get_listing` | Single listing detail, for follow-up questions about a specific result |

Each tool is a thin wrapper: it takes the caller's `userID` (never trusted from the
model, always injected by the service before the agent loop starts) plus model-supplied
arguments, and returns the same shapes the REST endpoints already return — no new
serialization to design.

Write tools (`create_inspection`, `log_treatment`, `log_harvest` from the original
Epic 7 sketch in BACKLOG.md) are **descoped from v1**: mixing a conversational
assistant with silent data mutation is a much bigger trust surface (confirmation UX,
undo, audit trail) than the voice-logging flow already solves for that exact need.
Voice logging is the write path; the chat assistant is read-only. This can be
revisited once v1 is validated.

### 3.4 In-app chat screen

New Flutter feature, `features/assistant/`, following the existing Cubit pattern:

- `AssistantScreen` — message list + input field, reached from the nav drawer (a new
  `AppSection` entry alongside Apiaries/Marketplace/Honey Batches), signed-in only.
- Streams the response so the beekeeper sees the answer arrive incrementally rather
  than waiting on the full agent loop (which may take several tool-call round trips).
- No conversation history persisted server-side for v1 — each message includes the
  full prior turns from the client, same stateless-server approach the rest of the
  API already uses (JWT auth, no server sessions). A `conversations` table is a
  natural v2 if beekeepers want history across app restarts.

### 3.5 Auth

- **In-app chat:** identical JWT `Authorization: Bearer` header as every other
  authenticated endpoint — `middleware.Auth` already injects `userID` into context;
  the assistant service uses that to scope every tool call, exactly like
  `GetMembership` scopes apiary/hive access today.
- **External MCP clients (future, not v1):** would need a separate long-lived
  credential (personal access token), distinct from the short-lived JWT used by the
  app, since an external AI client can't do the app's login/refresh dance. Deferred
  until there's a concrete external-client use case — v1 only needs the in-app path.

### 3.6 Marketplace analysis

"Analyze listings and help find what's being searched for" maps directly to the
`search_listings`/`get_listing` tools above — the assistant is doing structured
filtering + summarization over the *existing* public listing search, not a new
recommendation engine. E.g. "find cheap Dadant frames near Kraków" becomes a
`search_listings` call with `category`, `price_max`, and location args the model
extracts from the question, same filters the Marketplace screen's own filter sheet
already sends.

---

## 4. Shared building blocks

| Component | New? | Notes |
|---|---|---|
| Anthropic API client (Go) | Yes | `internal/llm/` — thin wrapper around the Messages API with tool-use support; used by both the voice worker's intent parser and the assistant's agent loop |
| Whisper (speech-to-text) | Yes | Called from the voice worker only; hosted API call (no local model — matches this project's "call external AI services from Go" pattern already used for blockchain RPC calls) |
| `GET /medicines`, `/feed-types`, `/feed-amounts` | No | Already exist for the manual forms' autocomplete (§2.2) — the voice worker calls them as regular internal reads, no new endpoint |
| `internal/mcp/` | Yes | Tool definitions + in-process registry; read-only repository calls only. `list_hives` is called directly (as a Go function, not over MCP transport) by the voice worker's hive-name resolution, so it's the one tool implementation shared by both features |
| Voice worker | Yes | New job type polling `voice_recordings` for `pending` rows, following the *exact same shape* as `internal/worker`'s existing blockchain certification worker (poll interval, `pending → processing → completed/failed` status column, stuck-job recovery) — likely lives in `internal/worker/voice.go` alongside it rather than a new package, since it's the same pattern applied to a different job |
| `AUDIO_STORAGE_PATH` / Docker volume | Yes | Server-side, transient audio storage between upload and transcription (§2.4) — same UUID-filename convention and Docker-volume approach as `IMAGE_STORAGE_PATH`, just for short-lived files the worker deletes once transcribed |
| Local audio storage (Flutter, `path_provider`) | Yes | Client-side only, separate from the server-side path above — recordings are saved to the device before upload so the pending-list dialog can play them back locally (§2.5); deleted once the recording leaves `pending`/`processing` |
| New env vars | Yes | `ANTHROPIC_API_KEY`, `WHISPER_API_KEY` (or one combined key if using Anthropic's own audio input once available), `AUDIO_STORAGE_PATH` — same `getEnv`-with-validation pattern as `LoadBlockchainConfig` |

One new migration is required: `voice_recordings` + `voice_actions` (§2.4). The chat
assistant needs none — it reads through existing tables only.

---

## 5. Open questions for the backlog pass

- Exact Whisper provider (OpenAI Whisper API vs. self-hosted vs. Anthropic native audio
  once available) — affects one config var and one client, not the architecture above.
- Whether `list_hives_needing_food` should trust the existing `needs_food` boolean flag
  only, or also infer from inspection frame data — affects one tool's implementation,
  not its interface.
- Streaming transport for the chat screen: SSE vs. chunked HTTP vs. WebSocket — pick
  during implementation based on what the existing `ApiClient` (Dio) supports most
  simply.
- How strict hive-name matching should be (§2.2) — left to Claude's judgment initially
  (given the full `list_hives` list, decide match vs. no-match vs. ambiguous itself)
  rather than a separate fuzzy-string-matching algorithm; revisit if that proves too
  loose or too strict once real recordings are tested.
- Worker concurrency (§2.1/§2.4) — process one `pending` recording at a time
  (simplest, matches the blockchain worker's current single-job-at-a-time model) vs.
  a small worker pool so several beekeepers' queued recordings don't serialize behind
  each other; start with one-at-a-time and only add concurrency if queue depth turns
  out to matter in practice.

---

## 6. Next step

Once this design is agreed, split it into a dedicated backlog file (`BACKLOG_AI_ASSISTANT.md` or
new epics in `BACKLOG.md`) with BE/FE tickets per the existing ID-prefix convention (e.g.
`VC-*` for voice, `AST-*` for the assistant), replacing the current placeholder rows in
BACKLOG.md's Epic 5 and Epic 7.
