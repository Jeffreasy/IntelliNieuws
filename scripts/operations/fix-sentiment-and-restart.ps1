# Fix Sentiment Analysis and Restart Backend
Write-Host "=== FIXING SENTIMENT ANALYSIS ===" -ForegroundColor Cyan

Write-Host "`nStep 1: Clear Redis Cache" -ForegroundColor Yellow
docker exec nieuws-scraper-redis redis-cli -a redis_password FLUSHDB
Write-Host "✅ Cache cleared" -ForegroundColor Green

Write-Host "`nStep 2: Rebuild and Restart Backend" -ForegroundColor Yellow
docker-compose up -d --build app
Write-Host "✅ Backend rebuilding..." -ForegroundColor Green

Write-Host "`nStep 3: Wait for backend to start (15 seconds)" -ForegroundColor Yellow
Start-Sleep -Seconds 15

Write-Host "`nStep 4: Test Sentiment Stats Endpoint" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/ai/sentiment/stats" -Method GET
    Write-Host "✅ API Response:" -ForegroundColor Green
    Write-Host ($response | ConvertTo-Json -Depth 10)
    
    if ($response.data.positive_count -gt 0 -or $response.data.negative_count -gt 0 -or $response.data.neutral_count -gt 0) {
        Write-Host "`n✅ SUCCESS! Sentiment labels are now being counted correctly!" -ForegroundColor Green
        Write-Host "  Positive: $($response.data.positive_count)" -ForegroundColor Green
        Write-Host "  Neutral: $($response.data.neutral_count)" -ForegroundColor Green
        Write-Host "  Negative: $($response.data.negative_count)" -ForegroundColor Green
    } else {
        Write-Host "`n⚠️  WARNING: Still showing 0 counts. Check logs." -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ API Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== FIX COMPLETE ===" -ForegroundColor Cyan
Write-Host "Check logs with: docker logs nieuws-scraper-app --tail 50" -ForegroundColor Yellow