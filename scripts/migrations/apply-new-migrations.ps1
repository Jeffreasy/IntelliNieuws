# ============================================================================
# PowerShell Script: Apply New Professional Migrations
# Description: Apply V001-V003 migrations to the database
# Version: 1.0.0
# Author: NieuwsScraper Team
# Date: 2025-10-30
# ============================================================================

param(
    [string]$Database = "nieuws_scraper",
    [string]$User = "postgres",
    [string]$Host = "localhost",
    [int]$Port = 5432,
    [switch]$DryRun = $false,
    [switch]$SkipLegacyMigration = $false
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "APPLY NEW MIGRATIONS (V001-V003)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$connectionString = "postgresql://${User}@${Host}:${Port}/${Database}"
$migrationsDir = "migrations"

# Function to execute SQL file
function Invoke-SqlFile {
    param(
        [string]$FilePath,
        [string]$Description
    )
    
    Write-Host "Applying: $Description" -ForegroundColor Yellow
    
    if ($DryRun) {
        Write-Host "  [DRY RUN] Would execute: $FilePath" -ForegroundColor Gray
        return $true
    }
    
    try {
        $result = docker exec -i nieuws-scraper-db psql -U $User -d $Database -f "/migrations/$(Split-Path $FilePath -Leaf)" 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  ✓ Success" -ForegroundColor Green
            return $true
        } else {
            Write-Host "  ✗ Failed" -ForegroundColor Red
            Write-Host "  Error: $result" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "  ✗ Exception: $_" -ForegroundColor Red
        return $false
    }
}

# Check if Docker container is running
Write-Host "Checking database connection..." -ForegroundColor Yellow
try {
    $dbCheck = docker exec nieuws-scraper-db psql -U $User -d $Database -c "SELECT version();" 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ Cannot connect to database" -ForegroundColor Red
        Write-Host "  Make sure Docker container 'nieuws-scraper-db' is running" -ForegroundColor Yellow
        exit 1
    }
    Write-Host "✓ Database connection OK" -ForegroundColor Green
} catch {
    Write-Host "✗ Error checking database: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Check if legacy migrations need to be migrated first
Write-Host "Checking migration status..." -ForegroundColor Yellow
$hasLegacyData = docker exec nieuws-scraper-db psql -U $User -d $Database -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'articles';" 2>&1
$hasSchemaVersioning = docker exec nieuws-scraper-db psql -U $User -d $Database -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'schema_migrations';" 2>&1

$hasLegacyData = $hasLegacyData.Trim()
$hasSchemaVersioning = $hasSchemaVersioning.Trim()

if ($hasLegacyData -eq "1" -and $hasSchemaVersioning -eq "0" -and -not $SkipLegacyMigration) {
    Write-Host "✓ Detected legacy schema" -ForegroundColor Yellow
    Write-Host "  Running legacy migration script first..." -ForegroundColor Yellow
    Write-Host ""
    
    $legacyResult = Invoke-SqlFile `
        -FilePath "$migrationsDir/utilities/01_migrate_from_legacy.sql" `
        -Description "Legacy Schema Migration"
    
    if (-not $legacyResult) {
        Write-Host ""
        Write-Host "✗ Legacy migration failed" -ForegroundColor Red
        exit 1
    }
    Write-Host ""
}

# Apply V001 (Base Schema)
Write-Host "Step 1: Applying V001 - Base Schema" -ForegroundColor Cyan
$v001Result = Invoke-SqlFile `
    -FilePath "$migrationsDir/V001__create_base_schema.sql" `
    -Description "V001: Base Schema (articles, sources, scraping_jobs)"

if (-not $v001Result) {
    Write-Host ""
    Write-Host "✗ V001 migration failed - stopping" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Apply V002 (Emails Table)
Write-Host "Step 2: Applying V002 - Emails Table" -ForegroundColor Cyan
$v002Result = Invoke-SqlFile `
    -FilePath "$migrationsDir/V002__create_emails_table.sql" `
    -Description "V002: Emails Table"

if (-not $v002Result) {
    Write-Host ""
    Write-Host "✗ V002 migration failed - stopping" -ForegroundColor Red
    Write-Host "  You can rollback V001 with: psql < migrations/rollback/V001__rollback.sql" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

# Apply V003 (Analytics Views)
Write-Host "Step 3: Applying V003 - Analytics Views" -ForegroundColor Cyan
$v003Result = Invoke-SqlFile `
    -FilePath "$migrationsDir/V003__create_analytics_views.sql" `
    -Description "V003: Analytics Materialized Views"

if (-not $v003Result) {
    Write-Host ""
    Write-Host "✗ V003 migration failed" -ForegroundColor Red
    Write-Host "  You can continue without analytics views" -ForegroundColor Yellow
    Write-Host "  To rollback: psql < migrations/rollback/V003__rollback.sql" -ForegroundColor Yellow
}
Write-Host ""

# Verify migrations
Write-Host "Verifying migrations..." -ForegroundColor Yellow
$migrationStatus = docker exec nieuws-scraper-db psql -U $User -d $Database -c "SELECT version, description, applied_at FROM schema_migrations ORDER BY version;" 2>&1

Write-Host ""
Write-Host "Migration Status:" -ForegroundColor Cyan
Write-Host $migrationStatus
Write-Host ""

# Run health check
Write-Host "Running health check..." -ForegroundColor Yellow
Write-Host ""

docker exec -i nieuws-scraper-db psql -U $User -d $Database -f /migrations/utilities/02_health_check.sql 2>&1

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✅ MIGRATION COMPLETE" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "  1. Review the health check results above" -ForegroundColor White
Write-Host "  2. Restart your application to use new schema" -ForegroundColor White
Write-Host "  3. Set up periodic maintenance:" -ForegroundColor White
Write-Host "     docker exec nieuws-scraper-db psql -U postgres -d nieuws_scraper -f /migrations/utilities/03_maintenance.sql" -ForegroundColor Gray
Write-Host "  4. Schedule materialized view refreshes (every 5-15 minutes):" -ForegroundColor White
Write-Host "     SELECT refresh_analytics_views(TRUE);" -ForegroundColor Gray
Write-Host ""