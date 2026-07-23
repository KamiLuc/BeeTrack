<#
.SYNOPSIS
  Wipes the local dev database and reseeds it with a fixed set of test users,
  promoting the first one to admin.

.DESCRIPTION
  Drops the Postgres/images Docker volumes, brings db/api/mailpit back up
  (migrations run automatically on API startup), seeds sample data for each
  user via cmd/seed, then promotes the first user to the admin role.

.PARAMETER Count
  How many extra numbered kamil emails to create, on top of kamil@op.pl
  (which is always created and promoted to admin). kamil1@op.pl..kamilN@op.pl
  are added for Count = N. Defaults to 0 (just kamil@op.pl). Ignored if
  -Emails is passed explicitly.

.PARAMETER Emails
  Explicit list of users to create, in order, overriding -Count. The first
  one becomes admin.

.PARAMETER Password
  Password set on every seeded user.

.EXAMPLE
  ./reset-dev-db.ps1
  ./reset-dev-db.ps1 -Count 3
  ./reset-dev-db.ps1 -Emails alice@example.com,bob@example.com -Password test1234
#>
param(
    [int]$Count = 0,
    [string[]]$Emails,
    [string]$Password = "lion12345"
)

if (-not $Emails) {
    $Emails = @("kamil@op.pl")
    if ($Count -gt 0) {
        $Emails += 1..$Count | ForEach-Object { "kamil$_@op.pl" }
    }
}

$ErrorActionPreference = "Stop"
$backendDir = Split-Path -Parent $PSScriptRoot
Push-Location $backendDir
try {

Write-Host "==> Wiping local db/images volumes" -ForegroundColor Cyan
docker compose down -v

Write-Host "==> Starting db, api, mailpit (migrations run automatically)" -ForegroundColor Cyan
docker compose up -d db api mailpit

Write-Host "==> Waiting for the API to come up" -ForegroundColor Cyan
$ready = $false
for ($i = 0; $i -lt 90; $i++) {
    try {
        Invoke-WebRequest -Uri "http://127.0.0.1:8080/" -UseBasicParsing -TimeoutSec 2 | Out-Null
        $ready = $true
        break
    } catch [System.Net.WebException] {
        if ($_.Exception.Response) { $ready = $true; break }
        Start-Sleep -Seconds 1
    } catch {
        Start-Sleep -Seconds 1
    }
}
if (-not $ready) {
    throw "API did not come up in time; check docker compose logs api"
}

$env:DATABASE_URL = "postgres://postgres:password@localhost:5432/beetrack?sslmode=disable"
$env:CGO_ENABLED = "0"

foreach ($email in $Emails) {
    Write-Host "==> Seeding $email" -ForegroundColor Cyan
    go run ./cmd/seed -email $email -password $Password -api-url "http://127.0.0.1:8080"
    if ($LASTEXITCODE -ne 0) { throw "Seeding failed for $email" }
}

$adminEmail = $Emails[0]
Write-Host "==> Promoting $adminEmail to admin" -ForegroundColor Cyan
docker exec backend-db-1 psql -U postgres -d beetrack -c "UPDATE users SET role='admin' WHERE email='$adminEmail';"

Write-Host "==> Done. Users:" -ForegroundColor Green
docker exec backend-db-1 psql -U postgres -d beetrack -c "SELECT id, email, role, verified FROM users ORDER BY id;"

} finally {
    Pop-Location
}
