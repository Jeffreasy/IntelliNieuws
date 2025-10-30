package email

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/emersion/go-message/mail"
	"github.com/jeffrey/intellinieuws/pkg/logger"
)

// AttachmentHandler handles email attachments
type AttachmentHandler struct {
	storageDir   string
	maxSizeBytes int64
	allowedTypes []string
	logger       *logger.Logger
}

// AttachmentInfo contains attachment metadata
type AttachmentInfo struct {
	Filename    string
	ContentType string
	Size        int64
	Path        string
}

// NewAttachmentHandler creates a new attachment handler
func NewAttachmentHandler(storageDir string, maxSizeMB int, log *logger.Logger) *AttachmentHandler {
	// Default allowed types (documents and images)
	allowedTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
		"text/csv",
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	return &AttachmentHandler{
		storageDir:   storageDir,
		maxSizeBytes: int64(maxSizeMB * 1024 * 1024),
		allowedTypes: allowedTypes,
		logger:       log.WithComponent("attachment-handler"),
	}
}

// ProcessAttachments extracts and saves attachments from an email part
func (h *AttachmentHandler) ProcessAttachments(mr *mail.Reader, emailID int64) ([]AttachmentInfo, error) {
	attachments := make([]AttachmentInfo, 0)

	// Ensure storage directory exists
	emailDir := filepath.Join(h.storageDir, fmt.Sprintf("email_%d", emailID))
	if err := os.MkdirAll(emailDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create attachment directory: %w", err)
	}

	// Read all parts
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			h.logger.WithError(err).Warn("Error reading message part")
			continue
		}

		// Check if this part is an attachment
		filename, isAttachment := h.isAttachment(part)
		if !isAttachment {
			continue
		}

		contentType := part.Header.Get("Content-Type")

		// Check if content type is allowed
		if !h.isAllowedType(contentType) {
			h.logger.Warnf("Skipping attachment with disallowed type: %s", contentType)
			continue
		}

		// Read attachment data
		data, err := io.ReadAll(part.Body)
		if err != nil {
			h.logger.WithError(err).Warnf("Failed to read attachment: %s", filename)
			continue
		}

		// Check size limit
		if int64(len(data)) > h.maxSizeBytes {
			h.logger.Warnf("Skipping attachment %s: exceeds size limit (%d MB)", filename, h.maxSizeBytes/(1024*1024))
			continue
		}

		// Save attachment
		filepath := filepath.Join(emailDir, h.sanitizeFilename(filename))
		if err := os.WriteFile(filepath, data, 0644); err != nil {
			h.logger.WithError(err).Warnf("Failed to save attachment: %s", filename)
			continue
		}

		attachments = append(attachments, AttachmentInfo{
			Filename:    filename,
			ContentType: contentType,
			Size:        int64(len(data)),
			Path:        filepath,
		})

		h.logger.Infof("Saved attachment: %s (%d bytes)", filename, len(data))
	}

	return attachments, nil
}

// isAttachment determines if a message part is an attachment
func (h *AttachmentHandler) isAttachment(part *mail.Part) (string, bool) {
	// Check Content-Disposition header
	disposition := part.Header.Get("Content-Disposition")
	if strings.Contains(disposition, "attachment") {
		// Extract filename from disposition
		if filename := extractFilename(disposition); filename != "" {
			return filename, true
		}
	}

	// Check if Content-Type has a name parameter
	contentType := part.Header.Get("Content-Type")
	if filename := extractFilename(contentType); filename != "" {
		return filename, true
	}

	return "", false
}

// isAllowedType checks if content type is in allowed list
func (h *AttachmentHandler) isAllowedType(contentType string) bool {
	// Extract main type (before semicolon)
	mainType := strings.Split(contentType, ";")[0]
	mainType = strings.TrimSpace(strings.ToLower(mainType))

	for _, allowed := range h.allowedTypes {
		if strings.EqualFold(mainType, allowed) {
			return true
		}
	}

	return false
}

// sanitizeFilename removes dangerous characters from filename
func (h *AttachmentHandler) sanitizeFilename(filename string) string {
	// Remove path separators and other dangerous characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")

	return filename
}

// extractFilename extracts filename from Content-Disposition or Content-Type header
func extractFilename(header string) string {
	// Look for filename parameter
	parts := strings.Split(header, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "filename=") {
			filename := strings.TrimPrefix(part, "filename=")
			filename = strings.TrimPrefix(filename, "\"")
			filename = strings.TrimSuffix(filename, "\"")
			return filename
		}
		if strings.HasPrefix(strings.ToLower(part), "name=") {
			filename := strings.TrimPrefix(part, "name=")
			filename = strings.TrimPrefix(filename, "\"")
			filename = strings.TrimSuffix(filename, "\"")
			return filename
		}
	}

	return ""
}

// GetAttachmentCount returns the number of attachments for an email
func (h *AttachmentHandler) GetAttachmentCount(emailID int64) int {
	emailDir := filepath.Join(h.storageDir, fmt.Sprintf("email_%d", emailID))

	entries, err := os.ReadDir(emailDir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	return count
}

// DeleteAttachments removes all attachments for an email
func (h *AttachmentHandler) DeleteAttachments(emailID int64) error {
	emailDir := filepath.Join(h.storageDir, fmt.Sprintf("email_%d", emailID))
	return os.RemoveAll(emailDir)
}
