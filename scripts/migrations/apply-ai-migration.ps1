# Apply AI Migration Script
# This script applies the AI columns migration to the database

Write-Host "=== AI Migration Script ===" -ForegroundColor Cyan

# Database connection parameters from .env
$dbHost = "localhost"
$dbPort = "5432"
$dbUser = "postgres"
$dbName = "nieuws_scraper"
$dbPassword = "postgres"

# Migration file
$migrationFile = "migrations/003_add_ai_columns.sql"

Write-Host ""
Write-Host "Checking migration file..." -ForegroundColor Yellow
if (-not (Test-Path $migrationFile)) {
    Write-Host "ERROR: Migration file not found: $migrationFile" -ForegroundColor Red
    exit 1
}

Write-Host "Migration file found: $migrationFile" -ForegroundColor Green

# Try to find psql in common PostgreSQL installation paths
$psqlPaths = @(
    "C:\Program Files\PostgreSQL\16\bin\psql.exe",
    "C:\Program Files\PostgreSQL\15\bin\psql.exe",
    "C:\Program Files\PostgreSQL\14\bin\psql.exe",
    "C:\Program Files\PostgreSQL\13\bin\psql.exe",
    "C:\Program Files (x86)\PostgreSQL\16\bin\psql.exe",
    "C:\Program Files (x86)\PostgreSQL\15\bin\psql.exe"
)

$psqlExe = $null
foreach ($path in $psqlPaths) {
    if (Test-Path $path) {
        $psqlExe = $path
        Write-Host "Found PostgreSQL at: $path" -ForegroundColor Green
        break
    }
}

if ($null -eq $psqlExe) {
    Write-Host ""
    Write-Host "ERROR: Could not find psql.exe" -ForegroundColor Red
    Write-Host "Please install PostgreSQL or add it to your PATH" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Alternatively, you can run the migration manually:" -ForegroundColor Yellow
    Write-Host "1. Open pgAdmin or DBeaver" -ForegroundColor White
    Write-Host "2. Connect to database: $dbName" -ForegroundColor White
    Write-Host "3. Run the SQL from: $migrationFile" -ForegroundColor White
    exit 1
}

# Set PGPASSWORD environment variable
$env:PGPASSWORD = $dbPassword

Write-Host ""
Write-Host "Connecting to database..." -ForegroundColor Yellow
Write-Host "Host: $dbHost" -ForegroundColor White
Write-Host "Port: $dbPort" -ForegroundColor White
Write-Host "Database: $dbName" -ForegroundColor White
Write-Host "User: $dbUser" -ForegroundColor White

try {
    # Execute migration
    Write-Host ""
    Write-Host "Applying migration..." -ForegroundColor Yellow
    
    $result = & $psqlExe -h $dbHost -p $dbPort -U $dbUser -d $dbName -f $migrationFile 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "SUCCESS: Migration applied successfully!" -ForegroundColor Green
        Write-Host ""
        Write-Host "The following was added:" -ForegroundColor Cyan
        Write-Host "- AI processing columns (ai_processed, ai_sentiment, etc.)" -ForegroundColor White
        Write-Host "- Indexes for efficient AI queries" -ForegroundColor White
        Write-Host "- Database functions for sentiment stats and trending topics" -ForegroundColor White
        Write-Host "- Views for AI-enriched articles" -ForegroundColor White
        Write-Host ""
        Write-Host "You can now start the server with:" -ForegroundColor Yellow
        Write-Host "go run cmd/api/main.go" -ForegroundColor White
    } else {
        Write-Host ""
        Write-Host "Migration failed with errors:" -ForegroundColor Red
        Write-Host $result -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host ""
    Write-Host "ERROR: Failed to apply migration" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
} finally {
    # Clear password from environment
    $env:PGPASSWORD = $null
}

Write-Host ""
Write-Host "=== Migration Complete ===" -ForegroundColor Cyan