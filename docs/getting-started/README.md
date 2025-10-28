# Getting Started met IntelliNieuws

Welkom! Deze gids helpt je om snel aan de slag te gaan met IntelliNieuws.

## 🎯 Wat is IntelliNieuws?

Een intelligente, AI-verrijkte nieuws aggregator voor Nederlandse nieuwsbronnen, gebouwd met Go en verrijkt met geavanceerde sentiment analyse, entity extraction en trending topic detection.

### Belangrijkste Features
- ✅ **Ethisch Scrapen** - Respecteert robots.txt en rate limiting
- ✅ **RSS Prioriteit** - Gebruikt RSS feeds waar mogelijk
- ✅ **AI-Verrijking** - Sentiment analyse, entity extraction, categorisatie
- ✅ **REST API** - Volledige CRUD operaties
- ✅ **Multi-layer Caching** - Redis + in-memory + materialized views
- ✅ **99.5% Uptime** - Circuit breakers en auto-recovery

## 🚀 Snelstart Opties

### Voor Backend Developers
Start hier: [Quick Start Guide](quick-start.md)  
→ In 5 minuten een werkende backend

### Voor Frontend Developers  
Start hier: [Frontend Quick Start](../frontend/quickstart.md)  
→ Integreer met de REST API

### Voor Operations
Start hier: [Deployment Guide](../deployment/deployment-guide.md)  
→ Production deployment

## 📋 Prerequisites

### Software Vereisten
- **Go 1.22+** - [Download](https://golang.org/dl/)
- **PostgreSQL 16+** - [Download](https://www.postgresql.org/download/)
- **Redis 7+** - Optioneel maar aanbevolen

### Accounts
- **OpenAI API Key** - Voor AI features ([Verkrijg hier](https://platform.openai.com/))

## 🏗️ Architectuur Overzicht

```
┌─────────────────────────────────────────────┐
│              CLIENT REQUESTS                 │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│         API LAYER (Fiber/Go)                │
│  • Articles  • AI Features  • Health        │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│          CACHING LAYER                      │
│  • Redis (API)  • In-Memory (AI)            │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│         SERVICE LAYER                       │
│  • Scraper  • AI Processor  • Scheduler    │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│       DATABASE (PostgreSQL)                 │
│  • Articles  • Materialized Views           │
└─────────────────────────────────────────────┘
```

## 📖 Documentatie Structuur

- **[Installation](installation.md)** - Gedetailleerde installatie instructies
- **[Quick Start](quick-start.md)** - 5-minuten setup
- **[Windows Setup](windows-setup.md)** - Windows-specifieke instructies

## 🎓 Learning Path

1. **Beginner** → Start met [Quick Start](quick-start.md)
2. **Intermediate** → Lees [Installation Guide](installation.md)
3. **Advanced** → Bestudeer [Architecture](../development/architecture.md)

## 💡 Volgende Stappen

Na de installatie:
1. Verken de [API Documentation](../api/README.md)
2. Bekijk [AI Features](../features/ai-processing.md)
3. Bouw een [Frontend](../frontend/README.md)

## 📞 Hulp Nodig?

- [Troubleshooting Guide](../operations/troubleshooting.md)
- [FAQ](installation.md#frequently-asked-questions)
- GitHub Issues

---

**Klaar om te beginnen?** → [Quick Start Guide](quick-start.md)