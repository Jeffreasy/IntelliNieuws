package rss

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/pkg/logger"
	"github.com/jeffrey/intellinieuws/pkg/utils"
	"github.com/mmcdole/gofeed"
)

// Scraper handles RSS feed scraping
type Scraper struct {
	parser        *gofeed.Parser
	robotsChecker *utils.RobotsChecker
	logger        *logger.Logger
	userAgent     string
}

// NewScraper creates a new RSS scraper
func NewScraper(userAgent string, log *logger.Logger) *Scraper {
	parser := gofeed.NewParser()
	parser.UserAgent = userAgent

	return &Scraper{
		parser:        parser,
		robotsChecker: utils.NewRobotsChecker(userAgent),
		logger:        log.WithComponent("rss-scraper"),
		userAgent:     userAgent,
	}
}

// ScrapeFeed scrapes articles from an RSS feed
func (s *Scraper) ScrapeFeed(ctx context.Context, feedURL string, source string) ([]*models.ArticleCreate, error) {
	s.logger.Infof("Scraping RSS feed: %s", feedURL)

	// Check robots.txt
	allowed, err := s.robotsChecker.IsAllowed(feedURL)
	if err != nil {
		s.logger.WithError(err).Warnf("Error checking robots.txt for %s", feedURL)
	}
	if !allowed {
		return nil, fmt.Errorf("robots.txt disallows scraping of %s", feedURL)
	}

	// Parse feed with timeout
	feed, err := s.parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	if feed == nil || len(feed.Items) == 0 {
		s.logger.Warnf("No items found in RSS feed: %s", feedURL)
		return []*models.ArticleCreate{}, nil
	}

	s.logger.Infof("Found %d items in RSS feed", len(feed.Items))

	// Convert feed items to articles
	articles := make([]*models.ArticleCreate, 0, len(feed.Items))
	for _, item := range feed.Items {
		article, err := s.convertFeedItem(item, source)
		if err != nil {
			s.logger.WithError(err).Warnf("Failed to convert feed item: %s", item.Link)
			continue
		}
		articles = append(articles, article)
	}

	s.logger.Infof("Successfully scraped %d articles from %s", len(articles), source)
	return articles, nil
}

// convertFeedItem converts a gofeed.Item to an ArticleCreate model
func (s *Scraper) convertFeedItem(item *gofeed.Item, source string) (*models.ArticleCreate, error) {
	if item.Link == "" {
		return nil, fmt.Errorf("feed item missing link")
	}

	// Parse published date
	var published time.Time
	if item.PublishedParsed != nil {
		published = *item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		published = *item.UpdatedParsed
	} else {
		published = time.Now()
	}

	// Extract summary/description
	summary := item.Description
	if summary == "" && item.Content != "" {
		summary = item.Content
	}
	summary = cleanHTML(summary)

	// Extract image URL
	imageURL := ""
	if item.Image != nil {
		imageURL = item.Image.URL
	} else if item.Enclosures != nil && len(item.Enclosures) > 0 {
		for _, enc := range item.Enclosures {
			if strings.HasPrefix(enc.Type, "image/") {
				imageURL = enc.URL
				break
			}
		}
	}

	// Extract author
	author := ""
	if item.Author != nil {
		author = item.Author.Name
	}

	// Extract categories as keywords
	keywords := make([]string, 0)
	if item.Categories != nil {
		keywords = item.Categories
	}

	// Extract category (first category if available)
	category := ""
	if len(item.Categories) > 0 {
		category = item.Categories[0]
	}

	article := &models.ArticleCreate{
		Title:     cleanText(item.Title),
		Summary:   truncateText(summary, 2000),
		URL:       item.Link,
		Published: published,
		Source:    source,
		Keywords:  keywords,
		ImageURL:  imageURL,
		Author:    author,
		Category:  category,
	}

	return article, nil
}

// cleanHTML removes HTML tags from text
func cleanHTML(text string) string {
	// Simple HTML tag removal - for production use a proper HTML parser
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<br />", "\n")
	text = strings.ReplaceAll(text, "<p>", "\n")
	text = strings.ReplaceAll(text, "</p>", "\n")

	// Remove all other HTML tags
	inTag := false
	result := strings.Builder{}
	for _, char := range text {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	return cleanText(result.String())
}

// cleanText cleans and normalizes text
func cleanText(text string) string {
	// Trim whitespace
	text = strings.TrimSpace(text)

	// Replace multiple spaces with single space
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	// Replace multiple newlines with double newline
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}

	return text
}

// truncateText truncates text to a maximum length
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

// ScrapeMultipleFeeds scrapes multiple RSS feeds concurrently
func (s *Scraper) ScrapeMultipleFeeds(ctx context.Context, feeds map[string]string) (map[string][]*models.ArticleCreate, map[string]error) {
	results := make(map[string][]*models.ArticleCreate)
	errors := make(map[string]error)

	// Create channels for results
	type result struct {
		source   string
		articles []*models.ArticleCreate
		err      error
	}
	resultChan := make(chan result, len(feeds))

	// Scrape feeds concurrently
	for source, feedURL := range feeds {
		go func(src, url string) {
			articles, err := s.ScrapeFeed(ctx, url, src)
			resultChan <- result{
				source:   src,
				articles: articles,
				err:      err,
			}
		}(source, feedURL)
	}

	// Collect results
	for i := 0; i < len(feeds); i++ {
		res := <-resultChan
		if res.err != nil {
			errors[res.source] = res.err
			s.logger.WithError(res.err).Errorf("Failed to scrape feed for source: %s", res.source)
		} else {
			results[res.source] = res.articles
		}
	}

	close(resultChan)
	return results, errors
}
