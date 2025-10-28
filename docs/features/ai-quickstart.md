# AI Processing Quick Start Guide

Deze gids helpt je om snel aan de slag te gaan met AI-verwerking in de NieuwsScraper.

## 1. Setup OpenAI API

### Stap 1: OpenAI API Key verkrijgen
1. Ga naar [OpenAI Platform](https://platform.openai.com/)
2. Maak een account aan of log in
3. Ga naar API Keys sectie
4. Genereer een nieuwe API key
5. Kopieer de key (je kunt hem maar één keer zien!)

### Stap 2: Configuratie instellen

Bewerk je `.env` bestand:

```env
# AI Processing inschakelen
AI_ENABLED=true
OPENAI_API_KEY=sk-your-api-key-here

# Model selectie (kies één)
OPENAI_MODEL=gpt-3.5-turbo    # Goedkoper, sneller
# OPENAI_MODEL=gpt-4           # Beter, duurder

# Features inschakelen
AI_ENABLE_SENTIMENT=true
AI_ENABLE_ENTITIES=true
AI_ENABLE_CATEGORIES=true
AI_ENABLE_KEYWORDS=true
AI_ENABLE_SUMMARY=false       # Optioneel, kost meer

# Async processing
AI_ASYNC_PROCESSING=true
AI_BATCH_SIZE=10
AI_PROCESS_INTERVAL_MINUTES=5

# Cost control
AI_MAX_DAILY_COST=10.0
AI_RATE_LIMIT_PER_MINUTE=60
```

## 2. Database Migratie

Voer de AI database migratie uit:

```bash
# PostgreSQL
psql -U scraper -d nieuws_scraper -f migrations/003_add_ai_columns.sql

# Of via script
.\scripts\apply-migrations.ps1
```

## 3. Server Starten

```bash
# Start de API server
go run cmd/api/main.go
```

Je zou dit moeten zien in de logs:
```
INFO Initializing AI processing service
INFO AI processor started with interval: 5m0s
INFO AI service initialized successfully
```

## 4. AI Features Testen

### Automatische Verwerking

Artikelen worden automatisch verwerkt door de background processor:
- Elke 5 minuten (configureerbaar)
- Batch van 10 artikelen (configureerbaar)
- Onverwerkte artikelen worden eerst gepikt

### Handmatige Verwerking

#### Enkel Artikel Verwerken
```bash
# POST /api/v1/articles/:id/process
curl -X POST "http://localhost:8080/api/v1/articles/123/process" \
  -H "X-API-Key: your-api-key"
```

#### Batch Processing Triggeren
```bash
# POST /api/v1/ai/process/trigger
curl -X POST "http://localhost:8080/api/v1/ai/process/trigger" \
  -H "X-API-Key: your-api-key"
```

### AI Data Ophalen

#### Enrichment voor Specifiek Artikel
```bash
# GET /api/v1/articles/:id/enrichment
curl "http://localhost:8080/api/v1/articles/123/enrichment"
```

Response:
```json
{
  "status": "success",
  "data": {
    "processed": true,
    "processed_at": "2025-01-28T14:30:00Z",
    "sentiment": {
      "score": 0.65,
      "label": "positive",
      "confidence": 0.89
    },
    "categories": {
      "Politics": 0.92,
      "Economy": 0.45
    },
    "entities": {
      "persons": ["Mark Rutte", "Geert Wilders"],
      "organizations": ["VVD", "PVV", "Tweede Kamer"],
      "locations": ["Den Haag", "Nederland"]
    },
    "keywords": [
      {"word": "verkiezingen", "score": 0.95},
      {"word": "coalitie", "score": 0.88},
      {"word": "formatie", "score": 0.82}
    ],
    "summary": "Korte AI-gegenereerde samenvatting van het artikel..."
  }
}
```

#### Sentiment Statistieken
```bash
# GET /api/v1/ai/sentiment/stats
curl "http://localhost:8080/api/v1/ai/sentiment/stats?source=nu.nl"
```

Response:
```json
{
  "status": "success",
  "data": {
    "total_articles": 150,
    "positive_count": 45,
    "neutral_count": 80,
    "negative_count": 25,
    "average_sentiment": 0.15,
    "most_positive_title": "Economie groeit sneller dan verwacht",
    "most_negative_title": "Grote zorgen over klimaatverandering"
  }
}
```

#### Trending Topics
```bash
# GET /api/v1/ai/trending?hours=24&min_articles=3
curl "http://localhost:8080/api/v1/ai/trending?hours=24&min_articles=3"
```

Response:
```json
{
  "status": "success",
  "data": {
    "topics": [
      {
        "keyword": "verkiezingen",
        "article_count": 15,
        "average_sentiment": 0.2,
        "sources": ["nu.nl", "ad.nl", "nos.nl"]
      },
      {
        "keyword": "klimaat",
        "article_count": 12,
        "average_sentiment": -0.3,
        "sources": ["nu.nl", "nos.nl"]
      }
    ],
    "hours_back": 24,
    "min_articles": 3,
    "count": 2
  }
}
```

#### Artikelen per Entiteit
```bash
# GET /api/v1/ai/entity/:name
curl "http://localhost:8080/api/v1/ai/entity/Mark%20Rutte?type=persons&limit=10"
```

#### Processor Status
```bash
# GET /api/v1/ai/processor/stats
curl "http://localhost:8080/api/v1/ai/processor/stats"
```

Response:
```json
{
  "status": "success",
  "data": {
    "is_running": true,
    "process_count": 145,
    "last_run": "2025-01-28T14:35:00Z"
  }
}
```

## 5. Kosten Monitoring

### Geschatte Kosten

**GPT-3.5-turbo** (aanbevolen):
- ~$0.002 per artikel (met alle features)
- 1000 artikelen/dag = ~$2/dag = ~$60/maand

**GPT-4**:
- ~$0.03 per artikel
- 1000 artikelen/dag = ~$30/dag = ~$900/maand

### Cost Control Features

1. **Daily Budget Limit**
   ```env
   AI_MAX_DAILY_COST=10.0  # Stop processing bij $10/dag
   ```

2. **Rate Limiting**
   ```env
   AI_RATE_LIMIT_PER_MINUTE=60  # Max 60 API calls/minuut
   ```

3. **Feature Toggles**
   ```env
   AI_ENABLE_SUMMARY=false  # Duurste feature uitschakelen
   ```

4. **Batch Size**
   ```env
   AI_BATCH_SIZE=10  # Kleinere batches = minder kosten per run
   ```

## 6. Best Practices

### Development
- Start met `AI_ENABLED=false` tijdens development
- Gebruik `gpt-3.5-turbo` voor testen
- Enable alleen de features die je nodig hebt
- Test eerst met kleine batches

### Production
- Monitor kosten dagelijks
- Set realistic `AI_MAX_DAILY_COST`
- Enable `AI_RETRY_FAILED=true`
- Gebruik caching waar mogelijk
- Schedule processing tijdens off-peak uren

### Performance
```env
# Optimale configuratie voor 1000 artikelen/dag
AI_BATCH_SIZE=20
AI_PROCESS_INTERVAL_MINUTES=15
AI_RATE_LIMIT_PER_MINUTE=60
OPENAI_MODEL=gpt-3.5-turbo
```

## 7. Troubleshooting

### "AI processor not started"
- Check `AI_ENABLED=true`
- Check `AI_ASYNC_PROCESSING=true`
- Check `OPENAI_API_KEY` is set

### "Failed to process article"
- Check OpenAI API key validity
- Check API rate limits niet overschreden
- Check daily cost limit niet bereikt
- Check logs voor specifieke errors

### "No pending articles to process"
- Alle artikelen zijn al verwerkt
- Run scraper eerst: `POST /api/v1/scrape`
- Of wacht op scheduled scrape

### Hoge Kosten
```env
# Verlaag kosten
AI_ENABLE_SUMMARY=false          # -60% kosten
OPENAI_MODEL=gpt-3.5-turbo      # -95% vs GPT-4
AI_BATCH_SIZE=5                  # Minder artikelen per run
AI_PROCESS_INTERVAL_MINUTES=60   # Minder frequent
```

## 8. Monitoring

### Logs Checken
```bash
# Kijk naar AI processing logs
grep "AI" api.log
grep "OpenAI" api.log
grep "processor" api.log
```

### Processor Status in Real-time
```bash
# Watch processor stats
watch -n 5 'curl -s http://localhost:8080/api/v1/ai/processor/stats | jq'
```

### Database Queries
```sql
-- Hoeveel artikelen zijn verwerkt?
SELECT COUNT(*) FROM articles WHERE ai_processed = TRUE;

-- Sentiment verdeling
SELECT ai_sentiment_label, COUNT(*) 
FROM articles 
WHERE ai_processed = TRUE 
GROUP BY ai_sentiment_label;

-- Meest voorkomende entities
SELECT jsonb_array_elements_text(ai_entities->'persons') as person, COUNT(*)
FROM articles
WHERE ai_entities IS NOT NULL
GROUP BY person
ORDER BY count DESC
LIMIT 10;

-- Trending keywords (laatste 24 uur)
SELECT * FROM get_trending_topics(24, 3);
```

## 9. Volgende Stappen

1. **Experimenteer met Features**
   - Enable/disable verschillende features
   - Test sentiment accuracy
   - Check entity extraction quality

2. **Optimaliseer Configuratie**
   - Monitor kosten vs kwaliteit
   - Tune batch sizes
   - Adjust processing intervals

3. **Bouw Applicaties**
   - Sentiment dashboard
   - Entity network grafen
   - Trending topics feed
   - Personalized recommendations

4. **Advanced Features** (toekomst)
   - Custom ML models training
   - Real-time processing met NATS
   - Advanced similarity detection
   - Fact-checking integratie

## Hulp Nodig?

- Check [AI_PROCESSING.md](./AI_PROCESSING.md) voor technische details
- Check logs voor error messages
- Test API endpoints met Postman/curl
- Monitor OpenAI dashboard voor usage/costs