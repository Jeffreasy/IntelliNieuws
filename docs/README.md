# IntelliNieuws Documentation

Welkom bij de officiële documentatie van IntelliNieuws - een intelligente, AI-verrijkte nieuws aggregator voor Nederlandse nieuwsbronnen met geavanceerde sentiment analyse, trending topic detection, en **production-ready Docker deployment**.

## 🆕 Nieuw in v2.2

- 🐳 **[Docker Setup (Production-Ready)](docker-setup.md)** - Complete containerized deployment met auto-backup
- 🔒 **Security Hardening** - Redis authentication, geen hardcoded credentials
- 💾 **Automatische Backups** - Dagelijkse database backups (7 dagen retentie)
- ⚡ **Redis Optimalisaties** - Connection pooling (20 conns) & persistence
- 📊 **Resource Management** - CPU/Memory limits per service
- 🛠️ **Cache Invalidation** - Intelligente cache management service

## 📚 Documentatie Overzicht

### 🚀 Getting Started
Kom snel op gang met IntelliNieuws
- **[Docker Setup](docker-setup.md)** - **✨ Aanbevolen: Production-ready deployment**
- [Quick Start Guide](getting-started/quick-start.md) - In 5 minuten aan de slag
- [Installation Guide](getting-started/installation.md) - Complete installatie instructies
- [Windows Setup](getting-started/windows-setup.md) - Windows-specifieke setup

### 🌐 API Documentation
Complete REST API referentie
- [API Overview](api/README.md) - API architectuur en overzicht
- [Endpoints Reference](api/endpoints.md) - Alle beschikbare endpoints
- [Authentication](api/authentication.md) - API keys en beveiliging
- [Examples](api/examples.md) - Praktische code voorbeelden

### 💻 Frontend Integration
Integreer met je frontend applicatie
- [Frontend Guide](frontend/README.md) - Complete integratie gids
- [Quick Start](frontend/quickstart.md) - Basis implementatie
- [AI Features](frontend/ai-features.md) - AI endpoints gebruiken
- [Stock Tickers Integration](frontend/stock-tickers-integration.md) - 📈 Aandelen integratie
- [Advanced Patterns](frontend/advanced-patterns.md) - Production-ready patterns

### ⚙️ Features & Capabilities
Diepgaande feature documentatie
- [AI Processing](features/ai-processing.md) - Sentiment, entities, keywords
- [Stock Tickers](features/stock-tickers.md) - 📈 Aandelen extraction & API integratie
- [Scraping Strategies](features/scraping.md) - RSS, HTML, Browser scraping
- [Content Extraction](features/content-extraction.md) - Volledige artikel tekst
- [Caching System](features/caching.md) - Multi-layer caching
- [Chat API](features/chat-api.md) - Conversational AI interface

### 🚀 Deployment & Infrastructure
Production deployment en configuratie
- **[Docker Setup](docker-setup.md)** - **✨ Preferred: Docker Compose deployment**
- [Deployment Guide](deployment/deployment-guide.md) - Stap-voor-stap manual deployment
- [Configuration](deployment/configuration.md) - Environment variables
- [Monitoring](deployment/monitoring.md) - Health checks en metrics
- [Maintenance](deployment/maintenance.md) - Dagelijkse operaties

### 🏗️ Development
Voor developers die bijdragen
- [Architecture](development/architecture.md) - System design en componenten
- [Optimizations](development/optimizations.md) - Performance verbeteringen
- [Agent Mapping](development/agents-mapping.md) - Service interacties
- [Contributing](development/contributing.md) - Bijdragen aan het project

### 🛠️ Operations
Dagelijkse operaties en troubleshooting
- [Quick Reference](operations/quick-reference.md) - Handige commando's
- [Troubleshooting](operations/troubleshooting.md) - Problemen oplossen
- [Scripts Guide](operations/scripts.md) - Utility scripts uitleg
- [Performance Tuning](operations/performance-tuning.md) - Optimalisatie tips

### ⚖️ Legal & Compliance
Juridische aspecten en naleving
- [Compliance Guide](legal/compliance.md) - Robots.txt en juridische richtlijnen

### 📝 Changelog
Versie geschiedenis en wijzigingen
- [Version 2.0](changelog/v2.0.md) - Optimalisatie release
- [Migration Guide](changelog/migration-guide.md) - Upgrade van v1 naar v2

## 🎯 Quick Navigation

### Ik wil...
- **Snel starten met Docker** → [Docker Setup](docker-setup.md) **✨ Nieuw!**
- **Lokaal starten** → [Quick Start](getting-started/quick-start.md)
- **API gebruiken** → [API Reference](api/endpoints.md)
- **Frontend bouwen** → [Frontend Guide](frontend/README.md)
- **Deployen** → [Docker Setup](docker-setup.md) of [Manual Deployment](deployment/deployment-guide.md)
- **Troubleshooting** → [Operations Guide](operations/troubleshooting.md)

### Voor verschillende rollen
- **👨‍💼 Management** → [Executive Summary](deployment/deployment-guide.md#executive-summary)
- **👨‍💻 Frontend Developer** → [Frontend Integration](frontend/README.md)
- **🔧 Backend Developer** → [Architecture](development/architecture.md)
- **🛠️ Operations** → [Quick Reference](operations/quick-reference.md)
- **📊 DevOps** → [Docker Setup](docker-setup.md) **✨ Start hier!**

## 📊 Project Status

**Versie:** 2.2 (Docker & Security Edition) **✨ Nieuw!**
**Status:** Production Ready ✅
**Performance:** 8x sneller dan v1.0
**Kosten:** 50-60% reductie
**Reliability:** 99.5% uptime
**Security:** Hardened with best practices
**Infrastructure:** Fully containerized with Docker

## 🤝 Support

Voor vragen en problemen:
1. Raadpleeg de relevante documentatie sectie
2. Check de [Troubleshooting Guide](operations/troubleshooting.md)
3. Bekijk de logs met de health endpoints
4. Open een GitHub issue voor bugs

## 📄 License

MIT License - zie LICENSE file voor details

---

**Made with ❤️ voor de Nederlandse nieuwsaggregatie gemeenschap**