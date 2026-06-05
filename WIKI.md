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
      home/        # HomeScreen (shell after login)
    l10n/          # ARB files (app_en.arb, app_pl.arb) + generated classes
    main.dart
  test/
    features/
      auth/        # auth_bloc_test.dart
      apiary/      # apiaries_cubit_test.dart
      hive/        # hives_cubit_test.dart, hive_detail_screen_test.dart
      inspection/  # inspections_cubit_test.dart

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
| `app/lib/core/theme/app_layout.dart` | `AppLayout.formConstraints(context)` — 85% width on phone, 40% on tablet |
| `app/lib/features/hive/view/hive_form_widgets.dart` | `HiveNameField`, `HiveTypeDropdown`, `HiveActiveToggle`, `HiveDiseasesSection`, `hiveDiseaseLabel()`, `hiveTypeLabels` map |
| `app/lib/features/apiary/view/apiary_form_widgets.dart` | `ApiaryGridSection`, `ApiaryLocationSection` |
| `app/lib/features/inspection/view/inspection_summary.dart` | Shared `InspectionSummary` widget — renders grouped observation/frame/note rows; used in hive detail card and inspection history cards |
| `app/lib/l10n/app_en.arb` | Source of truth for all UI strings |

---

## Data Models (current fields)

```
Hive          id, apiaryId, name, type, active, queenless, readyForHarvest,
              gridRow, gridCol, diseases (List<HiveDisease>), lastInspectedAt?
Apiary        id, name, lat?, lng?, gridRows, gridCols, hiveCount, userRole
HiveDisease   id, disease (string)
Inspection    id, hiveId, inspectedAt, queenSeen, broodPattern, aggressiveness,
              framesBrood?, framesHoney?, framesPollen?, framesAddedDrawn?,
              framesAddedFoundation?, framesAddedHoney?, queenCellsCount?,
              queenAdded, notes
```

Hive types (valid values): `dadant`, `langstroth`, `top_bar`, `wielkopolski`  
Display labels live in `hiveTypeLabels` map in `hive_form_widgets.dart`.

---

## Backend API Endpoints (implemented)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register user |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh tokens |
| POST | `/api/v1/auth/logout` | Logout |
| PATCH | `/api/v1/users/me/name` | Update display name |
| GET | `/api/v1/apiaries` | List apiaries |
| POST | `/api/v1/apiaries` | Create apiary |
| PATCH | `/api/v1/apiaries/{id}` | Update apiary |
| DELETE | `/api/v1/apiaries/{id}` | Delete apiary |
| GET | `/api/v1/apiaries/{id}/hives` | List hives (includes diseases + last_inspected_at) |
| POST | `/api/v1/apiaries/{id}/hives` | Create hive |
| GET | `/api/v1/apiaries/{id}/hives/{hiveId}` | Get hive |
| PATCH | `/api/v1/apiaries/{id}/hives/{hiveId}` | Update hive |
| PATCH | `/api/v1/apiaries/{id}/hives/{hiveId}/position` | Move hive |
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
      └── ApiaryGridScreen (tap apiary card)
          ├── AddHiveScreen (tap empty cell)
          └── HiveDetailScreen (tap hive cell)
              ├── EditHiveScreen (AppBar menu → Edit)
              ├── InspectionFormScreen (Add inspection button — direct, copies frames from last inspection)
              └── InspectionHistoryScreen (View all button — only shown when inspections exist)
                  └── InspectionFormScreen (empty state button / add button — copies frames from last inspection)
```
