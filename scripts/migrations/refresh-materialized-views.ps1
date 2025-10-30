# PowerShell script to refresh materialized views
# Run this periodically (e.g., every 5-15 minutes) via Task Scheduler or cron

param(
    [string]$DatabaseHost = "localhost",
    [int]$DatabasePort = 5432,
    [string]$DatabaseName = "nieuws_scraper",
    [string]$DatabaseUser = "postgres"
)

Write-Host "Refreshing materialized views..." -ForegroundColor Cyan

# Refresh trending keywords view (concurrent to avoid blocking)
$query = "REFRESH MATERIALIZED VIEW CONCURRENTLY mv_trending_keywords;"

try {
    $env:PGPASSWORD = $env:DATABASE_PASSWORD
    
    psql -h $DatabaseHost -p $DatabasePort -U $DatabaseUser -d $DatabaseName -c $query
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Successfully refreshed mv_trending_keywords" -ForegroundColor Green
    } else {
        Write-Host "✗ Failed to refresh mv_trending_keywords" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
} finally {
    Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue
}

Write-Host "Materialized view refresh completed!" -ForegroundColor Green