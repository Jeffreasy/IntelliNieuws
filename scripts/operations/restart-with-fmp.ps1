# Restart Backend with FMP Configuration
# This script stops the current backend and restarts with FMP API

Write-Host "Restarting backend with FMP API configuration..." -ForegroundColor Cyan
Write-Host ""

# Kill existing api.exe processes
Write-Host "Stopping existing backend processes..." -ForegroundColor Yellow
Get-Process -Name "api" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 2

# Rebuild
Write-Host "Rebuilding with new configuration..." -ForegroundColor Yellow
go build -o api.exe ./cmd/api

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "Build successful!" -ForegroundColor Green
Write-Host ""

# Start backend
Write-Host "Starting backend with FMP API..." -ForegroundColor Yellow
Write-Host "Configuration:" -ForegroundColor Cyan
Write-Host "  Provider: FMP (Financial Modeling Prep)" -ForegroundColor Cyan
Write-Host "  Rate Limit: 30 calls/min" -ForegroundColor Cyan
Write-Host "  Cache TTL: 5 minutes" -ForegroundColor Cyan
Write-Host "  Batch API: ENABLED" -ForegroundColor Green
Write-Host ""

Start-Process -FilePath ".\api.exe" -NoNewWindow

Start-Sleep -Seconds 3

# Test if backend is running
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -ErrorAction Stop
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Green
    Write-Host "Backend successfully started!" -ForegroundColor Green
    Write-Host "================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "API running on: http://localhost:8080" -ForegroundColor Cyan
    Write-Host "Stock API Provider: FMP" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "1. Run: .\scripts\test-fmp-integration.ps1" -ForegroundColor White
    Write-Host "2. Check logs for batch API messages" -ForegroundColor White
    Write-Host "3. Open http://localhost:8080/health/metrics" -ForegroundColor White
} catch {
    Write-Host "Failed to start backend!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
}