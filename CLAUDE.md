# BeeTrack

Beekeeping management app for hobbyist and professional beekeepers.

**Stack:** Flutter (Android + Web) · Go (backend API) · PostgreSQL · Docker

## Project Tracking

- Requirements and backlog: `REQUIREMENTS.md` — treat each item like a Jira ticket, update status as work progresses
- Version control: Git — commit changes as work is completed

## Code Conventions

- **Function order:** Declaration order in header/interface files must match definition order in implementation files
- **Includes/imports:** Sorted alphabetically
- **Comments:** Only when something is genuinely hard to understand — default to no comments

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

## Migrations

Migrations live in `backend/migrations/` and run automatically on API startup via goose.
Naming: `NNN_description.sql` (e.g. `002_create_users.sql`).

Each file must have:
```sql
-- +goose Up
-- +goose Down
```
