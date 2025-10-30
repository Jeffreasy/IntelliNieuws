# Start script voor Nieuws Scraper (Windows)
# Start de API zonder Docker

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Nieuws Scraper - API Starten" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Check if .env exists
if (-Not (Test-Path ".env")) {
    Write-Host "WAARSCHUWING .env bestand niet gevonden!" -ForegroundColor Yellow
    Write-Host "  Kopieren van .env.example..." -ForegroundColor White
    Copy-Item ".env.example" ".env"
    Write-Host "OK .env bestand aangemaakt" -ForegroundColor Green
    Write-Host ""
    Write-Host "WAARSCHUWING Pas de .env instellingen aan indien nodig!" -ForegroundColor Yellow
    Write-Host ""
}

# Check PostgreSQL connection
Write-Host "Controleren PostgreSQL verbinding..." -ForegroundColor Yellow
$pgCheck = Test-NetConnection -ComputerName localhost -Port 5432 -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
if (-Not $pgCheck.TcpTestSucceeded) {
    Write-Host "FOUT PostgreSQL is NIET bereikbaar!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Zorg dat PostgreSQL draait op poort 5432" -ForegroundColor Yellow
    Write-Host "Installeer PostgreSQL van: https://www.postgresql.org/download/windows/" -ForegroundColor Cyan
    Write-Host "Of via winget: winget install PostgreSQL.PostgreSQL.17" -ForegroundColor Cyan
    Write-Host ""
    Read-Host "Druk op Enter om toch te proberen te starten, of Ctrl+C om te stoppen"
} else {
    Write-Host "OK PostgreSQL is bereikbaar" -ForegroundColor Green
}
Write-Host ""

# Check Redis connection (optional)
Write-Host "Controleren Redis verbinding (optioneel)..." -ForegroundColor Yellow
$redisCheck = Test-NetConnection -ComputerName localhost -Port 6379 -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
if (-Not $redisCheck.TcpTestSucceeded) {
    Write-Host "WAARSCHUWING Redis is niet bereikbaar (optioneel - rate limiting werkt niet)" -ForegroundColor Yellow
} else {
    Write-Host "OK Redis is bereikbaar" -ForegroundColor Green
}
Write-Host ""

# Start the API
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "API starten..." -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "API draait op: http://localhost:8080" -ForegroundColor Cyan
Write-Host "Health check: http://localhost:8080/health" -ForegroundColor Cyan
Write-Host ""
Write-Host "Druk op Ctrl+C om te stoppen" -ForegroundColor Yellow
Write-Host ""

# Run the Go application
go run ./cmd/api/main.go