# BeeTrack Admin Panel

React + Vite + TypeScript SPA for reviewing pending marketplace listings and honey batch certification requests. Talks directly to the Go API's `/api/v1/admin/*` routes — no separate backend.

An account needs `role = 'admin'` to sign in here; that's set manually in the database (`UPDATE users SET role = 'admin' WHERE email = '...'`), there's no self-service promotion path.

## Setup

```
cp .env.example .env
npm install
npm run dev
```

Runs on http://localhost:5174 by default. `VITE_API_BASE_URL` in `.env` must point at the Go API (defaults to `http://localhost:8080/api/v1`).

## Resetting local test data

`backend/scripts/reset-dev-db.ps1` wipes the local Postgres/images volumes, brings db/api/mailpit back up, and seeds a fixed set of test users (`kamil@op.pl` / `kamil2@op.pl` / `kamil3@op.pl` / `kamil4@op.pl`, all password `lion12345`), promoting `kamil@op.pl` to admin. Run it from anywhere:

```
./backend/scripts/reset-dev-db.ps1
```

Pass `-Emails` / `-Password` to override the defaults.
