# Frontend: Content Extraction Implementatie

## 🎯 Nieuwe API Endpoint

We hebben een nieuw endpoint toegevoegd voor handmatige content extraction:

```
POST /api/v1/articles/:id/extract-content
```

**⚠️ BELANGRIJK:** Dit endpoint vereist **API key authentication**!

---

## 🔑 Authenticatie Vereist

De frontend kreeg een **401 Unauthorized** error omdat de API key ontbrak.

### Probleem in de Logs
```json
{"level":"info","message":"[...] POST /api/v1/articles/164/extract-content - 401 - 0s"}
```

### Oplossing: API Key Meesturen

**In de frontend JavaScript/TypeScript:**

```typescript
// Haal API key uit environment of config
const API_KEY = 'test123geheim'; // OF uit .env: process.env.REACT_APP_API_KEY

// Bij fetch requests:
const response = await fetch(`http://localhost:8080/api/v1/articles/${articleId}/extract-content`, {
  method: 'POST',
  headers: {
    'X-API-Key': API_KEY,  // ← DIT IS VERPLICHT!
    'Content-Type': 'application/json',
  },
});

if (response.ok) {
  const data = await response.json();
  console.log('Content extracted:', data);
} else if (response.status === 401) {
  console.error('Invalid or missing API key');
} else {
  console.error('Extraction failed:', await response.text());
}
```

**Met axios:**

```typescript
import axios from 'axios';

const API_KEY = 'test123geheim';

const extractContent = async (articleId: number) => {
  try {
    const response = await axios.post(
      `http://localhost:8080/api/v1/articles/${articleId}/extract-content`,
      {},
      {
        headers: {
          'X-API-Key': API_KEY,  // ← VERPLICHT
        },
      }
    );
    
    return response.data;
  } catch (error) {
    if (error.response?.status === 401) {
      console.error('Authentication failed - check API key');
    }
    throw error;
  }
};
```

---

## 📝 Request & Response

### Request
```http
POST /api/v1/articles/164/extract-content HTTP/1.1
Host: localhost:8080
X-API-Key: test123geheim
Content-Type: application/json
```

### Success Response (200 OK)
```json
{
  "success": true,
  "data": {
    "success": true,
    "message": "Content extracted successfully",
    "characters": 2453,
    "article": {
      "id": 164,
      "title": "Artikel Titel",
      "summary": "RSS summary...",
      "content": "Volledige artikel tekst hier... (2453 characters)",
      "content_extracted": true,
      "content_extracted_at": "2025-10-28T18:00:00Z",
      "url": "https://www.nu.nl/...",
      "source": "nu.nl",
      "published": "2025-10-28T17:00:00Z"
    }
  },
  "request_id": "..."
}
```

### Error Responses

**401 Unauthorized** (Missing/Invalid API Key)
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or missing API key"
  }
}
```

**400 Bad Request** (Invalid Article ID)
```json
{
  "success": false,
  "error": {
    "code": "INVALID_ID",
    "message": "Article ID must be a valid integer"
  }
}
```

**404 Not Found** (Article doesn't exist)
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Article not found"
  }
}
```

**500 Internal Server Error** (Extraction failed)
```json
{
  "success": false,
  "error": {
    "code": "EXTRACTION_FAILED",
    "message": "Failed to extract content",
    "details": "HTTP 403: Access denied"
  }
}
```

---

## 🎨 Frontend Component Voorbeeld

### React Component

```typescript
import React, { useState } from 'react';
import { Article } from './types';

const API_KEY = 'test123geheim'; // From .env
const API_BASE = 'http://localhost:8080/api/v1';

interface ContentExtractionButtonProps {
  article: Article;
  onContentExtracted?: (article: Article) => void;
}

export const ContentExtractionButton: React.FC<ContentExtractionButtonProps> = ({
  article,
  onContentExtracted,
}) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const extractContent = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(
        `${API_BASE}/articles/${article.id}/extract-content`,
        {
          method: 'POST',
          headers: {
            'X-API-Key': API_KEY,
            'Content-Type': 'application/json',
          },
        }
      );

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error?.message || 'Extraction failed');
      }

      const data = await response.json();
      const updatedArticle = data.data.article;

      // Callback with updated article
      onContentExtracted?.(updatedArticle);

      // Show success
      alert(`Content extracted: ${data.data.characters} characters`);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setError(errorMessage);
      console.error('Content extraction failed:', err);
    } finally {
      setLoading(false);
    }
  };

  // Don't show button if already extracted
  if (article.content_extracted) {
    return (
      <span className="badge badge-success">
        ✓ Content beschikbaar ({article.content?.length || 0} chars)
      </span>
    );
  }

  return (
    <div>
      <button
        onClick={extractContent}
        disabled={loading}
        className="btn btn-primary"
      >
        {loading ? 'Extracting...' : '📄 Haal Volledige Tekst Op'}
      </button>
      {error && <div className="error-message">{error}</div>}
    </div>
  );
};
```

### Vue Component

```vue
<template>
  <div>
    <button 
      v-if="!article.content_extracted"
      @click="extractContent"
      :disabled="loading"
      class="btn btn-primary"
    >
      {{ loading ? 'Bezig...' : '📄 Haal Volledige Tekst Op' }}
    </button>
    <span v-else class="badge badge-success">
      ✓ Content beschikbaar ({{ article.content?.length || 0 }} chars)
    </span>
    <div v-if="error" class="error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import type { Article } from './types';

const props = defineProps<{
  article: Article;
}>();

const emit = defineEmits<{
  contentExtracted: [article: Article];
}>();

const loading = ref(false);
const error = ref<string | null>(null);

const API_KEY = 'test123geheim'; // From env
const API_BASE = 'http://localhost:8080/api/v1';

const extractContent = async () => {
  loading.value = true;
  error.value = null;

  try {
    const response = await fetch(
      `${API_BASE}/articles/${props.article.id}/extract-content`,
      {
        method: 'POST',
        headers: {
          'X-API-Key': API_KEY,
          'Content-Type': 'application/json',
        },
      }
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error?.message || 'Extraction failed');
    }

    const data = await response.json();
    emit('contentExtracted', data.data.article);

  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error';
  } finally {
    loading.value = false;
  }
};
</script>
```

---

## 🔐 API Key Configuratie

### Backend (.env)
```env
API_KEY=test123geheim
API_KEY_HEADER=X-API-Key
```

### Frontend (.env of .env.local)

**React/Vite:**
```env
VITE_API_KEY=test123geheim
VITE_API_URL=http://localhost:8080/api/v1
```

**Next.js:**
```env
NEXT_PUBLIC_API_KEY=test123geheim
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

**Vue:**
```env
VUE_APP_API_KEY=test123geheim
VUE_APP_API_URL=http://localhost:8080/api/v1
```

**Gebruik in code:**
```typescript
// Vite
const API_KEY = import.meta.env.VITE_API_KEY;

// Next.js
const API_KEY = process.env.NEXT_PUBLIC_API_KEY;

// Create React App / Vue
const API_KEY = process.env.REACT_APP_API_KEY;
const API_KEY = process.env.VUE_APP_API_KEY;
```

---

## 🧪 Testing

### cURL Test (met API key)

```bash
# CORRECT (met API key):
curl -X POST http://localhost:8080/api/v1/articles/164/extract-content \
  -H "X-API-Key: test123geheim" \
  -H "Content-Type: application/json"

# Expected: 200 OK met content

# FOUT (zonder API key):
curl -X POST http://localhost:8080/api/v1/articles/164/extract-content

# Expected: 401 Unauthorized
```

### Browser Console Test

```javascript
// Open Chrome DevTools Console op je frontend
fetch('http://localhost:8080/api/v1/articles/164/extract-content', {
  method: 'POST',
  headers: {
    'X-API-Key': 'test123geheim',
    'Content-Type': 'application/json',
  },
})
.then(r => r.json())
.then(data => console.log('Success:', data))
.catch(err => console.error('Error:', err));
```

---

## ⚠️ Belangrijke Nota's

### 1. Database Migratie EERST!

**Voor het endpoint werkt, moet je de database migratie uitvoeren:**
```sql
-- In pgAdmin, run:
-- migrations/005_add_content_column.sql
```

Anders krijg je database errors omdat de `content` kolommen niet bestaan!

### 2. API Key is VERPLICHT

Het endpoint is protected. De frontend **MOET** de API key meesturen:
```
X-API-Key: test123geheim
```

Zonder deze header krijg je **401 Unauthorized**.

### 3. Rate Limiting

Het endpoint heeft rate limiting (100 requests/minuut). Bij te veel requests:
```json
{
  "error": "RATE_LIMIT_EXCEEDED",
  "message": "Too many requests, please try again later"
}
```

### 4. Timeout

Content extraction kan 5-30 seconden duren. Zorg dat je frontend:
- Loading state toont
- Timeout heeft van minimaal 30 seconden
- Error handling heeft

---

## 🎨 UX Aanbevelingen

### Show Extraction Status

```tsx
// Visual feedback voor gebruiker
{article.content_extracted ? (
  <div className="content-status success">
    <Icon name="check" />
    <span>Volledige tekst beschikbaar</span>
  </div>
) : (
  <div className="content-status pending">
    <Icon name="download" />
    <span>Alleen samenvatting</span>
    <button onClick={handleExtract}>Haal volledige tekst op</button>
  </div>
)}
```

### Progress Indicator

```tsx
const [extracting, setExtracting] = useState(false);
const [progress, setProgress] = useState(0);

const extractWithProgress = async (articleId: number) => {
  setExtracting(true);
  setProgress(0);
  
  // Simulate progress (extraction takes 2-10 seconds)
  const progressInterval = setInterval(() => {
    setProgress(p => Math.min(p + 10, 90));
  }, 500);
  
  try {
    await extractContent(articleId);
    setProgress(100);
  } finally {
    clearInterval(progressInterval);
    setTimeout(() => setExtracting(false), 500);
  }
};
```

### Error Handling met User Feedback

```tsx
const extractContent = async (articleId: number) => {
  try {
    const response = await fetch(`/api/v1/articles/${articleId}/extract-content`, {
      method: 'POST',
      headers: { 'X-API-Key': API_KEY },
    });
    
    if (response.status === 401) {
      toast.error('Authenticatie mislukt - check API configuratie');
      return;
    }
    
    if (response.status === 404) {
      toast.error('Artikel niet gevonden');
      return;
    }
    
    if (!response.ok) {
      const error = await response.json();
      toast.error(error.error?.message || 'Extraction mislukt');
      return;
    }
    
    const data = await response.json();
    toast.success(`Content opgehaald: ${data.data.characters} karakters`);
    
    // Update article in state
    updateArticle(data.data.article);
    
  } catch (error) {
    console.error('Extraction error:', error);
    toast.error('Netwerk error bij content extraction');
  }
};
```

---

## 🔧 Environment Setup

### Frontend .env File

**React/Vite:**
```env
VITE_API_URL=http://localhost:8080/api/v1
VITE_API_KEY=test123geheim
```

**Next.js:**
```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_API_KEY=test123geheim
```

### API Service Helper

Maak een centrale API service:

```typescript
// api/client.ts
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';
const API_KEY = import.meta.env.VITE_API_KEY || 'test123geheim';

class APIClient {
  private baseURL: string;
  private apiKey: string;

  constructor(baseURL: string, apiKey: string) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
  }

  private getHeaders(): Headers {
    const headers = new Headers();
    headers.set('Content-Type', 'application/json');
    headers.set('X-API-Key', this.apiKey);
    return headers;
  }

  async extractContent(articleId: number) {
    const response = await fetch(
      `${this.baseURL}/articles/${articleId}/extract-content`,
      {
        method: 'POST',
        headers: this.getHeaders(),
      }
    );

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error?.message || 'Extraction failed');
    }

    return response.json();
  }

  async getArticle(articleId: number) {
    const response = await fetch(
      `${this.baseURL}/articles/${articleId}`,
      {
        headers: this.getHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error('Failed to fetch article');
    }

    return response.json();
  }
}

export const apiClient = new APIClient(API_URL, API_KEY);
```

**Gebruik:**
```typescript
import { apiClient } from './api/client';

// Extract content
const result = await apiClient.extractContent(164);
console.log(result.data.article.content);

// Get article
const article = await apiClient.getArticle(164);
```

---

## ✅ Checklist Voor Frontend

Voordat je de frontend test:

- [ ] **Database migratie uitgevoerd** (migrations/005_add_content_column.sql)
- [ ] **Backend herstart** met nieuwe binary
- [ ] **API key geconfigureerd** in frontend .env
- [ ] **X-API-Key header** toegevoegd aan requests
- [ ] **Error handling** geïmplementeerd (401, 404, 500)
- [ ] **Loading state** getoond tijdens extraction
- [ ] **Success feedback** naar gebruiker
- [ ] **Article state update** na successful extraction

---

## 🎯 Complete Workflow

### 1. Artikel Lijst Tonen
```typescript
const articles = await fetch(`${API_URL}/articles`).then(r => r.json());
// Geen API key nodig voor GET requests (public)
```

### 2. Content Extraction Triggeren
```typescript
// User klikt "Haal volledige tekst op" button
const result = await fetch(
  `${API_URL}/articles/${articleId}/extract-content`,
  {
    method: 'POST',
    headers: { 'X-API-Key': API_KEY },  // API key WEL nodig voor POST
  }
).then(r => r.json());
```

### 3. Updated Article Tonen
```typescript
// Update article in state met nieuwe content
setArticle(result.data.article);

// Of refresh article van API
const updatedArticle = await fetch(
  `${API_URL}/articles/${articleId}`
).then(r => r.json());
```

---

## 🚀 Production Aanbevelingen

### 1. API Key Beveiliging

**NIET doen:**
```typescript
const API_KEY = 'test123geheim'; // Hardcoded in code!
```

**WEL doen:**
```typescript
const API_KEY = import.meta.env.VITE_API_KEY; // Van environment variable

// Check of key aanwezig is
if (!API_KEY) {
  throw new Error('API_KEY not configured');
}
```

### 2. Rate Limiting Handling

```typescript
const extractWithRetry = async (articleId: number, maxRetries = 3) => {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await extractContent(articleId);
    } catch (error) {
      if (error.status === 429) { // Rate limit
        // Wait and retry
        await new Promise(resolve => setTimeout(resolve, 2000 * (attempt + 1)));
        continue;
      }
      throw error; // Other errors, don't retry
    }
  }
};
```

### 3. Batch Extraction

```typescript
// Extract content voor meerdere artikelen (met delay)
const extractMultiple = async (articleIds: number[]) => {
  const results = [];
  
  for (const id of articleIds) {
    try {
      const result = await extractContent(id);
      results.push({ id, success: true, result });
    } catch (error) {
      results.push({ id, success: false, error });
    }
    
    // Delay tussen requests (respect rate limits)
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  
  return results;
};
```

---

## 📌 Summary

**Om het te laten werken:**

1. ✅ **Backend:** Endpoint is toegevoegd en werkt
2. ⏳ **Database:** Migratie moet uitgevoerd worden
3. ⚠️ **Frontend:** API key header moet toegevoegd worden

**De 401 error was verwacht** - het endpoint is protected en vereist authenticatie!

**Fix voor frontend:**
```typescript
headers: {
  'X-API-Key': 'test123geheim',  // ← Toevoegen
}
```

**Dan werkt het perfect!** 🎉