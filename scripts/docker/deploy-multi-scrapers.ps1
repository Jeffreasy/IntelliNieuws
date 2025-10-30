# Deploy Multi-Scraper Profiles
# This script deploys multiple scraper instances with different profiles

Write-Host "=== Multi-Scraper Profiles Deployment ===" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
$dockerRunning = docker ps 2>$null
if (!$dockerRunning) {
    Write-Host "ERROR: Docker is not running!" -ForegroundColor Red
    Write-Host "Please start Docker Desktop and try again." -ForegroundColor Yellow
    exit 1
}

Write-Host "✓ Docker is running" -ForegroundColor Green
Write-Host ""

# Stop existing containers
Write-Host "Stopping existing containers..." -ForegroundColor Yellow
docker-compose down 2>$null
Write-Host "✓ Existing containers stopped" -ForegroundColor Green
Write-Host ""

# Build image
Write-Host "Building scraper image..." -ForegroundColor Cyan
docker-compose build
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Image built successfully" -ForegroundColor Green
Write-Host ""

# Start infrastructure (postgres, redis)
Write-Host "Starting infrastructure (postgres, redis)..." -ForegroundColor Cyan
docker-compose up -d postgres redis
Start-Sleep -Seconds 5
Write-Host "✓ Infrastructure started" -ForegroundColor Green
Write-Host ""

# Deploy all scraper profiles
Write-Host "Deploying scraper profiles..." -ForegroundColor Cyan

Write-Host "  1. Fast Profile (5 min interval, nu.nl only)" -ForegroundColor White
docker-compose -f docker-compose.profiles.yml up -d scraper-fast
Start-Sleep -Seconds 2

Write-Host "  2. Balanced Profile (15 min interval, all sources, MAIN)" -ForegroundColor White
docker-compose -f docker-compose.profiles.yml up -d scraper-balanced
Start-Sleep -Seconds 2

Write-Host "  3. Deep Profile (60 min interval, max quality)" -ForegroundColor White
docker-compose -f docker-compose.profiles.yml up -d scraper-deep
Start-Sleep -Seconds 2

Write-Host "  4. Conservative Profile (30 min interval, minimal load)" -ForegroundColor White
docker-compose -f docker-compose.profiles.yml up -d scraper-conservative

Write-Host ""
Write-Host "✓ All profiles deployed!" -ForegroundColor Green
Write-Host ""

# Wait for containers to be healthy
Write-Host "Waiting for containers to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Check status
Write-Host ""
Write-Host "=== Container Status ===" -ForegroundColor Cyan
docker-compose -f docker-compose.profiles.yml ps

Write-Host ""
Write-Host "=== Profile Summary ===" -ForegroundColor Cyan
Write-Host "Fast Profile:         http://localhost:8081/health (every 5 min)" -ForegroundColor Green
Write-Host "Balanced Profile:     http://localhost:8080/health (every 15 min)" -ForegroundColor Green
Write-Host "Deep Profile:         http://localhost:8082/health (every 60 min)" -ForegroundColor Green
Write-Host "Conservative Profile: http://localhost:8083/health (every 30 min)" -ForegroundColor Green
Write-Host ""

Write-Host "=== Expected Coverage ===" -ForegroundColor Cyan
Write-Host "Fast:         12 scrapes/hour × 30 articles = ~360 articles/hour" -ForegroundColor White
Write-Host "Balanced:     4 scrapes/hour × 80 articles = ~320 articles/hour" -ForegroundColor White
Write-Host "Deep:         1 scrape/hour × 100 articles = ~100 articles/hour" -ForegroundColor White
Write-Host "Conservative: 2 scrapes/hour × 40 articles = ~80 articles/hour" -ForegroundColor White
Write-Host "TOTAL:        ~860 articles/hour, ~20,000/day" -ForegroundColor Yellow
Write-Host ""

Write-Host "=== Quick Commands ===" -ForegroundColor Cyan
Write-Host "View logs (all):      docker-compose -f docker-compose.profiles.yml logs -f" -ForegroundColor White
Write-Host "View logs (fast):     docker-compose -f docker-compose.profiles.yml logs -f scraper-fast" -ForegroundColor White
Write-Host "Stop all profiles:    docker-compose -f docker-compose.profiles.yml down" -ForegroundColor White
Write-Host "Restart profile:      docker-compose -f docker-compose.profiles.yml restart scraper-fast" -ForegroundColor White
Write-Host ""

Write-Host "Deployment complete!" -ForegroundColor Green