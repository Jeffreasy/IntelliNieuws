# Apply Email Integration Migration
# This script applies the email table migration to the database

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Email Integration Migration" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Load environment variables from .env file
if (Test-Path .env) {
    Write-Host "Loading environment variables from .env..." -ForegroundColor Green
    Get-Content .env | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]*)\s*=\s*(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
} else {
    Write-Host "Warning: .env file not found" -ForegroundColor Yellow
}

# Get database connection details from environment variables
$dbHost = $env:POSTGRES_HOST
$dbPort = $env:POSTGRES_PORT
$dbUser = $env:POSTGRES_USER
$dbPassword = $env:POSTGRES_PASSWORD
$dbName = $env:POSTGRES_DB

if (-not $dbHost -or -not $dbPort -or -not $dbUser -or -not $dbPassword -or -not $dbName) {
    Write-Host "Error: Database connection details not found in environment variables" -ForegroundColor Red
    Write-Host "Please ensure .env file exists with database configuration" -ForegroundColor Red
    exit 1
}

Write-Host "Database: $dbName@$dbHost:$dbPort" -ForegroundColor Cyan
Write-Host ""

# Set PGPASSWORD environment variable for psql
$env:PGPASSWORD = $dbPassword

# Migration file
$migrationFile = "migrations/007_create_emails_table.sql"

if (-not (Test-Path $migrationFile)) {
    Write-Host "Error: Migration file not found: $migrationFile" -ForegroundColor Red
    exit 1
}

Write-Host "Applying migration: $migrationFile" -ForegroundColor Green
Write-Host ""

# Apply migration using psql
$psqlCommand = "psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -f $migrationFile"

try {
    Invoke-Expression $psqlCommand
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "==================================" -ForegroundColor Green
        Write-Host "Migration applied successfully!" -ForegroundColor Green
        Write-Host "==================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "The emails table has been created with the following features:" -ForegroundColor Cyan
        Write-Host "  - Email storage with deduplication (message_id)" -ForegroundColor White
        Write-Host "  - Processing status tracking" -ForegroundColor White
        Write-Host "  - Article linkage for processed emails" -ForegroundColor White
        Write-Host "  - Error tracking and retry support" -ForegroundColor White
        Write-Host "  - Indexed for efficient querying" -ForegroundColor White
        Write-Host ""
        Write-Host "Next steps:" -ForegroundColor Yellow
        Write-Host "  1. Configure email settings in .env file:" -ForegroundColor White
        Write-Host "     EMAIL_ENABLED=true" -ForegroundColor Gray
        Write-Host "     EMAIL_HOST=outlook.office365.com" -ForegroundColor Gray
        Write-Host "     EMAIL_USERNAME=your-email@outlook.com" -ForegroundColor Gray
        Write-Host "     EMAIL_PASSWORD=your-password" -ForegroundColor Gray
        Write-Host "     EMAIL_ALLOWED_SENDERS=noreply@x.ai" -ForegroundColor Gray
        Write-Host ""
        Write-Host "  2. Restart the application to enable email integration" -ForegroundColor White
        Write-Host "     .\scripts\restart-with-fmp.ps1" -ForegroundColor Gray
        Write-Host ""
    } else {
        Write-Host ""
        Write-Host "Migration failed with exit code: $LASTEXITCODE" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error executing migration: $_" -ForegroundColor Red
    exit 1
} finally {
    # Clear password from environment
    Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
}