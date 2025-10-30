-- Migration: Decode HTML entities in existing articles data
-- This migration fixes garbled text by decoding common HTML entities

-- Update title field
UPDATE articles 
SET title = REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
    title,
    '&amp;', '&'),
    '&lt;', '<'),
    '&gt;', '>'),
    '&quot;', '"'),
    '&apos;', ''''),
    '&#34;', '"'),
    '&#39;', ''''),
    '&#8220;', '"'),
    '&#8221;', '"'),
    '&#8216;', ''''),
    '&#8217;', ''''),
    '&#8230;', '…'),
    '&nbsp;', ' '),
    '&ndash;', '–'),
    '&mdash;', '—')
WHERE title LIKE '%&%' OR title LIKE '%&#%';

-- Update summary field
UPDATE articles 
SET summary = REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
    summary,
    '&amp;', '&'),
    '&lt;', '<'),
    '&gt;', '>'),
    '&quot;', '"'),
    '&apos;', ''''),
    '&#34;', '"'),
    '&#39;', ''''),
    '&#8220;', '"'),
    '&#8221;', '"'),
    '&#8216;', ''''),
    '&#8217;', ''''),
    '&#8230;', '…'),
    '&nbsp;', ' '),
    '&ndash;', '–'),
    '&mdash;', '—')
WHERE summary LIKE '%&%' OR summary LIKE '%&#%';

-- Update content field if it exists
UPDATE articles 
SET content = REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(
    content,
    '&amp;', '&'),
    '&lt;', '<'),
    '&gt;', '>'),
    '&quot;', '"'),
    '&apos;', ''''),
    '&#34;', '"'),
    '&#39;', ''''),
    '&#8220;', '"'),
    '&#8221;', '"'),
    '&#8216;', ''''),
    '&#8217;', ''''),
    '&#8230;', '…'),
    '&nbsp;', ' '),
    '&ndash;', '–'),
    '&mdash;', '—')
WHERE content IS NOT NULL AND (content LIKE '%&%' OR content LIKE '%&#%');

-- Log results
DO $$
DECLARE
    title_count INTEGER;
    summary_count INTEGER;
    content_count INTEGER;
    total_updated INTEGER;
BEGIN
    SELECT COUNT(*) INTO title_count FROM articles WHERE title LIKE '%&%' OR title LIKE '%&#%';
    SELECT COUNT(*) INTO summary_count FROM articles WHERE summary LIKE '%&%' OR summary LIKE '%&#%';
    SELECT COUNT(*) INTO content_count FROM articles WHERE content IS NOT NULL AND (content LIKE '%&%' OR content LIKE '%&#%');
    
    total_updated := title_count + summary_count + content_count;
    
    RAISE NOTICE 'HTML entity decoding migration completed';
    RAISE NOTICE 'Titles with entities: %', title_count;
    RAISE NOTICE 'Summaries with entities: %', summary_count;
    RAISE NOTICE 'Contents with entities: %', content_count;
    RAISE NOTICE 'Total fields still with entities: %', total_updated;
END $$;