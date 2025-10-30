package email

import (
	"regexp"
	"strings"

	"github.com/jeffrey/intellinieuws/internal/models"
)

// SpamDetector provides spam detection functionality
type SpamDetector struct {
	spamKeywords []string
	spamPatterns []*regexp.Regexp
}

// NewSpamDetector creates a new spam detector
func NewSpamDetector() *SpamDetector {
	// Common spam keywords
	keywords := []string{
		"viagra", "cialis", "casino", "lottery", "winner",
		"congratulations", "prize", "claim now", "act now",
		"limited time", "urgent", "click here", "buy now",
		"free money", "make money fast", "work from home",
		"weight loss", "lose weight", "diet pills",
	}

	// Spam patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)click\s+here`),
		regexp.MustCompile(`(?i)act\s+now`),
		regexp.MustCompile(`(?i)\$\d+[,\d]*\s*(million|thousand)`),
		regexp.MustCompile(`(?i)100%\s+(free|guaranteed)`),
		regexp.MustCompile(`(?i)no\s+risk`),
		regexp.MustCompile(`(?i)order\s+now`),
	}

	return &SpamDetector{
		spamKeywords: keywords,
		spamPatterns: patterns,
	}
}

// CalculateSpamScore calculates spam score for an email (0.0 to 1.0)
func (d *SpamDetector) CalculateSpamScore(email *models.EmailCreate) float64 {
	score := 0.0

	// Combine subject and body for analysis
	content := strings.ToLower(email.Subject + " " + email.BodyText + " " + email.BodyHTML)

	// Check for spam keywords (each keyword adds 0.1)
	keywordMatches := 0
	for _, keyword := range d.spamKeywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			keywordMatches++
		}
	}
	score += float64(keywordMatches) * 0.1

	// Check for spam patterns (each pattern adds 0.15)
	patternMatches := 0
	for _, pattern := range d.spamPatterns {
		if pattern.MatchString(content) {
			patternMatches++
		}
	}
	score += float64(patternMatches) * 0.15

	// Check for excessive capitalization (adds up to 0.2)
	caps := countCaps(email.Subject)
	if len(email.Subject) > 0 {
		capsRatio := float64(caps) / float64(len(email.Subject))
		if capsRatio > 0.5 {
			score += 0.2
		}
	}

	// Check for excessive punctuation (adds up to 0.1)
	exclamationCount := strings.Count(email.Subject, "!")
	if exclamationCount > 2 {
		score += 0.1
	}

	// Cap score at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// IsSpam determines if email is spam based on threshold
func (d *SpamDetector) IsSpam(email *models.EmailCreate, threshold float64) bool {
	return d.CalculateSpamScore(email) >= threshold
}

// countCaps counts capital letters in a string
func countCaps(s string) int {
	count := 0
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			count++
		}
	}
	return count
}

// GetSpamReason returns a human-readable reason why email is spam
func (d *SpamDetector) GetSpamReason(email *models.EmailCreate) string {
	reasons := []string{}
	content := strings.ToLower(email.Subject + " " + email.BodyText)

	// Check keywords
	matchedKeywords := []string{}
	for _, keyword := range d.spamKeywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			matchedKeywords = append(matchedKeywords, keyword)
			if len(matchedKeywords) >= 3 {
				break
			}
		}
	}
	if len(matchedKeywords) > 0 {
		reasons = append(reasons, "Contains spam keywords: "+strings.Join(matchedKeywords, ", "))
	}

	// Check patterns
	for _, pattern := range d.spamPatterns {
		if pattern.MatchString(content) {
			reasons = append(reasons, "Matches spam pattern: "+pattern.String())
			break
		}
	}

	// Check capitalization
	caps := countCaps(email.Subject)
	if len(email.Subject) > 0 {
		capsRatio := float64(caps) / float64(len(email.Subject))
		if capsRatio > 0.5 {
			reasons = append(reasons, "Excessive capitalization in subject")
		}
	}

	// Check punctuation
	if strings.Count(email.Subject, "!") > 2 {
		reasons = append(reasons, "Excessive exclamation marks")
	}

	if len(reasons) == 0 {
		return "No specific spam indicators found"
	}

	return strings.Join(reasons, "; ")
}
