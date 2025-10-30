-- ============================================================================
-- Rollback Script: V002__create_emails_table.sql
-- Description: Rollback emails table creation
-- Version: 1.0.0
-- Author: NieuwsScraper Team
-- Date: 2025-10-30
-- WARNING: This will delete all email data
-- ============================================================================

-- Confirm before execution
DO $$
BEGIN
    RAISE NOTICE '⚠️  WARNING: This rollback will DROP the emails table and DELETE ALL DATA!';
    RAISE NOTICE 'Press Ctrl+C within 5 seconds to cancel...';
    PERFORM pg_sleep(5);
END $$;

-- ============================================================================
-- DROP VIEWS
-- ============================================================================

DROP VIEW IF EXISTS v_recent_email_activity CASCADE;
DROP VIEW IF EXISTS v_email_sender_stats CASCADE;
DROP VIEW IF EXISTS v_email_stats CASCADE;
DROP VIEW IF EXISTS v_emails_pending_processing CASCADE;

RAISE NOTICE 'Dropped 4 views';

-- ============================================================================
-- DROP FUNCTIONS
-- ============================================================================

DROP FUNCTION IF EXISTS cleanup_old_emails(INTEGER, BOOLEAN) CASCADE;
DROP FUNCTION IF EXISTS mark_email_failed(BIGINT, TEXT, VARCHAR) CASCADE;
DROP FUNCTION IF EXISTS mark_email_processed(BIGINT, BIGINT) CASCADE;
DROP FUNCTION IF EXISTS get_emails_for_retry(INTEGER, INTEGER) CASCADE;

RAISE NOTICE 'Dropped 4 functions';

-- ============================================================================
-- DROP TRIGGERS
-- ============================================================================

DROP TRIGGER IF EXISTS trg_emails_article_validation ON emails;
DROP TRIGGER IF EXISTS trg_emails_snippet ON emails;
DROP TRIGGER IF EXISTS trg_emails_updated_at ON emails;

RAISE NOTICE 'Dropped 3 triggers';

DROP FUNCTION IF EXISTS trigger_validate_email_article() CASCADE;
DROP FUNCTION IF EXISTS trigger_update_email_snippet() CASCADE;

RAISE NOTICE 'Dropped 2 trigger functions';

-- ============================================================================
-- DROP TABLES
-- ============================================================================

DROP TABLE IF EXISTS emails CASCADE;

RAISE NOTICE 'Dropped table: emails';

-- ============================================================================
-- REMOVE MIGRATION RECORD
-- ============================================================================

DELETE FROM schema_migrations WHERE version = 'V002';
RAISE NOTICE 'Removed migration record: V002';

-- ============================================================================
-- FINALIZE ROLLBACK
-- ============================================================================

DO $$ 
BEGIN 
    RAISE NOTICE '✅ Rollback V002 completed successfully';
    RAISE NOTICE 'Emails table and all related objects have been removed';
    RAISE NOTICE 'Database is now in post-V001 state';
END $$;