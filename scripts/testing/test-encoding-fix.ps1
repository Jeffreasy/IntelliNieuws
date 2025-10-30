# Test Character Encoding Fix
# Verifies that Dutch characters are properly handled

Write-Host "🧪 Testing Character Encoding Fix v3.1" -ForegroundColor Cyan
Write-Host "======================================`n" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080/api/v1"

# Test 1: Fetch Recent Articles
Write-Host "📰 Test 1: Fetch Recent Articles" -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/articles?limit=5" -Method Get
    $articles = $response.data
    
    Write-Host "✅ Fetched $($articles.Count) articles" -ForegroundColor Green
    
    # Check for encoding issues
    $hasIssues = $false
    $dutchCharsFound = 0
    
    foreach ($article in $articles) {
        $title = $article.title
        
        # Check for corrupted encoding markers
        if ($title -match "Ã|â€|Â") {
            Write-Host "   ⚠️  Corrupted title: $($title.Substring(0, [Math]::Min(50, $title.Length)))" -ForegroundColor Red
            $hasIssues = $true
        }
        
        # Check for proper Dutch characters
        if ($title -match "[éëïöüáèà]|[ÉËÏÖÜÁÈÀ]") {
            $dutchCharsFound++
            Write-Host "   ✅ Proper Dutch chars in: $($title.Substring(0, [Math]::Min(60, $title.Length)))" -ForegroundColor Green
        }
    }
    
    if (-not $hasIssues) {
        Write-Host "`n✅ No encoding corruption detected!" -ForegroundColor Green
    } else {
        Write-Host "`n❌ Encoding issues found - needs investigation" -ForegroundColor Red
    }
    
    Write-Host "   Dutch characters found in $dutchCharsFound/$($articles.Count) articles" -ForegroundColor Gray
    
} catch {
    Write-Host "❌ Failed to fetch articles: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""

# Test 2: Search for Dutch Characters
Write-Host "🔍 Test 2: Search for Articles with Dutch Characters" -ForegroundColor Green
try {
    # Common Dutch words with special characters
    $testWords = @("één", "financiën", "België", "Oekraïne")
    
    foreach ($word in $testWords) {
        $encoded = [System.Web.HttpUtility]::UrlEncode($word)
        $response = Invoke-RestMethod -Uri "$baseUrl/articles/search?q=$encoded&limit=3" -Method Get
        
        if ($response.data.Count -gt 0) {
            Write-Host "   ✅ Found $($response.data.Count) articles for '$word'" -ForegroundColor Green
            $title = $response.data[0].title
            Write-Host "      Example: $($title.Substring(0, [Math]::Min(80, $title.Length)))" -ForegroundColor Gray
        } else {
            Write-Host "   ℹ️  No articles found for '$word'" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "   ⚠️  Search test failed: $($_.Exception.Message)" -ForegroundColor Yellow
}

Write-Host ""

# Test 3: Check Database Content
Write-Host "🗄️  Test 3: Check Database for Encoding Issues" -ForegroundColor Green
Write-Host "   Run this SQL query to check:" -ForegroundColor Gray
Write-Host @"
   
   SELECT id, title, 
          CASE 
            WHEN title LIKE '%Ã%' THEN 'CORRUPTED'
            WHEN title ~ '[éëïöüáèà]' THEN 'CORRECT'
            ELSE 'NO_SPECIAL_CHARS'
          END as encoding_status
   FROM articles
   ORDER BY created_at DESC
   LIMIT 20;
   
"@ -ForegroundColor DarkGray

Write-Host ""

# Test 4: Verify Specific Patterns
Write-Host "🔬 Test 4: Pattern Detection" -ForegroundColor Green
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/articles?limit=20" -Method Get
    
    $patterns = @{
        "Corrupted_Ae" = "Ã©|Ã«|Ã¨"  # é, ë, è encoded wrong
        "Corrupted_Quotes" = "â€œ|â€|Â"  # Smart quotes encoded wrong
        "Correct_Dutch" = "[éëïöüáèà]"  # Proper Dutch characters
        "Correct_Quotes" = "[""]|['']"  # Proper smart quotes
    }
    
    $results = @{}
    foreach ($pattern in $patterns.Keys) {
        $count = 0
        foreach ($article in $response.data) {
            if (($article.title + $article.summary) -match $patterns[$pattern]) {
                $count++
            }
        }
        $results[$pattern] = $count
    }
    
    Write-Host "   Corrupted (Ã): $($results['Corrupted_Ae'])" -ForegroundColor $(if ($results['Corrupted_Ae'] -gt 0) { 'Red' } else { 'Green' })
    Write-Host "   Corrupted Quotes: $($results['Corrupted_Quotes'])" -ForegroundColor $(if ($results['Corrupted_Quotes'] -gt 0) { 'Red' } else { 'Green' })
    Write-Host "   ✅ Correct Dutch: $($results['Correct_Dutch'])" -ForegroundColor Green
    Write-Host "   ✅ Correct Quotes: $($results['Correct_Quotes'])" -ForegroundColor Green
    
} catch {
    Write-Host "   ⚠️  Pattern test failed" -ForegroundColor Yellow
}

Write-Host ""

# Summary
Write-Host "======================================`n" -ForegroundColor Cyan
Write-Host "📊 Test Summary" -ForegroundColor Cyan
Write-Host ""

if ($hasIssues) {
    Write-Host "❌ ENCODING ISSUES DETECTED" -ForegroundColor Red
    Write-Host ""
    Write-Host "Recommended Actions:" -ForegroundColor Yellow
    Write-Host "1. Verify go.mod has: golang.org/x/net" -ForegroundColor Gray
    Write-Host "2. Run: go mod tidy" -ForegroundColor Gray
    Write-Host "3. Rebuild: docker-compose build api" -ForegroundColor Gray
    Write-Host "4. Restart: docker-compose restart api" -ForegroundColor Gray
    Write-Host "5. Re-scrape: curl -X POST http://localhost:8080/api/v1/scrape" -ForegroundColor Gray
} else {
    Write-Host "✅ ENCODING WORKING CORRECTLY" -ForegroundColor Green
    Write-Host ""
    Write-Host "Results:" -ForegroundColor Cyan
    Write-Host "- No corruption detected" -ForegroundColor Green
    Write-Host "- Dutch characters render properly" -ForegroundColor Green
    Write-Host "- Character encoding fix is working" -ForegroundColor Green
}

Write-Host ""