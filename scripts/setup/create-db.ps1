# Database setup script voor Nieuws Scraper (Windows)
# Maakt de database en tabellen aan

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Database Setup" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Find PostgreSQL installation
$pgPaths = @(
    "C:\Program Files\PostgreSQL\17\bin",
    "C:\Program Files\PostgreSQL\16\bin",
    "C:\Program Files\PostgreSQL\15\bin",
    "C:\Program Files (x86)\PostgreSQL\17\bin",
    "C:\Program Files (x86)\PostgreSQL\16\bin"
)

$psqlPath = $null
foreach ($path in $pgPaths) {
    if (Test-Path "$path\psql.exe") {
        $psqlPath = "$path\psql.exe"
        Write-Host "PostgreSQL gevonden in: $path" -ForegroundColor Green
        break
    }
}

if (-Not $psqlPath) {
    Write-Host "FOUT: psql.exe niet gevonden!" -ForegroundColor Red
    Write-Host "PostgreSQL lijkt niet correct geinstalleerd te zijn." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Probeer handmatig:" -ForegroundColor Yellow
    Write-Host '  1. Open pgAdmin of SQL Shell (psql)' -ForegroundColor White
    Write-Host '  2. Voer uit: CREATE DATABASE nieuws_scraper;' -ForegroundColor White
    Write-Host '  3. Voer uit: \c nieuws_scraper' -ForegroundColor White
    Write-Host '  4. Kopieer en plak de inhoud van migrations/001_create_tables.sql' -ForegroundColor White
    exit 1
}
Write-Host ""

# Database credentials
$DB_USER = "postgres"
$DB_NAME = "nieuws_scraper"
$DB_PASSWORD = "postgres"

Write-Host "Database aanmaken..." -ForegroundColor Yellow
Write-Host "  Database: $DB_NAME" -ForegroundColor White
Write-Host "  User: $DB_USER" -ForegroundColor White
Write-Host ""

# Set password environment variable
$env:PGPASSWORD = $DB_PASSWORD

# Create database
Write-Host "Stap 1: Database aanmaken" -ForegroundColor Cyan
$createDbOutput = & $psqlPath -U $DB_USER -h localhost -c "CREATE DATABASE $DB_NAME;" 2>&1
if ($LASTEXITCODE -eq 0 -or $createDbOutput -like "*already exists*") {
    Write-Host "OK Database bestaat of is aangemaakt" -ForegroundColor Green
} else {
    Write-Host "FOUT bij aanmaken database: $createDbOutput" -ForegroundColor Red
    Write-Host ""
    Write-Host "Controleer het wachtwoord in .env (POSTGRES_PASSWORD)" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

# Run migrations
Write-Host "Stap 2: Database tabellen aanmaken" -ForegroundColor Cyan
$migrateOutput = & $psqlPath -U $DB_USER -h localhost -d $DB_NAME -f "migrations/001_create_tables.sql" 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "OK Tabellen aangemaakt" -ForegroundColor Green
} else {
    Write-Host "FOUT bij aanmaken tabellen: $migrateOutput" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Clear password
$env:PGPASSWORD = ""

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Database setup voltooid!" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Start nu de API met: .\scripts\start.ps1" -ForegroundColor Cyan
Write-Host ""