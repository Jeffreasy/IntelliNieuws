# AI Processing voor IntelliNieuws

## Overzicht

Deze documentatie beschrijft hoe AI wordt gebruikt om nieuwsartikelen te verrijken met intelligente analyses.

## Architectuur

### 1. AI Processing Pipeline

```
Nieuws Scraper → Database → AI Processor → Verrijkte Data → API
                                  ↓
                          OpenAI API / Lokale NLP
```

### 2. AI Functionaliteiten

#### A. **Sentiment Analyse**
- Detecteert of een artikel positief, negatief of neutraal is
- Score van -1.0 (zeer negatief) tot +1.0 (zeer positief)
- Gebruikt voor trending analyse en emotionele context

#### B. **Named Entity Recognition (NER)**
- Extraheert personen, organisaties, locaties
- Gebruikt voor gerelateerde artikelen en filtering
- Voorbeelden:
  - Personen: "Mark Rutte", "Donald Trump"
  - Organisaties: "Tweede Kamer", "EU", "Tesla"
  - Locaties: "Amsterdam", "Oekraïne", "Gaza"

#### C. **Slimme Categorisering**
- Automatische categorie toewijzing op basis van inhoud
- Categorieën: Politiek, Economie, Sport, Tech, Gezondheid, etc.
- Confidence score per categorie

#### D. **Automatische Samenvatting**
- Genereert korte samenvattingen (2-3 zinnen)
- Extraheert kernpunten uit het artikel
- Gebruikt voor preview en quick reading

#### E. **Gelijkenis Detectie**
- Vindt gerelateerde artikelen
- Detecteert duplicate content (anders dan URL check)
- Groepering van artikelen over hetzelfde onderwerp

#### F. **Keyword Extractie**
- Intelligente keyword extractie met relevantie scores
- Beter dan simple tags
- Gebruikt voor search en recommendations

#### G. **Trending Topics**
- Detecteert wat trending is
- Clustering van gerelateerde artikelen
- Time-based trending analysis

## Database Schema Uitbreidingen

```sql
-- Nieuwe kolommen in articles tabel
ALTER TABLE articles ADD COLUMN ai_processed BOOLEAN DEFAULT FALSE;
ALTER TABLE articles ADD COLUMN ai_sentiment FLOAT; -- -1.0 tot 1.0
ALTER TABLE articles ADD COLUMN ai_sentiment_label VARCHAR(20); -- positive, negative, neutral
ALTER TABLE articles ADD COLUMN ai_categories JSONB; -- {"category": confidence}
ALTER TABLE articles ADD COLUMN ai_entities JSONB; -- {persons: [], orgs: [], locations: []}
ALTER TABLE articles ADD COLUMN ai_summary TEXT;
ALTER TABLE articles ADD COLUMN ai_keywords JSONB; -- [{"keyword": "...", "score": 0.9}]
ALTER TABLE articles ADD COLUMN ai_processed_at TIMESTAMP;
ALTER TABLE articles ADD COLUMN ai_error TEXT;

-- Index voor queries
CREATE INDEX idx_articles_ai_processed ON articles(ai_processed);
CREATE INDEX idx_articles_ai_sentiment ON articles(ai_sentiment);
CREATE INDEX idx_articles_ai_categories ON articles USING GIN(ai_categories);
CREATE INDEX idx_articles_ai_entities ON articles USING GIN(ai_entities);
```

## API Integratie Opties

### Optie 1: OpenAI API (Aanbevolen)
**Voordelen:**
- Beste kwaliteit
- GPT-4 of GPT-3.5-turbo
- Eenvoudig te integreren
- Ondersteunt alle functionaliteiten

**Nadelen:**
- Kosten per API call (~$0.002-0.03 per artikel)
- Afhankelijk van externe service
- Rate limits

**Kosten inschatting:**
- 1000 artikelen/dag: ~$2-30/dag
- Met caching: aanzienlijk lager

### Optie 2: Lokale NLP Models
**Voordelen:**
- Geen externe kosten
- Volledige controle
- Geen privacy concerns
- Geen rate limits

**Nadelen:**
- Lagere kwaliteit
- Server resources nodig
- Meer implementatiewerk

**Modellen:**
- Sentiment: `cardiffnlp/twitter-roberta-base-sentiment`
- NER: `flair/ner-dutch` of `pdelobelle/robbert-v2-dutch-ner`
- Summarization: `facebook/bart-large-cnn`

### Optie 3: Hybrid Aanpak (Aanbevolen voor productie)
- Gebruik lokale models voor basis taken (sentiment, NER)
- OpenAI voor complexe taken (samenvatting, categorisering)
- Best of both worlds
- Optimale kosten/kwaliteit ratio

## Implementatie Strategie

### Fase 1: Basis Infrastructuur
1. AI service package maken
2. OpenAI client integratie
3. Database schema uitbreiden
4. Basis processing pipeline

### Fase 2: Core Features
1. Sentiment analyse
2. Entity extraction
3. Automatische categorisering
4. Keyword extractie

### Fase 3: Geavanceerde Features
1. Samenvatting generatie
2. Similarity detection
3. Trending analysis
4. Gerelateerde artikelen

### Fase 4: Optimalisatie
1. Batch processing
2. Caching strategy
3. Asynchrone verwerking
4. Cost optimization

## Processing Workflow

```go
1. Artikel wordt gescraped
   ↓
2. Opgeslagen in database (ai_processed = false)
   ↓
3. Background worker pakt onverwerkte artikelen
   ↓
4. AI Processing:
   - Sentiment analysis
   - Entity extraction
   - Categorization
   - Keyword extraction
   - Summary generation (optioneel)
   ↓
5. Update artikel met AI data (ai_processed = true)
   ↓
6. Beschikbaar via API
```

## API Endpoints

### Nieuwe Endpoints
```
GET /api/v1/articles/trending          - Trending artikelen
GET /api/v1/articles/:id/related       - Gerelateerde artikelen
GET /api/v1/articles/by-entity/:entity - Artikelen over persoon/org
GET /api/v1/analytics/sentiment        - Sentiment trends
GET /api/v1/analytics/topics           - Trending topics
```

### Uitgebreide Article Response
```json
{
  "id": 123,
  "title": "Nieuws titel",
  "summary": "Original summary",
  "ai_enrichment": {
    "sentiment": {
      "score": 0.65,
      "label": "positive"
    },
    "categories": {
      "Politics": 0.89,
      "Economy": 0.45
    },
    "entities": {
      "persons": ["Mark Rutte", "Donald Trump"],
      "organizations": ["EU", "VVD"],
      "locations": ["Amsterdam", "Brussel"]
    },
    "keywords": [
      {"word": "verkiezingen", "score": 0.92},
      {"word": "coalitie", "score": 0.87}
    ],
    "ai_summary": "Korte AI-gegenereerde samenvatting..."
  }
}
```

## Configuratie

### Environment Variabelen
```env
# OpenAI
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-3.5-turbo  # of gpt-4
OPENAI_MAX_TOKENS=1000

# AI Processing
AI_ENABLED=true
AI_ASYNC_PROCESSING=true
AI_BATCH_SIZE=10
AI_PROCESS_INTERVAL_MINUTES=5
AI_RETRY_FAILED=true

# Feature Toggles
AI_ENABLE_SENTIMENT=true
AI_ENABLE_ENTITIES=true
AI_ENABLE_CATEGORIES=true
AI_ENABLE_KEYWORDS=true
AI_ENABLE_SUMMARY=true
AI_ENABLE_SIMILARITY=false  # Duur

# Cost Control
AI_MAX_DAILY_COST=10.00  # USD
AI_RATE_LIMIT_PER_MINUTE=60
```

## Kosten Optimalisatie

### Strategieën
1. **Selectieve Processing**
   - Alleen belangrijke bronnen
   - Skip duplicates
   - Prioriteit op basis van recency

2. **Caching**
   - Cache entity results
   - Cache category predictions
   - Redis voor temporary storage

3. **Batch Processing**
   - Verwerk meerdere artikelen tegelijk
   - Gebruik GPT-3.5 in plaats van GPT-4 waar mogelijk

4. **Rate Limiting**
   - Max aantal API calls per minuut
   - Daily budget limits
   - Graceful degradation

## Monitoring & Analytics

### Metrics
- AI processing success rate
- Average processing time
- Cost per article
- API usage
- Sentiment distribution
- Top entities
- Category distribution

### Dashboards
- Real-time sentiment trends
- Trending topics
- Entity network graphs
- Processing queue status

## Code Structuur

```
internal/
  ai/
    service.go           # Main AI service
    openai_client.go     # OpenAI integration
    sentiment.go         # Sentiment analysis
    entities.go          # Entity extraction
    categories.go        # Categorization
    keywords.go          # Keyword extraction
    summary.go           # Summarization
    similarity.go        # Similarity detection
    processor.go         # Background processor
    models.go            # AI data models
    cache.go             # AI caching layer
```

## Gebruik Voorbeelden

### Backend Processing
```go
// Automatische verwerking van nieuwe artikelen
processor := ai.NewProcessor(aiService, articleRepo)
processor.StartBackgroundProcessing(ctx)

// Handmatig een artikel verwerken
enrichment, err := aiService.ProcessArticle(ctx, article)
```

### API Queries
```bash
# Positieve nieuws
GET /api/v1/articles?sentiment=positive&limit=10

# Artikelen over een persoon
GET /api/v1/articles/by-entity/Mark%20Rutte

# Trending topics
GET /api/v1/analytics/trending

# Gerelateerde artikelen
GET /api/v1/articles/123/related
```

## Best Practices

1. **Altijd error handling** - AI services kunnen falen
2. **Rate limiting** - Voorkom budget overschrijding
3. **Caching** - Cache expensive operations
4. **Async processing** - Block de scraper niet
5. **Monitoring** - Track kosten en performance
6. **Graceful degradation** - App werkt zonder AI
7. **Privacy** - Geen persoonlijke data naar externe APIs

## Toekomstige Uitbreidingen

1. **Custom ML Models** - Train eigen modellen op Dutch news
2. **Real-time Processing** - Stream processing met NATS
3. **Advanced Analytics** - Predictive trending, audience targeting
4. **Multi-language** - Support voor Engelse bronnen
5. **Image Analysis** - Analyseer artikel afbeeldingen
6. **Fact Checking** - Integratie met fact-checking APIs
7. **Recommendation Engine** - Personalized news recommendations