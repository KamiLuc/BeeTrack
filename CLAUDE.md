# BeeTrack

Beekeeping management app for hobbyist and professional beekeepers.

**Stack:** Flutter (Android + Web) · Go (backend API) · PostgreSQL · Docker

## Project Tracking

- Backlog: `BACKLOG.md` — treat each item like a Jira ticket, update status as work progresses
- Version control: Git — always ask the user before committing

## Code Conventions

- **Function order:** Declaration order in header/interface files must match definition order in implementation files
- **Includes/imports:** Sorted alphabetically
- **Comments:** Only when something is genuinely hard to understand - default to no comments
- **Backend API:** All handler, service, and repository functions must have doc comments explaining what they do
- **Tests:** Write unit tests for each feature before moving to the next ticket
- **Review:** When the user says "let me review", list changed files with a short note on what changed in each

## Project Structure (planned)

```
backend/       # Go API (cmd/, internal/, pkg/)
app/           # Flutter app (Android + Web)
docker/        # Docker Compose + related config
```

## Key Commands

- Backend: `docker compose up` (Go API + PostgreSQL)
- Flutter (Android): `flutter run`
- Flutter (Web): `flutter run -d chrome`

> Flutter SDK: `C:\Users\Kamil\Prog\Flutter\flutter\bin\flutter`

## Migrations

Migrations live in `backend/migrations/` and run automatically on API startup via goose.
Naming: `NNN_description.sql` (e.g. `002_create_users.sql`).

Each file must have:
```sql
-- +goose Up
-- +goose Down
```
