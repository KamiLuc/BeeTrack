# BeeTrack — Developer Wiki

Quick reference for AI-assisted sessions. Read this before grepping the codebase.

---

## Project Layout

```
backend/           # Go API
  cmd/api/         # main.go entry point
  internal/
    handler/       # HTTP handlers (one file per domain)
    service/       # Business logic
    repository/    # DB queries (sqlx + raw SQL)
    middleware/    # Auth JWT middleware
  migrations/      # goose SQL migrations (NNN_description.sql)
  pkg/             # Shared types (e.g. apierror)

app/               # Flutter app
  lib/
    core/
      api/         # ApiClient (Dio wrapper), ApiException
      storage/     # TokenStorage (JWT in SharedPreferences)
      theme/       # AppColors, AppTextStyles, AppLayout, AppTheme
    features/
      auth/        # login/register — BLoC pattern
      apiary/      # apiary CRUD — Cubit pattern
      hive/        # hive CRUD + grid view + filter — Cubit pattern
      inspection/  # inspection CRUD + history — Cubit pattern
      treatment/   # treatment CRUD + history — Cubit pattern
      harvest/     # harvest CRUD + history — Cubit pattern
      home/        # HomeScreen (shell after login)
    l10n/          # ARB files (app_en.arb, app_pl.arb) + generated classes
    main.dart
  test/
    features/
      auth/        # auth_bloc_test.dart
      apiary/      # apiaries_cubit_test.dart
      hive/        # hives_cubit_test.dart, hive_detail_screen_test.dart
      inspection/  # inspections_cubit_test.dart
      treatment/   # treatments_cubit_test.dart, treatment_form_fields_test.dart
      harvest/     # harvests_cubit_test.dart, harvest_form_fields_test.dart

docker/            # Docker Compose config
```

---

## Backend Architecture

### Request flow

```
HTTP request → middleware (Auth JWT) → Handler → Service → Repository → DB
```

- **Handler** (`internal/handler/`) — decodes JSON, extracts path params and user ID from context, calls service, maps errors to HTTP codes, writes JSON response via `respond.JSON` / `respond.Error`.
- **Service** (`internal/service/`) — owns all business logic and validation. Never touches HTTP. Returns sentinel errors; callers `errors.Is()` to distinguish them.
- **Repository** (`internal/repository/`) — raw DB queries via GORM. No business logic. Returns GORM errors directly (e.g. `gorm.ErrRecordNotFound`).

### Error handling pattern

Sentinel errors are declared as package-level `var` in the service file:

```go
var (
    ErrApiaryNotFound  = errors.New("apiary not found")
    ErrForbidden       = errors.New("forbidden")
    ErrGridTooSmall    = errors.New("grid is too small to fit all existing hives")
    ErrInvalidGridSize = errors.New("grid rows and cols must be at least 1")
    ErrNameRequired    = errors.New("name is required")
)
```

Handler maps them:
```go
switch {
case errors.Is(err, service.ErrGridTooSmall):
    respond.Error(w, http.StatusUnprocessableEntity, "GRID_TOO_SMALL", err.Error())
case errors.Is(err, service.ErrForbidden):
    respond.Error(w, http.StatusForbidden, "FORBIDDEN", "...")
default:
    respond.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "...")
}
```

Wrapped errors propagate with `fmt.Errorf("context: %w", err)`.

### Service interface pattern (dependency inversion)

Each service defines its own repository interface — only the methods it actually needs:

```go
// In service/apiary.go
type ApiaryRepository interface { ... }
type HiveRelocator interface {          // subset of HiveRepository, only what ApiaryService needs
    ListByApiaryID(...) ([]*model.Hive, error)
    Move(...) error
}

type ApiaryService struct {
    apiaries ApiaryRepository
    hives    HiveRelocator
}
```

This makes services independently testable with small mock structs.

### Testing pattern (Go)

Tests live in the same package (`package service`). Mocks are plain structs, not generated:

```go
type mockApiaryRepo struct {
    apiary   *model.Apiary
    role     string
    updated  *model.Apiary
}
func (m *mockApiaryRepo) GetMembership(...) (*model.Apiary, string, error) { ... }
// implement only the methods the test exercises
```

Helper creates service + repos together:
```go
func newTestApiaryService() (*ApiaryService, *mockApiaryRepo, *mockHiveRelocator) { ... }
```

### Wiring (`cmd/api/main.go`)

Repositories are concrete structs; services accept interfaces:
```go
apiaryRepo := repository.NewApiaryRepository(db)
hiveRepo   := repository.NewHiveRepository(db)

apiarySvc := service.NewApiaryService(apiaryRepo, hiveRepo)
hiveSvc   := service.NewHiveService(apiaryRepo, hiveRepo)
```

### Middleware

- `middleware.Auth(jwtSecret)` — validates Bearer token, injects `userID` into context.
- `middleware.CORS(allowedOrigins)` — wraps the whole mux.
- `middleware.UserIDFromContext(ctx)` — extracts userID; returns `(int64, bool)`.

### Doc comment convention

Every exported function in handler, service, and repository must have a doc comment:
```go
// Create handles POST /api/v1/apiaries — creates a new apiary owned by the authenticated user.
func (h *ApiaryHandler) Create(w http.ResponseWriter, r *http.Request) { ... }
```

### Migrations

Files live in `backend/migrations/`, run automatically on startup via goose:
```sql
-- +goose Up
CREATE TABLE ...;

-- +goose Down
DROP TABLE ...;
```
Naming: `NNN_description.sql` (e.g. `003_create_hives.sql`).

### Models (`internal/model/`)

```go
// user.go
type User struct {
    ID           int64
    Email        string
    Name         string
    PasswordHash string
    Verified     bool      // false until email verification link is clicked
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// email_token.go
type EmailVerificationToken struct {
    ID        int64
    UserID    int64
    Token     string    // base64url random, expires in 24h
    ExpiresAt time.Time
    CreatedAt time.Time
}
type PasswordResetToken struct {
    ID        int64
    UserID    int64
    Token     string    // base64url random, expires in 1h
    ExpiresAt time.Time
    CreatedAt time.Time
}

// apiary.go
type Apiary struct {
    ID          int64
    OwnerUserID int64
    Name        string
    Lat, Lng    *float64
    GridRows    int
    GridCols    int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
type ApiaryMembership struct {
    Apiary    *Apiary
    UserRole  string
    HiveCount int
}

// hive.go
type Hive struct {
    ID              int64
    ApiaryID        int64
    Name            string
    Type            string   // dadant | langstroth | top_bar | wielkopolski
    Active          bool
    Queenless       bool
    ReadyForHarvest bool
    Frames          int      // total frame capacity; 0 means not set
    GridRow         int
    GridCol         int
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
type HiveDisease struct {
    ID        int64
    HiveID    int64
    Disease   string
    CreatedAt time.Time
}
type Inspection struct {
    ID                    int64
    HiveID                int64
    InspectedBy           int64
    InspectedAt           time.Time
    QueenStatus           string   // seen | not_seen
    BroodPattern          string   // excellent | good | poor | none
    FramesBrood           *int
    FramesHoney           *int
    FramesPollen          *int
    QueenCellsCount       *int
    Aggressiveness        string   // calm | mild | aggressive | very_aggressive
    FramesAddedFoundation *int
    FramesAddedDrawn      *int
    FramesAddedHoney      *int
    QueenAdded            bool
    Notes                 string
    CreatedAt             time.Time
    UpdatedAt             time.Time
}
type InspectionImage struct {
    ID           int64
    InspectionID int64
    Filename     string   // UUID-based, stored under IMAGE_STORAGE_PATH
    MimeType     string   // image/jpeg | image/png | image/webp
    CreatedAt    time.Time
}
```

### Adding a new backend feature (checklist)

1. `internal/model/x.go` — struct(s)
2. `internal/repository/x.go` — DB queries, no logic
3. `internal/service/x.go` — interface + service struct + sentinel errors + business logic
4. `internal/service/x_test.go` — mock structs + unit tests
5. `internal/handler/x.go` — HTTP handler, error mapping
6. `cmd/api/main.go` — wire repo → service → handler, register routes
7. `backend/migrations/NNN_x.sql` — schema changes if needed

---

## Flutter Architecture

### State management
- **Cubit** for data features (apiaries, hives). Sealed state classes: `Initial | Loading | Loaded | Error`.
- **BLoC** for auth only.
- Cubits are scoped per-screen via `BlocProvider` in the screen's `build()`.
- Global providers (in `main.dart`): `TokenStorage`, `ApiClient`.

### Navigation
- Imperative `Navigator.push` / `MaterialPageRoute` everywhere — no named routes, no GoRouter.
- Pattern after returning from a pushed screen:
  ```dart
  await Navigator.of(context).push(...);
  if (context.mounted) context.read<SomeCubit>().load();
  ```
- `EditHiveScreen` pops with the updated `Hive` object as result so callers can update state without a refetch.
- Logout: `AuthWrapper` in `main.dart` uses a `BlocListener` (alongside `BlocBuilder`) that calls `Navigator.of(context).popUntil((route) => route.isFirst)` on `AuthUnauthenticated`, clearing any pushed screens before the BlocBuilder swaps home to `LoginScreen`.

### Repository pattern
- Each feature has a `XRepository` class that wraps `ApiClient`.
- Repositories are instantiated inline: `XRepository(api: context.read())`.
- They are NOT in the widget tree as providers.

### Adding a new feature (checklist)
1. `data/x_model.dart` — model + `fromJson`
2. `data/x_repository.dart` — API calls via `context.read<ApiClient>()`
3. `cubit/x_cubit.dart` + `cubit/x_state.dart` — sealed states
4. `view/x_screen.dart` — screen widget with `BlocProvider`
5. `l10n/app_en.arb` + `app_pl.arb` — new strings (Flutter regenerates on build)
6. `test/features/x/x_cubit_test.dart` — unit tests (bloc_test + mocktail)

---

## Key Files

| File | Purpose |
|------|---------|
| `app/lib/core/theme/app_colors.dart` | Color constants (primary = amber `#FBBF24`, background = cream `#FFFBF2`, onSurfaceVariant = warm grey `#6F6961`) |
| `app/lib/core/theme/app_text_styles.dart` | Text style constants (headlineLarge 28px … caption 12px) |
| `app/lib/core/theme/app_layout.dart` | `AppLayout.formConstraints(context)` — 85% width on phone, 40% on tablet; `AppLayout.bannerWidth(context)` — shared amber banner width (85% phone, min(440, 40%) tablet) |
| `app/lib/features/hive/view/hive_form_widgets.dart` | `HiveNameField`, `HiveTypeDropdown`, `HiveActiveToggle`, `HiveDiseasesSection`, `HiveFramesField`, `hiveDiseaseLabel()`, `hiveTypeLabels` map |
| `app/lib/features/apiary/view/apiary_form_widgets.dart` | `ApiaryGridSection`, `ApiaryLocationSection` |
| `app/lib/features/treatment/view/treatment_form_fields.dart` | `TreatmentFormFields` — shared form body (date picker, medicine autocomplete, dose, notes); used by both `TreatmentFormScreen` and `BulkTreatmentFormScreen` |
| `app/lib/features/apiary/view/apiaries_map_screen.dart` | Full-screen `FlutterMap` showing all located apiaries; three concentric circles per pin (green 1.5 km, orange 3 km, red 5 km, drawn outermost-first); marker tooltip shows apiary name |
| `app/lib/core/widgets/delete_dialog.dart` | `showDeleteDialog()` — simple confirm or math-puzzle confirm (`withPuzzle: true`); puzzle is a proper `StatefulWidget` (`_PuzzleDialog`) — clears error on typing, disposes controller in `dispose()` |
| `app/lib/features/inspection/view/inspection_summary.dart` | Shared `InspectionSummary` widget — renders labelled sections (Observations, Frames with added sub-row, queen cells, Notes); each section header uses `labelStyle` (small, primary-coloured); used in hive detail card and inspection history cards |
| `app/lib/features/inspection/data/inspection_image_model.dart` | `InspectionImage` — id, inspectionId, mimeType, createdAt; `fromJson` factory |
| `app/lib/features/inspection/data/inspection_image_repository.dart` | `listImages`, `uploadImage` (multipart via Dio), `deleteImage`, `imageUrl()` (builds full URL from `apiClient.baseUrl`), `authHeaders()` |
| `app/lib/l10n/app_en.arb` | Source of truth for all UI strings |

---

## Data Models (current fields)

```
Hive             id, apiaryId, name, type, active, queenless, readyForHarvest,
                 frames (int, 0 = not set), gridRow, gridCol,
                 diseases (List<HiveDisease>), lastInspectedAt?
Apiary           id, name, lat?, lng?, gridRows, gridCols, hiveCount, userRole, lastInspectedAt?
HiveDisease      id, disease (string)
Inspection       id, hiveId, inspectedAt, queenSeen, broodPattern, aggressiveness,
                 framesBrood?, framesHoney?, framesPollen?, framesAddedDrawn?,
                 framesAddedFoundation?, framesAddedHoney?, queenCellsCount?,
                 queenAdded, notes, photoCount (int, default 0)
InspectionImage  id, inspectionId, mimeType, createdAt
                 (URL built from apiClient.baseUrl + REST path)
Treatment        id, hiveId, treatedAt, medicineName, dose (string, default "1"),
                 notes, treatedByName? (populated via JOIN, not stored in table)
Harvest          id, hiveId, harvestedAt, frames (int, default 1), halfFrames (int, default 0),
                 kilograms (double, 2dp), notes, harvestedByName? (populated via JOIN)
```

Hive types (valid values): `dadant`, `langstroth`, `top_bar`, `wielkopolski`  
Display labels live in `hiveTypeLabels` map in `hive_form_widgets.dart`.

---

## Backend API Endpoints (implemented)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register user — sends verification email (`lang` field sets email language) |
| POST | `/api/v1/auth/login` | Login (requires verified account; returns `EMAIL_NOT_VERIFIED` if not) |
| POST | `/api/v1/auth/refresh` | Refresh tokens |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/auth/verify-email?token=&lang=` | Verify email — returns HTML page (called from email link) |
| POST | `/api/v1/auth/resend-verification` | Resend verification email (`lang` field sets email language) |
| POST | `/api/v1/auth/forgot-password` | Request password reset email (`lang` field sets email language) |
| GET | `/api/v1/auth/reset-password-form?token=&lang=` | HTML password reset form (called from email link) |
| POST | `/api/v1/auth/reset-password-form` | Submit HTML reset form (form-encoded) |
| POST | `/api/v1/auth/reset-password` | Reset password via JSON — for mobile clients |
| PATCH | `/api/v1/users/me/name` | Update display name |
| GET | `/api/v1/apiaries` | List apiaries (includes hive_count + last_inspected_at per apiary) |
| POST | `/api/v1/apiaries` | Create apiary |
| PATCH | `/api/v1/apiaries/{id}` | Update apiary |
| DELETE | `/api/v1/apiaries/{id}` | Delete apiary |
| GET | `/api/v1/apiaries/{id}/hives` | List hives (includes diseases + last_inspected_at) |
| POST | `/api/v1/apiaries/{id}/hives` | Create hive |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}` | Get hive |
| PATCH | `/api/v1/apiaries/{id}/hives/{hiveId}` | Update hive |
| PATCH | `/api/v1/apiaries/{id}/hives/{hiveId}/position` | Move hive |
| PATCH | `/api/v1/apiaries/{id}/hives/{hiveId}/frames` | Atomically increment hive frame count by `delta` |
| DELETE | `/api/v1/apiaries/{id}/hives/{hiveId}` | Delete hive |
| POST | `/api/v1/apiaries/{id}/hives/{hiveId}/diseases` | Add hive disease |
| DELETE | `/api/v1/apiaries/{id}/hives/{hiveId}/diseases/{diseaseId}` | Remove hive disease |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections` | List inspections (paginated) |
| POST | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections` | Create inspection |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}` | Get inspection |
| PATCH | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}` | Update inspection |
| DELETE | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}` | Delete inspection |
| POST | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/diseases` | Add inspection disease |
| DELETE | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/diseases/{diseaseId}` | Remove inspection disease |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images` | List inspection images |
| POST | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images` | Upload image (multipart, field `image`; max 10 MB; jpeg/png/webp; max 6 per inspection) |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images/{imageId}/file` | Serve image bytes (auth-gated, `Cache-Control: private, max-age=86400`) |
| DELETE | `/api/v1/apiaries/{id}/hives/{hiveId}/inspections/{inspectionId}/images/{imageId}` | Delete image |
| GET | `/api/v1/medicines` | List known medicine names (no auth required) — 13 predefined varroa/disease treatments |
| POST | `/api/v1/apiaries/{id}/treatments/bulk` | Create one treatment per hive in the apiary (same body as single; wrapped in a transaction; returns `{"count": N}`) |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}/treatments` | List treatments (paginated, `limit`/`offset` query params) |
| POST | `/api/v1/apiaries/{id}/hives/{hiveId}/treatments` | Create treatment (`treated_at`, `medicine_name`, `dose`, `notes`; dose defaults to "1") |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}/treatments/{treatmentId}` | Get treatment |
| PATCH | `/api/v1/apiaries/{id}/hives/{hiveId}/treatments/{treatmentId}` | Update treatment |
| DELETE | `/api/v1/apiaries/{id}/hives/{hiveId}/treatments/{treatmentId}` | Delete treatment |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}/harvests` | List harvests (paginated, `limit`/`offset` query params) |
| POST | `/api/v1/apiaries/{id}/hives/{hiveId}/harvests` | Create harvest (`harvested_at`, `frames`, `half_frames`, `kilograms`, `notes`; frames must sum > 0, kilograms > 0) |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}/harvests/{harvestId}` | Get harvest |
| PATCH | `/api/v1/apiaries/{id}/hives/{hiveId}/harvests/{harvestId}` | Update harvest |
| DELETE | `/api/v1/apiaries/{id}/hives/{hiveId}/harvests/{harvestId}` | Delete harvest |

---

## Flutter Widget Patterns

### Confirmation dialog
```dart
final confirmed = await showDialog<bool>(
  context: context,
  builder: (ctx) => AlertDialog(
    title: Text(...),
    content: Text(...),
    actions: [
      TextButton(onPressed: () => Navigator.of(ctx).pop(false), child: Text(l10n.generalCancel)),
      TextButton(
        onPressed: () => Navigator.of(ctx).pop(true),
        child: Text(l10n.generalDelete, style: TextStyle(color: Theme.of(ctx).colorScheme.error)),
      ),
    ],
  ),
);
```

### Screen with form
- `SafeArea > Center > SingleChildScrollView > ConstrainedBox(AppLayout.formConstraints) > Form > Column`
- Submit button: fixed `width: 200`, shows `CircularProgressIndicator` while loading.

### Cubit screen shell
```dart
BlocProvider(
  create: (_) => XCubit(repo: XRepository(api: context.read()), ...)..load(),
  child: _XView(...),
)
```

---

## Localization

- Add strings to both `app_en.arb` and `app_pl.arb`.
- Also manually update the abstract class in `app_localizations.dart` and both `app_localizations_en.dart` / `app_localizations_pl.dart` — auto-generation is not wired up.
- Access: `AppLocalizations.of(context)!.yourKey`
- Plurals use ICU format; Polish needs `=1 / few / many / other` forms.

---

## Testing Conventions

- **Cubit tests**: `bloc_test` + `mocktail`. Mock the repository, test state transitions.
- **Widget tests**: `flutter_test`. Wrap with `MaterialApp` + `AppLocalizations` delegates + `locale: Locale('en')`. No network mocking needed for render-only tests.
- Test file lives at `test/features/<domain>/`.
- Run all tests: `flutter test` from `app/` directory.

---

## Screen Navigation Map

```
LoginScreen / RegisterScreen
  └── ApiariesScreen (after login)
      │   Shows apiary cards (name, hive count, last inspection date).
      │   Empty state: centered Add button only.
      │   Non-empty: bottom amber banner with two buttons:
      │     • + (add) — always enabled → CreateApiaryScreen
      │     • map — enabled if any apiary has GPS → ApiariesMapScreen
      ├── ApiariesMapScreen (map banner → map button)
      │     FlutterMap showing pins + 3 concentric circles per apiary
      │     (green 1.5 km, orange 3 km, red 5 km, drawn outermost-first)
      └── ApiaryGridScreen (tap apiary card)
          │   Grid is zoomable/pannable via InteractiveViewer (pinch or trackpad scroll).
          │   Bottom amber banner has three icon buttons:
          │     • Filter (Icons.tune) — dialog with FilterChip toggles; badge shows active count;
          │       × close button in header
          │     • Hive list (Icons.format_list_bulleted, disabled when no hives) — dialog listing
          │       hives with last-inspection date, active/disease subtitle, status icons;
          │       × close button in header; "Treat all hives" OutlinedButton.icon at 60% width
          │       shown at bottom when apiary has > 1 hive — closes dialog then opens
          │       BulkTreatmentFormScreen; on return shows "Treatment logged for N hives" snackbar
          │     • Center view (Icons.center_focus_strong_outlined) — resets TransformationController
          │       to Matrix4.identity(), snapping pan/zoom back to initial position
          ├── AddHiveScreen (tap empty cell)
          ├── HiveDetailScreen (tap hive cell  OR  bottom-bar hive list → tap hive)
              └── EditHiveScreen (AppBar edit icon)
                  │   Delete button at the bottom; shows math puzzle if hive has inspections.
                  │   On delete: pops both EditHiveScreen and HiveDetailScreen (ApiaryGridScreen reloads).
                  ├── InspectionFormScreen (Add inspection button — copies frames from last inspection)
                  ├── TreatmentFormScreen (Log treatment button in hive detail)
                  │   Shared form body via TreatmentFormFields (date, medicine Autocomplete, dose, notes).
                  │   Bottom amber banner with ✓ saves.
                  ├── BulkTreatmentFormScreen (Treat all hives — opened from hive list dialog)
                  │   Same TreatmentFormFields body; on save POSTs to /apiaries/{id}/treatments/bulk;
                  │   pops with count (int) so caller can show snackbar.
                  ├── TreatmentHistoryScreen (View all — opened when last treatment exists)
                  │   Same amber banner + pagination pattern as InspectionHistoryScreen.
                  │   TreatmentCard: date · treatedByName (only when != current user) · medicine · dose
                  │   (dose shown as "1 dawka / 2 dawki / 5 dawek" when dose is a plain integer) ·
                  │   Note label + note text (labelStyle/bodyStyle same as InspectionSummary).
                  ├── InspectionHistoryScreen (View all — only shown when inspections exist)
                      │   Page-based pagination (10 per page). Bottom amber banner:
                      │     • + (add) — opens InspectionFormScreen
                      │     • ← prev / page number buttons / next → (hidden when only 1 page)
                      │   Page numbers show: 1 … currentPage … lastPage (ellipsis condenses middle).
                          └── InspectionFormScreen (add button in banner)

#### Hive list dialog (`_HiveListDialog`)
Opened via the list icon in the bottom banner. Shows all hives sorted by last inspection date:
- No inspection → top (needs attention)
- Oldest inspection → next
- Most recently inspected → bottom

Each row subtitle: `"d MMM yyyy · <note>"` (note truncated with ellipsis) or `"d MMM yyyy · Queen seen · Brood: good"` if note is empty. Diseases are shown as icons only (🦠) — never in the subtitle text. Inspection data is lazy-loaded concurrently (one `GET inspections?limit=1` per hive that has `lastInspectedAt`) when the dialog opens; rows update as responses arrive.
```

#### InspectionFormScreen — bottom amber banner
The form has a fixed amber banner at the bottom (same style as apiary/grid banners):
- **+** icon button — picks a photo from gallery (web) or gallery/camera (mobile); badge shows pending-upload count; disabled at the 6-photo limit.
- **✓** icon button — saves the form and uploads any pending photos in sequence; shows a spinner while running.

Photo gallery (below the Diseases section, only rendered when photos exist):
- Horizontal scrollable strip of `120×120` thumbnails (`180×180` on web), centred via `LayoutBuilder + ConstrainedBox(minWidth)`.
- Tap thumbnail → full-screen swipeable `PageView` viewer (`InteractiveViewer` per page, tap anywhere to close).
- × button on each thumbnail → delete from server (existing) or discard (pending).

### Delete confirmation pattern
- Simple `AlertDialog` for entities with no data (apiary with 0 hives, hive with 0 inspections).
- Math-puzzle dialog (`showDeleteDialog(..., withPuzzle: true)` from `delete_dialog.dart`) for entities with data — user must solve a random `a + b = ?` before the delete button becomes effective.

### Async context safety in StatelessWidget
Always capture a cubit/provider reference **before** an `await` in a `StatelessWidget` method — the element may be reassigned between the await completing and the next frame:
```dart
// ✓ correct
final cubit = context.read<SomeCubit>();
final confirmed = await showDialog(...);
if (confirmed) cubit.doSomething();

// ✗ wrong — context.read after await in StatelessWidget
final confirmed = await showDialog(...);
if (confirmed && context.mounted) context.read<SomeCubit>().doSomething();
```

### Docker — image storage
Images uploaded to inspections are stored on disk in a Docker volume:
- Volume `images_data` mounted at `/data/images` inside the `api` container.
- Path configurable via `IMAGE_STORAGE_PATH` env var (default `/data/images`).
- Files are UUID-named (e.g. `550e8400-e29b-41d4-a716-446655440000.jpg`).
- Cascade DB delete (via FK) cleans DB records; the service also removes files from disk before the parent inspection row is deleted.

### Docker — email (MailPit)
In development, all outgoing emails are caught by MailPit instead of being delivered:
- SMTP on port `1025` (used by the `api` container via `SMTP_HOST=mailpit`).
- Web UI at `http://localhost:8025` — inspect all sent emails here.
- No authentication required for MailPit.

Relevant env vars for the `api` container:

| Var | Dev default | Description |
|-----|-------------|-------------|
| `API_URL` | `http://localhost:8080` | Base URL of the API — used in verification/reset email links |
| `APP_URL` | `http://localhost:5000` | Flutter web app URL — reserved for future mobile deep links |
| `SMTP_HOST` | `mailpit` | SMTP server host |
| `SMTP_PORT` | `1025` | SMTP server port |
| `SMTP_USER` | _(empty)_ | SMTP username — leave empty for MailPit, set for production |
| `SMTP_PASS` | _(empty)_ | SMTP password |
| `SMTP_FROM` | `noreply@beetrack.app` | Sender address |
