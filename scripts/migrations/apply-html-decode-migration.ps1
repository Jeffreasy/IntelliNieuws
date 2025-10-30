# Apply HTML Entity Decoding Migration
# This script decodes HTML entities in existing articles data

Write-Host "================================" -ForegroundColor Cyan
Write-Host "HTML Entity Decoding Migration" -ForegroundColor Cyan
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

# Count articles before migration
Write-Host "Counting articles with HTML entities..." -ForegroundColor Yellow
$beforeCount = docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -t -c "SELECT COUNT(*) FROM articles WHERE title LIKE '%&%' OR title LIKE '%&#%' OR summary LIKE '%&%' OR summary LIKE '%&#%' OR (content IS NOT NULL AND (content LIKE '%&%' OR content LIKE '%&#%'));"

Write-Host "Articles with HTML entities: $($beforeCount.Trim())" -ForegroundColor Cyan
Write-Host ""

# Apply migration
Write-Host "Applying migration 009_decode_html_entities.sql..." -ForegroundColor Yellow
$migrationPath = "migrations/009_decode_html_entities.sql"

if (-not (Test-Path $migrationPath)) {
    Write-Host "ERROR: Migration file not found at $migrationPath" -ForegroundColor Red
    exit 1
}

# Execute migration
Get-Content $migrationPath | docker exec -i nieuws-scraper-postgres psql -U scraper -d nieuws_scraper

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Migration applied successfully!" -ForegroundColor Green
    
    # Count after migration
    Write-Host ""
    Write-Host "Counting remaining articles with HTML entities..." -ForegroundColor Yellow
    $afterCount = docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -t -c "SELECT COUNT(*) FROM articles WHERE title LIKE '%&%' OR title LIKE '%&#%' OR summary LIKE '%&%' OR summary LIKE '%&#%' OR (content IS NOT NULL AND (content LIKE '%&%' OR content LIKE '%&#%'));"
    
    Write-Host "Articles with HTML entities after migration: $($afterCount.Trim())" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Decoded articles: $($beforeCount.Trim() - $afterCount.Trim())" -ForegroundColor Green
    
    # Show sample of decoded articles
    Write-Host ""
    Write-Host "Sample of decoded titles:" -ForegroundColor Yellow
    docker exec nieuws-scraper-postgres psql -U scraper -d nieuws_scraper -c "SELECT id, LEFT(title, 80) as title FROM articles ORDER BY created_at DESC LIMIT 5;"
    
    Write-Host ""
    Write-Host "================================" -ForegroundColor Cyan
    Write-Host "Migration completed successfully!" -ForegroundColor Green
    Write-Host "================================" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "ERROR: Migration failed!" -ForegroundColor Red
    Write-Host "Please check the error messages above" -ForegroundColor Yellow
    exit 1
}