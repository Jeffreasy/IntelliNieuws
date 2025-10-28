package ai

import (
	"github.com/jeffrey/intellinieuws/internal/models"
)

// ChatRequest represents a user's chat message
type ChatRequest struct {
	Message        string `json:"message" validate:"required,min=1,max=1000"`
	Context        string `json:"context,omitempty"`         // Optional context (conversation history)
	ArticleContent string `json:"article_content,omitempty"` // Optional article content for context (no length limit)
	ArticleID      int64  `json:"article_id,omitempty"`      // Optional article ID for context
}

// ChatResponse represents the AI's response
type ChatResponse struct {
	Message  string           `json:"message"`
	Articles []models.Article `json:"articles,omitempty"`
	Stats    interface{}      `json:"stats,omitempty"`
	Sources  []string         `json:"sources,omitempty"`
}

// FunctionCall represents OpenAI function calling
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Available functions for OpenAI
const (
	FunctionSearchArticles      = "search_articles"
	FunctionGetSentimentStats   = "get_sentiment_stats"
	FunctionGetTrendingTopics   = "get_trending_topics"
	FunctionGetArticlesByEntity = "get_articles_by_entity"
	FunctionGetRecentArticles   = "get_recent_articles"
)

// Function definitions for OpenAI
var ChatFunctions = []map[string]interface{}{
	{
		"name":        FunctionSearchArticles,
		"description": "Search for articles based on keywords, topics, or content. Returns articles matching the search query.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query (keywords, topics, or phrases)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of articles to return (default: 10, max: 50)",
				},
			},
			"required": []string{"query"},
		},
	},
	{
		"name":        FunctionGetSentimentStats,
		"description": "Get sentiment statistics for articles. Can be filtered by source or date range. Returns positive, neutral, and negative counts.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Filter by news source (e.g., 'nu.nl', 'nos.nl')",
				},
				"hours_back": map[string]interface{}{
					"type":        "integer",
					"description": "Look back X hours (default: 24)",
				},
			},
			"required": []string{},
		},
	},
	{
		"name":        FunctionGetTrendingTopics,
		"description": "Get currently trending topics based on article keywords and frequency. Returns trending keywords with article counts.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"hours_back": map[string]interface{}{
					"type":        "integer",
					"description": "Look back X hours (default: 24)",
				},
				"min_articles": map[string]interface{}{
					"type":        "integer",
					"description": "Minimum articles per topic (default: 3)",
				},
			},
			"required": []string{},
		},
	},
	{
		"name":        FunctionGetArticlesByEntity,
		"description": "Get articles mentioning a specific person, organization, or location. Useful for finding news about specific entities.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"entity_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the person, organization, or location",
				},
				"entity_type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"persons", "organizations", "locations"},
					"description": "Type of entity",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of articles (default: 20, max: 50)",
				},
			},
			"required": []string{"entity_name"},
		},
	},
	{
		"name":        FunctionGetRecentArticles,
		"description": "Get the most recent articles. Can be filtered by source, category, or sentiment.",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "Filter by news source",
				},
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Filter by category",
				},
				"sentiment": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"positive", "neutral", "negative"},
					"description": "Filter by sentiment",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of articles (default: 10, max: 50)",
				},
			},
			"required": []string{},
		},
	},
}
