# ============================================================================
# Test Script: Professional Schema & Analytics API
# Description: Comprehensive testing of V001-V003 migrations and analytics
# Version: 1.0.0
# Author: NieuwsScraper Team
# Date: 2025-10-30
# ============================================================================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "PROFESSIONAL SCHEMA TEST SUITE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080"
$dbContainer = "nieuws-scraper-postgres"
$dbUser = "scraper"
$dbName = "nieuws_scraper"

$passed = 0
$failed = 0

# Helper function to run test
function Test-Feature {
    param(
        [string]$Name,
        [scriptblock]$Test
    )
    
    Write-Host "Testing: $Name" -ForegroundColor Yellow -NoNewline
    
    try {
        $result = & $Test
        if ($result) {
            Write-Host " ✓ PASS" -ForegroundColor Green
            $script:passed++
            return $true
        } else {
            Write-Host " ✗ FAIL" -ForegroundColor Red
            $script:failed++
            return $false
        }
    } catch {
        Write-Host " ✗ ERROR: $_" -ForegroundColor Red
        $script:failed++
        return $false
    }
}

Write-Host "1. DATABASE TESTS" -ForegroundColor Cyan
Write-Host "-----------------" -ForegroundColor Cyan

# Test 1.1: Schema migrations applied
Test-Feature "Schema migrations applied" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM schema_migrations;" 2>&1
    $count = [int]$result.Trim()
    return $count -ge 3
}

# Test 1.2: All tables exist
Test-Feature "Core tables exist" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('articles', 'sources', 'scraping_jobs', 'emails', 'schema_migrations');" 2>&1
    $count = [int]$result.Trim()
    return $count -eq 5
}

# Test 1.3: Materialized views exist
Test-Feature "Materialized views exist" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM pg_matviews WHERE schemaname = 'public';" 2>&1
    $count = [int]$result.Trim()
    return $count -ge 1
}

# Test 1.4: Articles preserved
Test-Feature "All articles preserved" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM articles;" 2>&1
    $count = [int]$result.Trim()
    return $count -gt 0
}

# Test 1.5: Indexes created
Test-Feature "Indexes created" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public';" 2>&1
    $count = [int]$result.Trim()
    return $count -ge 50
}

Write-Host ""
Write-Host "2. ANALYTICS API TESTS" -ForegroundColor Cyan
Write-Host "----------------------" -ForegroundColor Cyan

# Test 2.1: Trending keywords endpoint
Test-Feature "GET /analytics/trending" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/trending?limit=5" -Method Get -TimeoutSec 10
        return $response.trending -ne $null
    } catch {
        return $false
    }
}

# Test 2.2: Sentiment trends endpoint
Test-Feature "GET /analytics/sentiment-trends" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/sentiment-trends" -Method Get -TimeoutSec 10
        return $response.trends -ne $null
    } catch {
        return $false
    }
}

# Test 2.3: Hot entities endpoint
Test-Feature "GET /analytics/hot-entities" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/hot-entities?limit=10" -Method Get -TimeoutSec 10
        return $response.entities -ne $null
    } catch {
        return $false
    }
}

# Test 2.4: Analytics overview endpoint
Test-Feature "GET /analytics/overview" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/overview" -Method Get -TimeoutSec 10
        return $response.trending_keywords -ne $null -and $response.hot_entities -ne $null
    } catch {
        return $false
    }
}

# Test 2.5: Article stats endpoint
Test-Feature "GET /analytics/article-stats" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/article-stats" -Method Get -TimeoutSec 10
        return $response.sources -ne $null
    } catch {
        return $false
    }
}

# Test 2.6: Database health endpoint
Test-Feature "GET /analytics/database-health" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/database-health" -Method Get -TimeoutSec 10
        return $response.status -eq "healthy"
    } catch {
        return $false
    }
}

# Test 2.7: Maintenance schedule endpoint
Test-Feature "GET /analytics/maintenance-schedule" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/maintenance-schedule" -Method Get -TimeoutSec 10
        return $response.tasks -ne $null
    } catch {
        return $false
    }
}

Write-Host ""
Write-Host "3. DATABASE FUNCTION TESTS" -ForegroundColor Cyan
Write-Host "---------------------------" -ForegroundColor Cyan

# Test 3.1: get_trending_topics function
Test-Feature "get_trending_topics() function" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM get_trending_topics(24, 3, 10);" 2>&1
    return $? -eq $true
}

# Test 3.2: refresh_analytics_views function
Test-Feature "refresh_analytics_views() function" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM refresh_analytics_views(TRUE);" 2>&1
    return $? -eq $true
}

# Test 3.3: get_maintenance_schedule function
Test-Feature "get_maintenance_schedule() function" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM get_maintenance_schedule();" 2>&1
    return $? -eq $true
}

Write-Host ""
Write-Host "4. VIEW TESTS" -ForegroundColor Cyan
Write-Host "-------------" -ForegroundColor Cyan

# Test 4.1: v_trending_keywords_24h view
Test-Feature "v_trending_keywords_24h view" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM v_trending_keywords_24h;" 2>&1
    return $? -eq $true
}

# Test 4.2: v_article_stats view
Test-Feature "v_article_stats view" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM v_article_stats;" 2>&1
    $count = [int]$result.Trim()
    return $count -gt 0
}

# Test 4.3: v_active_sources view
Test-Feature "v_active_sources view" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM v_active_sources;" 2>&1
    $count = [int]$result.Trim()
    return $count -gt 0
}

# Test 4.4: v_email_stats view
Test-Feature "v_email_stats view" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT total_emails FROM v_email_stats;" 2>&1
    return $? -eq $true
}

Write-Host ""
Write-Host "5. PERFORMANCE TESTS" -ForegroundColor Cyan
Write-Host "--------------------" -ForegroundColor Cyan

# Test 5.1: Trending query performance (should be < 1s)
Test-Feature "Trending query < 1 second" {
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT * FROM v_trending_keywords_24h LIMIT 10;" 2>&1
    $stopwatch.Stop()
    return $stopwatch.ElapsedMilliseconds -lt 1000
}

# Test 5.2: Article list query performance
Test-Feature "Article list query < 200ms" {
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT * FROM articles ORDER BY published DESC LIMIT 50;" 2>&1
    $stopwatch.Stop()
    return $stopwatch.ElapsedMilliseconds -lt 200
}

# Test 5.3: Cache hit ratio > 95%
Test-Feature "Cache hit ratio > 95%" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT ROUND(100.0 * SUM(heap_blks_hit) / NULLIF(SUM(heap_blks_hit) + SUM(heap_blks_read), 0), 2) FROM pg_statio_user_tables;" 2>&1
    $ratio = [decimal]$result.Trim()
    return $ratio -gt 95
}

Write-Host ""
Write-Host "6. DATA INTEGRITY TESTS" -ForegroundColor Cyan
Write-Host "-----------------------" -ForegroundColor Cyan

# Test 6.1: No duplicate articles
Test-Feature "No duplicate article URLs" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM (SELECT url, COUNT(*) FROM articles GROUP BY url HAVING COUNT(*) > 1) duplicates;" 2>&1
    $count = [int]$result.Trim()
    return $count -eq 0
}

# Test 6.2: All AI processed articles have sentiment
Test-Feature "AI processed articles have sentiment" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM articles WHERE ai_processed = TRUE AND ai_sentiment IS NULL;" 2>&1
    $count = [int]$result.Trim()
    return $count -eq 0
}

# Test 6.3: Content hash uniqueness
Test-Feature "Content hash uniqueness" {
    $result = docker exec $dbContainer psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM (SELECT content_hash, COUNT(*) FROM articles WHERE content_hash IS NOT NULL GROUP BY content_hash HAVING COUNT(*) > 1) duplicates;" 2>&1
    $count = [int]$result.Trim()
    return $count -eq 0
}

Write-Host ""
Write-Host "7. API RESPONSE TIME TESTS" -ForegroundColor Cyan
Write-Host "---------------------------" -ForegroundColor Cyan

# Test 7.1: Analytics overview < 300ms
Test-Feature "Analytics overview < 300ms" {
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/overview" -Method Get -TimeoutSec 5
        $stopwatch.Stop()
        return $stopwatch.ElapsedMilliseconds -lt 300
    } catch {
        $stopwatch.Stop()
        return $false
    }
}

# Test 7.2: Trending keywords < 100ms
Test-Feature "Trending keywords < 100ms" {
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/trending?limit=5" -Method Get -TimeoutSec 5
        $stopwatch.Stop()
        return $stopwatch.ElapsedMilliseconds -lt 100
    } catch {
        $stopwatch.Stop()
        return $false
    }
}

# Test 7.3: Article stats < 100ms
Test-Feature "Article stats < 100ms" {
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/article-stats" -Method Get -TimeoutSec 5
        $stopwatch.Stop()
        return $stopwatch.ElapsedMilliseconds -lt 100
    } catch {
        $stopwatch.Stop()
        return $false
    }
}

Write-Host ""
Write-Host "8. FEATURE VALIDATION TESTS" -ForegroundColor Cyan
Write-Host "----------------------------" -ForegroundColor Cyan

# Test 8.1: Trending keywords have data
Test-Feature "Trending keywords populated" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/trending?limit=1" -Method Get
        return $response.trending.Count -gt 0
    } catch {
        return $false
    }
}

# Test 8.2: Article stats have all sources
Test-Feature "Article stats for all sources" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/article-stats" -Method Get
        return $response.sources.Count -ge 3
    } catch {
        return $false
    }
}

# Test 8.3: Database health returns metrics
Test-Feature "Database health metrics" {
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/analytics/database-health" -Method Get
        return $response.table_sizes -ne $null -and $response.cache_hit_ratio -ne $null
    } catch {
        return $false
    }
}

Write-Host ""
Write-Host "9. ROLLBACK SAFETY TESTS" -ForegroundColor Cyan
Write-Host "------------------------" -ForegroundColor Cyan

# Test 9.1: Rollback scripts exist
Test-Feature "Rollback scripts exist" {
    $v001 = Test-Path "migrations\rollback\V001__rollback.sql"
    $v002 = Test-Path "migrations\rollback\V002__rollback.sql"
    $v003 = Test-Path "migrations\rollback\V003__rollback.sql"
    return $v001 -and $v002 -and $v003
}

# Test 9.2: Utility scripts exist
Test-Feature "Utility scripts exist" {
    $legacy = Test-Path "migrations\utilities\01_migrate_from_legacy.sql"
    $health = Test-Path "migrations\utilities\02_health_check.sql"
    $maint = Test-Path "migrations\utilities\03_maintenance.sql"
    return $legacy -and $health -and $maint
}

# Test 9.3: Documentation exists
Test-Feature "Documentation complete" {
    $readme = Test-Path "migrations\README.md"
    $guide = Test-Path "migrations\MIGRATION-GUIDE.md"
    $ref = Test-Path "migrations\QUICK-REFERENCE.md"
    return $readme -and $guide -and $ref
}

Write-Host ""
Write-Host "10. CODE QUALITY TESTS" -ForegroundColor Cyan
Write-Host "----------------------" -ForegroundColor Cyan

# Test 10.1: Constants file exists
Test-Feature "Constants file created" {
    return Test-Path "internal\models\constants.go"
}

# Test 10.2: Analytics handler exists
Test-Feature "Analytics handler created" {
    return Test-Path "internal\api\handlers\analytics_handler.go"
}

# Test 10.3: Models updated
Test-Feature "Enhanced models exist" {
    $email = Test-Path "internal\models\email.go"
    $article = Test-Path "internal\models\article.go"
    return $email -and $article
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "TEST RESULTS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Passed: $passed" -ForegroundColor Green
Write-Host "Failed: $failed" -ForegroundColor Red
Write-Host "Total:  $($passed + $failed)" -ForegroundColor White
Write-Host ""

$percentage = [math]::Round(($passed / ($passed + $failed)) * 100, 2)
Write-Host "Success Rate: $percentage%" -ForegroundColor $(if ($percentage -ge 90) { "Green" } elseif ($percentage -ge 70) { "Yellow" } else { "Red" })
Write-Host ""

if ($failed -eq 0) {
    Write-Host "✅ ALL TESTS PASSED!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Professional schema is fully operational!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Quick Links:" -ForegroundColor Yellow
    Write-Host "  - Analytics Overview: $baseUrl/api/v1/analytics/overview" -ForegroundColor White
    Write-Host "  - Trending Keywords:  $baseUrl/api/v1/analytics/trending" -ForegroundColor White
    Write-Host "  - Database Health:    $baseUrl/api/v1/analytics/database-health" -ForegroundColor White
    Write-Host ""
    exit 0
} else {
    Write-Host "⚠️  SOME TESTS FAILED" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please review failed tests above and check:" -ForegroundColor Yellow
    Write-Host "  1. Is the application running? (docker-compose ps)" -ForegroundColor White
    Write-Host "  2. Are migrations applied? (SELECT * FROM schema_migrations;)" -ForegroundColor White
    Write-Host "  3. Check application logs (docker-compose logs app)" -ForegroundColor White
    Write-Host ""
    exit 1
}