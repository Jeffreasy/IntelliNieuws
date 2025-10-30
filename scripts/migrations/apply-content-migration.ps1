# Apply Content Extraction Migration
# This script adds the content columns to the articles table

Write-Host "=================================" -ForegroundColor Cyan
Write-Host "Content Extraction Migration" -ForegroundColor Cyan
Write-Host "=================================" -ForegroundColor Cyan
Write-Host ""

# Database connection settings (from .env)
$env:PGPASSWORD = "postgres"
$dbHost = "localhost"
$dbPort = "5432"
$dbName = "nieuws_scraper"
$dbUser = "postgres"

# Check if PostgreSQL is accessible
Write-Host "Checking PostgreSQL connection..." -ForegroundColor Yellow

# Try to find psql
$psqlPath = $null
$possiblePaths = @(
    "C:\Program Files\PostgreSQL\16\bin\psql.exe",
    "C:\Program Files\PostgreSQL\15\bin\psql.exe",
    "C:\Program Files\PostgreSQL\14\bin\psql.exe",
    "C:\Program Files (x86)\PostgreSQL\16\bin\psql.exe",
    "C:\Program Files (x86)\PostgreSQL\15\bin\psql.exe"
)

foreach ($path in $possiblePaths) {
    if (Test-Path $path) {
        $psqlPath = $path
        Write-Host "Found psql at: $psqlPath" -ForegroundColor Green
        break
    }
}

if ($null -eq $psqlPath) {
    Write-Host ""
    Write-Host "❌ psql not found in common locations" -ForegroundColor Red
    Write-Host ""
    Write-Host "MANUAL MIGRATION REQUIRED:" -ForegroundColor Yellow
    Write-Host "1. Open pgAdmin or your PostgreSQL client" -ForegroundColor White
    Write-Host "2. Connect to database: $dbName" -ForegroundColor White
    Write-Host "3. Run the SQL file: migrations/005_add_content_column.sql" -ForegroundColor White
    Write-Host ""
    Write-Host "OR add psql to your PATH and run this script again" -ForegroundColor White
    Write-Host ""
    
    # Show the SQL content for easy copy-paste
    Write-Host "SQL Content (you can copy this):" -ForegroundColor Cyan
    Write-Host "=================================" -ForegroundColor Cyan
    Get-Content "migrations\005_add_content_column.sql"
    
    exit 1
}

# Execute migration
Write-Host ""
Write-Host "Executing migration..." -ForegroundColor Yellow

try {
    & $psqlPath -h $dbHost -p $dbPort -U $dbUser -d $dbName -f "migrations\005_add_content_column.sql"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "✅ Migration completed successfully!" -ForegroundColor Green
        Write-Host ""
        
        # Show statistics
        Write-Host "Checking migration results..." -ForegroundColor Yellow
        $query = "SELECT COUNT(*) as total, COUNT(*) FILTER (WHERE content_extracted = FALSE) as needs_content FROM articles;"
        & $psqlPath -h $dbHost -p $dbPort -U $dbUser -d $dbName -c $query
        
    } else {
        Write-Host ""
        Write-Host "❌ Migration failed with exit code: $LASTEXITCODE" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host ""
    Write-Host "❌ Error executing migration: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=================================" -ForegroundColor Cyan
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "1. Restart the backend to use new columns" -ForegroundColor White
Write-Host "2. Enable content extraction in .env:" -ForegroundColor White
Write-Host "   ENABLE_FULL_CONTENT_EXTRACTION=true" -ForegroundColor White
Write-Host "=================================" -ForegroundColor Cyan