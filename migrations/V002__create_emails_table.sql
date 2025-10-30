-- ============================================================================
-- Migration: V002__create_emails_table.sql
-- Description: Email integration for newsletter and automated content ingestion
-- Version: 1.0.0
-- Author: NieuwsScraper Team
-- Date: 2025-10-30
-- Dependencies: V001__create_base_schema.sql
-- ============================================================================

-- ============================================================================
-- EMAILS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS emails (
    id BIGSERIAL PRIMARY KEY,
    
    -- Email identification
    message_id VARCHAR(255) NOT NULL UNIQUE,
    message_uid VARCHAR(100), -- IMAP UID for tracking
    thread_id VARCHAR(255), -- Email thread identifier
    
    -- Email metadata
    sender VARCHAR(255) NOT NULL,
    sender_name VARCHAR(200),
    recipient VARCHAR(255),
    subject TEXT NOT NULL,
    
    -- Email content
    body_text TEXT,
    body_html TEXT,
    snippet TEXT, -- First 200 characters for preview
    
    -- Timestamps
    received_date TIMESTAMPTZ NOT NULL,
    sent_date TIMESTAMPTZ,
    
    -- Processing status
    status VARCHAR(20) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'processing', 'processed', 'failed', 'ignored', 'spam')),
    processed_at TIMESTAMPTZ,
    
    -- Article linkage
    article_id BIGINT,
    article_created BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Error handling
    error TEXT,
    error_code VARCHAR(50),
    retry_count INTEGER NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    max_retries INTEGER NOT NULL DEFAULT 3 CHECK (max_retries >= 0),
    last_retry_at TIMESTAMPTZ,
    
    -- Email properties
    has_attachments BOOLEAN NOT NULL DEFAULT FALSE,
    attachment_count INTEGER DEFAULT 0 CHECK (attachment_count >= 0),
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    is_flagged BOOLEAN NOT NULL DEFAULT FALSE,
    is_spam BOOLEAN NOT NULL DEFAULT FALSE,
    importance VARCHAR(20) CHECK (importance IN ('low', 'normal', 'high')),
    
    -- Additional metadata (headers, labels, etc.)
    metadata JSONB,
    headers JSONB,
    labels TEXT[],
    
    -- Email size and quality
    size_bytes INTEGER CHECK (size_bytes >= 0),
    spam_score DECIMAL(5,2) CHECK (spam_score BETWEEN 0 AND 100),
    
    -- Audit fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100) DEFAULT 'email_processor',
    
    -- Foreign key to articles
    CONSTRAINT fk_emails_article 
        FOREIGN KEY (article_id) 
        REFERENCES articles(id) 
        ON DELETE SET NULL 
        ON UPDATE CASCADE,
    
    -- Business logic constraints
    CONSTRAINT chk_emails_article_status CHECK (
        (article_created = TRUE AND article_id IS NOT NULL AND status = 'processed') OR
        (article_created = FALSE)
    ),
    CONSTRAINT chk_emails_retry_logic CHECK (
        retry_count <= max_retries
    )
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Primary lookup indexes
CREATE INDEX idx_emails_message_id ON emails(message_id);
CREATE INDEX idx_emails_sender ON emails(sender);
CREATE INDEX idx_emails_subject ON emails USING gin(to_tsvector('english', subject));

-- Status and processing indexes
CREATE INDEX idx_emails_status ON emails(status);
CREATE INDEX idx_emails_pending ON emails(received_date DESC) 
    WHERE status = 'pending';
CREATE INDEX idx_emails_processing ON emails(received_date DESC) 
    WHERE status = 'processing';
CREATE INDEX idx_emails_failed ON emails(retry_count, received_date DESC) 
    WHERE status = 'failed' AND retry_count < max_retries;

-- Timestamp indexes
CREATE INDEX idx_emails_received_date_desc ON emails(received_date DESC);
CREATE INDEX idx_emails_processed_at ON emails(processed_at DESC) 
    WHERE processed_at IS NOT NULL;

-- Article linkage
CREATE INDEX idx_emails_article_id ON emails(article_id) 
    WHERE article_id IS NOT NULL;
CREATE INDEX idx_emails_article_created ON emails(article_created, received_date DESC);

-- Sender analysis
CREATE INDEX idx_emails_sender_status ON emails(sender, status, received_date DESC);

-- Spam detection
CREATE INDEX idx_emails_spam ON emails(is_spam, received_date DESC) 
    WHERE is_spam = TRUE;

-- JSONB indexes
CREATE INDEX idx_emails_metadata_gin ON emails USING gin(metadata) 
    WHERE metadata IS NOT NULL;
CREATE INDEX idx_emails_headers_gin ON emails USING gin(headers) 
    WHERE headers IS NOT NULL;

-- Array index for labels
CREATE INDEX idx_emails_labels_gin ON emails USING gin(labels) 
    WHERE labels IS NOT NULL;

-- Full-text search across all text fields
CREATE INDEX idx_emails_fulltext_search ON emails USING gin(
    to_tsvector('english', 
        subject || ' ' || 
        COALESCE(sender_name, '') || ' ' || 
        COALESCE(body_text, '') || ' ' ||
        COALESCE(snippet, '')
    )
);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Auto-update updated_at timestamp
CREATE TRIGGER trg_emails_updated_at
    BEFORE UPDATE ON emails
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- Trigger to update snippet when body_text changes
CREATE OR REPLACE FUNCTION trigger_update_email_snippet()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.body_text IS NOT NULL AND NEW.body_text != '' THEN
        NEW.snippet = LEFT(NEW.body_text, 200);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_emails_snippet
    BEFORE INSERT OR UPDATE OF body_text ON emails
    FOR EACH ROW
    EXECUTE FUNCTION trigger_update_email_snippet();

-- Trigger to validate article linkage
CREATE OR REPLACE FUNCTION trigger_validate_email_article()
RETURNS TRIGGER AS $$
BEGIN
    -- If article_created is TRUE, ensure article_id is set and exists
    IF NEW.article_created = TRUE THEN
        IF NEW.article_id IS NULL THEN
            RAISE EXCEPTION 'article_id cannot be NULL when article_created is TRUE';
        END IF;
        
        -- Verify article exists
        IF NOT EXISTS (SELECT 1 FROM articles WHERE id = NEW.article_id) THEN
            RAISE EXCEPTION 'Referenced article_id % does not exist', NEW.article_id;
        END IF;
        
        -- Ensure status is 'processed'
        IF NEW.status != 'processed' THEN
            NEW.status = 'processed';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_emails_article_validation
    BEFORE INSERT OR UPDATE ON emails
    FOR EACH ROW
    EXECUTE FUNCTION trigger_validate_email_article();

-- ============================================================================
-- VIEWS
-- ============================================================================

-- View: Pending emails for processing
CREATE OR REPLACE VIEW v_emails_pending_processing AS
SELECT 
    e.id,
    e.message_id,
    e.sender,
    e.sender_name,
    e.subject,
    e.snippet,
    e.received_date,
    e.retry_count,
    e.max_retries,
    e.last_retry_at,
    e.has_attachments,
    e.size_bytes
FROM emails e
WHERE e.status = 'pending'
   OR (e.status = 'failed' AND e.retry_count < e.max_retries)
ORDER BY 
    CASE WHEN e.status = 'pending' THEN 0 ELSE 1 END,
    e.received_date DESC;

COMMENT ON VIEW v_emails_pending_processing IS 'Emails ready for processing, prioritized by status and date';

-- View: Email processing statistics
CREATE OR REPLACE VIEW v_email_stats AS
SELECT 
    COUNT(*) AS total_emails,
    COUNT(*) FILTER (WHERE status = 'pending') AS pending_count,
    COUNT(*) FILTER (WHERE status = 'processing') AS processing_count,
    COUNT(*) FILTER (WHERE status = 'processed') AS processed_count,
    COUNT(*) FILTER (WHERE status = 'failed') AS failed_count,
    COUNT(*) FILTER (WHERE status = 'ignored') AS ignored_count,
    COUNT(*) FILTER (WHERE status = 'spam') AS spam_count,
    COUNT(*) FILTER (WHERE article_created = TRUE) AS articles_created,
    COUNT(*) FILTER (WHERE received_date >= CURRENT_DATE - INTERVAL '24 hours') AS emails_today,
    COUNT(*) FILTER (WHERE received_date >= CURRENT_DATE - INTERVAL '7 days') AS emails_week,
    AVG(size_bytes) FILTER (WHERE size_bytes IS NOT NULL) AS avg_size_bytes,
    AVG(retry_count) FILTER (WHERE status = 'failed') AS avg_retry_count
FROM emails;

COMMENT ON VIEW v_email_stats IS 'Overall email processing statistics';

-- View: Email sender statistics
CREATE OR REPLACE VIEW v_email_sender_stats AS
SELECT 
    e.sender,
    MAX(e.sender_name) AS sender_name,
    COUNT(*) AS total_emails,
    COUNT(*) FILTER (WHERE e.article_created = TRUE) AS articles_created,
    COUNT(*) FILTER (WHERE e.status = 'processed') AS processed_count,
    COUNT(*) FILTER (WHERE e.status = 'failed') AS failed_count,
    COUNT(*) FILTER (WHERE e.is_spam = TRUE) AS spam_count,
    MAX(e.received_date) AS last_email_date,
    MIN(e.received_date) AS first_email_date,
    AVG(e.size_bytes) FILTER (WHERE e.size_bytes IS NOT NULL) AS avg_size_bytes
FROM emails e
GROUP BY e.sender
ORDER BY total_emails DESC;

COMMENT ON VIEW v_email_sender_stats IS 'Statistics grouped by email sender';

-- View: Recent email activity
CREATE OR REPLACE VIEW v_recent_email_activity AS
SELECT 
    e.id,
    e.message_id,
    e.sender,
    e.sender_name,
    e.subject,
    e.snippet,
    e.received_date,
    e.status,
    e.article_created,
    e.article_id,
    a.title AS article_title,
    a.url AS article_url,
    e.processed_at,
    e.error
FROM emails e
LEFT JOIN articles a ON e.article_id = a.id
ORDER BY e.received_date DESC
LIMIT 100;

COMMENT ON VIEW v_recent_email_activity IS 'Last 100 emails with processing status';

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Function: Get emails ready for retry
CREATE OR REPLACE FUNCTION get_emails_for_retry(
    max_age_hours INTEGER DEFAULT 24,
    batch_size INTEGER DEFAULT 50
)
RETURNS TABLE (
    email_id BIGINT,
    message_id VARCHAR,
    sender VARCHAR,
    subject TEXT,
    retry_count INTEGER,
    last_retry_at TIMESTAMPTZ,
    error TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        e.id,
        e.message_id,
        e.sender,
        e.subject,
        e.retry_count,
        e.last_retry_at,
        e.error
    FROM emails e
    WHERE e.status = 'failed'
      AND e.retry_count < e.max_retries
      AND (
          e.last_retry_at IS NULL 
          OR e.last_retry_at < CURRENT_TIMESTAMP - (max_age_hours || ' hours')::INTERVAL
      )
    ORDER BY e.retry_count ASC, e.received_date DESC
    LIMIT batch_size;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_emails_for_retry IS 'Returns failed emails eligible for retry';

-- Function: Mark email as processed and link to article
CREATE OR REPLACE FUNCTION mark_email_processed(
    p_email_id BIGINT,
    p_article_id BIGINT
)
RETURNS BOOLEAN AS $$
DECLARE
    v_updated BOOLEAN;
BEGIN
    UPDATE emails
    SET 
        status = 'processed',
        processed_at = CURRENT_TIMESTAMP,
        article_id = p_article_id,
        article_created = TRUE,
        error = NULL,
        error_code = NULL
    WHERE id = p_email_id
      AND status IN ('pending', 'processing', 'failed');
    
    GET DIAGNOSTICS v_updated = FOUND;
    RETURN v_updated;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION mark_email_processed IS 'Marks an email as successfully processed';

-- Function: Mark email as failed
CREATE OR REPLACE FUNCTION mark_email_failed(
    p_email_id BIGINT,
    p_error TEXT,
    p_error_code VARCHAR DEFAULT NULL
)
RETURNS BOOLEAN AS $$
DECLARE
    v_updated BOOLEAN;
BEGIN
    UPDATE emails
    SET 
        status = 'failed',
        error = p_error,
        error_code = p_error_code,
        retry_count = retry_count + 1,
        last_retry_at = CURRENT_TIMESTAMP
    WHERE id = p_email_id;
    
    GET DIAGNOSTICS v_updated = FOUND;
    RETURN v_updated;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION mark_email_failed IS 'Marks an email as failed with error details';

-- Function: Clean up old processed emails
CREATE OR REPLACE FUNCTION cleanup_old_emails(
    days_to_keep INTEGER DEFAULT 90,
    keep_with_articles BOOLEAN DEFAULT TRUE
)
RETURNS TABLE (
    emails_deleted BIGINT,
    space_freed_bytes BIGINT
) AS $$
DECLARE
    v_deleted BIGINT;
    v_space_freed BIGINT;
BEGIN
    -- Calculate space to be freed
    SELECT COALESCE(SUM(size_bytes), 0)
    INTO v_space_freed
    FROM emails
    WHERE received_date < CURRENT_DATE - (days_to_keep || ' days')::INTERVAL
      AND status IN ('processed', 'ignored', 'spam')
      AND (NOT keep_with_articles OR article_id IS NULL);
    
    -- Delete old emails
    WITH deleted AS (
        DELETE FROM emails
        WHERE received_date < CURRENT_DATE - (days_to_keep || ' days')::INTERVAL
          AND status IN ('processed', 'ignored', 'spam')
          AND (NOT keep_with_articles OR article_id IS NULL)
        RETURNING id
    )
    SELECT COUNT(*) INTO v_deleted FROM deleted;
    
    RETURN QUERY SELECT v_deleted, v_space_freed;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_emails IS 'Removes old processed emails to free up space';

-- ============================================================================
-- PERMISSIONS
-- ============================================================================

GRANT SELECT ON v_emails_pending_processing TO PUBLIC;
GRANT SELECT ON v_email_stats TO PUBLIC;
GRANT SELECT ON v_email_sender_stats TO PUBLIC;
GRANT SELECT ON v_recent_email_activity TO PUBLIC;

-- ============================================================================
-- FINALIZE MIGRATION
-- ============================================================================

-- Update statistics
ANALYZE emails;

-- Record migration
INSERT INTO schema_migrations (version, description, checksum) 
VALUES (
    'V002',
    'Create emails table with enterprise features',
    'emails_v1'
) ON CONFLICT (version) DO NOTHING;

-- Success notification
DO $$ 
DECLARE
    v_index_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_index_count
    FROM pg_indexes 
    WHERE tablename = 'emails';
    
    RAISE NOTICE '✅ Migration V002 completed successfully';
    RAISE NOTICE 'Created table: emails';
    RAISE NOTICE 'Created % indexes on emails table', v_index_count;
    RAISE NOTICE 'Created 4 views for email monitoring';
    RAISE NOTICE 'Created 4 helper functions';
    RAISE NOTICE 'Added 3 triggers for data integrity';
END $$;