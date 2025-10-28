# FMP Free Tier Integration Test
# Tests only endpoints available in free tier

$baseUrl = "http://localhost:8080"

Write-Host "Testing FMP Free Tier Integration..." -ForegroundColor Cyan
Write-Host "Note: Only US stocks supported in free tier" -ForegroundColor Yellow
Write-Host ""

# Test 1: Single Quote (US Stock)
Write-Host "Test 1: Single Quote (US Stock - AAPL)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/stocks/quote/AAPL" -Method Get
    Write-Host "SUCCESS: Quote for AAPL: $($response.price)" -ForegroundColor Green
    Write-Host "  Name: $($response.name)" -ForegroundColor Cyan
    Write-Host "  Change: $($response.change_percent)%" -ForegroundColor Cyan
} catch {
    Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 2: Another US Stock
Write-Host "Test 2: Single Quote (US Stock - MSFT)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/stocks/quote/MSFT" -Method Get
    Write-Host "SUCCESS: Quote for MSFT: $($response.price)" -ForegroundColor Green
} catch {
    Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 3: Company Profile
Write-Host "Test 3: Company Profile (AAPL)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/stocks/profile/AAPL" -Method Get
    Write-Host "SUCCESS: Profile for $($response.company_name)" -ForegroundColor Green
    Write-Host "  Exchange: $($response.exchange)" -ForegroundColor Cyan
} catch {
    Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 4: Earnings Calendar
Write-Host "Test 4: Earnings Calendar" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/stocks/earnings" -Method Get
    Write-Host "SUCCESS: Found $($response.total) upcoming earnings" -ForegroundColor Green
    if ($response.earnings.Count -gt 0) {
        $first = $response.earnings[0]
        Write-Host "  Next: $($first.symbol) on $(Get-Date $first.date -Format 'yyyy-MM-dd')" -ForegroundColor Cyan
    }
} catch {
    Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 5: Symbol Search
Write-Host "Test 5: Symbol Search (query: apple)" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/stocks/search?q=apple&limit=5" -Method Get
    Write-Host "SUCCESS: Found $($response.total) results" -ForegroundColor Green
    if ($response.results.Count -gt 0) {
        Write-Host "  Top result: $($response.results[0].symbol) - $($response.results[0].company_name)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 6: Cache Stats
Write-Host "Test 6: Cache Statistics" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/stocks/stats" -Method Get
    Write-Host "SUCCESS: Cache enabled: $($response.cache.enabled)" -ForegroundColor Green
} catch {
    Write-Host "FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Summary
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "FMP FREE TIER TEST SUMMARY" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Working Endpoints:" -ForegroundColor Green
Write-Host "  - Single quotes (US stocks only)" -ForegroundColor White
Write-Host "  - Company profiles" -ForegroundColor White  
Write-Host "  - Earnings calendar" -ForegroundColor White
Write-Host "  - Symbol search" -ForegroundColor White
Write-Host ""
Write-Host "Limitations:" -ForegroundColor Yellow
Write-Host "  - No batch quotes (premium required)" -ForegroundColor White
Write-Host "  - No non-US stocks like ASML (premium required)" -ForegroundColor White
Write-Host "  - No market performance data (premium required)" -ForegroundColor White
Write-Host "  - No historical/metrics/news (premium required)" -ForegroundColor White
Write-Host ""
Write-Host "For full features, consider FMP Starter: $14/month" -ForegroundColor Cyan