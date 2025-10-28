# Frontend Quick Start Guide

Deze gids laat zien hoe je snel een frontend kunt bouwen die met de Nieuws Scraper API communiceert.

## 1. Start de Backend

```bash
# Start de API server
go run cmd/api/main.go

# Server draait op: http://localhost:8080
# API base: http://localhost:8080/api/v1
```

## 2. Basis Frontend Voorbeeld (Vanilla JavaScript)

### HTML Pagina
```html
<!DOCTYPE html>
<html lang="nl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nieuws Scraper</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        .article { border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .article h3 { margin-top: 0; }
        .meta { color: #666; font-size: 0.9em; }
        button { padding: 10px 20px; margin: 5px; cursor: pointer; }
        .loading { color: #666; font-style: italic; }
        .error { color: red; padding: 10px; background: #fee; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>📰 Nieuws Scraper</h1>
    
    <div>
        <button onclick="loadArticles()">Laad Artikelen</button>
        <button onclick="loadStats()">Toon Statistieken</button>
        <button onclick="loadTrending()">Trending Topics</button>
    </div>
    
    <div id="content"></div>
    
    <script src="app.js"></script>
</body>
</html>
```

### JavaScript (app.js)
```javascript
const API_BASE = 'http://localhost:8080/api/v1';

// Helper functie voor API calls
async function apiCall(endpoint) {
    const content = document.getElementById('content');
    content.innerHTML = '<p class="loading">Laden...</p>';
    
    try {
        const response = await fetch(`${API_BASE}${endpoint}`);
        const data = await response.json();
        
        if (!data.success) {
            throw new Error(data.error?.message || 'API Error');
        }
        
        return data;
    } catch (error) {
        content.innerHTML = `<div class="error">Fout: ${error.message}</div>`;
        throw error;
    }
}

// Laad artikelen
async function loadArticles() {
    try {
        const data = await apiCall('/articles?limit=10&sort_by=published&sort_order=desc');
        const content = document.getElementById('content');
        
        const articlesHTML = data.data.map(article => `
            <div class="article">
                <h3>${article.title}</h3>
                <p>${article.summary}</p>
                <div class="meta">
                    <strong>Bron:</strong> ${article.source} | 
                    <strong>Categorie:</strong> ${article.category || 'N/A'} | 
                    <strong>Datum:</strong> ${new Date(article.published).toLocaleString('nl-NL')}
                </div>
                <a href="${article.url}" target="_blank">Lees meer →</a>
            </div>
        `).join('');
        
        const paginationInfo = `
            <p><strong>Totaal artikelen:</strong> ${data.meta.pagination.total} | 
            <strong>Pagina:</strong> ${data.meta.pagination.current_page}/${data.meta.pagination.total_pages}</p>
        `;
        
        content.innerHTML = `
            <h2>📄 Recente Artikelen</h2>
            ${paginationInfo}
            ${articlesHTML}
        `;
    } catch (error) {
        console.error('Fout bij laden artikelen:', error);
    }
}

// Laad statistieken
async function loadStats() {
    try {
        const data = await apiCall('/articles/stats');
        const content = document.getElementById('content');
        const stats = data.data;
        
        const sourcesList = Object.entries(stats.articles_by_source)
            .map(([source, count]) => `<li><strong>${source}:</strong> ${count} artikelen</li>`)
            .join('');
        
        const categoriesList = Object.values(stats.categories)
            .map(cat => `<li><strong>${cat.name}:</strong> ${cat.article_count} artikelen</li>`)
            .join('');
        
        content.innerHTML = `
            <h2>📊 Statistieken</h2>
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 20px;">
                <div>
                    <h3>Algemeen</h3>
                    <p><strong>Totaal artikelen:</strong> ${stats.total_articles}</p>
                    <p><strong>Laatste 24 uur:</strong> ${stats.recent_articles_24h}</p>
                    
                    <h3>Per Bron</h3>
                    <ul>${sourcesList}</ul>
                </div>
                <div>
                    <h3>Categorieën</h3>
                    <ul>${categoriesList}</ul>
                </div>
            </div>
        `;
    } catch (error) {
        console.error('Fout bij laden statistieken:', error);
    }
}

// Laad trending topics (AI feature)
async function loadTrending() {
    try {
        const data = await apiCall('/ai/trending?limit=10');
        const content = document.getElementById('content');
        
        if (!data.data || data.data.length === 0) {
            content.innerHTML = '<p>Geen trending topics gevonden. Start AI processing eerst.</p>';
            return;
        }
        
        const trendingHTML = data.data.map(topic => `
            <div class="article">
                <h3>🔥 ${topic.entity_text}</h3>
                <p><strong>Type:</strong> ${topic.entity_type}</p>
                <p><strong>Aantal artikelen:</strong> ${topic.article_count}</p>
                <p><strong>Sentiment:</strong> ${topic.avg_sentiment?.toFixed(2) || 'N/A'}</p>
            </div>
        `).join('');
        
        content.innerHTML = `
            <h2>🔥 Trending Topics (AI)</h2>
            ${trendingHTML}
        `;
    } catch (error) {
        console.error('Fout bij laden trending topics:', error);
    }
}

// Laad artikelen bij opstarten
window.onload = () => loadArticles();
```

## 3. React Voorbeeld

### Installatie
```bash
npx create-react-app nieuws-frontend
cd nieuws-frontend
npm install
```

### API Service (src/services/api.js)
```javascript
const API_BASE = 'http://localhost:8080/api/v1';

class NewsAPI {
    async request(endpoint) {
        const response = await fetch(`${API_BASE}${endpoint}`);
        const data = await response.json();
        
        if (!data.success) {
            throw new Error(data.error?.message || 'API Error');
        }
        
        return data;
    }
    
    // Artikelen
    async getArticles(params = {}) {
        const query = new URLSearchParams({
            limit: params.limit || 20,
            offset: params.offset || 0,
            sort_by: params.sortBy || 'published',
            sort_order: params.sortOrder || 'desc',
            ...params
        });
        return this.request(`/articles?${query}`);
    }
    
    async getArticle(id) {
        return this.request(`/articles/${id}`);
    }
    
    async searchArticles(query, params = {}) {
        const searchParams = new URLSearchParams({
            q: query,
            limit: params.limit || 20,
            offset: params.offset || 0
        });
        return this.request(`/articles/search?${searchParams}`);
    }
    
    async getStats() {
        return this.request('/articles/stats');
    }
    
    // AI Features
    async getTrendingTopics(limit = 10) {
        return this.request(`/ai/trending?limit=${limit}`);
    }
    
    async getSentimentStats() {
        return this.request('/ai/sentiment/stats');
    }
    
    async getEnrichment(articleId) {
        return this.request(`/articles/${articleId}/enrichment`);
    }
    
    // Sources & Categories
    async getSources() {
        return this.request('/sources');
    }
    
    async getCategories() {
        return this.request('/categories');
    }
}

export default new NewsAPI();
```

### React Component (src/App.js)
```javascript
import React, { useState, useEffect } from 'react';
import api from './services/api';
import './App.css';

function App() {
    const [articles, setArticles] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    
    useEffect(() => {
        loadArticles();
    }, [page]);
    
    const loadArticles = async () => {
        try {
            setLoading(true);
            const data = await api.getArticles({
                limit: 20,
                offset: (page - 1) * 20
            });
            setArticles(data.data);
            setTotalPages(data.meta.pagination.total_pages);
            setError(null);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };
    
    if (loading) return <div className="loading">Laden...</div>;
    if (error) return <div className="error">Fout: {error}</div>;
    
    return (
        <div className="App">
            <h1>📰 Nieuws Scraper</h1>
            
            <div className="articles">
                {articles.map(article => (
                    <article key={article.id} className="article-card">
                        <h2>{article.title}</h2>
                        <p>{article.summary}</p>
                        <div className="meta">
                            <span>🏢 {article.source}</span>
                            <span>📁 {article.category || 'N/A'}</span>
                            <span>📅 {new Date(article.published).toLocaleDateString('nl-NL')}</span>
                        </div>
                        <a href={article.url} target="_blank" rel="noopener noreferrer">
                            Lees meer →
                        </a>
                    </article>
                ))}
            </div>
            
            <div className="pagination">
                <button 
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page === 1}
                >
                    ← Vorige
                </button>
                <span>Pagina {page} van {totalPages}</span>
                <button 
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                    disabled={page === totalPages}
                >
                    Volgende →
                </button>
            </div>
        </div>
    );
}

export default App;
```

## 4. TypeScript Types

```typescript
// types.ts
export interface Article {
    id: number;
    title: string;
    summary: string;
    url: string;
    published: string;
    source: string;
    keywords: string[];
    image_url?: string;
    author?: string;
    category?: string;
    created_at: string;
    update