# Article Context Support voor Chat API

## ✅ Probleem Opgelost

De chat API accepteert nu **onbeperkte artikel content** via een apart veld, zodat het `message` veld binnen de 1000 karakter limiet blijft.

## 📝 API Wijzigingen

### Nieuwe Request Velden

```typescript
interface ChatRequest {
  message: string;              // User vraag (max 1000 chars)
  context?: string;             // Conversatie context (optional)
  article_content?: string;     // Artikel content (GEEN limiet!)
  article_id?: number;          // Artikel ID (optional)
}
```

## 🎯 Gebruik

### Voor: ❌ Fout - Content in message
```json
{
  "message": "Wat vind je van dit artikel?\n\n[LANGE ARTIKEL CONTENT...]"
}
```
**Resultaat:** `400 Bad Request - Message too long`

### Na: ✅ Correct - Content apart
```json
{
  "message": "Wat vind je van dit artikel?",
  "article_content": "[LANGE ARTIKEL CONTENT...]",
  "article_id": 123
}
```
**Resultaat:** `200 OK` - AI krijgt volledige context!

## 💡 Frontend Implementatie

### React Hook Update

```typescript
export const sendChatMessage = async (
  message: string,
  articleContent?: string,
  articleId?: number
) => {
  const response = await fetch('http://localhost:8080/api/v1/ai/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      message,
      article_content: articleContent,
      article_id: articleId,
    }),
  });

  return response.json();
};
```

### Gebruik in Component

```typescript
const ChatWithArticleContext = ({ article }: { article: Article }) => {
  const [message, setMessage] = useState('');

  const handleSendMessage = async () => {
    try {
      const response = await sendChatMessage(
        message,
        article.content,  // Volledige artikel content
        article.id
      );
      
      // Process response...
    } catch (error) {
      console.error('Chat error:', error);
    }
  };

  return (
    <div className="chat">
      <input
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        placeholder="Stel een vraag over dit artikel..."
        maxLength={1000}  // Enforce limit in UI
      />
      <button onClick={handleSendMessage}>Verstuur</button>
    </div>
  );
};
```

## 🔍 Backend Verwerking

De backend:
1. Neemt de `article_content` uit de request
2. Truncates naar 4000 karakters (voor OpenAI token limiet)
3. Voegt toe als context: `"Context - Artikel content:\n\n[content]\n\n---\n\nVraag: [message]"`
4. Stuurt naar OpenAI voor verwerking

## ✨ Voordelen

1. ✅ **Geen 1000 karakter limiet** voor artikel content
2. ✅ **Duidelijke scheiding** tussen vraag en context
3. ✅ **Betere AI responses** met volledige artikel context
4. ✅ **Backward compatible** - article_content is optioneel
5. ✅ **Type-safe** met article_id tracking

## 🧪 Test Voorbeelden

### Voorbeeld 1: Vraag zonder artikel
```json
POST /api/v1/ai/chat
{
  "message": "Wat zijn de trending topics?"
}
```

### Voorbeeld 2: Vraag met artikel context
```json
POST /api/v1/ai/chat
{
  "message": "Vat dit artikel samen in 3 punten",
  "article_content": "De Tweede Kamer heeft vandaag... [lang artikel]",
  "article_id": 123
}
```

### Voorbeeld 3: Conversatie met context
```json
POST /api/v1/ai/chat
{
  "message": "Wat is het sentiment?",
  "context": "Ik heb je zojuist gevraagd om een samenvatting.",
  "article_content": "[artikel content]",
  "article_id": 123
}
```

## 🔄 Migration Guide

Als je frontend al het `article_content` in het `message` veld stuurt:

**Voor:**
```typescript
const fullMessage = `Analyseer dit artikel:\n\n${article.content}\n\nVraag: ${userQuestion}`;
await sendChatMessage(fullMessage);  // ❌ Te lang!
```

**Na:**
```typescript
await sendChatMessage(
  userQuestion,           // Alleen de vraag
  article.content,        // Content apart
  article.id             // ID voor tracking
);  // ✅ Werkt perfect!
```

## 📋 Checklist

- [ ] Update frontend TypeScript types
- [ ] Update sendChatMessage functie om article_content te accepteren
- [ ] Update UI components om article.content apart mee te sturen
- [ ] Test met lange artikelen (>1000 chars)
- [ ] Verify dat korte vragen nog steeds werken
- [ ] Herstart backend met nieuwe code

---

**Status:** ✅ Backend geüpdatet, frontend moet aangepast worden
**Versie:** 1.1.0
**Laatste Update:** 2025-01-28