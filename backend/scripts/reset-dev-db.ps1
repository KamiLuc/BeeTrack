<#
.SYNOPSIS
  Wipes the local dev database (except honey batches already confirmed
  on-chain) and reseeds it with a fixed set of test users, promoting the
  first one to admin.

.DESCRIPTION
  Deletes all app data except honey batches with a confirmed on-chain
  certification (and the handful of rows those need to stay valid — see
  reset-dev-data.sql) via a targeted SQL wipe against the running database,
  rather than dropping the Postgres volume outright. Confirmed batches are
  permanent on-chain records; keeping them locally means a fresh batch can
  never collide with an old certified id (which the contract would reject
  as "already certified" — see writer.go's isAlreadyCertifiedRevert). Then
  seeds sample data for each user via cmd/seed and promotes the first one to
  the admin role.

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

Write-Host "==> Wiping dev data (keeping honey batches confirmed on-chain)" -ForegroundColor Cyan
Get-Content (Join-Path $PSScriptRoot "reset-dev-data.sql") -Raw |
    docker exec -i backend-db-1 psql -U postgres -d beetrack -v ON_ERROR_STOP=1 -f -
if ($LASTEXITCODE -ne 0) { throw "Data wipe failed; see psql output above" }

$preservedCount = (docker exec backend-db-1 psql -U postgres -d beetrack -tAc "SELECT COUNT(*) FROM honey_batches;").Trim()
Write-Host "==> Preserved $preservedCount honey batch(es) already confirmed on-chain" -ForegroundColor Green

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
