# BeeTrack Deployment Guide

Production backend runs on DigitalOcean, served over HTTPS via Caddy, with the Android app connecting to it directly.

---

## Infrastructure

| Component | Details |
|-----------|---------|
| Server | DigitalOcean Droplet — Ubuntu 24.04, 1 vCPU / 1 GB RAM, $6/month, region FRA1 |
| IP | `165.227.162.182` |
| Domain | `beetrack.duckdns.org` (free DuckDNS subdomain — https://duckdns.org) |
| TLS | Let's Encrypt via Caddy (auto-issued, auto-renewed) |
| Container registry | GitHub Container Registry — `ghcr.io/kamiluc/beetrack-api` |

---

## Server Layout

```
/opt/beetrack/
  docker-compose.prod.yml   # production compose (db + api + caddy)
  Caddyfile                 # reverse proxy config
  .env.prod                 # secrets (DB_PASSWORD, JWT_SECRET) — not in git
```

## Files

**Caddyfile**
```
beetrack.duckdns.org {
    reverse_proxy api:8080
}
```

**docker-compose.prod.yml** — runs three services:
- `db` — postgres:16-alpine, data persisted in `postgres_data` volume
- `api` — Go binary, pulls from `ghcr.io/kamiluc/beetrack-api:latest`
- `caddy` — reverse proxy, ports 80/443, TLS certs in `caddy_data` volume

**.env.prod** — never commit this file:
```
DB_PASSWORD=<openssl rand -hex 32>
JWT_SECRET=<openssl rand -hex 32>
```

---

## Deploying a New Backend Version

On your local machine (from `backend/`):

```bash
docker build -t ghcr.io/kamiluc/beetrack-api:latest .
docker push ghcr.io/kamiluc/beetrack-api:latest
```

On the server (SSH into `165.227.162.182`):

```bash
cd /opt/beetrack
docker compose -f docker-compose.prod.yml --env-file .env.prod pull api
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d api
```

Migrations run automatically on startup — no manual steps needed.

---

## Checking Server Status

```bash
# All containers
docker compose -f docker-compose.prod.yml ps

# Live logs
docker compose -f docker-compose.prod.yml logs -f

# API logs only
docker compose -f docker-compose.prod.yml logs -f api
```

---

## Creating / Verifying a User Account

Register (no email needed — just make the API call):
```bash
curl -X POST https://beetrack.duckdns.org/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","name":"Name","password":"yourpassword","lang":"pl"}'
```

Then mark the account as verified directly in the DB (email sending is not configured in production):
```bash
docker compose -f docker-compose.prod.yml exec db \
  psql -U postgres -d beetrack \
  -c "UPDATE users SET verified = true WHERE email = 'you@example.com';"
```

---

## Android Release Build

The Flutter app connects to the production backend on Android and to localhost on web:

```dart
// app/lib/main.dart
baseUrl: kIsWeb ? 'http://localhost:8080' : 'https://beetrack.duckdns.org',
```

Build and install:
```powershell
cd app
flutter build apk --release
flutter install --release
```

The APK is at `build/app/outputs/flutter-apk/app-release.apk`.

### Important: INTERNET permission

The `android.permission.INTERNET` permission must be explicitly declared in `app/android/app/src/main/AndroidManifest.xml`. Flutter adds it automatically for debug builds but not always for release builds. Without it, all DNS lookups fail silently with `EAI_NODATA (errno=7)`.

```xml
<uses-permission android:name="android.permission.INTERNET"/>
```

---

## DuckDNS

The domain is managed at https://duckdns.org (login with Google/GitHub).

- Subdomain: `beetrack`
- Points to: `165.227.162.182`

If the DigitalOcean IP ever changes, update it on the DuckDNS dashboard and click **update ip**.

---

## Useful Links

- DigitalOcean dashboard: https://cloud.digitalocean.com
- DuckDNS dashboard: https://duckdns.org
- GitHub Container Registry: https://github.com/KamiLuc?tab=packages
- Server SSH: `ssh root@165.227.162.182`
