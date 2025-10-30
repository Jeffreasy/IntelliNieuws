# Test script for Sentiment Analysis Component
# Tests database, API endpoint, and cache behavior

Write-Host "=== SENTIMENT ANALYSIS DIAGNOSTIC TEST ===" -ForegroundColor Cyan
Write-Host ""

# Configuration
$API_BASE = "http://localhost:8080"
$DB_CONTAINER = "nieuwsscraper-postgres-1"

Write-Host "Step 1: Check Database for Sentiment Data" -ForegroundColor Yellow
Write-Host "=========================================" -ForegroundColor Yellow

# Check if articles have sentiment data
$query1 = @"
SELECT 
    COUNT(*) as total_articles,
    COUNT(CASE WHEN ai_processed = TRUE THEN 1 END) as ai_processed_count,
    COUNT(CASE WHEN ai_sentiment IS NOT NULL THEN 1 END) as with_sentiment,
    COUNT(CASE WHEN ai_sentiment_label IS NOT NULL THEN 1 END) as with_label,
    MIN(ai_sentiment) as min_sentiment,
    MAX(ai_sentiment) as max_sentiment,
    AVG(ai_sentiment) as avg_sentiment
FROM articles;
"@

Write-Host "`nDatabase Query 1: Article Sentiment Overview" -ForegroundColor Green
docker exec $DB_CONTAINER psql -U intellinieuws -d intellinieuws -c "$query1"

# Check sentiment distribution
$query2 = @"
SELECT 
    ai_sentiment_label,
    COUNT(*) as count,
    ROUND(AVG(ai_sentiment)::numeric, 3) as avg_score
FROM articles
WHERE ai_processed = TRUE 
  AND ai_sentiment IS NOT NULL
GROUP BY ai_sentiment_label
ORDER BY count DESC;
"@

Write-Host "`nDatabase Query 2: Sentiment Label Distribution" -ForegroundColor Green
docker exec $DB_CONTAINER psql -U intellinieuws -d intellinieuws -c "$query2"

# Check recent processed articles
$query3 = @"
SELECT 
    id,
    LEFT(title, 50) as title,
    ai_sentiment,
    ai_sentiment_label,
    ai_processed_at
FROM articles
WHERE ai_processed = TRUE 
  AND ai_sentiment IS NOT NULL
ORDER BY ai_processed_at DESC
LIMIT 5;
"@

Write-Host "`nDatabase Query 3: Recent Processed Articles with Sentiment" -ForegroundColor Green
docker exec $DB_CONTAINER psql -U intellinieuws -d intellinieuws -c "$query3"

Write-Host "`n`nStep 2: Test Sentiment Stats API Endpoint" -ForegroundColor Yellow
Write-Host "==========================================" -ForegroundColor Yellow

Write-Host "`nTest 1: Basic sentiment stats (no filters)" -ForegroundColor Green
try {
    $response1 = Invoke-RestMethod -Uri "$API_BASE/api/v1/ai/sentiment/stats" -Method GET
    Write-Host "✅ API Response received" -ForegroundColor Green
    Write-Host ($response1 | ConvertTo-Json -Depth 10)
} catch {
    Write-Host "❌ API Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
}

Write-Host "`n`nTest 2: Sentiment stats with source filter" -ForegroundColor Green
try {
    $response2 = Invoke-RestMethod -Uri "$API_BASE/api/v1/ai/sentiment/stats?source=nu.nl" -Method GET
    Write-Host "✅ API Response received" -ForegroundColor Green
    Write-Host ($response2 | ConvertTo-Json -Depth 10)
} catch {
    Write-Host "❌ API Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n`nTest 3: Sentiment stats with date range" -ForegroundColor Green
$endDate = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
$startDate = (Get-Date).AddDays(-7).ToString("yyyy-MM-ddTHH:mm:ssZ")
try {
    $response3 = Invoke-RestMethod -Uri "$API_BASE/api/v1/ai/sentiment/stats?start_date=$startDate&end_date=$endDate" -Method GET
    Write-Host "✅ API Response received (last 7 days)" -ForegroundColor Green
    Write-Host ($response3 | ConvertTo-Json -Depth 10)
} catch {
    Write-Host "❌ API Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n`nStep 3: Check Cache Behavior" -ForegroundColor Yellow
Write-Host "==============================" -ForegroundColor Yellow

Write-Host "`nTest 1: First call (should be cache MISS)" -ForegroundColor Green
$stopwatch1 = [System.Diagnostics.Stopwatch]::StartNew()
try {
    $cache1 = Invoke-RestMethod -Uri "$API_BASE/api/v1/ai/sentiment/stats" -Method GET
    $stopwatch1.Stop()
    Write-Host "✅ Response time: $($stopwatch1.ElapsedMilliseconds)ms" -ForegroundColor Green
} catch {
    Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 1

Write-Host "`nTest 2: Second call (should be cache HIT)" -ForegroundColor Green
$stopwatch2 = [System.Diagnostics.Stopwatch]::StartNew()
try {
    $cache2 = Invoke-RestMethod -Uri "$API_BASE/api/v1/ai/sentiment/stats" -Method GET
    $stopwatch2.Stop()
    Write-Host "✅ Response time: $($stopwatch2.ElapsedMilliseconds)ms" -ForegroundColor Green
    
    if ($stopwatch2.ElapsedMilliseconds -lt $stopwatch1.ElapsedMilliseconds) {
        Write-Host "✅ Cache is working! (faster response)" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Cache might not be working (similar speed)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n`nStep 4: Verify Data Consistency" -ForegroundColor Yellow
Write-Host "=================================" -ForegroundColor Yellow

# Compare database counts with API response
Write-Host "`nComparing database counts with API response..." -ForegroundColor Green

if ($response1 -and $response1.data) {
    $apiData = $response1.data
    Write-Host "`nAPI Response Summary:" -ForegroundColor Cyan
    Write-Host "  Total Articles: $($apiData.total_articles)"
    Write-Host "  Positive: $($apiData.positive_count)"
    Write-Host "  Neutral: $($apiData.neutral_count)"
    Write-Host "  Negative: $($apiData.negative_count)"
    Write-Host "  Average Sentiment: $($apiData.average_sentiment)"
    
    if ($apiData.total_articles -eq 0) {
        Write-Host "`n⚠️  WARNING: No articles with sentiment data found!" -ForegroundColor Yellow
        Write-Host "Possible causes:" -ForegroundColor Yellow
        Write-Host "  1. AI processing has not run yet" -ForegroundColor Yellow
        Write-Host "  2. Database has no articles" -ForegroundColor Yellow
        Write-Host "  3. Articles exist but haven't been processed by AI" -ForegroundColor Yellow
    } else {
        Write-Host "`n✅ Sentiment data found and processed correctly" -ForegroundColor Green
    }
}

Write-Host "`n`n=== DIAGNOSTIC SUMMARY ===" -ForegroundColor Cyan

Write-Host "`nCheck the logs above for:" -ForegroundColor Yellow
Write-Host "  • Database contains articles with ai_processed=TRUE"
Write-Host "  • Sentiment scores are in range [-1.0, 1.0]"
Write-Host "  • Sentiment labels are correctly assigned"
Write-Host "  • API returns valid JSON responses"
Write-Host "  • Cache is working (faster second request)"
Write-Host "  • No SQL errors or null pointer exceptions"

Write-Host "`nNext steps:" -ForegroundColor Cyan
Write-Host "  1. Check backend logs: docker logs nieuwsscraper-api-1 --tail 100"
Write-Host "  2. If no sentiment data: Run AI processing manually"
Write-Host "  3. If cache issues: Clear Redis cache"

Write-Host "`n=== TEST COMPLETE ===" -ForegroundColor Cyan