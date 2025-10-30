# Docker Cleanup and Restart Script
# Dit script stopt alle containers, ruimt oude containers op, en start alles opnieuw

Write-Host "=== Docker Cleanup and Restart ===" -ForegroundColor Cyan
Write-Host ""

# Stap 1: Stop alle services
Write-Host "STAP 1: Stopping all services..." -ForegroundColor Yellow
docker-compose down

# Stap 2: Verwijder orphan containers
Write-Host ""
Write-Host "STAP 2: Removing orphan containers..." -ForegroundColor Yellow
docker-compose down --remove-orphans

# Stap 3: Check wat er nog draait
Write-Host ""
Write-Host "STAP 3: Checking for remaining nieuws-scraper containers..." -ForegroundColor Yellow
$remainingContainers = docker ps -a --filter "name=nieuws" --format "{{.Names}}"
if ($remainingContainers) {
    Write-Host "Found remaining containers:" -ForegroundColor Red
    $remainingContainers | ForEach-Object { Write-Host "  - $_" -ForegroundColor Red }
    
    Write-Host ""
    $response = Read-Host "Do you want to force remove these? (y/n)"
    if ($response -eq 'y') {
        $remainingContainers | ForEach-Object {
            Write-Host "Removing $_..." -ForegroundColor Yellow
            docker rm -f $_
        }
    }
} else {
    Write-Host "✅ No remaining containers found" -ForegroundColor Green
}

# Stap 4: Rebuild en start
Write-Host ""
Write-Host "STAP 4: Rebuilding and starting services..." -ForegroundColor Yellow
docker-compose up -d --build

# Stap 5: Wacht tot services healthy zijn
Write-Host ""
Write-Host "STAP 5: Waiting for services to become healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Stap 6: Toon status
Write-Host ""
Write-Host "STAP 6: Checking container status..." -ForegroundColor Yellow
Write-Host ""
docker-compose ps

Write-Host ""
Write-Host "=== Expected Containers ===" -ForegroundColor Cyan
Write-Host "✅ nieuws-scraper-postgres (Status: Up, healthy)" -ForegroundColor Green
Write-Host "✅ nieuws-scraper-redis    (Status: Up, healthy)" -ForegroundColor Green
Write-Host "✅ nieuws-scraper-app      (Status: Up, healthy)" -ForegroundColor Green
Write-Host "✅ nieuws-scraper-backup   (Status: Up)" -ForegroundColor Green

Write-Host ""
Write-Host "=== Health Check ===" -ForegroundColor Cyan

# Check health
$appHealth = docker inspect --format='{{.State.Health.Status}}' nieuws-scraper-app 2>$null
$postgresHealth = docker inspect --format='{{.State.Health.Status}}' nieuws-scraper-postgres 2>$null
$redisHealth = docker inspect --format='{{.State.Health.Status}}' nieuws-scraper-redis 2>$null

if ($appHealth -eq "healthy") {
    Write-Host "✅ App is healthy" -ForegroundColor Green
} else {
    Write-Host "⚠️  App health: $appHealth" -ForegroundColor Yellow
}

if ($postgresHealth -eq "healthy") {
    Write-Host "✅ PostgreSQL is healthy" -ForegroundColor Green
} else {
    Write-Host "⚠️  PostgreSQL health: $postgresHealth" -ForegroundColor Yellow
}

if ($redisHealth -eq "healthy") {
    Write-Host "✅ Redis is healthy" -ForegroundColor Green
} else {
    Write-Host "⚠️  Redis health: $redisHealth" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=== Test API ===" -ForegroundColor Cyan
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -TimeoutSec 5
    if ($response.status -eq "ok") {
        Write-Host "✅ API is responding correctly" -ForegroundColor Green
    } else {
        Write-Host "⚠️  API status: $($response.status)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ API is not responding" -ForegroundColor Red
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Logs (last 20 lines) ===" -ForegroundColor Cyan
docker-compose logs --tail=20 app

Write-Host ""
Write-Host "✅ Cleanup and restart complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Check that all 4 containers are running and healthy"
Write-Host "2. Test the API endpoints"
Write-Host "3. Check your frontend components"
Write-Host ""