# Setup script voor Nieuws Scraper (Windows)
# Dit script helpt met het opzetten van de lokale ontwikkelomgeving

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Nieuws Scraper - Lokale Setup" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Check if .env exists
if (-Not (Test-Path ".env")) {
    Write-Host "[STAP 1] .env bestand aanmaken..." -ForegroundColor Yellow
    Copy-Item ".env.example" ".env"
    Write-Host "OK .env bestand aangemaakt" -ForegroundColor Green
} else {
    Write-Host "[STAP 1] .env bestand bestaat al" -ForegroundColor Green
}
Write-Host ""

# Check Go installation
Write-Host "[STAP 2] Go installatie controleren..." -ForegroundColor Yellow
$goVersion = go version 2>$null
if ($?) {
    Write-Host "OK Go is geinstalleerd: $goVersion" -ForegroundColor Green
} else {
    Write-Host "FOUT Go is NIET geinstalleerd!" -ForegroundColor Red
    Write-Host "   Download Go van: https://golang.org/dl/" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

# Download Go dependencies
Write-Host "[STAP 3] Go dependencies downloaden..." -ForegroundColor Yellow
go mod download
if ($?) {
    Write-Host "OK Dependencies gedownload" -ForegroundColor Green
} else {
    Write-Host "FOUT bij downloaden dependencies" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Check PostgreSQL
Write-Host "[STAP 4] PostgreSQL controleren..." -ForegroundColor Yellow
$pgCheck = Test-NetConnection -ComputerName localhost -Port 5432 -WarningAction SilentlyContinue
if ($pgCheck.TcpTestSucceeded) {
    Write-Host "OK PostgreSQL draait op poort 5432" -ForegroundColor Green
} else {
    Write-Host "FOUT PostgreSQL is NIET bereikbaar op poort 5432" -ForegroundColor Red
    Write-Host "   Download PostgreSQL van: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
    Write-Host "   Of installeer via winget: winget install PostgreSQL.PostgreSQL" -ForegroundColor Yellow
}
Write-Host ""

# Check Redis
Write-Host "[STAP 5] Redis controleren..." -ForegroundColor Yellow
$redisCheck = Test-NetConnection -ComputerName localhost -Port 6379 -WarningAction SilentlyContinue
if ($redisCheck.TcpTestSucceeded) {
    Write-Host "OK Redis draait op poort 6379" -ForegroundColor Green
} else {
    Write-Host "WAARSCHUWING Redis is NIET bereikbaar op poort 6379 (optioneel)" -ForegroundColor Yellow
    Write-Host "   Download Redis van: https://github.com/microsoftarchive/redis/releases" -ForegroundColor Yellow
    Write-Host "   Of gebruik Memurai: https://www.memurai.com/" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Setup voltooid!" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "VOLGENDE STAPPEN:" -ForegroundColor Yellow
Write-Host "  1. Zorg dat PostgreSQL en Redis draaien" -ForegroundColor White
Write-Host "  2. Run database setup script" -ForegroundColor White
Write-Host "  3. Start de API" -ForegroundColor White
Write-Host ""