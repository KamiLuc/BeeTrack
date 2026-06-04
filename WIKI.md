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
      hive/        # hive CRUD + detail — Cubit pattern
      home/        # HomeScreen (shell after login)
    l10n/          # ARB files (app_en.arb, app_pl.arb) + generated classes
    main.dart
  test/
    features/
      auth/        # auth_bloc_test.dart
      apiary/      # apiaries_cubit_test.dart
      hive/        # hives_cubit_test.dart, hive_detail_screen_test.dart

docker/            # Docker Compose config
```

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
| `app/lib/core/theme/app_colors.dart` | Color constants (primary = amber `#FBBF24`, background = cream `#FFFBF2`) |
| `app/lib/core/theme/app_text_styles.dart` | Text style constants (headlineLarge 28px … caption 12px) |
| `app/lib/core/theme/app_layout.dart` | `AppLayout.formConstraints(context)` — 85% width on phone, 40% on tablet |
| `app/lib/features/hive/view/hive_form_widgets.dart` | `HiveNameField`, `HiveTypeDropdown`, `HiveActiveToggle`, `hiveTypeLabels` map |
| `app/lib/features/apiary/view/apiary_form_widgets.dart` | `ApiaryGridSection`, `ApiaryLocationSection` |
| `app/lib/l10n/app_en.arb` | Source of truth for all UI strings |

---

## Data Models (current fields)

```
Hive          id, apiaryId, name, type, active, gridRow, gridCol
Apiary        id, name, lat?, lng?, gridRows, gridCols, hiveCount, userRole
```

Hive types (valid values): `dadant`, `langstroth`, `top_bar`, `wielkopolski`  
Display labels live in `hiveTypeLabels` map in `hive_form_widgets.dart`.

---

## Backend API Endpoints (implemented)

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/v1/auth/register` | RegisterHandler |
| POST | `/api/v1/auth/login` | LoginHandler |
| POST | `/api/v1/auth/refresh` | RefreshHandler |
| GET | `/api/v1/apiaries` | ListApiariesHandler |
| POST | `/api/v1/apiaries` | CreateApiaryHandler |
| PATCH | `/api/v1/apiaries/:id` | UpdateApiaryHandler |
| DELETE | `/api/v1/apiaries/:id` | DeleteApiaryHandler |
| GET | `/api/v1/apiaries/:id/hives` | ListHivesHandler |
| POST | `/api/v1/apiaries/:id/hives` | CreateHiveHandler |
| PATCH | `/api/v1/apiaries/:id/hives/:hiveId` | UpdateHiveHandler |
| DELETE | `/api/v1/apiaries/:id/hives/:hiveId` | DeleteHiveHandler |

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
- Flutter regenerates `app_localizations_en.dart` / `app_localizations_pl.dart` automatically on build.
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
              └── EditHiveScreen (AppBar menu → Edit)
```
