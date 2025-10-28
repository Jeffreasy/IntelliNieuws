# Frontend Integratie: Stock Tickers

Deze guide laat zien hoe je stock ticker data in de frontend applicatie kunt integreren.

## API Endpoints

### 1. Artikelen met Stock Tickers

```typescript
// GET /api/v1/articles/:id
interface Article {
  id: number;
  title: string;
  summary: string;
  url: string;
  published: string;
  source: string;
  ai_enrichment?: {
    entities?: {
      persons?: string[];
      organizations?: string[];
      locations?: string[];
      stock_tickers?: StockTicker[];
    };
    sentiment?: {
      score: number;
      label: 'positive' | 'negative' | 'neutral';
    };
    categories?: Record<string, number>;
    keywords?: Array<{word: string; score: number}>;
  };
  stock_data?: Record<string, StockQuote>;
}

interface StockTicker {
  symbol: string;        // "ASML", "AAPL", etc.
  name?: string;         // "ASML Holding"
  exchange?: string;     // "AEX", "NASDAQ"
  mentions?: number;     // Times mentioned in article
  context?: string;      // Context snippet
}

interface StockQuote {
  symbol: string;
  name: string;
  price: number;
  change: number;
  change_percent: number;
  volume: number;
  market_cap?: number;
  exchange: string;
  currency: string;
  last_updated: string;
  previous_close?: number;
  day_high?: number;
  day_low?: number;
}
```

### 2. Query Artikelen per Ticker

```typescript
// GET /api/v1/articles/by-ticker/:symbol?limit=10
const fetchArticlesByTicker = async (symbol: string, limit: number = 10) => {
  const response = await fetch(
    `${API_BASE}/articles/by-ticker/${symbol}?limit=${limit}`
  );
  return response.json();
};

// Voorbeeld gebruik
const asmlArticles = await fetchArticlesByTicker('ASML', 20);
```

### 3. Stock Quote Data

```typescript
// GET /api/v1/stocks/quote/:symbol
const fetchStockQuote = async (symbol: string): Promise<StockQuote> => {
  const response = await fetch(`${API_BASE}/stocks/quote/${symbol}`);
  return response.json();
};

// Voorbeeld
const asmlQuote = await fetchStockQuote('ASML');
console.log(`ASML: €${asmlQuote.price} (${asmlQuote.change_percent}%)`);
```

### 4. Multiple Quotes (Batch)

```typescript
// POST /api/v1/stocks/quotes
const fetchMultipleQuotes = async (symbols: string[]) => {
  const response = await fetch(`${API_BASE}/stocks/quotes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ symbols })
  });
  return response.json();
};

// Voorbeeld
const quotes = await fetchMultipleQuotes(['ASML', 'SHELL', 'ING']);
```

### 5. Company Profile

```typescript
// GET /api/v1/stocks/profile/:symbol
interface StockProfile {
  symbol: string;
  company_name: string;
  currency: string;
  exchange: string;
  industry?: string;
  sector?: string;
  website?: string;
  description?: string;
  ceo?: string;
  country?: string;
  ipo_date?: string;
}

const fetchStockProfile = async (symbol: string): Promise<StockProfile> => {
  const response = await fetch(`${API_BASE}/stocks/profile/${symbol}`);
  return response.json();
};
```

## React Components

### StockTickerBadge Component

```tsx
// components/StockTickerBadge.tsx
import React from 'react';
import { StockTicker } from '@/types';

interface Props {
  ticker: StockTicker;
  onClick?: (symbol: string) => void;
}

export const StockTickerBadge: React.FC<Props> = ({ ticker, onClick }) => {
  return (
    <button
      onClick={() => onClick?.(ticker.symbol)}
      className="inline-flex items-center gap-1 px-2 py-1 bg-blue-100 
                 hover:bg-blue-200 text-blue-800 rounded-md text-sm 
                 font-medium transition-colors"
    >
      <span className="font-mono">{ticker.symbol}</span>
      {ticker.exchange && (
        <span className="text-xs text-blue-600">({ticker.exchange})</span>
      )}
    </button>
  );
};
```

### StockQuoteCard Component

```tsx
// components/StockQuoteCard.tsx
import React, { useEffect, useState } from 'react';
import { StockQuote } from '@/types';
import { fetchStockQuote } from '@/api/stocks';

interface Props {
  symbol: string;
}

export const StockQuoteCard: React.FC<Props> = ({ symbol }) => {
  const [quote, setQuote] = useState<StockQuote | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadQuote = async () => {
      try {
        const data = await fetchStockQuote(symbol);
        setQuote(data);
      } catch (error) {
        console.error('Failed to load quote:', error);
      } finally {
        setLoading(false);
      }
    };
    loadQuote();
  }, [symbol]);

  if (loading) {
    return <div className="animate-pulse">Loading quote...</div>;
  }

  if (!quote) {
    return <div className="text-gray-500">Quote not available</div>;
  }

  const isPositive = quote.change >= 0;
  const changeColor = isPositive ? 'text-green-600' : 'text-red-600';
  const changeBg = isPositive ? 'bg-green-50' : 'bg-red-50';

  return (
    <div className="border rounded-lg p-4 space-y-2">
      <div className="flex items-baseline justify-between">
        <div>
          <h3 className="text-lg font-bold">{quote.symbol}</h3>
          <p className="text-sm text-gray-600">{quote.name}</p>
        </div>
        <div className="text-right">
          <div className="text-2xl font-bold">
            {quote.currency === 'USD' ? '$' : '€'}{quote.price.toFixed(2)}
          </div>
          <div className={`text-sm font-medium ${changeColor} ${changeBg} 
                          px-2 py-1 rounded`}>
            {isPositive ? '+' : ''}{quote.change.toFixed(2)} 
            ({isPositive ? '+' : ''}{quote.change_percent.toFixed(2)}%)
          </div>
        </div>
      </div>
      
      <div className="grid grid-cols-2 gap-2 text-sm text-gray-600 pt-2 
                      border-t">
        <div>
          <span className="font-medium">Volume:</span>{' '}
          {quote.volume.toLocaleString()}
        </div>
        <div>
          <span className="font-medium">Exchange:</span> {quote.exchange}
        </div>
        {quote.day_high && (
          <div>
            <span className="font-medium">Day High:</span>{' '}
            {quote.day_high.toFixed(2)}
          </div>
        )}
        {quote.day_low && (
          <div>
            <span className="font-medium">Day Low:</span>{' '}
            {quote.day_low.toFixed(2)}
          </div>
        )}
      </div>
    </div>
  );
};
```

### ArticleStockTickers Component

```tsx
// components/ArticleStockTickers.tsx
import React from 'react';
import { StockTicker } from '@/types';
import { StockTickerBadge } from './StockTickerBadge';
import { useRouter } from 'next/navigation';

interface Props {
  tickers: StockTicker[];
}

export const ArticleStockTickers: React.FC<Props> = ({ tickers }) => {
  const router = useRouter();

  if (!tickers || tickers.length === 0) {
    return null;
  }

  const handleTickerClick = (symbol: string) => {
    router.push(`/stocks/${symbol}`);
  };

  return (
    <div className="flex flex-wrap gap-2">
      <span className="text-sm font-medium text-gray-700">
        📈 Aandelen:
      </span>
      {tickers.map((ticker) => (
        <StockTickerBadge
          key={ticker.symbol}
          ticker={ticker}
          onClick={handleTickerClick}
        />
      ))}
    </div>
  );
};
```

### StockTickerPage Component

```tsx
// app/stocks/[symbol]/page.tsx
import React from 'react';
import { StockQuoteCard } from '@/components/StockQuoteCard';
import { ArticleList } from '@/components/ArticleList';
import { fetchArticlesByTicker, fetchStockProfile } from '@/api/stocks';

interface Props {
  params: { symbol: string };
}

export default async function StockTickerPage({ params }: Props) {
  const { symbol } = params;
  
  // Fetch data server-side
  const [articles, profile] = await Promise.all([
    fetchArticlesByTicker(symbol, 20),
    fetchStockProfile(symbol).catch(() => null)
  ]);

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left sidebar: Stock info */}
        <div className="lg:col-span-1">
          <StockQuoteCard symbol={symbol} />
          
          {profile && (
            <div className="mt-4 border rounded-lg p-4">
              <h3 className="font-bold mb-2">Company Info</h3>
              <dl className="space-y-2 text-sm">
                <div>
                  <dt className="font-medium text-gray-700">Industry</dt>
                  <dd className="text-gray-600">{profile.industry}</dd>
                </div>
                <div>
                  <dt className="font-medium text-gray-700">Sector</dt>
                  <dd className="text-gray-600">{profile.sector}</dd>
                </div>
                {profile.website && (
                  <div>
                    <dt className="font-medium text-gray-700">Website</dt>
                    <dd>
                      <a 
                        href={profile.website}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 hover:underline"
                      >
                        {profile.website}
                      </a>
                    </dd>
                  </div>
                )}
              </dl>
            </div>
          )}
        </div>

        {/* Right content: News articles */}
        <div className="lg:col-span-2">
          <h2 className="text-2xl font-bold mb-4">
            Nieuws over {symbol}
          </h2>
          <ArticleList articles={articles} />
        </div>
      </div>
    </div>
  );
}
```

## Custom Hooks

### useStockQuote Hook

```typescript
// hooks/useStockQuote.ts
import { useState, useEffect } from 'react';
import { StockQuote } from '@/types';
import { fetchStockQuote } from '@/api/stocks';

export const useStockQuote = (symbol: string | null) => {
  const [quote, setQuote] = useState<StockQuote | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    if (!symbol) {
      setQuote(null);
      return;
    }

    setLoading(true);
    setError(null);

    fetchStockQuote(symbol)
      .then(setQuote)
      .catch(setError)
      .finally(() => setLoading(false));
  }, [symbol]);

  return { quote, loading, error };
};

// Gebruik
const MyComponent = () => {
  const { quote, loading, error } = useStockQuote('ASML');
  
  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  if (!quote) return null;
  
  return <div>ASML: €{quote.price}</div>;
};
```

### useArticleStockTickers Hook

```typescript
// hooks/useArticleStockTickers.ts
import { useMemo } from 'react';
import { Article } from '@/types';

export const useArticleStockTickers = (article: Article | null) => {
  const tickers = useMemo(() => {
    return article?.ai_enrichment?.entities?.stock_tickers || [];
  }, [article]);

  const symbols = useMemo(() => {
    return tickers.map(t => t.symbol);
  }, [tickers]);

  return { tickers, symbols };
};
```

## API Service Layer

```typescript
// api/stocks.ts
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export const stocksApi = {
  async getQuote(symbol: string): Promise<StockQuote> {
    const response = await fetch(`${API_BASE}/stocks/quote/${symbol}`, {
      next: { revalidate: 300 } // Cache for 5 minutes
    });
    
    if (!response.ok) {
      throw new Error(`Failed to fetch quote for ${symbol}`);
    }
    
    return response.json();
  },

  async getMultipleQuotes(symbols: string[]): Promise<Record<string, StockQuote>> {
    const response = await fetch(`${API_BASE}/stocks/quotes`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ symbols }),
      next: { revalidate: 300 }
    });
    
    if (!response.ok) {
      throw new Error('Failed to fetch quotes');
    }
    
    return response.json();
  },

  async getProfile(symbol: string): Promise<StockProfile> {
    const response = await fetch(`${API_BASE}/stocks/profile/${symbol}`, {
      next: { revalidate: 86400 } // Cache for 24 hours
    });
    
    if (!response.ok) {
      throw new Error(`Failed to fetch profile for ${symbol}`);
    }
    
    return response.json();
  },

  async getArticlesByTicker(symbol: string, limit: number = 10): Promise<Article[]> {
    const response = await fetch(
      `${API_BASE}/articles/by-ticker/${symbol}?limit=${limit}`,
      { next: { revalidate: 300 } }
    );
    
    if (!response.ok) {
      throw new Error(`Failed to fetch articles for ${symbol}`);
    }
    
    return response.json();
  }
};
```

## Real-time Updates (Optional)

Voor real-time stock updates kun je een polling mechanisme gebruiken:

```typescript
// hooks/useRealtimeStockQuote.ts
import { useState, useEffect, useRef } from 'react';
import { StockQuote } from '@/types';
import { stocksApi } from '@/api/stocks';

export const useRealtimeStockQuote = (
  symbol: string | null,
  intervalMs: number = 30000 // 30 seconds
) => {
  const [quote, setQuote] = useState<StockQuote | null>(null);
  const [loading, setLoading] = useState(false);
  const intervalRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    if (!symbol) return;

    const fetchQuote = async () => {
      try {
        const data = await stocksApi.getQuote(symbol);
        setQuote(data);
      } catch (error) {
        console.error('Failed to fetch quote:', error);
      }
    };

    // Initial fetch
    setLoading(true);
    fetchQuote().finally(() => setLoading(false));

    // Set up polling
    intervalRef.current = setInterval(fetchQuote, intervalMs);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [symbol, intervalMs]);

  return { quote, loading };
};
```

## Performance Optimizations

### 1. Lazy Loading Stock Data

```tsx
import dynamic from 'next/dynamic';

const StockQuoteCard = dynamic(() => import('@/components/StockQuoteCard'), {
  loading: () => <div className="animate-pulse h-32 bg-gray-200 rounded" />,
  ssr: false // Client-side only to avoid SSR overhead
});
```

### 2. Batch Loading

```typescript
// Load multiple quotes in one request
const symbols = ['ASML', 'SHELL', 'ING'];
const quotes = await stocksApi.getMultipleQuotes(symbols);
```

### 3. Client-side Caching

```typescript
// Using React Query
import { useQuery } from '@tanstack/react-query';

export const useStockQuote = (symbol: string) => {
  return useQuery({
    queryKey: ['stock', symbol],
    queryFn: () => stocksApi.getQuote(symbol),
    staleTime: 5 * 60 * 1000, // 5 minutes
    cacheTime: 10 * 60 * 1000, // 10 minutes
  });
};
```

## Complete Example: Article Detail Page

```tsx
// app/articles/[id]/page.tsx
import { ArticleStockTickers } from '@/components/ArticleStockTickers';
import { StockQuoteCard } from '@/components/StockQuoteCard';

export default async function ArticleDetailPage({ params }: { params: { id: string } }) {
  const article = await fetchArticle(params.id);
  const tickers = article.ai_enrichment?.entities?.stock_tickers || [];

  return (
    <div className="container mx-auto px-4 py-8">
      <article className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-4">{article.title}</h1>
        
        {/* Stock Tickers */}
        {tickers.length > 0 && (
          <div className="mb-6">
            <ArticleStockTickers tickers={tickers} />
          </div>
        )}
        
        {/* Article Content */}
        <div className="prose max-w-none">
          <p>{article.summary}</p>
          {article.content && <div dangerouslySetInnerHTML={{ __html: article.content }} />}
        </div>
        
        {/* Stock Quotes Sidebar */}
        {tickers.length > 0 && (
          <aside className="mt-8 grid grid-cols-1 md:grid-cols-2 gap-4">
            {tickers.slice(0, 4).map((ticker) => (
              <StockQuoteCard key={ticker.symbol} symbol={ticker.symbol} />
            ))}
          </aside>
        )}
      </article>
    </div>
  );
}
```

## Styling Tips

### Tailwind Classes voor Stock Components

```css
/* Positive/Negative Indicators */
.stock-positive {
  @apply text-green-600 bg-green-50;
}

.stock-negative {
  @apply text-red-600 bg-red-50;
}

/* Ticker Badge */
.ticker-badge {
  @apply inline-flex items-center px-3 py-1 rounded-full text-sm font-medium;
  @apply bg-blue-100 text-blue-800 hover:bg-blue-200 transition-colors;
}

/* Quote Card */
.quote-card {
  @apply border rounded-lg p-4 shadow-sm hover:shadow-md transition-shadow;
}
```

## Testing

```typescript
// __tests__/StockQuoteCard.test.tsx
import { render, screen, waitFor } from '@testing-library/react';
import { StockQuoteCard } from '@/components/StockQuoteCard';
import { stocksApi } from '@/api/stocks';

jest.mock('@/api/stocks');

describe('StockQuoteCard', () => {
  it('renders stock quote data', async () => {
    const mockQuote = {
      symbol: 'ASML',
      name: 'ASML Holding',
      price: 745.30,
      change: 12.50,
      change_percent: 1.71,
      // ...
    };

    (stocksApi.getQuote as jest.Mock).mockResolvedValue(mockQuote);

    render(<StockQuoteCard symbol="ASML" />);

    await waitFor(() => {
      expect(screen.getByText('ASML')).toBeInTheDocument();
      expect(screen.getByText('€745.30')).toBeInTheDocument();
      expect(screen.getByText('+12.50 (+1.71%)')).toBeInTheDocument();
    });
  });
});
```

## Resources

- [Next.js Data Fetching](https://nextjs.org/docs/app/building-your-application/data-fetching)
- [React Query](https://tanstack.com/query/latest)
- [TailwindCSS](https://tailwindcss.com/docs)
- [Backend API Docs](../features/stock-tickers.md)