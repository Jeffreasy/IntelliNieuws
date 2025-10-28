-- Migration 007: Create emails table for Outlook integration
-- This table stores emails from noreply@x.ai and other configured senders

-- Create emails table
CREATE TABLE IF NOT EXISTS emails (
    id SERIAL PRIMARY KEY,
    message_id VARCHAR(255) UNIQUE NOT NULL,  -- Email message ID for deduplication
    sender VARCHAR(255) NOT NULL,              -- Email sender address
    subject TEXT NOT NULL,                     -- Email subject
    body_text TEXT,                            -- Plain text email body
    body_html TEXT,                            -- HTML email body
    received_date TIMESTAMP NOT NULL,          -- Date email was received
    processed BOOLEAN DEFAULT FALSE,           -- Whether email has been processed
    processed_at TIMESTAMP,                    -- When email was processed
    article_id INTEGER,                        -- Link to articles table if converted
    error TEXT,                                -- Any processing errors
    retry_count INTEGER DEFAULT 0,             -- Number of retry attempts
    metadata JSONB,                            -- Additional email metadata (headers, etc.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_article FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE SET NULL
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_emails_sender ON emails(sender);
CREATE INDEX IF NOT EXISTS idx_emails_received_date ON emails(received_date DESC);
CREATE INDEX IF NOT EXISTS idx_emails_processed ON emails(processed) WHERE NOT processed;
CREATE INDEX IF NOT EXISTS idx_emails_message_id ON emails(message_id);
CREATE INDEX IF NOT EXISTS idx_emails_article_id ON emails(article_id) WHERE article_id IS NOT NULL;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_emails_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_emails_updated_at
    BEFORE UPDATE ON emails
    FOR EACH ROW
    EXECUTE FUNCTION update_emails_updated_at();

-- Add comment to table
COMMENT ON TABLE emails IS 'Stores emails from configured senders (e.g., noreply@x.ai) for processing into news articles';
COMMENT ON COLUMN emails.message_id IS 'Unique email message ID for deduplication';
COMMENT ON COLUMN emails.processed IS 'Indicates if email has been processed into an article';
COMMENT ON COLUMN emails.article_id IS 'References the article created from this email';
COMMENT ON COLUMN emails.metadata IS 'Additional email metadata stored as JSON';