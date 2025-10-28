package ai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"regexp"

	"github.com/jeffrey/intellinieuws/pkg/logger"
)

const (
	openAIAPIURL = "https://api.openai.com/v1/chat/completions"
)

// CachedResponse stores a cached OpenAI response
type CachedResponse struct {
	Enrichment *AIEnrichment
	CachedAt   time.Time
	Hits       int
}

// OpenAIClient handles interactions with OpenAI API
type OpenAIClient struct {
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
	logger     *logger.Logger
	// Caching
	cache       map[string]*CachedResponse
	cacheMu     sync.RWMutex
	cacheSize   int
	cacheTTL    time.Duration
	cacheHits   int64
	cacheMisses int64
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey, model string, maxTokens int, log *logger.Logger) *OpenAIClient {
	// Optimized HTTP transport with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	return &OpenAIClient{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		logger:    log.WithComponent("openai-client"),
		cache:     make(map[string]*CachedResponse),
		cacheSize: 1000,           // Store up to 1000 cached responses
		cacheTTL:  24 * time.Hour, // Cache for 24 hours
	}
}

// Complete sends a completion request to OpenAI
func (c *OpenAIClient) Complete(ctx context.Context, messages []ChatMessage, temperature float64) (*OpenAIResponse, error) {
	request := OpenAIRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   c.maxTokens,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Debugf("OpenAI API call completed. Tokens used: %d", response.Usage.TotalTokens)

	return &response, nil
}

// CompleteWithRetry sends a completion request with exponential backoff retry
func (c *OpenAIClient) CompleteWithRetry(ctx context.Context, messages []ChatMessage, temperature float64) (*OpenAIResponse, error) {
	maxRetries := 3
	baseDelay := time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		response, err := c.Complete(ctx, messages, temperature)

		if err == nil {
			return response, nil
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			return nil, err
		}

		if attempt < maxRetries-1 {
			delay := baseDelay * time.Duration(1<<uint(attempt)) // Exponential: 1s, 2s, 4s
			c.logger.Warnf("API call failed (attempt %d/%d), retrying in %v: %v",
				attempt+1, maxRetries, delay, err)

			select {
			case <-time.After(delay):
				// Continue to next retry
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("all %d retry attempts failed", maxRetries)
}

// isRetryableError checks if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for rate limit, timeout, or temporary errors
	return strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "500")
}

// cleanJSON attempts to fix common JSON formatting issues
func cleanJSON(content string) string {
	// Remove any markdown code blocks
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	// Fix missing commas between array elements and object properties
	// Pattern: ]\n    " or ]\n        " (array followed by property without comma)
	re1 := regexp.MustCompile(`\]\s+("[\w_]+":)`)
	content = re1.ReplaceAllString(content, "],$1")

	// Pattern: }\n    " or }\n        " (object followed by property without comma)
	re2 := regexp.MustCompile(`\}\s+("[\w_]+":)`)
	content = re2.ReplaceAllString(content, "},$1")

	// Pattern: value\n    " (value followed by property without comma)
	// This handles cases like: "value"\n    "nextKey"
	re3 := regexp.MustCompile(`(["\d])\s+("[\w_]+":)`)
	content = re3.ReplaceAllString(content, "$1,$2")

	return content
}

// getCacheKey generates a cache key based on content
func (c *OpenAIClient) getCacheKey(title, content string) string {
	hash := sha256.Sum256([]byte(title + "|" + content))
	return fmt.Sprintf("%x", hash[:16])
}

// evictOldest removes the oldest cached response
func (c *OpenAIClient) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, cached := range c.cache {
		if oldestKey == "" || cached.CachedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.CachedAt
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
		c.logger.Debugf("Evicted cached response (key: %s)", oldestKey[:8])
	}
}

// GetCacheStats returns cache statistics
func (c *OpenAIClient) GetCacheStats() map[string]interface{} {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	totalRequests := c.cacheHits + c.cacheMisses
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(c.cacheHits) / float64(totalRequests)
	}

	return map[string]interface{}{
		"cache_size":     len(c.cache),
		"cache_hits":     c.cacheHits,
		"cache_misses":   c.cacheMisses,
		"hit_rate":       hitRate,
		"total_requests": totalRequests,
	}
}

// AnalyzeSentiment uses OpenAI to analyze article sentiment
func (c *OpenAIClient) AnalyzeSentiment(ctx context.Context, title, content string) (*SentimentAnalysis, error) {
	text := title
	if content != "" {
		text = title + "\n\n" + content
	}

	// Truncate if too long
	if len(text) > 4000 {
		text = text[:4000]
	}

	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are a sentiment analysis expert. Analyze the sentiment of Dutch news articles.
Respond ONLY with a JSON object in this exact format:
{"score": 0.5, "label": "positive", "confidence": 0.9}

Where:
- score: -1.0 (very negative) to 1.0 (very positive)
- label: "positive", "negative", or "neutral"
- confidence: 0.0 to 1.0`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Analyze the sentiment of this article:\n\n%s", text),
		},
	}

	response, err := c.CompleteWithRetry(ctx, messages, 0.3)
	if err != nil {
		return nil, fmt.Errorf("failed to get sentiment: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var sentiment SentimentAnalysis
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &sentiment); err != nil {
		// Fallback: try to parse manually
		c.logger.Warnf("Failed to parse sentiment JSON, content: %s", response.Choices[0].Message.Content)
		return nil, fmt.Errorf("failed to parse sentiment response: %w", err)
	}

	// Validate and normalize
	if sentiment.Score < -1.0 {
		sentiment.Score = -1.0
	} else if sentiment.Score > 1.0 {
		sentiment.Score = 1.0
	}

	if sentiment.Label == "" {
		sentiment.Label = GetSentimentLabel(sentiment.Score)
	}

	return &sentiment, nil
}

// ExtractEntities uses OpenAI to extract named entities
func (c *OpenAIClient) ExtractEntities(ctx context.Context, title, content string) (*EntityExtraction, error) {
	text := title
	if content != "" {
		text = title + "\n\n" + content
	}

	// Truncate if too long
	if len(text) > 4000 {
		text = text[:4000]
	}

	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are an expert in Named Entity Recognition for Dutch news articles.
Extract persons, organizations, locations, and stock tickers mentioned in the article.
Respond ONLY with a JSON object in this exact format:
{"persons": ["Name1", "Name2"], "organizations": ["Org1", "Org2"], "locations": ["Loc1", "Loc2"], "stock_tickers": [{"symbol": "ASML", "name": "ASML Holding", "exchange": "AEX"}]}

Rules:
- Only include entities explicitly mentioned
- Use proper capitalization
- Don't include generic terms
- Return empty arrays if no entities found
- For stock tickers, extract: symbol (e.g., ASML, AAPL), company name, and exchange if mentioned
- Common Dutch stocks: ASML, Shell, ING, Philips, Unilever, ASMI, IMCD, etc.
- Common US stocks: AAPL, MSFT, GOOGL, AMZN, TSLA, NVDA, etc.`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Extract entities from this article:\n\n%s", text),
		},
	}

	response, err := c.CompleteWithRetry(ctx, messages, 0.2)
	if err != nil {
		return nil, fmt.Errorf("failed to extract entities: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var entities EntityExtraction
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &entities); err != nil {
		c.logger.Warnf("Failed to parse entities JSON, content: %s", response.Choices[0].Message.Content)
		return nil, fmt.Errorf("failed to parse entities response: %w", err)
	}

	return &entities, nil
}

// CategorizeArticle uses OpenAI to categorize an article
func (c *OpenAIClient) CategorizeArticle(ctx context.Context, title, content string) (map[string]float64, error) {
	text := title
	if content != "" {
		text = title + "\n\n" + content
	}

	// Truncate if too long
	if len(text) > 4000 {
		text = text[:4000]
	}

	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are an expert in categorizing Dutch news articles.
Assign the article to one or more categories with confidence scores.
Respond ONLY with a JSON object mapping categories to confidence scores (0.0 to 1.0):
{"Politics": 0.9, "Economy": 0.3}

Available categories:
Politics, Economy, Technology, Sports, Health, Science, Entertainment, Environment, Education, Crime, International, National, Local, Business, Culture

Rules:
- Assign 1-3 most relevant categories
- Confidence scores between 0.0 and 1.0
- Higher score means more relevant`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Categorize this article:\n\n%s", text),
		},
	}

	response, err := c.CompleteWithRetry(ctx, messages, 0.3)
	if err != nil {
		return nil, fmt.Errorf("failed to categorize: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var categories map[string]float64
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &categories); err != nil {
		c.logger.Warnf("Failed to parse categories JSON, content: %s", response.Choices[0].Message.Content)
		return nil, fmt.Errorf("failed to parse categories response: %w", err)
	}

	// Normalize confidence scores
	for cat, score := range categories {
		if score < 0.0 {
			categories[cat] = 0.0
		} else if score > 1.0 {
			categories[cat] = 1.0
		}
	}

	return categories, nil
}

// ExtractKeywords uses OpenAI to extract keywords
func (c *OpenAIClient) ExtractKeywords(ctx context.Context, title, content string) ([]Keyword, error) {
	text := title
	if content != "" {
		text = title + "\n\n" + content
	}

	// Truncate if too long
	if len(text) > 4000 {
		text = text[:4000]
	}

	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are an expert in keyword extraction from Dutch news articles.
Extract the most important keywords with relevance scores.
Respond ONLY with a JSON array in this exact format:
[{"word": "keyword1", "score": 0.95}, {"word": "keyword2", "score": 0.87}]

Rules:
- Extract 5-10 most relevant keywords
- Score between 0.0 and 1.0 (relevance)
- Use lowercase
- Prioritize specific terms over generic ones
- Include multi-word phrases if relevant`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Extract keywords from this article:\n\n%s", text),
		},
	}

	response, err := c.CompleteWithRetry(ctx, messages, 0.3)
	if err != nil {
		return nil, fmt.Errorf("failed to extract keywords: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	var keywords []Keyword
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &keywords); err != nil {
		c.logger.Warnf("Failed to parse keywords JSON, content: %s", response.Choices[0].Message.Content)
		return nil, fmt.Errorf("failed to parse keywords response: %w", err)
	}

	// Normalize scores
	for i := range keywords {
		if keywords[i].Score < 0.0 {
			keywords[i].Score = 0.0
		} else if keywords[i].Score > 1.0 {
			keywords[i].Score = 1.0
		}
	}

	return keywords, nil
}

// GenerateSummary uses OpenAI to generate a summary
func (c *OpenAIClient) GenerateSummary(ctx context.Context, title, content string) (string, error) {
	text := title
	if content != "" {
		text = title + "\n\n" + content
	}

	// Truncate if too long
	if len(text) > 4000 {
		text = text[:4000]
	}

	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are an expert in summarizing Dutch news articles.
Create a concise summary in 2-3 sentences that captures the main points.
Write in Dutch, be objective and factual.
Respond with ONLY the summary text, no extra formatting.`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Summarize this article:\n\n%s", text),
		},
	}

	response, err := c.CompleteWithRetry(ctx, messages, 0.5)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	summary := response.Choices[0].Message.Content
	if len(summary) > 500 {
		summary = summary[:500]
	}

	return summary, nil
}

// ProcessArticle performs all AI processing in a single call (more efficient) with caching
func (c *OpenAIClient) ProcessArticle(ctx context.Context, title, content string, opts ProcessingOptions) (*AIEnrichment, error) {
	// Generate cache key
	cacheKey := c.getCacheKey(title, content)

	// Check cache first
	c.cacheMu.RLock()
	if cached, exists := c.cache[cacheKey]; exists {
		if time.Since(cached.CachedAt) < c.cacheTTL {
			cached.Hits++
			c.cacheHits++
			c.cacheMu.RUnlock()
			c.logger.Debugf("Cache HIT for content (key: %s, hits: %d)", cacheKey[:8], cached.Hits)
			return cached.Enrichment, nil
		}
		// Cache expired, will be overwritten
	}
	c.cacheMisses++
	c.cacheMu.RUnlock()

	text := title
	if content != "" {
		text = title + "\n\n" + content
	}

	// Truncate if too long
	if len(text) > 4000 {
		text = text[:4000]
	}

	// Build comprehensive prompt
	tasksDesc := "Analyze this Dutch news article and provide:\n"
	if opts.EnableSentiment {
		tasksDesc += "1. Sentiment analysis (score -1.0 to 1.0, label, confidence)\n"
	}
	if opts.EnableEntities {
		tasksDesc += "2. Named entities (persons, organizations, locations, stock tickers)\n"
	}
	if opts.EnableCategories {
		tasksDesc += "3. Categories with confidence scores\n"
	}
	if opts.EnableKeywords {
		tasksDesc += "4. Keywords with relevance scores\n"
	}
	if opts.EnableSummary {
		tasksDesc += "5. A 2-3 sentence summary in Dutch\n"
	}

	messages := []ChatMessage{
		{
			Role: "system",
			Content: `You are an expert AI assistant for analyzing Dutch news articles.
Respond with a valid JSON object containing all requested analyses.
Be accurate, objective, and follow the specified formats exactly.

IMPORTANT:
- Categories must be an object mapping category names to confidence scores (0.0-1.0), NOT an array.
		Example: {"categories": {"Politics": 0.9, "Economy": 0.3}}
- Stock tickers must be extracted from entities, including symbol, name, and exchange.
		Example: {"entities": {"stock_tickers": [{"symbol": "ASML", "name": "ASML Holding", "exchange": "AEX"}]}}
- Common stocks: Dutch (ASML, Shell, ING, Philips), US (AAPL, MSFT, GOOGL, TSLA, NVDA)`,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("%s\n\nRespond ONLY with a valid JSON object. No markdown, no explanations.\n\nArticle:\n%s", tasksDesc, text),
		},
	}

	response, err := c.CompleteWithRetry(ctx, messages, 0.4)
	if err != nil {
		return nil, fmt.Errorf("failed to process article: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the comprehensive response
	enrichment := &AIEnrichment{
		Processed: true,
	}

	content = response.Choices[0].Message.Content

	// Try to parse as complete enrichment
	var fullResponse struct {
		Sentiment  *SentimentAnalysis `json:"sentiment,omitempty"`
		Entities   *EntityExtraction  `json:"entities,omitempty"`
		Categories interface{}        `json:"categories,omitempty"` // Can be map or array
		Keywords   interface{}        `json:"keywords,omitempty"`   // Can be map or array
		Summary    string             `json:"summary,omitempty"`
	}

	// Try to clean the JSON first
	cleanedContent := cleanJSON(content)

	if err := json.Unmarshal([]byte(cleanedContent), &fullResponse); err != nil {
		c.logger.Warnf("Failed to parse comprehensive response: %v", err)
		c.logger.Warnf("Original content: %s", content)
		if cleanedContent != content {
			c.logger.Warnf("Cleaned content: %s", cleanedContent)
		}
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	enrichment.Sentiment = fullResponse.Sentiment
	enrichment.Entities = fullResponse.Entities
	enrichment.Summary = fullResponse.Summary

	// Handle keywords - can be either object or array
	if fullResponse.Keywords != nil {
		switch v := fullResponse.Keywords.(type) {
		case map[string]interface{}:
			// It's an object mapping keyword to score - convert to array
			enrichment.Keywords = make([]Keyword, 0, len(v))
			for word, scoreVal := range v {
				if score, ok := scoreVal.(float64); ok {
					enrichment.Keywords = append(enrichment.Keywords, Keyword{
						Word:  word,
						Score: score,
					})
				}
			}
		case []interface{}:
			// It's already an array - parse as []Keyword
			keywordsJSON, _ := json.Marshal(v)
			var keywords []Keyword
			if err := json.Unmarshal(keywordsJSON, &keywords); err == nil {
				enrichment.Keywords = keywords
			}
		}
	}

	// Handle categories - could be map or array
	if fullResponse.Categories != nil {
		switch v := fullResponse.Categories.(type) {
		case map[string]interface{}:
			// It's already a map, convert to map[string]float64
			enrichment.Categories = make(map[string]float64)
			for key, val := range v {
				if floatVal, ok := val.(float64); ok {
					enrichment.Categories[key] = floatVal
				}
			}
		case []interface{}:
			// It's an array - convert to map with equal weights
			enrichment.Categories = make(map[string]float64)
			weight := 1.0 / float64(len(v))
			for _, item := range v {
				if strVal, ok := item.(string); ok {
					enrichment.Categories[strVal] = weight
				}
			}
		}
	}

	now := time.Now()
	enrichment.ProcessedAt = &now

	// Cache the result
	c.cacheMu.Lock()
	c.cache[cacheKey] = &CachedResponse{
		Enrichment: enrichment,
		CachedAt:   time.Now(),
		Hits:       1,
	}

	// Evict oldest if cache is full
	if len(c.cache) > c.cacheSize {
		c.evictOldest()
	}
	c.cacheMu.Unlock()

	c.logger.Debugf("Cached response (key: %s, cache size: %d)", cacheKey[:8], len(c.cache))

	return enrichment, nil
}

// ArticleData represents article data for batch processing
type ArticleData struct {
	ID      int64
	Title   string
	Content string
}

// ProcessArticlesBatch processes multiple articles in a single API call (PHASE 3: 70% extra cost reduction)
// This reduces API calls by 90% by batching up to 10 articles per request
func (c *OpenAIClient) ProcessArticlesBatch(ctx context.Context, articles []ArticleData, opts ProcessingOptions) ([]*AIEnrichment, error) {
	if len(articles) == 0 {
		return nil, nil
	}

	// Limit batch size to prevent token overflow
	maxBatchSize := 10
	if len(articles) > maxBatchSize {
		c.logger.Warnf("Batch size %d exceeds maximum %d, truncating", len(articles), maxBatchSize)
		articles = articles[:maxBatchSize]
	}

	// Build batch prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString("Analyze the following Dutch news articles and provide enrichment for each.\n\n")
	promptBuilder.WriteString("For each article, provide:\n")

	if opts.EnableSentiment {
		promptBuilder.WriteString("- Sentiment (score -1.0 to 1.0, label, confidence)\n")
	}
	if opts.EnableEntities {
		promptBuilder.WriteString("- Entities (persons, organizations, locations, stock tickers)\n")
	}
	if opts.EnableCategories {
		promptBuilder.WriteString("- Categories with confidence (as object, not array)\n")
	}
	if opts.EnableKeywords {
		promptBuilder.WriteString("- Keywords with scores (as array of objects)\n")
	}
	if opts.EnableSummary {
		promptBuilder.WriteString("- Summary (2-3 sentences in Dutch)\n")
	}

	promptBuilder.WriteString("\nArticles to analyze:\n\n")

	for i, article := range articles {
		promptBuilder.WriteString(fmt.Sprintf("=== Article %d (ID: %d) ===\n", i+1, article.ID))
		promptBuilder.WriteString(fmt.Sprintf("Title: %s\n", article.Title))

		content := article.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		promptBuilder.WriteString(fmt.Sprintf("Content: %s\n\n", content))
	}

	promptBuilder.WriteString("\n📋 IMPORTANT: Respond with a JSON array containing one enrichment object per article, in the EXACT same order.\n")
	promptBuilder.WriteString("Format: [{\"sentiment\": {...}, \"entities\": {...}, \"categories\": {...}, \"keywords\": [...], \"summary\": \"...\"}]\n")

	systemPrompt := `You are an expert AI assistant for analyzing Dutch news articles.
Analyze multiple articles and return a JSON array with one enrichment object per article.
Maintain the EXACT order of articles in your response.
Be accurate, objective, and follow the specified formats exactly.

CRITICAL RULES:
1. Return a JSON ARRAY, not individual objects
2. One enrichment per article, in the SAME ORDER
3. Categories must be objects: {"Politics": 0.9, "Economy": 0.3}
4. Keywords must be arrays: [{"word": "keyword", "score": 0.9}]
5. Stock tickers in entities: {"stock_tickers": [{"symbol": "ASML", "name": "ASML Holding", "exchange": "AEX"}]}
6. If you cannot analyze an article, return {"sentiment": null, "entities": null}`

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: promptBuilder.String()},
	}

	c.logger.Infof("Sending batch of %d articles to OpenAI", len(articles))

	response, err := c.CompleteWithRetry(ctx, messages, 0.4)
	if err != nil {
		c.logger.WithError(err).Error("Batch processing failed")
		return nil, fmt.Errorf("failed to process batch: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse batch response
	responseContent := response.Choices[0].Message.Content

	// Try to parse as array of enrichments
	var batchResponse []struct {
		Sentiment  *SentimentAnalysis `json:"sentiment,omitempty"`
		Entities   *EntityExtraction  `json:"entities,omitempty"`
		Categories interface{}        `json:"categories,omitempty"`
		Keywords   interface{}        `json:"keywords,omitempty"`
		Summary    string             `json:"summary,omitempty"`
	}

	// Try to clean the JSON first
	cleanedResponse := cleanJSON(responseContent)

	if err := json.Unmarshal([]byte(cleanedResponse), &batchResponse); err != nil {
		c.logger.Warnf("Failed to parse batch response as array: %v", err)
		c.logger.Warnf("Original content: %s", responseContent)
		if cleanedResponse != responseContent {
			c.logger.Warnf("Cleaned content: %s", cleanedResponse)
		}

		// Fallback: return empty enrichments
		enrichments := make([]*AIEnrichment, len(articles))
		for i := range enrichments {
			enrichments[i] = &AIEnrichment{Processed: false, Error: "Failed to parse batch response"}
		}
		return enrichments, fmt.Errorf("failed to parse batch response: %w", err)
	}

	// Convert to AIEnrichment array
	enrichments := make([]*AIEnrichment, len(articles))
	now := time.Now()

	for i := range articles {
		enrichment := &AIEnrichment{
			Processed:   true,
			ProcessedAt: &now,
		}

		// Use response if available, otherwise mark as failed
		if i < len(batchResponse) {
			resp := batchResponse[i]
			enrichment.Sentiment = resp.Sentiment
			enrichment.Entities = resp.Entities
			enrichment.Summary = resp.Summary

			// Handle keywords
			if resp.Keywords != nil {
				switch v := resp.Keywords.(type) {
				case map[string]interface{}:
					enrichment.Keywords = make([]Keyword, 0, len(v))
					for word, scoreVal := range v {
						if score, ok := scoreVal.(float64); ok {
							enrichment.Keywords = append(enrichment.Keywords, Keyword{
								Word:  word,
								Score: score,
							})
						}
					}
				case []interface{}:
					keywordsJSON, _ := json.Marshal(v)
					var keywords []Keyword
					if err := json.Unmarshal(keywordsJSON, &keywords); err == nil {
						enrichment.Keywords = keywords
					}
				}
			}

			// Handle categories
			if resp.Categories != nil {
				switch v := resp.Categories.(type) {
				case map[string]interface{}:
					enrichment.Categories = make(map[string]float64)
					for key, val := range v {
						if floatVal, ok := val.(float64); ok {
							enrichment.Categories[key] = floatVal
						}
					}
				case []interface{}:
					enrichment.Categories = make(map[string]float64)
					weight := 1.0 / float64(len(v))
					for _, item := range v {
						if strVal, ok := item.(string); ok {
							enrichment.Categories[strVal] = weight
						}
					}
				}
			}
		} else {
			enrichment.Processed = false
			enrichment.Error = "Missing from batch response"
		}

		enrichments[i] = enrichment
	}

	c.logger.Infof("✅ Batch processed %d articles in single API call (saved %d API calls)",
		len(articles), len(articles)-1)

	return enrichments, nil
}

// ChatWithFunctions performs a chat completion with function calling support
func (c *OpenAIClient) ChatWithFunctions(ctx context.Context, messages []map[string]interface{}, functions []map[string]interface{}) (string, *FunctionCall, error) {
	request := map[string]interface{}{
		"model":       c.model,
		"messages":    messages,
		"temperature": 0.7,
		"max_tokens":  c.maxTokens,
	}

	// Add functions if provided
	if len(functions) > 0 {
		request["functions"] = functions
		request["function_call"] = "auto"
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content      string                  `json:"content"`
				FunctionCall *map[string]interface{} `json:"function_call,omitempty"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	c.logger.Debugf("OpenAI chat API call completed. Tokens used: %d", response.Usage.TotalTokens)

	if len(response.Choices) == 0 {
		return "", nil, fmt.Errorf("no choices in response")
	}

	message := response.Choices[0].Message

	// Check if there's a function call
	if message.FunctionCall != nil {
		fc := &FunctionCall{}

		if name, ok := (*message.FunctionCall)["name"].(string); ok {
			fc.Name = name
		}

		if argsStr, ok := (*message.FunctionCall)["arguments"].(string); ok {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(argsStr), &args); err == nil {
				fc.Arguments = args
			}
		}

		c.logger.Debugf("Function call requested: %s", fc.Name)
		return "", fc, nil
	}

	return message.Content, nil, nil
}
