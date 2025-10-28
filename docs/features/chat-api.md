# AI Chat API Documentatie

Complete documentatie voor de conversational AI chat endpoint waarmee de frontend natuurlijke taal vragen kan stellen over de nieuws database.

## 📋 Inhoudsopgave

1. [Overzicht](#overzicht)
2. [Chat Endpoint](#chat-endpoint)
3. [Beschikbare Functies](#beschikbare-functies)
4. [Frontend Implementatie](#frontend-implementatie)
5. [Voorbeelden](#voorbeelden)
6. [Best Practices](#best-practices)

## Overzicht

De AI Chat API biedt een conversational interface voor het bevragen van de nieuws database. De AI assistent kan:

- 🔍 Artikelen zoeken op basis van keywords
- 📊 Sentiment statistieken ophalen
- 🔥 Trending topics identificeren  
- 👤 Artikelen vinden over specifieke personen/organisaties/locaties
- 📰 Recente artikelen ophalen met filters

De AI gebruikt **OpenAI Function Calling** om automatisch de juiste database queries uit te voeren op basis van natuurlijke taal vragen.

## Chat Endpoint

### POST /api/v1/ai/chat

Stuur een natuurlijke taal vraag en ontvang een intelligent antwoord met relevante data.

**Authenticatie:** Niet vereist (public endpoint)

**Request Body:**

```json
{
  "message": "Wat zijn de meest negatieve artikelen van vandaag?",
  "context": "Optionele conversatie context"
}
```

**Parameters:**

- `message` (string, required): De vraag/bericht van de gebruiker (max 1000 karakters)
- `context` (string, optional): Conversatie context voor follow-up vragen

**Response:**

```json
{
  "success": true,
  "data": {
    "message": "Hier zijn de meest negatieve artikelen van vandaag...",
    "articles": [
      {
        "id": 123,
        "title": "Artikel titel",
        "summary": "Samenvatting...",
        "url": "https://...",
        "published": "2025-01-28T12:00:00Z",
        "source": "nu.nl",
        "sentiment": {
          "score": -0.8,
          "label": "negative"
        }
      }
    ],
    "stats": {
      "total_articles": 150,
      "positive_count": 30,
      "negative_count": 50,
      "neutral_count": 70
    }
  },
  "request_id": "abc123",
  "timestamp": "2025-01-28T14:30:00Z"
}
```

**Response Velden:**

- `message` (string): AI gegenereerd antwoord in natuurlijke taal
- `articles` (array, optional): Relevante artikelen indien van toepassing
- `stats` (object, optional): Statistieken indien van toepassing
- `sources` (array, optional): Lijst van bronnen indien van toepassing

## Beschikbare Functies

De AI kan automatisch de volgende functies aanroepen:

### 1. Artikelen Zoeken

**Voorbeeldvragen:**
- "Zoek artikelen over klimaat"
- "Geef me nieuws over verkiezingen"
- "Wat zijn de laatste artikelen over technologie?"

**Functie:** `search_articles`

### 2. Sentiment Statistieken

**Voorbeeldvragen:**
- "Wat is het algemene sentiment van het nieuws?"
- "Hoeveel positieve artikelen zijn er vandaag?"
- "Laat sentiment stats zien voor NU.nl"

**Functie:** `get_sentiment_stats`

### 3. Trending Topics

**Voorbeeldvragen:**
- "Wat is trending?"
- "Welke onderwerpen zijn populair vandaag?"
- "Toon trending topics van de laatste 24 uur"

**Functie:** `get_trending_topics`

### 4. Artikelen per Entity

**Voorbeeldvragen:**
- "Laat artikelen over Mark Rutte zien"
- "Wat zegt het nieuws over Tesla?"
- "Artikelen over Amsterdam"

**Functie:** `get_articles_by_entity`

### 5. Recente Artikelen

**Voorbeeldvragen:**
- "Wat zijn de laatste artikelen?"
- "Toon recente positieve artikelen van NOS"
- "Geef me sport nieuws van vandaag"

**Functie:** `get_recent_articles`

## Frontend Implementatie

### React Hook Voorbeeld

```typescript
import { useState } from 'react';

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  data?: any;
}

export const useAIChat = () => {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);

  const sendMessage = async (message: string) => {
    setLoading(true);
    
    // Add user message
    const userMessage: ChatMessage = { role: 'user', content: message };
    setMessages(prev => [...prev, userMessage]);

    try {
      const response = await fetch('http://localhost:8080/api/v1/ai/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message }),
      });

      const result = await response.json();
      
      if (result.success) {
        const assistantMessage: ChatMessage = {
          role: 'assistant',
          content: result.data.message,
          data: {
            articles: result.data.articles,
            stats: result.data.stats,
          },
        };
        setMessages(prev => [...prev, assistantMessage]);
      }
    } catch (error) {
      console.error('Chat error:', error);
    } finally {
      setLoading(false);
    }
  };

  return { messages, loading, sendMessage };
};
```

### Chat Component Voorbeeld

```typescript
const AIChatModal = () => {
  const { messages, loading, sendMessage } = useAIChat();
  const [input, setInput] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim()) {
      sendMessage(input);
      setInput('');
    }
  };

  return (
    <div className="chat-modal">
      <div className="chat-header">
        <h2>🤖 AI Assistent</h2>
        <p>Stel vragen over het nieuws</p>
      </div>

      <div className="chat-messages">
        {messages.map((msg, idx) => (
          <ChatMessage key={idx} message={msg} />
        ))}
        {loading && <LoadingIndicator />}
      </div>

      <form onSubmit={handleSubmit} className="chat-input">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Stel een vraag..."
          maxLength={1000}
        />
        <button type="submit" disabled={loading || !input.trim()}>
          Verstuur
        </button>
      </form>
    </div>
  );
};

const ChatMessage = ({ message }: { message: ChatMessage }) => {
  return (
    <div className={`message ${message.role}`}>
      <div className="message-content">
        {message.content}
      </div>
      
      {/* Render articles if present */}
      {message.data?.articles && (
        <div className="message-articles">
          {message.data.articles.map((article: any) => (
            <ArticleCard key={article.id} article={article} compact />
          ))}
        </div>
      )}
      
      {/* Render stats if present */}
      {message.data?.stats && (
        <div className="message-stats">
          <SentimentChart stats={message.data.stats} />
        </div>
      )}
    </div>
  );
};
```

## Voorbeelden

### Voorbeeld 1: Artikelen Zoeken

**Request:**
```json
{
  "message": "Zoek artikelen over klimaat"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Ik heb 8 artikelen over klimaat gevonden van de afgelopen dagen. Hier zijn de meest recente:",
    "articles": [
      {
        "id": 456,
        "title": "Nieuwe klimaatmaatregelen aangekondigd",
        "summary": "De regering heeft vandaag...",
        "source": "nu.nl",
        "published": "2025-01-28T10:00:00Z"
      }
    ]
  }
}
```

### Voorbeeld 2: Sentiment Analyse

**Request:**
```json
{
  "message": "Wat is het sentiment van het nieuws vandaag?"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Het sentiment van het nieuws vandaag is overwegend neutraal. Van de 150 artikelen zijn er 45 positief (30%), 80 neutraal (53%), en 25 negatief (17%). De gemiddelde sentiment score is 0.15, wat licht positief is.",
    "stats": {
      "total_articles": 150,
      "positive_count": 45,
      "neutral_count": 80,
      "negative_count": 25,
      "average_sentiment": 0.15
    }
  }
}
```

### Voorbeeld 3: Trending Topics

**Request:**
```json
{
  "message": "Wat is er trending?"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Dit zijn de trending onderwerpen van de laatste 24 uur:\n\n1. verkiezingen (15 artikelen, sentiment: 0.2)\n2. klimaat (12 artikelen, sentiment: -0.3)\n3. economie (10 artikelen, sentiment: 0.4)",
    "stats": {
      "topics": [
        {
          "keyword": "verkiezingen",
          "article_count": 15,
          "average_sentiment": 0.2,
          "sources": ["nu.nl", "nos.nl", "ad.nl"]
        }
      ]
    }
  }
}
```

### Voorbeeld 4: Multi-turn Conversatie

**Request 1:**
```json
{
  "message": "Wat zijn de trending topics?"
}
```

**Response 1:**
```json
{
  "data": {
    "message": "De top 3 trending topics zijn: verkiezingen, klimaat, en economie."
  }
}
```

**Request 2 (met context):**
```json
{
  "message": "Geef me meer details over verkiezingen",
  "context": "De top 3 trending topics zijn: verkiezingen, klimaat, en economie."
}
```

**Response 2:**
```json
{
  "data": {
    "message": "Hier zijn 15 artikelen over verkiezingen...",
    "articles": [...]
  }
}
```

## Best Practices

### 1. Error Handling

```typescript
try {
  const response = await sendChatMessage(message);
  
  if (!response.success) {
    showError(response.error.message);
    return;
  }
  
  displayResponse(response.data);
} catch (error) {
  showError('Kon geen verbinding maken met de AI service');
}
```

### 2. Loading States

```typescript
const [isTyping, setIsTyping] = useState(false);

const sendMessage = async (msg: string) => {
  setIsTyping(true);
  try {
    const response = await fetch('/api/v1/ai/chat', {
      method: 'POST',
      body: JSON.stringify({ message: msg }),
    });
    // Process response
  } finally {
    setIsTyping(false);
  }
};
```

### 3. Message Validation

```typescript
const validateMessage = (message: string): boolean => {
  if (!message.trim()) {
    showError('Bericht mag niet leeg zijn');
    return false;
  }
  
  if (message.length > 1000) {
    showError('Bericht mag maximaal 1000 karakters zijn');
    return false;
  }
  
  return true;
};
```

### 4. Conversatie Context

```typescript
// Bewaar laatste paar berichten als context
const getConversationContext = (messages: ChatMessage[]): string => {
  return messages
    .slice(-3) // Laatste 3 berichten
    .filter(m => m.role === 'assistant')
    .map(m => m.content)
    .join('\n');
};

const sendMessage = async (message: string) => {
  const context = getConversationContext(messages);
  
  const response = await fetch('/api/v1/ai/chat', {
    method: 'POST',
    body: JSON.stringify({ message, context }),
  });
};
```

### 5. Suggested Questions

```typescript
const suggestedQuestions = [
  "Wat is het sentiment van het nieuws?",
  "Welke onderwerpen zijn trending?",
  "Zoek artikelen over klimaat",
  "Toon recente positieve artikelen",
  "Wat zegt het nieuws over de economie?",
];

<div className="suggested-questions">
  {suggestedQuestions.map(q => (
    <button key={q} onClick={() => sendMessage(q)}>
      {q}
    </button>
  ))}
</div>
```

### 6. Response Formatting

```typescript
const formatResponse = (data: ChatResponse) => {
  return (
    <div className="ai-response">
      {/* Tekst antwoord */}
      <div className="response-text">
        {data.message}
      </div>
      
      {/* Artikelen indien aanwezig */}
      {data.articles && data.articles.length > 0 && (
        <div className="response-articles">
          <h4>📰 Gevonden artikelen ({data.articles.length})</h4>
          <ArticleList articles={data.articles} />
        </div>
      )}
      
      {/* Statistieken indien aanwezig */}
      {data.stats && (
        <div className="response-stats">
          <h4>📊 Statistieken</h4>
          <StatsDisplay stats={data.stats} />
        </div>
      )}
    </div>
  );
};
```

## Technische Details

### Caching

Responses worden 2 minuten gecached voor identieke vragen:
- Snellere responses voor veel gestelde vragen
- Vermindert API kosten
- Automatische cache invalidatie

### Rate Limiting

- Standaard rate limiting van toepassing (zoals geconfigureerd)
- Geen specifieke chat rate limits
- Gebruik debouncing bij typing

### Timeouts

- OpenAI API timeout: 30 seconden
- Function call execution: variabel per functie
- Totale request timeout: ~45 seconden

---

**Status**: ✅ AI Chat API volledig geïmplementeerd en gedocumenteerd  
**Endpoint**: `POST /api/v1/ai/chat`  
**Versie**: 1.0.0  
**Laatste Update**: 2025-01-28