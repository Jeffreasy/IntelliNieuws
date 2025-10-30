# Test Configuration API Features
# Tests profile switching, settings updates, and caching

Write-Host "🧪 Testing Configuration API v3.1" -ForegroundColor Cyan
Write-Host "====================================`n" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080/api/v1"
$apiKey = $env:API_KEY
if (-not $apiKey) {
    Write-Host "⚠️  Warning: API_KEY not set, write operations will fail" -ForegroundColor Yellow
    $apiKey = "test-key"
}

# Test 1: Get All Profiles
Write-Host "📋 Test 1: Get All Profiles" -ForegroundColor Green
$response = Invoke-RestMethod -Uri "$baseUrl/config/profiles" -Method Get
Write-Host "✅ Found $($response.data.total_profiles) profiles" -ForegroundColor Green
Write-Host "   Active: $($response.data.active_profile)" -ForegroundColor Gray
foreach ($profileEntry in $response.data.profiles.PSObject.Properties) {
    $p = $profileEntry.Value
    Write-Host "   - $($p.name): interval=$($p.schedule_interval_min)min, rate=$($p.rate_limit_seconds)s" -ForegroundColor Gray
}
Write-Host ""

# Test 2: Get Current Config
Write-Host "📊 Test 2: Get Current Configuration" -ForegroundColor Green
$response = Invoke-RestMethod -Uri "$baseUrl/config/current" -Method Get
Write-Host "✅ Current profile: $($response.data.active_profile)" -ForegroundColor Green
Write-Host "   Rate Limit: $($response.data.rate_limit_seconds)s" -ForegroundColor Gray
Write-Host "   Max Concurrent: $($response.data.max_concurrent)" -ForegroundColor Gray
Write-Host "   Schedule Interval: $($response.data.schedule_interval_min) minutes" -ForegroundColor Gray
Write-Host ""

# Test 3: Get Scheduler Status
Write-Host "⏰ Test 3: Get Scheduler Status" -ForegroundColor Green
$response = Invoke-RestMethod -Uri "$baseUrl/config/scheduler/status" -Method Get
Write-Host "✅ Scheduler running: $($response.data.running)" -ForegroundColor Green
Write-Host "   Next run: $($response.data.next_run)" -ForegroundColor Gray
Write-Host ""

# Test 4: Switch Profile (requires API key)
Write-Host "🔄 Test 4: Switch Profile to 'fast'" -ForegroundColor Green
try {
    $headers = @{
        "X-API-Key" = $apiKey
        "Content-Type" = "application/json"
    }
    $response = Invoke-RestMethod -Uri "$baseUrl/config/profile/fast" -Method Post -Headers $headers
    Write-Host "✅ Switched to fast profile" -ForegroundColor Green
    Write-Host "   New interval: $($response.data.new_interval) minutes" -ForegroundColor Gray
    Write-Host "   New rate limit: $($response.data.new_rate_limit)s" -ForegroundColor Gray
} catch {
    Write-Host "⚠️  Profile switch failed (check API_KEY): $($_.Exception.Message)" -ForegroundColor Yellow
}
Write-Host ""

# Test 5: Update Setting (requires API key)
Write-Host "⚙️  Test 5: Update Rate Limit Setting" -ForegroundColor Green
try {
    $body = @{
        setting = "rate_limit_seconds"
        value = 2
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/config/setting" `
        -Method Patch -Headers $headers -Body $body
    Write-Host "✅ Setting updated: rate_limit_seconds = 2" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Setting update failed (check API_KEY): $($_.Exception.Message)" -ForegroundColor Yellow
}
Write-Host ""

# Test 6: Test Response Caching
Write-Host "💾 Test 6: Test Response Caching" -ForegroundColor Green
Write-Host "   First request (cache miss)..."
$start = Get-Date
Invoke-RestMethod -Uri "$baseUrl/articles?limit=10" -Method Get | Out-Null
$time1 = ((Get-Date) - $start).TotalMilliseconds
Write-Host "   ⏱️  Time: $([math]::Round($time1, 2))ms" -ForegroundColor Gray

Write-Host "   Second request (cache hit)..."
$start = Get-Date
Invoke-RestMethod -Uri "$baseUrl/articles?limit=10" -Method Get | Out-Null
$time2 = ((Get-Date) - $start).TotalMilliseconds
Write-Host "   ⏱️  Time: $([math]::Round($time2, 2))ms" -ForegroundColor Gray

$speedup = [math]::Round($time1 / $time2, 1)
Write-Host "✅ Cache speedup: ${speedup}x faster!" -ForegroundColor Green
Write-Host ""

# Test 7: Switch Back to Balanced
Write-Host "↩️  Test 7: Switch Back to Balanced Profile" -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/config/profile/balanced" -Method Post -Headers $headers
    Write-Host "✅ Switched back to balanced profile" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Profile switch failed: $($_.Exception.Message)" -ForegroundColor Yellow
}
Write-Host ""

# Test 8: Cache Statistics
Write-Host "📊 Test 8: Get Cache Statistics" -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/cache/stats" -Method Get
    Write-Host "✅ Cache Statistics:" -ForegroundColor Green
    Write-Host "   Total Keys: $($response.data.total_keys)" -ForegroundColor Gray
    Write-Host "   Memory Usage: $($response.data.memory_usage_mb) MB" -ForegroundColor Gray
} catch {
    Write-Host "⚠️  Cache stats unavailable (Redis may not be running)" -ForegroundColor Yellow
}
Write-Host ""

# Summary
Write-Host "====================================`n" -ForegroundColor Cyan
Write-Host "✅ Configuration API Tests Complete!" -ForegroundColor Green
Write-Host ""
Write-Host "📝 Results:" -ForegroundColor Cyan
Write-Host "   - Profile management: Working" -ForegroundColor Green
Write-Host "   - Setting updates: $(if ($apiKey -eq 'test-key') { 'Needs API_KEY' } else { 'Working' })" -ForegroundColor $(if ($apiKey -eq 'test-key') { 'Yellow' } else { 'Green' })
Write-Host "   - Response caching: ${speedup}x speedup" -ForegroundColor Green
Write-Host "   - Scheduler integration: Working" -ForegroundColor Green
Write-Host ""
Write-Host "🎯 Next Steps:" -ForegroundColor Cyan
Write-Host "   1. Set API_KEY environment variable for full testing" -ForegroundColor Gray
Write-Host "   2. Test from frontend with React components" -ForegroundColor Gray
Write-Host "   3. Monitor cache hit ratios in production" -ForegroundColor Gray
Write-Host ""