# Test script voor de Nieuws Scraper

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Nieuws Scraper - Test Script" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# API configuratie
$API_URL = "http://localhost:8080"
$API_KEY = "test123geheim"

Write-Host "1. Health Check..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$API_URL/health" -Method Get
    Write-Host "OK API is actief" -ForegroundColor Green
    Write-Host "   Status: $($health.status)" -ForegroundColor White
} catch {
    Write-Host "FOUT API is niet bereikbaar!" -ForegroundColor Red
    Write-Host "   Zorg dat de API draait met: .\scripts\start.ps1" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

Write-Host "2. Scraping triggeren (alle bronnen)..." -ForegroundColor Yellow
try {
    $scrapeResult = Invoke-RestMethod -Uri "$API_URL/api/v1/scrape" -Method Post -Headers @{"X-API-Key"=$API_KEY}
    Write-Host "OK Scraping gestart" -ForegroundColor Green
    Write-Host "   Message: $($scrapeResult.message)" -ForegroundColor White
} catch {
    Write-Host "FOUT bij starten scraping: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "Wacht 5 seconden voor scraping..." -ForegroundColor Yellow
Start-Sleep -Seconds 5
Write-Host ""

Write-Host "3. Artikelen ophalen..." -ForegroundColor Yellow
try {
    $articles = Invoke-RestMethod -Uri "$API_URL/api/v1/articles?limit=5" -Method Get
    Write-Host "OK Artikelen opgehaald" -ForegroundColor Green
    Write-Host "   Aantal: $($articles.data.Count)" -ForegroundColor White
    Write-Host "   Totaal: $($articles.total)" -ForegroundColor White
    Write-Host ""
    
    if ($articles.data.Count -gt 0) {
        Write-Host "Laatste artikelen:" -ForegroundColor Cyan
        foreach ($article in $articles.data) {
            Write-Host "   - $($article.title)" -ForegroundColor White
            Write-Host "     Bron: $($article.source) | $($article.published)" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "FOUT bij ophalen artikelen: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "4. Statistieken ophalen..." -ForegroundColor Yellow
try {
    $stats = Invoke-RestMethod -Uri "$API_URL/api/v1/articles/stats" -Method Get
    Write-Host "OK Statistieken:" -ForegroundColor Green
    Write-Host "   Totaal artikelen: $($stats.total_articles)" -ForegroundColor White
    Write-Host "   Bronnen: $($stats.sources.Count)" -ForegroundColor White
} catch {
    Write-Host "Kon geen statistieken ophalen" -ForegroundColor Yellow
}
Write-Host ""

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "Test voltooid!" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Andere test opties:" -ForegroundColor Yellow
Write-Host "  - Specifieke bron scrapen:" -ForegroundColor White
Write-Host '    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/scrape" -Method Post -Headers @{"X-API-Key"="test123geheim"} -Body (ConvertTo-Json @{source="nu.nl"}) -ContentType "application/json"' -ForegroundColor Gray
Write-Host ""
Write-Host "  - Artikelen filteren:" -ForegroundColor White
Write-Host '    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/articles?source=nu.nl&limit=10"' -ForegroundColor Gray
Write-Host ""
Write-Host "  - Browser gebruiken:" -ForegroundColor White
Write-Host "    http://localhost:8080/api/v1/articles" -ForegroundColor Gray
Write-Host ""