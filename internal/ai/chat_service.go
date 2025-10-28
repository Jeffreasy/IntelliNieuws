package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// ChatService handles conversational AI interactions
type ChatService struct {
	aiService    *Service
	openAIClient *OpenAIClient
	logger       *logger.Logger
}

// NewChatService creates a new chat service
func NewChatService(aiService *Service, openAIClient *OpenAIClient, log *logger.Logger) *ChatService {
	return &ChatService{
		aiService:    aiService,
		openAIClient: openAIClient,
		logger:       log.WithComponent("chat-service"),
	}
}

// ProcessChatMessageWithContext processes a user's chat message with optional article context
func (cs *ChatService) ProcessChatMessageWithContext(ctx context.Context, message string, conversationContext string, articleContent string, articleID int64) (*ChatResponse, error) {
	cs.logger.Infof("Processing chat message: %s (article_id: %d, has_content: %v)", message, articleID, articleContent != "")

	// Build system prompt
	systemPrompt := cs.buildSystemPrompt()

	// Build messages
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": systemPrompt,
		},
	}

	// Add conversation context if provided
	if conversationContext != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "assistant",
			"content": conversationContext,
		})
	}

	// Build user message with optional article context
	userMessage := message
	if articleContent != "" {
		// Truncate content if too long (keep first 4000 chars)
		content := articleContent
		if len(content) > 4000 {
			content = content[:4000] + "..."
		}
		userMessage = fmt.Sprintf("Context - Artikel content:\n\n%s\n\n---\n\nVraag: %s", content, message)
	}

	// Add user message
	messages = append(messages, map[string]interface{}{
		"role":    "user",
		"content": userMessage,
	})

	// Call OpenAI with function calling
	response, functionCall, err := cs.openAIClient.ChatWithFunctions(ctx, messages, ChatFunctions)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	// If no function call, return text response
	if functionCall == nil {
		return &ChatResponse{
			Message: response,
		}, nil
	}

	// Execute function call
	functionResult, err := cs.executeFunctionCall(ctx, functionCall)
	if err != nil {
		cs.logger.WithError(err).Errorf("Failed to execute function: %s", functionCall.Name)
		return &ChatResponse{
			Message: fmt.Sprintf("Sorry, ik kon die informatie niet ophalen: %s", err.Error()),
		}, nil
	}

	// Build final response with function results
	finalResponse, err := cs.buildFinalResponse(ctx, messages, functionCall, functionResult)
	if err != nil {
		return nil, fmt.Errorf("failed to build final response: %w", err)
	}

	return finalResponse, nil
}

// executeFunctionCall executes the requested function
func (cs *ChatService) executeFunctionCall(ctx context.Context, fc *FunctionCall) (interface{}, error) {
	cs.logger.Infof("Executing function: %s with args: %+v", fc.Name, fc.Arguments)

	switch fc.Name {
	case FunctionSearchArticles:
		query, _ := fc.Arguments["query"].(string)
		limit := 10
		if l, ok := fc.Arguments["limit"].(float64); ok {
			limit = int(l)
		}
		return cs.aiService.SearchArticlesForChat(ctx, query, limit)

	case FunctionGetSentimentStats:
		source, _ := fc.Arguments["source"].(string)
		hoursBack := 24
		if h, ok := fc.Arguments["hours_back"].(float64); ok {
			hoursBack = int(h)
		}
		return cs.aiService.GetSentimentStatsForChat(ctx, source, hoursBack)

	case FunctionGetTrendingTopics:
		hoursBack := 24
		if h, ok := fc.Arguments["hours_back"].(float64); ok {
			hoursBack = int(h)
		}
		minArticles := 3
		if m, ok := fc.Arguments["min_articles"].(float64); ok {
			minArticles = int(m)
		}
		return cs.aiService.GetTrendingTopics(ctx, hoursBack, minArticles)

	case FunctionGetArticlesByEntity:
		entityName, _ := fc.Arguments["entity_name"].(string)
		entityType, _ := fc.Arguments["entity_type"].(string)
		limit := 20
		if l, ok := fc.Arguments["limit"].(float64); ok {
			limit = int(l)
		}
		return cs.aiService.GetArticlesByEntity(ctx, entityName, entityType, limit)

	case FunctionGetRecentArticles:
		source, _ := fc.Arguments["source"].(string)
		category, _ := fc.Arguments["category"].(string)
		sentiment, _ := fc.Arguments["sentiment"].(string)
		limit := 10
		if l, ok := fc.Arguments["limit"].(float64); ok {
			limit = int(l)
		}
		return cs.aiService.GetRecentArticlesForChat(ctx, source, category, sentiment, limit)

	default:
		return nil, fmt.Errorf("unknown function: %s", fc.Name)
	}
}

// buildFinalResponse builds the final response with function results
func (cs *ChatService) buildFinalResponse(ctx context.Context, messages []map[string]interface{}, fc *FunctionCall, result interface{}) (*ChatResponse, error) {
	// Add function call to messages
	messages = append(messages, map[string]interface{}{
		"role":    "assistant",
		"content": nil,
		"function_call": map[string]interface{}{
			"name":      fc.Name,
			"arguments": mustMarshalJSON(fc.Arguments),
		},
	})

	// Add function result
	messages = append(messages, map[string]interface{}{
		"role":    "function",
		"name":    fc.Name,
		"content": cs.formatFunctionResult(result),
	})

	// Get final response from OpenAI
	finalMessage, _, err := cs.openAIClient.ChatWithFunctions(ctx, messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get final response: %w", err)
	}

	// Build response
	response := &ChatResponse{
		Message: finalMessage,
	}

	// Add structured data based on function type
	switch fc.Name {
	case FunctionSearchArticles, FunctionGetArticlesByEntity, FunctionGetRecentArticles:
		if articles, ok := result.([]models.Article); ok {
			response.Articles = articles
		}
	case FunctionGetSentimentStats:
		response.Stats = result
	case FunctionGetTrendingTopics:
		response.Stats = result
	}

	return response, nil
}

// formatFunctionResult formats function result for OpenAI
func (cs *ChatService) formatFunctionResult(result interface{}) string {
	switch v := result.(type) {
	case []models.Article:
		if len(v) == 0 {
			return "Geen artikelen gevonden."
		}
		summary := fmt.Sprintf("Gevonden: %d artikelen\n\n", len(v))
		for i, article := range v {
			if i >= 5 {
				summary += fmt.Sprintf("... en nog %d artikelen\n", len(v)-5)
				break
			}

			// Include content preview if available
			contentPreview := ""
			if article.ContentExtracted && article.Content != "" {
				preview := article.Content
				if len(preview) > 200 {
					preview = preview[:200] + "..."
				}
				contentPreview = fmt.Sprintf("\n  Inhoud: %s", preview)
			}

			summary += fmt.Sprintf("- %s (bron: %s, datum: %s)%s\n",
				article.Title,
				article.Source,
				article.Published.Format("2006-01-02"),
				contentPreview,
			)
		}
		return summary

	case *SentimentStats:
		return fmt.Sprintf(`Sentiment Statistieken:
- Totaal artikelen: %d
- Positief: %d (%.1f%%)
- Neutraal: %d (%.1f%%)
- Negatief: %d (%.1f%%)
- Gemiddelde score: %.2f`,
			v.TotalArticles,
			v.PositiveCount,
			float64(v.PositiveCount)/float64(v.TotalArticles)*100,
			v.NeutralCount,
			float64(v.NeutralCount)/float64(v.TotalArticles)*100,
			v.NegativeCount,
			float64(v.NegativeCount)/float64(v.TotalArticles)*100,
			v.AverageSentiment,
		)

	case []TrendingTopic:
		if len(v) == 0 {
			return "Geen trending topics gevonden."
		}
		summary := fmt.Sprintf("Trending Topics (top %d):\n\n", len(v))
		for i, topic := range v {
			if i >= 10 {
				break
			}
			summary += fmt.Sprintf("%d. %s (%d artikelen, sentiment: %.2f)\n",
				i+1,
				topic.Keyword,
				topic.ArticleCount,
				topic.AverageSentiment,
			)
		}
		return summary

	default:
		data, _ := json.Marshal(v)
		return string(data)
	}
}

// buildSystemPrompt creates the system prompt for the AI
func (cs *ChatService) buildSystemPrompt() string {
	return fmt.Sprintf(`Je bent een behulpzame AI assistent voor een Nederlandse nieuwsaggregator.
Je helpt gebruikers met het vinden en analyseren van nieuwsartikelen uit verschillende bronnen.

Huidige datum en tijd: %s

Capabilities:
- Zoeken naar artikelen op basis van keywords of onderwerpen
- Sentiment analyse van artikelen (positief, negatief, neutraal)
- Trending topics identificeren
- Artikelen vinden over specifieke personen, organisaties of locaties
- Recente artikelen ophalen met filters

Beschikbare bronnen: NU.nl, NOS.nl, AD.nl, Telegraaf.nl, Trouw.nl, Volkskrant.nl

Richtlijnen:
- Geef altijd duidelijke, informatieve antwoorden in het Nederlands
- Gebruik de beschikbare functies om actuele data op te halen
- Wees specifiek over bronnen en datums
- Leg sentiment scores uit (positief > 0.2, negatief < -0.2)
- Geef context bij trending topics
- Wees vriendelijk en professioneel

Als je geen relevante data kunt vinden, vertel dit eerlijk en suggereer alternatieven.`, time.Now().Format("2006-01-02 15:04:05"))
}

// mustMarshalJSON marshals to JSON or panics
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
