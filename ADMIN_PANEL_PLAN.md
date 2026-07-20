# Admin Panel & Moderation Implementation Plan

## Context

BeeTrack currently has no admin/moderation concept anywhere: `users` has no role field, marketplace listings publish instantly (`IsHidden` is the only status-like field, owner-controlled), and honey batch certification (Epic 9, see [HONEY_BLOCKCHAIN_PLAN.md](HONEY_BLOCKCHAIN_PLAN.md)) enqueues a blockchain job the moment an owner requests it — fully automatic, no human in the loop.

This feature adds:
1. An **ADMIN** user role — a flag on the existing `users` table, set manually in the DB only (no self-signup/promotion path in the product).
2. A **moderation workflow for marketplace listings** — new and edited listings are hidden from the public marketplace until an admin approves them.
3. An **approval gate in front of honey batch certification** — a batch owner's request to certify creates a review request; only after admin approval does the existing blockchain-jobs pipeline (worker, idempotency, retries — all unchanged) take over.
4. A **React admin panel** (new `admin/` directory) — a small, plain-looking, functional SPA that talks to new `/api/v1/admin/*` REST endpoints on the existing Go API. No new backend framework, no new auth system: same JWT login, same Postgres, gated by the new admin flag.

**Scope note (matches the honey-blockchain thesis framing):** this is being built alongside the same BeeTrack thesis project. Production-hardening concerns (rate limiting admin actions, audit-log retention policy, multi-admin conflict handling beyond "last write wins") are noted but not required for the feature to be considered complete — tagged **(production-hardening — optional)** below.

### Key decisions (confirmed)

- **Edit-review behavior:** editing an already-approved listing immediately unpublishes it (flips back to `pending`) until an admin re-approves it — no separate revision/diff table. Simpler, at the cost of the seller losing visibility while an edit is in review.
- **Rejections require a reason:** both listing rejections and certification-request rejections require admin-entered free text, shown back to the owner.
- **Admin auth reuses the existing JWT login** (`POST /api/v1/auth/login`), gated by a new `role` column on `users`. No separate admin login endpoint.
- **Certification approval sits before job enqueue:** a new `honey_batch_certification_requests` table is the review queue. Approving a request is the *only* thing that creates a `blockchain_jobs` row — the existing worker, idempotency guarantees (HC-BE-25), and retry logic (HC-10-02) are untouched.

---

## Implementation Phases

### Phase 1: Database Foundation (3 tasks)

1. **AP-DB-01: Add `role` to `users`**
   - File: `backend/migrations/036_add_user_role.sql`
   - `ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin'))`
   - No index needed (never queried at scale; every admin check is a single-row lookup by the already-indexed `id`).
   - **How an admin is created:** manually, via `UPDATE users SET role = 'admin' WHERE email = '...'` — no migration seeds one, no API path sets it. Document this in `docs/api.md` so it isn't rediscovered by trial and error.

2. **AP-DB-02: Add moderation fields to `listings`**
   - File: `backend/migrations/037_add_listing_moderation.sql`
   - Columns: `status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected'))`, `rejection_reason TEXT`, `first_approved_at TIMESTAMPTZ` (nullable — set once, on a listing's *first* approval; never cleared by later edits, so admin UI can tell "new listing" (`first_approved_at IS NULL`) apart from "edit of a previously-approved listing" (`first_approved_at IS NOT NULL`) even though both currently sit at `status='pending'`), `reviewed_by BIGINT REFERENCES users(id)`, `reviewed_at TIMESTAMPTZ`.
   - **Backfill:** existing rows get `status = 'approved'`, `first_approved_at = created_at` in the same migration — this feature must not retroactively de-list every listing that already exists.
   - Index: `(status, created_at DESC)` — the admin queue's primary query (`WHERE status = 'pending' ORDER BY created_at`).

3. **AP-DB-03: Create `honey_batch_certification_requests` table**
   - File: `backend/migrations/038_create_honey_batch_certification_requests.sql`
   - Columns: `id BIGSERIAL PRIMARY KEY`, `batch_id BIGINT NOT NULL REFERENCES honey_batches(id) ON DELETE CASCADE`, `requested_by BIGINT NOT NULL REFERENCES users(id)`, `status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected'))`, `rejection_reason TEXT`, `reviewed_by BIGINT REFERENCES users(id)`, `reviewed_at TIMESTAMPTZ`, `blockchain_job_id BIGINT REFERENCES blockchain_jobs(id)` (set once approval creates the job, so the request row links forward to the existing certification lifecycle), `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`.
   - Indexes: `(status, created_at DESC)` — admin queue; `(batch_id, created_at DESC)` — a batch may accumulate multiple requests if rejected and resubmitted.
   - Partial unique index `(batch_id) WHERE status = 'pending'` — mirrors the idempotency pattern from HC-DB-04: a batch can't have two pending review requests at once.

---

### Phase 2: Backend Models & Persistence (5 tasks)

4. **AP-BE-01: Extend `User` model**
   - File: `backend/internal/model/user.go`
   - Add `Role string` field. Add `IsAdmin() bool` helper (`Role == "admin"`) so callers never compare the raw string.

5. **AP-BE-02: Extend `Listing` model**
   - File: `backend/internal/model/listing.go`
   - Add `Status string`, `RejectionReason *string`, `FirstApprovedAt *time.Time`, `ReviewedBy *int64`, `ReviewedAt *time.Time`.
   - Constants in the same file: `ListingStatusPending`, `ListingStatusApproved`, `ListingStatusRejected`.

6. **AP-BE-03: Create `HoneyBatchCertificationRequest` model**
   - File: `backend/internal/model/honey_batch_certification_request.go`
   - Fields mirror the migration. Reuses the existing tri-state pattern (`pending`/`approved`/`rejected`) rather than inventing a fourth status enum in the codebase.

7. **AP-BE-04: Repositories**
   - `backend/internal/repository/user.go` — add `GetByID` if not already present (needed by `RequireAdmin` middleware), `ListAdmins` not needed.
   - `backend/internal/repository/listing.go` — extend `ListingFilter` with a `Status` field; add `ListPendingReview(ctx, limit, offset)`, `Approve(ctx, id, reviewerID)`, `Reject(ctx, id, reviewerID, reason)`.
   - File: `backend/internal/repository/honey_batch_certification_request.go` — `Create`, `GetByID`, `ListPending(ctx, limit, offset)`, `GetPendingForBatch(ctx, batchID)` (idempotency check before creating a new one), `Approve(ctx, id, reviewerID)` (also stamps `blockchain_job_id`), `Reject(ctx, id, reviewerID, reason)`.

8. **AP-BE-04b: Publish existing data + seed pending listings for admin QA**
   - File: `backend/cmd/seed/main.go`
   - **Publish existing data:** already covered by AP-DB-02's backfill (`status = 'approved'` for every pre-existing row in the same migration) — this task is the operational follow-through, not new logic: after AP-DB-02 lands, re-running the migration against dev/staging DBs is what actually "publishes" today's listings. No separate script needed beyond the migration itself.
   - **Seed script changes** (`seedListings`, `main`):
     - After creating the existing seed set (`seedListings`'s 12 specs), explicitly approve every one of them (`listingRepo.SetStatus(ctx, l.ID, model.ListingStatusApproved)`) — the seed script's whole point is a realistic, immediately-usable marketplace, so seeded listings shouldn't sit stuck in the new `pending` default forever.
     - Add 2–3 new listing specs left at the default `pending` status (don't call `SetStatus` on them) — realistic-looking new/edited-style entries so the admin panel's review queue (AP-FE-06) has something to show immediately after seeding, without requiring a human to manually create a pending listing first.
     - Log counts on completion, e.g. `log.Printf("created %d listings (%d approved, %d pending review)", ...)`, mirroring the existing summary logging pattern already used for hives/inspections/etc.
   - Depends on AP-BE-02 (the `Status` field existing on `model.Listing`) and the `SetStatus` repository method from AP-BE-04 above.

---

### Phase 3: Backend Business Logic (5 tasks)

9. **AP-BE-05: `RequireAdmin` middleware**
   - File: `backend/internal/middleware/admin.go`
   - `RequireAdmin(users UserStore) func(http.Handler) http.Handler` — wraps a handler that already ran through `Auth`; reads `userID` from context (`middleware.UserIDFromContext`), loads the user, checks `IsAdmin()`, else `403 FORBIDDEN`. A DB lookup (not a JWT claim) is deliberate: revoking admin access takes effect on the next request, not after every outstanding token expires.
   - Composed in `main.go` as `admin := func(h http.Handler) http.Handler { return auth(middleware.RequireAdmin(userRepo)(h)) }`, matching the existing `auth`/`optionalAuth` composition pattern.

10. **AP-BE-06: Listing moderation defaults**
   - File: `backend/internal/service/listing.go`
   - `Create` now sets `Status: ListingStatusPending` (was implicitly always-visible).
   - `Update` sets `Status: ListingStatusPending`, clears `ReviewedBy`/`ReviewedAt`/`RejectionReason`, but **leaves `FirstApprovedAt` untouched** (the signal that distinguishes "edit" from "new" in the admin queue).
   - `Search`/`Get` (public/anonymous paths) filter to `Status = ListingStatusApproved` in addition to the existing `IsHidden` check. Owner-scoped views (`My Listings`) are unfiltered by status so the seller can see `pending`/`rejected` items and the rejection reason.

11. **AP-BE-07: `ListingModerationService`**
    - File: `backend/internal/service/listing_moderation.go`
    - `ListPending(ctx, limit, offset) ([]*model.Listing, int64, error)`
    - `Approve(ctx, adminID, listingID) error` — sets `status='approved'`, `reviewed_by`/`reviewed_at`, and `first_approved_at` if it was still nil.
    - `Reject(ctx, adminID, listingID, reason string) error` — `reason` required, non-empty (mirrors the confirmed decision).

12. **AP-BE-08: Certification review gate**
    - File: `backend/internal/service/honey_batch.go` (extends the existing `HoneyBatchService`, not a new package — it already owns `RequestCertification`)
    - `RequestCertification` (the code behind HC-BE-24c) is changed: instead of creating a `blockchain_jobs` row directly, it creates a `honey_batch_certification_requests` row (`status='pending'`) — after the same live-certification idempotency check that already exists (HC-BE-25 layer 2), reused as-is. Response to the owner changes from "queued" to "pending admin review".
    - New file: `backend/internal/service/certification_review.go` — `CertificationReviewService`:
      - `ListPending(ctx, limit, offset)`
      - `Approve(ctx, adminID, requestID) error` — in one transaction: mark the request `approved`, then run the *exact* existing enqueue logic from the old HC-BE-24c (`blockchain_jobs` insert, `status='queued'`, `next_retry_at=NOW()`), and stamp `blockchain_job_id` on the request row. From this point on, the worker (HC-BE-15b/15c) is completely unaware anything changed — it just sees a new queued job, same as before.
      - `Reject(ctx, adminID, requestID, reason string) error` — reason required; batch stays uncertified, owner can resubmit (creates a fresh request row, since the partial-unique index only blocks concurrent *pending* requests, not sequential ones).

13. **AP-BE-09: Admin PDF/photo access**
    - File: `backend/internal/handler/honey_batch.go`, `backend/internal/handler/listing_image.go`
    - The existing owner-scoped PDF handler (`GET /api/v1/honey-batches/{id}/pdf`) gets an admin bypass: ownership check passes if `userID == batch.UserID` **or** the caller is an admin (same pattern as the moderation gate — DB-checked, not JWT-claimed).
    - Listing images are already served unauthenticated (`GET /api/v1/listings/{id}/images/{imageId}/file`), so no change needed there — the admin panel can load them directly, including for `pending` listings, since that route has no status filter today (confirmed: this is not a new information leak, pending listing images were already fetchable by anyone with the direct URL, same as before this feature).

---

### Phase 4: Backend API Handlers (10 tasks)

14. **AP-BE-10: `GET /api/v1/admin/listings`**
    - File: `backend/internal/handler/admin_listing.go`
    - Admin-only. Query params: `status` (default `pending`), `limit`, `offset`.
    - Response includes `is_edit: bool` computed as `first_approved_at != null` — so the React panel doesn't need to reimplement that logic.

15. **AP-BE-11: `GET /api/v1/admin/listings/{id}`**
    - Full listing detail including images — admin can view regardless of `status`/`IsHidden`.

16. **AP-BE-12: `POST /api/v1/admin/listings/{id}/approve`**
    - Admin-only. 404 if not found, 409 if already `approved`.

17. **AP-BE-13: `POST /api/v1/admin/listings/{id}/reject`**
    - Body: `{reason: string}`, `reason` required (400 if empty).

18. **AP-BE-14: `GET /api/v1/admin/certification-requests`**
    - File: `backend/internal/handler/admin_certification.go`
    - Query params: `status` (default `pending`), `limit`, `offset`. Each item includes batch summary fields (gathering date, amount, honey type, owner email) inline — the admin queue view shouldn't need a second round-trip per row.

19. **AP-BE-15: `GET /api/v1/admin/certification-requests/{id}`**
    - Full detail: batch fields, PDF URL (points at the admin-bypass PDF route from AP-BE-09), requester info.

20. **AP-BE-16: `POST /api/v1/admin/certification-requests/{id}/approve`**
    - 404 if not found, 409 if not `pending`.

21. **AP-BE-17: `POST /api/v1/admin/certification-requests/{id}/reject`**
    - Body: `{reason: string}`, required.

22. **AP-BE-18: `GET /api/v1/admin/honey-batches/{id}/pdf`**
    - Thin alias onto the admin-bypass route from AP-BE-09, under the `/admin/` prefix for consistency with the rest of this API surface (avoids the React panel needing to special-case one non-`/admin/`-prefixed URL).

23. **AP-BE-19: `GET /api/v1/users/me` role field**
    - File: `backend/internal/handler/user.go` (existing handler, minor extension)
    - Include `role` in the response so the React panel can confirm admin status client-side (in addition to the server-side `RequireAdmin` check on every actual admin route — this is UX only, never a security boundary).

---

### Phase 5: Backend Integration & Wiring (1 task)

24. **AP-BE-20: Wire into `main.go`**
    - File: `backend/cmd/api/main.go`
    - Construct `admin := func(h http.Handler) http.Handler { return auth(middleware.RequireAdmin(userRepo)(h)) }`
    - Register routes: `mux.Handle("GET /api/v1/admin/listings", admin(...))`, and the rest of AP-BE-10 through AP-BE-18, following the existing `mux.Handle("METHOD /path", wrapper(http.HandlerFunc(handler.Method)))` convention already used throughout the file.
    - CORS: the React panel runs on a different origin/port than the Flutter web build during development — extend the existing CORS middleware's allowed-origins list (or add a dedicated one scoped to `/api/v1/admin/*`) to include the admin panel's dev/deploy origin.

---

### Phase 6: React Admin Panel (10 tasks)

**Stack:** Vite + React + TypeScript. No CSS framework required (functional over pretty, per the confirmed scope) — plain CSS modules or a minimal utility stylesheet. No new backend-for-frontend layer; calls the Go API directly with a Bearer token.

25. **AP-FE-01: Project scaffold**
    - Directory: `admin/` (new top-level directory, sibling to `app/` and `backend/`)
    - `admin/package.json`, `admin/vite.config.ts`, `admin/tsconfig.json`, `admin/index.html`
    - `admin/.env.example` — `VITE_API_BASE_URL` (points at the Go API)

26. **AP-FE-02: API client**
    - File: `admin/src/api/client.ts`
    - Thin `fetch` wrapper: attaches `Authorization: Bearer <token>` from `localStorage`, JSON encode/decode, throws a typed `ApiError` on non-2xx (mirrors the Dart repository pattern's `ApiException`, for consistency in spirit even though it's a different language).
    - File: `admin/src/api/auth.ts` — `login(email, password)`, `logout()`, `getStoredToken()`.
    - File: `admin/src/api/listings.ts` — `listPendingListings`, `getListing`, `approveListing`, `rejectListing`.
    - File: `admin/src/api/certifications.ts` — `listPendingCertificationRequests`, `getCertificationRequest`, `approveCertificationRequest`, `rejectCertificationRequest`.

27. **AP-FE-03: Auth state + route guard**
    - File: `admin/src/auth/AuthContext.tsx`
    - On login, calls `POST /api/v1/auth/login`, then `GET /api/v1/users/me` to confirm `role === 'admin'`; rejects (client-side toast, "not an admin account") if not, even though the actual routes are already server-enforced — avoids a confusing silent-403 UX for a non-admin who tries to log in.
    - `admin/src/auth/RequireAuth.tsx` — route wrapper redirecting to `/login` when unauthenticated.

28. **AP-FE-04: Login page**
    - File: `admin/src/pages/LoginPage.tsx`
    - Email + password form, calls `auth.login`, redirects to `/listings` on success.

29. **AP-FE-05: App shell + nav**
    - File: `admin/src/App.tsx`, `admin/src/components/Layout.tsx`
    - Simple top nav: "Listings" | "Certifications" | logout button. React Router for the two queue sections.

30. **AP-FE-06: Listings queue page**
    - File: `admin/src/pages/ListingsQueuePage.tsx`
    - Table: title, category, owner, submitted date, "New" or "Edited" badge (from `is_edit`), link to detail.
    - Pagination controls (reuses `limit`/`offset` from AP-BE-10).

31. **AP-FE-07: Listing detail/review page**
    - File: `admin/src/pages/ListingDetailPage.tsx`
    - All listing fields (title, description, category, price, quantity, address, contact info, apiary link if any), photo gallery (`<img>` tags pointed directly at the existing public image-file URLs from AP-BE-09), Approve button, Reject button opening a reason textarea (required, disabled submit until non-empty).

32. **AP-FE-08: Certification queue page**
    - File: `admin/src/pages/CertificationQueuePage.tsx`
    - Table: batch honey type, amount, gathering date, owner, submitted date, link to detail.

33. **AP-FE-09: Certification detail/review page**
    - File: `admin/src/pages/CertificationDetailPage.tsx`
    - Batch fields, embedded PDF viewer (`<iframe>` or `<embed>` pointed at AP-BE-18's URL with the admin's Bearer token — since that route is authenticated, the iframe src needs either a short-lived signed link or a fetch-then-blob-URL approach; **fetch-then-blob-URL** is simpler and avoids adding a signing mechanism), Approve/Reject (reason required) buttons.

34. **AP-FE-10: Docker/dev wiring**
    - File: `docker/docker-compose.yml` (extend) or `admin/Dockerfile` — a dev-only Vite dev server target is enough for the thesis; a production nginx-served static build is **(production-hardening — optional)**, noted but not required.
    - `README.md` or `docs/api.md` gets a short "Admin Panel" section: how to run it (`cd admin && npm install && npm run dev`), how to create the first admin (the manual `UPDATE users SET role='admin' ...` from AP-DB-01).

---

### Phase 7: Flutter App Changes (3 tasks)

**Goal:** the regular mobile/web app needs to surface the new pending/rejected states — sellers and beekeepers need to know why their listing vanished or their certification hasn't started.

35. **AP-10-01: Listing status badge (My Listings)**
    - File: `app/lib/features/marketplace/view/my_listings_screen.dart`, `app/lib/features/marketplace/data/listing_model.dart` (add `status`, `rejectionReason` fields)
    - Badge: "Pending review" (amber), "Rejected" (red, tap to see reason), "Live" (green) — approved listings show no badge or a subtle "Live" tag, since that's the expected default state.
    - Public/other-user-facing screens (`marketplace_home_screen.dart`, `marketplace_map_screen.dart`) need no change — the backend already excludes non-approved listings from those queries (AP-BE-06).

36. **AP-10-02: Certification request status (Honey batch card)**
    - File: `app/lib/features/honey/widgets/blockchain_status_widget.dart`, `app/lib/features/honey/data/honey_batch_certification_model.dart`
    - New pre-blockchain state surfaced to the UI: "Pending admin review" (distinct from `queued`, which now only appears *after* approval) and "Rejected by admin" (with reason) if the review request was rejected. These aren't values of the existing `CertificationStatus` enum (HC-FE-08b) — they describe the certification *request*, not the on-chain job — so the card needs to check for an active/rejected `certification_request` in addition to the existing nullable `certification`.
    - "Certify" button behavior unchanged from the user's perspective (HC-BE-24c's contract stays the same); the response now reflects "pending review" instead of "queued" immediately after tapping it.

37. **AP-10-03: Localization**
    - File: `app/lib/l10n/app_en.arb` + `app_pl.arb`
    - Keys: `listingStatus_pending`, `listingStatus_rejected`, `listingStatus_rejectionReason`, `certificationRequest_pendingReview`, `certificationRequest_rejected`.
    - Run `flutter gen-l10n`.

---

## Critical Files Summary

### New Files
- **Backend:**
  - `backend/migrations/036_add_user_role.sql`
  - `backend/migrations/037_add_listing_moderation.sql`
  - `backend/migrations/038_create_honey_batch_certification_requests.sql`
  - `backend/internal/model/honey_batch_certification_request.go`
  - `backend/internal/repository/honey_batch_certification_request.go`
  - `backend/internal/middleware/admin.go`
  - `backend/internal/service/listing_moderation.go`
  - `backend/internal/service/certification_review.go`
  - `backend/internal/handler/admin_listing.go`
  - `backend/internal/handler/admin_certification.go`
- **Admin panel (all new):** `admin/` — see Phase 6 file list above.

### Files to Modify
- **Backend:**
  - `backend/internal/model/user.go` — add `Role`
  - `backend/internal/model/listing.go` — add moderation fields
  - `backend/internal/repository/listing.go` — status filtering, moderation queries
  - `backend/internal/service/listing.go` — default-to-pending on create/edit, filter public reads
  - `backend/internal/service/honey_batch.go` — `RequestCertification` creates a review request instead of a job
  - `backend/internal/handler/honey_batch.go` — admin PDF bypass
  - `backend/internal/handler/listing_image.go` — no functional change, confirmed already admin-accessible
  - `backend/internal/handler/user.go` — include `role` in `/users/me`
  - `backend/cmd/api/main.go` — `admin` middleware composition, route registration, CORS
- **Frontend (Flutter):**
  - `app/lib/features/marketplace/data/listing_model.dart`, `my_listings_screen.dart`
  - `app/lib/features/honey/data/honey_batch_certification_model.dart`, `blockchain_status_widget.dart`
  - `app/lib/l10n/app_en.arb`, `app_pl.arb`

---

## Testing Strategy

### Backend
- **Unit:** `listing_moderation_test.go` (create → pending, edit resets to pending but keeps `first_approved_at`, approve/reject transitions, reason required on reject), `certification_review_test.go` (request → approve creates exactly one `blockchain_jobs` row with the same shape the old direct-enqueue path produced, reject leaves no job, resubmission after rejection works), `admin_middleware_test.go` (non-admin gets 403, admin passes, unauthenticated gets 401 same as any other route).
- **Integration:** approving a certification request end-to-end through to the existing worker picking it up and reaching `confirmed` (reuses the existing worker test harness from the honey blockchain feature — this is the regression check that the gate didn't break anything downstream).

### Admin Panel
- Manual testing only for the thesis scope (per CLAUDE.md, `flutter analyze`-equivalent automated frontend tests aren't required here since this is a new stack); a smoke-test checklist substitutes:
  1. Login as non-admin → rejected with a clear message.
  2. Login as admin → see pending listings queue.
  3. Approve a new listing → confirm it appears in the Flutter app's public marketplace.
  4. Reject a listing with a reason → confirm the Flutter app's My Listings screen shows the reason.
  5. Edit an approved listing as a normal user → confirm it disappears from public marketplace and reappears in the admin queue tagged "Edited".
  6. Approve a certification request → confirm a `blockchain_jobs` row appears and the existing worker drains it to `confirmed`, same as before this feature existed.
  7. Reject a certification request → confirm the batch owner sees the reason and can resubmit.

---

## Known Constraints & Decisions

1. **Manual admin provisioning only:** no signup/promotion flow exists or is planned — `role='admin'` is set by direct SQL. Acceptable because BeeTrack has a small, known set of operators (per the thesis scope), same reasoning as the honey-blockchain plan's testnet-only scope.
2. **DB-checked, not JWT-claimed, admin status:** every admin route re-checks `users.role` per request rather than trusting a claim baked into the JWT at login time — revocation is immediate, at the cost of one extra row lookup per admin request (negligible).
3. **No revision/diff table for listing edits:** an edit simply flips the live listing back to `pending`; the previous approved version isn't preserved for the admin to diff against. Simpler at the cost of the seller's edit being all-or-nothing to review. Matches the confirmed decision.
4. **Certification approval reuses 100% of the existing worker/job infrastructure:** the new review-request table sits strictly upstream of `blockchain_jobs`; nothing in `blockchain_worker.go`, the idempotency guarantees (HC-BE-25), or the retry logic (HC-10-02) changes. This was a deliberate constraint to avoid destabilizing an already-shipped, tested feature.
5. **Listing images remain unauthenticated:** no new access-control change there — pending-listing photos were already fetchable via direct URL before this feature (nothing currently gates that route by listing status), so the admin panel loading them introduces no new exposure.
6. **PDF viewing in the admin panel uses fetch-then-blob-URL**, not a signed link scheme, to avoid adding a new short-lived-token mechanism for what is a thesis-scope, low-traffic feature.
