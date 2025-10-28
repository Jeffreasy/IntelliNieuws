# AI Samenvattingen Inschakelen

## Huidige Situatie

Samenvattingen zijn **uitgeschakeld** om kosten te besparen.

**Wat de AI NU doet per artikel:**
- ✅ Sentiment analyse (positief/negatief/neutraal)
- ✅ Named entities (personen, organisaties, locaties)
- ✅ Categorieën (Politics, Sports, Technology, etc.)
- ✅ Keywords met relevantie scores
- ❌ Samenvattingen (2-3 zinnen in Nederlands)
- ❌ Similarity detection (gelijksoortige artikelen)

## Hoe Samenvattingen Inschakelen

### Stap 1: Update .env bestand

Wijzig in `.env`:
```env
# WAS:
AI_ENABLE_SUMMARY=false

# WORDT:
AI_ENABLE_SUMMARY=true
```

### Stap 2: Herstart backend

```powershell
# Stop huidige backend (Ctrl+C)
# Start opnieuw
.\scripts\start.ps1
```

**Dat is alles!** De AI zal nu automatisch samenvattingen toevoegen.

## Wat je krijgt

De AI maakt voor elk artikel een **2-3 zins samenvatting in Nederlands**:

**Voorbeeld van samenvatting:**
```json
{
  "summary": "Marco Borsato staat terecht voor beschuldigingen van ontucht. Het Openbaar Ministerie eist een gevangenisstraf van vijf maanden. Borsato ontkent de beschuldigingen."
}
```

De samenvatting wordt opgeslagen in de database kolom `ai_summary`.

## Cost Impact

**Kosten inschatting per artikel:**

**ZONDER samenvatting:**
- Input: ~300 tokens (artikel tekst)
- Output: ~150 tokens (sentiment + entities + categories + keywords)
- **Totaal: ~450 tokens = $0.0007 per artikel**

**MET samenvatting:**
- Input: ~300 tokens (artikel tekst)
- Output: ~250 tokens (alles hierboven + samenvatting)
- **Totaal: ~550 tokens = $0.0009 per artikel**

**Extra kosten: +$0.0002 per artikel (~25% duurder)**

Voor 1000 artikelen per dag:
- Zonder: ~$0.70/dag
- Met: ~$0.90/dag
- **Extra: $0.20/dag = $6/maand**

## Andere Opties

### Optie 1: Alleen belangrijke artikelen samenvatten

Wijzig in code om alleen hoog-prioriteit artikelen samen te vatten:
- Artikelen van belangrijke bronnen
- Artikelen met veel keywords
- Trending artikelen

### Optie 2: Gebruik bestaande RSS summary

Veel RSS feeds hebben al een summary/description. Die wordt al opgeslagen in de `summary` kolom. Je hoeft geen AI samenvatting als je die al hebt!

Check in database:
```sql
SELECT title, summary, ai_summary 
FROM articles 
WHERE summary IS NOT NULL 
LIMIT 10;
```

Als `summary` al goede content heeft, heb je misschien geen `ai_summary` nodig!

## API Endpoints

Na inschakelen krijg je samenvattingen in deze endpoints:

```bash
# Get article met AI enrichment
curl http://localhost:8080/api/v1/articles/123

# Response bevat nu:
{
  "id": 123,
  "title": "...",
  "summary": "RSS feed summary",  # Van de bron
  "ai_summary": "AI gegenereerde Nederlandse samenvatting",  # Van OpenAI
  "ai_sentiment": 0.6,
  "ai_categories": {"Politics": 0.9},
  ...
}
```

## Aanbeveling

**Voor nu: LAAT UIT** ❌
- RSS feeds geven al goede summaries
- Bespaart 25% op AI kosten
- Functionaliteit bestaat als je het later wilt

**Schakel IN als:** ✅
- Je RSS summaries niet goed zijn
- Je consistente Nederlandse samenvattingen wilt
- Je budget hebt voor de extra $6/maand

## Test het eerst!

Wil je testen zonder alles in te schakelen?

```bash
# Test samenvatting voor één artikel via API:
curl -X POST http://localhost:8080/api/v1/ai/process/123 \
  -H "X-API-Key: test123geheim" \
  -H "Content-Type: application/json" \
  -d '{"enable_summary": true}'
```

Dit verwerkt artikel 123 met samenvatting, zonder de globale setting te wijzigen.