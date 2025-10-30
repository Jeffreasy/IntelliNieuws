package email

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
	"github.com/jeffrey/intellinieuws/internal/models"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// Service handles email operations via IMAP
type Service struct {
	config *Config
	logger *logger.Logger
}

// Config holds email service configuration
type Config struct {
	Host            string
	Port            int
	Username        string
	Password        string
	UseTLS          bool
	AllowedSenders  []string
	PollInterval    time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	MarkAsRead      bool
	DeleteAfterRead bool
	FetchExisting   bool // Fetch existing emails on first run
	MaxDaysBack     int  // How many days back to fetch (default: 30)
}

// NewService creates a new email service
func NewService(config *Config, log *logger.Logger) *Service {
	return &Service{
		config: config,
		logger: log.WithComponent("email"),
	}
}

// FetchNewEmails fetches unread emails from allowed senders
func (s *Service) FetchNewEmails(ctx context.Context) ([]*models.EmailCreate, error) {
	s.logger.Info("Connecting to email server")

	// Connect to IMAP server
	client, err := s.connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Select INBOX
	mailbox, err := client.Select("INBOX", nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}

	s.logger.Infof("Connected to INBOX, %d messages total", mailbox.NumMessages)

	// Build search criteria
	criteria := &imap.SearchCriteria{}

	// If FetchExisting is true, search by sender and date range
	// Otherwise, only fetch unread messages
	if s.config.FetchExisting {
		// Search by date (last N days)
		if s.config.MaxDaysBack > 0 {
			since := time.Now().AddDate(0, 0, -s.config.MaxDaysBack)
			criteria.Since = since
		}
		s.logger.Infof("Fetching all emails from last %d days (including read)", s.config.MaxDaysBack)
	} else {
		// Only unread messages
		criteria.NotFlag = []imap.Flag{imap.FlagSeen}
		s.logger.Info("Fetching unread messages only")
	}

	// Search for messages
	searchData, err := client.Search(criteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if len(searchData.AllSeqNums()) == 0 {
		s.logger.Info("No messages found matching criteria")
		return []*models.EmailCreate{}, nil
	}

	s.logger.Infof("Found %d messages matching criteria", len(searchData.AllSeqNums()))

	// Fetch message details
	emails := make([]*models.EmailCreate, 0)
	seqSet := imap.SeqSetNum(searchData.AllSeqNums()...)

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		BodySection: []*imap.FetchItemBodySection{
			{Specifier: imap.PartSpecifierHeader},
			{Specifier: imap.PartSpecifierText},
		},
		UID: true,
	}

	fetchCmd := client.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		// Get buffered message data
		msgBuffer, err := msg.Collect()
		if err != nil {
			s.logger.WithError(err).Warn("Failed to collect message buffer")
			continue
		}

		// Parse email
		email, err := s.parseMessage(msgBuffer)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to parse message")
			continue
		}

		// Filter by allowed senders
		if !s.isAllowedSender(email.Sender) {
			s.logger.Debugf("Skipping email from non-allowed sender: %s", email.Sender)
			continue
		}

		emails = append(emails, email)

		// Mark as read if configured
		if s.config.MarkAsRead {
			if err := s.markAsRead(client, msg.SeqNum); err != nil {
				s.logger.WithError(err).Warn("Failed to mark message as read")
			}
		}
	}

	if err := fetchCmd.Close(); err != nil {
		s.logger.WithError(err).Warn("Error closing fetch command")
	}

	s.logger.Infof("Fetched %d emails from allowed senders", len(emails))
	return emails, nil
}

// connect establishes connection to IMAP server with retry logic
func (s *Service) connect(ctx context.Context) (*imapclient.Client, error) {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var client *imapclient.Client
	var err error

	// Retry connection with exponential backoff
	for attempt := 0; attempt < s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			s.logger.Infof("Retry attempt %d/%d", attempt+1, s.config.MaxRetries)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(s.config.RetryDelay * time.Duration(attempt)):
			}
		}

		// Connect based on TLS configuration
		options := &imapclient.Options{
			WordDecoder: &mime.WordDecoder{},
		}

		if s.config.UseTLS {
			client, err = imapclient.DialTLS(addr, options)
		} else {
			client, err = imapclient.DialInsecure(addr, options)
		}

		if err == nil {
			break
		}

		s.logger.WithError(err).Warnf("Connection attempt %d failed", attempt+1)
	}

	if err != nil {
		return nil, fmt.Errorf("failed after %d attempts: %w", s.config.MaxRetries, err)
	}

	// Login
	if err := client.Login(s.config.Username, s.config.Password).Wait(); err != nil {
		client.Close()
		return nil, fmt.Errorf("login failed: %w", err)
	}

	s.logger.Info("Successfully authenticated")
	return client, nil
}

// parseMessage parses an IMAP message into EmailCreate
func (s *Service) parseMessage(msgBuffer *imapclient.FetchMessageBuffer) (*models.EmailCreate, error) {
	env := msgBuffer.Envelope
	if env == nil {
		return nil, fmt.Errorf("message envelope is nil")
	}

	// Extract sender
	var sender string
	if len(env.From) > 0 {
		sender = env.From[0].Addr()
	}

	// Get message ID
	messageID := env.MessageID
	if messageID == "" {
		messageID = fmt.Sprintf("generated-%d-%d", msgBuffer.UID, time.Now().Unix())
	}

	// Get subject
	subject := env.Subject

	// Get received date
	receivedDate := env.Date
	if receivedDate.IsZero() {
		receivedDate = time.Now()
	}

	// Parse body sections
	var bodyText, bodyHTML string
	metadata := make(models.EmailMetadata)

	for sectionID, data := range msgBuffer.BodySection {
		if strings.Contains(fmt.Sprintf("%d", sectionID), "TEXT") {
			// FetchBodySectionBuffer is a []byte alias - use type assertion
			dataBytes, ok := interface{}(data).([]byte)
			if !ok {
				s.logger.Warn("Could not convert body section to bytes")
				continue
			}

			// Try to parse as multipart message
			mr, err := mail.CreateReader(bytes.NewReader(dataBytes))
			if err != nil {
				// If not multipart, use as plain text
				bodyText = string(dataBytes)
				continue
			}

			// Read parts
			for {
				part, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					s.logger.WithError(err).Warn("Error reading message part")
					break
				}

				contentType := part.Header.Get("Content-Type")
				body, err := io.ReadAll(part.Body)
				if err != nil {
					s.logger.WithError(err).Warn("Error reading part body")
					continue
				}

				if strings.Contains(contentType, "text/plain") {
					bodyText = string(body)
				} else if strings.Contains(contentType, "text/html") {
					bodyHTML = string(body)
				}
			}
		}
	}

	// Fallback: use subject as body if no body found
	if bodyText == "" && bodyHTML == "" {
		bodyText = subject
	}

	// Store additional metadata
	if len(env.To) > 0 {
		toAddrs := make([]string, len(env.To))
		for i, addr := range env.To {
			toAddrs[i] = addr.Addr()
		}
		metadata["to"] = toAddrs
	}

	if len(env.Cc) > 0 {
		ccAddrs := make([]string, len(env.Cc))
		for i, addr := range env.Cc {
			ccAddrs[i] = addr.Addr()
		}
		metadata["cc"] = ccAddrs
	}

	if len(env.InReplyTo) > 0 {
		metadata["in_reply_to"] = env.InReplyTo
	}

	metadata["uid"] = msgBuffer.UID

	return &models.EmailCreate{
		MessageID:    messageID,
		Sender:       sender,
		Subject:      subject,
		BodyText:     bodyText,
		BodyHTML:     bodyHTML,
		ReceivedDate: receivedDate,
		Metadata:     metadata,
	}, nil
}

// isAllowedSender checks if sender is in the allowed list
func (s *Service) isAllowedSender(sender string) bool {
	if len(s.config.AllowedSenders) == 0 {
		return true // No filter, allow all
	}

	sender = strings.ToLower(strings.TrimSpace(sender))
	for _, allowed := range s.config.AllowedSenders {
		if strings.ToLower(strings.TrimSpace(allowed)) == sender {
			return true
		}
	}

	return false
}

// markAsRead marks a message as read
func (s *Service) markAsRead(client *imapclient.Client, seqNum uint32) error {
	seqSet := imap.SeqSetNum(seqNum)
	storeFlags := imap.StoreFlags{
		Op:     imap.StoreFlagsAdd,
		Flags:  []imap.Flag{imap.FlagSeen},
		Silent: true,
	}

	storeCmd := client.Store(seqSet, &storeFlags, nil)
	return storeCmd.Close()
}

// FetchExistingEmails fetches existing emails from the inbox based on date range
func (s *Service) FetchExistingEmails(ctx context.Context) ([]*models.EmailCreate, error) {
	s.logger.Info("Fetching existing emails from inbox...")

	// Connect to IMAP server
	client, err := s.connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Select INBOX
	mailbox, err := client.Select("INBOX", nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}

	s.logger.Infof("Connected to INBOX, %d messages total", mailbox.NumMessages)

	// Build search criteria for existing emails
	criteria := &imap.SearchCriteria{}

	// Search by date (last N days)
	if s.config.MaxDaysBack > 0 {
		since := time.Now().AddDate(0, 0, -s.config.MaxDaysBack)
		criteria.Since = since
	} else {
		// Default to 30 days if not specified
		since := time.Now().AddDate(0, 0, -30)
		criteria.Since = since
	}

	s.logger.Infof("Fetching existing emails from last %d days", s.config.MaxDaysBack)

	// Search for messages
	searchData, err := client.Search(criteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	if len(searchData.AllSeqNums()) == 0 {
		s.logger.Info("No existing messages found")
		return []*models.EmailCreate{}, nil
	}

	s.logger.Infof("Found %d existing messages", len(searchData.AllSeqNums()))

	// Fetch message details
	emails := make([]*models.EmailCreate, 0)
	seqSet := imap.SeqSetNum(searchData.AllSeqNums()...)

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		BodySection: []*imap.FetchItemBodySection{
			{Specifier: imap.PartSpecifierHeader},
			{Specifier: imap.PartSpecifierText},
		},
		UID: true,
	}

	fetchCmd := client.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		// Get buffered message data
		msgBuffer, err := msg.Collect()
		if err != nil {
			s.logger.WithError(err).Warn("Failed to collect message buffer")
			continue
		}

		// Parse email
		email, err := s.parseMessage(msgBuffer)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to parse message")
			continue
		}

		// Filter by allowed senders
		if !s.isAllowedSender(email.Sender) {
			s.logger.Debugf("Skipping email from non-allowed sender: %s", email.Sender)
			continue
		}

		emails = append(emails, email)

		// Mark as read if configured (but don't delete existing emails)
		if s.config.MarkAsRead {
			if err := s.markAsRead(client, msg.SeqNum); err != nil {
				s.logger.WithError(err).Warn("Failed to mark message as read")
			}
		}
	}

	if err := fetchCmd.Close(); err != nil {
		s.logger.WithError(err).Warn("Error closing fetch command")
	}

	s.logger.Infof("Fetched %d existing emails from allowed senders", len(emails))
	return emails, nil
}

// TestConnection tests the IMAP connection
func (s *Service) TestConnection(ctx context.Context) error {
	client, err := s.connect(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// Try to select INBOX to verify access
	_, err = client.Select("INBOX", nil).Wait()
	if err != nil {
		return fmt.Errorf("failed to access INBOX: %w", err)
	}

	s.logger.Info("Email connection test successful")
	return nil
}
