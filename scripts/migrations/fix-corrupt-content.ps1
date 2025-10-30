# Fix Corrupt Content Script
# This script resets corrupt article content and triggers re-scraping

Write-Host "================================" -ForegroundColor Cyan
Write-Host "Corrupt Content Fix & Re-scrape" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker container is running
Write-Host "Checking if PostgreSQL container is running..." -ForegroundColor Yellow
$containerStatus = docker ps --filter "name=nieuws-scraper-postgres" --format "{{.Status}}"

if (-not $containerStatus) {
    Write-Host "ERROR: PostgreSQL container is not running!" -ForegroundColor Red
    Write-Host "Please start the containers first with: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

Write-Host "PostgreSQL container is running" -ForegroundColor Green
Write-Host ""

# Count articles with potentially corrupt content
Write-Host "Identifying articles with corrupt content..." -ForegroundColor Yellow
$corruptCountQuery = @"
SELECT COUNT(*) FROM articles 
WHERE content_extracted = true
  AND content IS NOT NULL
  AND (
    content ~ '[^\x20-\x7E\x0A\x0D\t\u00A0-\uFFFF]'
    OR LENGTH(content) < 100
  );
"@

$corruptCount = docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -t -c $corruptCountQuery

Write-Host "Articles with potentially corrupt content: $($corruptCount.Trim())" -ForegroundColor Cyan
Write-Host ""

if ($corruptCount.Trim() -eq "0") {
    Write-Host "No corrupt content found! All articles are clean." -ForegroundColor Green
    exit 0
}

# Apply migration to reset corrupt content
Write-Host "Applying migration 010_reset_corrupt_content.sql..." -ForegroundColor Yellow
$migrationPath = "migrations/010_reset_corrupt_content.sql"

if (-not (Test-Path $migrationPath)) {
    Write-Host "ERROR: Migration file not found at $migrationPath" -ForegroundColor Red
    exit 1
}

# Execute migration
Get-Content $migrationPath | docker exec -i nieuws-scraper-postgres psql -U scraper -d nieuws_scraper

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Migration applied successfully!" -ForegroundColor Green
    
    # Count articles marked for re-scraping
    Write-Host ""
    Write-Host "Verifying reset..." -ForegroundColor Yellow
    $resetQuery = @"
SELECT COUNT(*) FROM articles 
WHERE content_extracted = false 
  AND content IS NULL
  AND published > NOW() - INTERVAL '7 days';
"@
    
    $resetCount = docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -t -c $resetQuery
    
    Write-Host "Articles marked for re-scraping (last 7 days): $($resetCount.Trim())" -ForegroundColor Cyan
    Write-Host ""
    
    # Show sample of articles to be re-scraped
    Write-Host "Sample articles to be re-scraped:" -ForegroundColor Yellow
    docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -c "SELECT id, LEFT(title, 60) as title, source FROM articles WHERE content_extracted = false AND content IS NULL AND published > NOW() - INTERVAL '7 days' ORDER BY published DESC LIMIT 5;"
    
    Write-Host ""
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host "Next Steps:" -ForegroundColor Yellow
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "1. The corrupt content has been reset" -ForegroundColor Green
    Write-Host "2. These articles will be automatically re-scraped by the content processor" -ForegroundColor Green
    Write-Host "3. Or manually trigger re-scraping with:" -ForegroundColor Yellow
    Write-Host "   curl -X POST http://localhost:8080/api/v1/articles/{id}/extract-content" -ForegroundColor White
    Write-Host ""
    Write-Host "The new HTML decoder will be used for re-scraping!" -ForegroundColor Green
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "ERROR: Migration failed!" -ForegroundColor Red
    Write-Host "Please check the error messages above" -ForegroundColor Yellow
    exit 1
}