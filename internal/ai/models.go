package ai

import (
	"time"
)

// AIEnrichment represents all AI-processed data for an article
type AIEnrichment struct {
	Processed   bool               `json:"processed"`
	ProcessedAt *time.Time         `json:"processed_at,omitempty"`
	Sentiment   *SentimentAnalysis `json:"sentiment,omitempty"`
	Categories  map[string]float64 `json:"categories,omitempty"`
	Entities    *EntityExtraction  `json:"entities,omitempty"`
	Keywords    []Keyword          `json:"keywords,omitempty"`
	Summary     string             `json:"summary,omitempty"`
	Error       string             `json:"error,omitempty"`
}

// SentimentAnalysis contains sentiment detection results
type SentimentAnalysis struct {
	Score      float64 `json:"score"`                // -1.0 to 1.0
	Label      string  `json:"label"`                // positive, negative, neutral
	Confidence float64 `json:"confidence,omitempty"` // 0.0 to 1.0
}

// EntityExtraction contains extracted named entities
type EntityExtraction struct {
	Persons       []string `json:"persons,omitempty"`
	Organizations []string `json:"organizations,omitempty"`
	Locations     []string `json:"locations,omitempty"`
}

// Keyword represents a keyword with relevance score
type Keyword struct {
	Word  string  `json:"word"`
	Score float64 `json:"score"` // 0.0 to 1.0
}

// ProcessingRequest represents a request to process an article
type ProcessingRequest struct {
	ArticleID int64
	Title     string
	Content   string
	Summary   string
	URL       string
	Source    string
}

// ProcessingResult contains the result of AI processing
type ProcessingResult struct {
	ArticleID   int64
	Enrichment  *AIEnrichment
	Success     bool
	Error       error
	ProcessedAt time.Time
}

// BatchProcessingResult contains results for multiple articles
type BatchProcessingResult struct {
	Results        []*ProcessingResult
	TotalProcessed int
	SuccessCount   int
	FailureCount   int
	Duration       time.Duration
}

// SentimentStats represents sentiment statistics
type SentimentStats struct {
	TotalArticles     int     `json:"total_articles"`
	PositiveCount     int     `json:"positive_count"`
	NeutralCount      int     `json:"neutral_count"`
	NegativeCount     int     `json:"negative_count"`
	AverageSentiment  float64 `json:"average_sentiment"`
	MostPositiveTitle string  `json:"most_positive_title,omitempty"`
	MostNegativeTitle string  `json:"most_negative_title,omitempty"`
}

// TrendingTopic represents a trending topic with metadata
type TrendingTopic struct {
	Keyword          string   `json:"keyword"`
	ArticleCount     int      `json:"article_count"`
	AverageSentiment float64  `json:"average_sentiment"`
	Sources          []string `json:"sources"`
}

// CategoryPrediction represents a category with confidence score
type CategoryPrediction struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
}

// Config holds AI service configuration
type Config struct {
	// OpenAI settings
	OpenAIAPIKey    string
	OpenAIModel     string
	OpenAIMaxTokens int

	// Processing settings
	Enabled         bool
	AsyncProcessing bool
	BatchSize       int
	ProcessInterval time.Duration
	RetryFailed     bool
	MaxRetries      int

	// Feature toggles
	EnableSentiment  bool
	EnableEntities   bool
	EnableCategories bool
	EnableKeywords   bool
	EnableSummary    bool
	EnableSimilarity bool

	// Cost control
	MaxDailyCost       float64
	RateLimitPerMinute int
	Timeout            time.Duration
}

// ProcessingOptions allows fine-grained control over processing
type ProcessingOptions struct {
	EnableSentiment  bool
	EnableEntities   bool
	EnableCategories bool
	EnableKeywords   bool
	EnableSummary    bool
	Force            bool // Reprocess even if already processed
}

// DefaultProcessingOptions returns default processing options
func DefaultProcessingOptions() ProcessingOptions {
	return ProcessingOptions{
		EnableSentiment:  true,
		EnableEntities:   true,
		EnableCategories: true,
		EnableKeywords:   true,
		EnableSummary:    false, // More expensive
		Force:            false,
	}
}

// SimilarityResult represents similar articles
type SimilarityResult struct {
	ArticleID      int64    `json:"article_id"`
	Title          string   `json:"title"`
	Similarity     float64  `json:"similarity"`
	SharedKeywords []string `json:"shared_keywords,omitempty"`
}

// OpenAIRequest represents a request to OpenAI API
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// OpenAIResponse represents a response from OpenAI API
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GetSentimentLabel returns the sentiment label based on score
func GetSentimentLabel(score float64) string {
	if score >= 0.2 {
		return "positive"
	} else if score <= -0.2 {
		return "negative"
	}
	return "neutral"
}

// ValidateEnrichment checks if enrichment data is valid
func (e *AIEnrichment) ValidateEnrichment() bool {
	if e == nil {
		return false
	}

	// At least one field should be populated
	hasData := e.Sentiment != nil ||
		len(e.Categories) > 0 ||
		e.Entities != nil ||
		len(e.Keywords) > 0 ||
		e.Summary != ""

	return hasData
}

// GetTopCategories returns top N categories by confidence
func (e *AIEnrichment) GetTopCategories(n int) []CategoryPrediction {
	if len(e.Categories) == 0 {
		return nil
	}

	predictions := make([]CategoryPrediction, 0, len(e.Categories))
	for cat, conf := range e.Categories {
		predictions = append(predictions, CategoryPrediction{
			Category:   cat,
			Confidence: conf,
		})
	}

	// Sort by confidence (simple bubble sort for small n)
	for i := 0; i < len(predictions)-1; i++ {
		for j := i + 1; j < len(predictions); j++ {
			if predictions[j].Confidence > predictions[i].Confidence {
				predictions[i], predictions[j] = predictions[j], predictions[i]
			}
		}
	}

	if n > len(predictions) {
		n = len(predictions)
	}

	return predictions[:n]
}

// GetTopKeywords returns top N keywords by score
func (e *AIEnrichment) GetTopKeywords(n int) []Keyword {
	if len(e.Keywords) == 0 {
		return nil
	}

	keywords := make([]Keyword, len(e.Keywords))
	copy(keywords, e.Keywords)

	// Sort by score
	for i := 0; i < len(keywords)-1; i++ {
		for j := i + 1; j < len(keywords); j++ {
			if keywords[j].Score > keywords[i].Score {
				keywords[i], keywords[j] = keywords[j], keywords[i]
			}
		}
	}

	if n > len(keywords) {
		n = len(keywords)
	}

	return keywords[:n]
}

// GetAllEntities returns all entities as a flat list
func (e *EntityExtraction) GetAllEntities() []string {
	if e == nil {
		return nil
	}

	all := make([]string, 0)
	all = append(all, e.Persons...)
	all = append(all, e.Organizations...)
	all = append(all, e.Locations...)

	return all
}
