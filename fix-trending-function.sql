-- Fix get_trending_topics function to handle type mismatch and filter empty keywords
CREATE OR REPLACE FUNCTION public.get_trending_topics(
    p_hours_back integer DEFAULT 24, 
    p_min_articles integer DEFAULT 3, 
    p_limit integer DEFAULT 20
)
RETURNS TABLE(
    keyword text, 
    article_count bigint, 
    source_count bigint, 
    sources text[], 
    avg_sentiment numeric, 
    avg_relevance numeric, 
    trending_score numeric, 
    latest_mention timestamp with time zone
)
LANGUAGE plpgsql
STABLE
AS $function$
BEGIN
    RETURN QUERY
    SELECT 
        tk.keyword,
        SUM(tk.article_count)::BIGINT AS article_count,
        MAX(tk.source_count)::BIGINT AS source_count,
        ARRAY_AGG(DISTINCT unnest ORDER BY unnest)::TEXT[] AS sources,
        ROUND(AVG(tk.avg_sentiment)::NUMERIC, 3) AS avg_sentiment,
        ROUND(AVG(tk.avg_relevance)::NUMERIC, 3) AS avg_relevance,
        ROUND(AVG(tk.trending_score)::NUMERIC, 2) AS trending_score,
        MAX(tk.latest_article_date)::TIMESTAMP WITH TIME ZONE AS latest_mention
    FROM mv_trending_keywords tk
    CROSS JOIN LATERAL unnest(tk.sources) AS unnest
    WHERE tk.hour_bucket >= CURRENT_TIMESTAMP - (p_hours_back || ' hours')::INTERVAL
      AND tk.keyword IS NOT NULL 
      AND tk.keyword != ''
      AND TRIM(tk.keyword) != ''
    GROUP BY tk.keyword
    HAVING SUM(tk.article_count) >= p_min_articles
    ORDER BY trending_score DESC NULLS LAST, article_count DESC
    LIMIT p_limit;
END;
$function$;