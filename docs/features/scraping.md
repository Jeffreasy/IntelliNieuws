# Scraping Opties & Mogelijkheden

## Huidige Implementatie: RSS Scraping

**Status:** ✅ **ACTIEF** en werkt perfect

### Wat gebeurt er nu?

Het systeem scraped automatisch RSS feeds van:
- **nu.nl** - https://www.nu.nl/rss
- **ad.nl** - https://www.ad.nl/rss.xml
- **nos.nl** - https://feeds.nos.nl/nosnieuwsalgemeen

**Frequentie:** Elke 15 minuten (zie [`.env`](.env:30))

**Wat krijgen we uit RSS feeds:**
- ✅ Titel
- ✅ Samenvatting/beschrijving
- ✅ URL naar volledig artikel
- ✅ Publicatiedatum
- ✅ Categorie/tags
- ✅ Auteur (soms)
- ✅ Afbeelding URL (soms)

## Dynamic (HTML) Scraping

**Status:** ❌ **NIET GEÏMPLEMENTEERD** (feature flag bestaat, code niet)

### Waarom niet geïmplementeerd?

RSS scraping is **superieur** voor nieuws sites:

**RSS Voordelen:**
1. **Officieel aangeboden** - Sites willen dat je hun RSS gebruikt
2. **Gestructureerde data** - Geen HTML parsing nodig
3. **Stabiel** - RSS formaat verandert niet bij redesigns
4. **Snel** - Klein bestand, niet hele pagina
5. **Compleet** - Bevat al de metadata die je nodig hebt
6. **Legaal** - Expliciet voor content distributie bedoeld

**HTML Scraping Nadelen:**
1. **Breekt vaak** - Elke website update = kapotte scraper
2. **Langzaam** - Moet hele HTML pagina's downloaden
3. **Geblokkeerd** - Anti-scraping maatregelen
4. **Complex** - Per site custom CSS selectors
5. **Grijze zone** - Juridisch discutabel
6. **Resource intensief** - Meer bandwidth, CPU, geheugen

### Wanneer zou je HTML scraping WILLEN?

**Use case 1: Volledige artikel tekst**
RSS feeds geven vaak alleen samenvatting. Voor de VOLLEDIGE tekst moet je:
1. RSS scrapen (krijg URL)
2. HTML van die URL scrapen (krijg volledige tekst)

**Use case 2: Sites zonder RSS**
Sommige nieuws sites hebben geen RSS feed.

**Use case 3: Extra metadata**
Reacties, views, shares, etc. staan niet in RSS.

## Wat We NU Kunnen Doen

### Optie 1: Meer RSS Feeds Toevoegen ✅ AANBEVOLEN

Voeg meer nieuws bronnen toe door RSS feeds te configureren:

**Nederlandse nieuws sites met RSS:**
- Trouw - `https://www.trouw.nl/rss.xml`
- Volkskrant - `https://www.volkskrant.nl/rss.xml`
- Telegraaf - `https://www.telegraaf.nl/rss`
- RTL Nieuws - `https://www.rtlnieuws.nl/rss.xml`
- Metro - `https://www.metronieuws.nl/feed/`
- NRC - `https://www.nrc.nl/rss/`

**Toevoegen in `.env`:**
```env
TARGET_SITES=nu.nl,ad.nl,nos.nl,trouw.nl,volkskrant.nl,rtl.nl
```

En in [`service.go`](internal/scraper/service.go:46-50):
```go
ScrapeSources = map[string]string{
    "nu.nl":         "https://www.nu.nl/rss",
    "ad.nl":         "https://www.ad.nl/rss.xml",
    "nos.nl":        "https://feeds.nos.nl/nosnieuwsalgemeen",
    "trouw.nl":      "https://www.trouw.nl/rss.xml",
    "volkskrant.nl": "https://www.volkskrant.nl/rss.xml",
    // etc.
}
```

### Optie 2: Volledige Artikel Tekst Scrapen (Hybrid)

**Strategie:**
1. RSS scrapt titel + samenvatting + URL ✅ (hebben we al)
2. Optioneel: Download HTML van artikel URL voor volledige tekst
3. Gebruik AI om volledige tekst te analyseren (beter dan alleen summary)

**Voordeel:** Best of both worlds
**Nadeel:** Langzamer, meer bandwidth

**Implementatie:**
```go
// Pseudo-code
func (s *Service) ScrapeFullArticle(url string) (string, error) {
    // 1. Download HTML
    resp, err := http.Get(url)
    
    // 2. Parse met goquery
    doc, _ := goquery.NewDocumentFromReader(resp.Body)
    
    // 3. Extract main content (per-site selectors nodig)
    content := doc.Find("article").Text()
    
    return content, nil
}
```

### Optie 3: Custom HTML Scraper Bouwen

Als je ECHT HTML scraping wilt:

**Benodigde libraries:**
```go
import (
    "github.com/PuerkitoBio/goquery"  // jQuery-like HTML parsing
    "github.com/gocolly/colly/v2"     // Web scraping framework
)
```

**Per site moet je definiëren:**
- URL patterns
- CSS selectors voor titel, tekst, datum, etc.
- Anti-blocking strategie
- Error handling

**Tijd investering:** ~2-4 dagen per site

## Aanbeveling

**BLIJF BIJ RSS SCRAPING** ✅

Redenen:
1. Werkt perfect nu
2. Alle grote Nederlandse nieuws sites hebben RSS
3. RSS is designed voor dit doel
4. Juridisch veilig
5. Snel en betrouwbaar
6. Minimaal onderhoud

**Voeg optioneel toe:**
- Meer RSS bronnen (easy!)
- Volledige artikel tekst via hybrid benadering (medium)

**Vermijd:**
- Pure HTML scraping zonder RSS fallback
- Per-site custom scrapers bouwen
- Anti-scraping bypass trucs

## Huidige RSS Coverage

Met NU.nl, AD.nl en NOS.nl heb je:
- ✅ **~75 artikelen per 15 minuten**
- ✅ **Grootste Nederlandse nieuws sites**
- ✅ **Brede categorie coverage**
- ✅ **Stabiel en betrouwbaar**

Dit is meer dan genoeg voor de meeste use cases!

## Conclusie

Je hebt momenteel **GEEN** HTML scraping, **ALLEEN** RSS scraping. Dat is een **bewuste keuze** en de **juiste aanpak** voor een nieuws aggregator. RSS scraping is sneller, betrouwbaarder en completer dan HTML scraping voor nieuws content.

Wil je meer content? Voeg meer RSS feeds toe! 
Wil je volledige tekst? Implementeer hybrid benadering.
Wil je HTML scraping? Denk goed na of het de moeite waard is.