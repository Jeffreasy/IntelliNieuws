# Content Field Update voor Chat API

## ✅ Wat is Gefixt

De chat API haalt nu het volledige `content` veld op van artikelen, zodat de AI toegang heeft tot de complete artikel tekst.

## 📝 Wijzigingen

### 1. SearchArticlesForChat (service.go:706-755)
**Voor:**
- Haalde alleen title, summary op
- Zocht alleen in title, summary, keywords

**Na:**
- Haalt ook `content`, `content_extracted`, `content_extracted_at` op
- Zoekt ook in het `content` veld: `WHERE (title ILIKE $1 OR summary ILIKE $1 OR keywords ILIKE $1 OR content ILIKE $1)`

### 2. GetRecentArticlesForChat (service.go:758-824)
**Voor:**
- Haalde alleen basis artikel velden op

**Na:**
- Haalt ook `content`, `content_extracted`, `content_extracted_at` op
- Alle artikel data inclusief volledige content beschikbaar voor AI

### 3. GetArticlesByEntity (service.go:306-365)
**Voor:**
- Haalde alleen basis artikel velden op

**Na:**
- Haalt ook `content`, `content_extracted`, `content_extracted_at` op
- Entity zoeken werkt nu ook met volledige content

### 4. formatFunctionResult (chat_service.go:178-205)
**Voor:**
- Toonde alleen titel, bron, datum

**Na:**
- Toont ook content preview (eerste 200 karakters) als beschikbaar
- Format: `"- Titel (bron: X, datum: Y)\n  Inhoud: [preview]..."`

## 🎯 Voordelen

1. **Betere Context**: AI heeft toegang tot volledige artikel tekst
2. **Nauwkeurigere Zoekresultaten**: Zoekt ook in de content, niet alleen title/summary
3. **Rijkere Responses**: Content preview wordt getoond aan gebruiker
4. **Slimmere AI**: Kan vragen beantwoorden op basis van volledige artikelen

## 📊 Voorbeeld Response

**Zonder content:**
```
- "Klimaatwet aangenomen" (bron: nu.nl, datum: 2025-01-28)
```

**Met content:**
```
- "Klimaatwet aangenomen" (bron: nu.nl, datum: 2025-01-28)
  Inhoud: De Tweede Kamer heeft vandaag de nieuwe klimaatwet goedgekeurd. 
  Het wetsvoorstel bevat ambitieuze doelstellingen voor CO2-reductie en...
```

## 🔄 Volgende Stap

**Herstart de backend** om de wijzigingen actief te maken:

```powershell
# Stop backend (Ctrl+C)
go build -o api.exe ./cmd/api
.\api.exe
```

Na herstart kan de AI:
- Zoeken in volledige artikel content
- Rijkere antwoorden geven met content previews
- Betere context gebruiken voor vragen