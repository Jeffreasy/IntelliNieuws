-- Migration: Reset corrupt content for re-scraping
-- This identifies and resets articles with garbled/corrupt content
-- so they can be re-scraped with the new HTML decoder

-- Reset content_extracted flag for articles with potentially corrupt content
-- Corrupt content typically contains unusual byte sequences or encoding issues
UPDATE articles 
SET 
    content_extracted = false,
    content = NULL
WHERE 
    content_extracted = true
    AND content IS NOT NULL
    AND (
        -- Contains unusual control characters or byte sequences
        content ~ '[^\x20-\x7E\x0A\x0D\t\u00A0-\uFFFF]'
        -- Or very short content (likely extraction failed)
        OR LENGTH(content) < 100
        -- Or content that looks like binary/garbled data
        OR content ~ '[\x00-\x08\x0B-\x0C\x0E-\x1F]'
    );

-- Log results
DO $$
DECLARE
    reset_count INTEGER;
BEGIN
    GET DIAGNOSTICS reset_count = ROW_COUNT;
    
    RAISE NOTICE 'Corrupt content reset migration completed';
    RAISE NOTICE 'Articles marked for re-scraping: %', reset_count;
    RAISE NOTICE 'These articles will be re-scraped with the new HTML decoder';
END $$;