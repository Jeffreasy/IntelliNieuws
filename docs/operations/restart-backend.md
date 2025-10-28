# Backend Herstarten voor Chat API

De chat endpoint is toegevoegd maar de backend moet opnieuw gebuild worden.

## Windows (PowerShell):

```powershell
# 1. Stop de huidige backend (Ctrl+C in de terminal)

# 2. Build opnieuw
go build -o api.exe ./cmd/api

# 3. Start de backend
.\api.exe
```

## Of gebruik de bestaande scripts:

```powershell
# In de project root directory:
.\scripts\start.ps1
```

## Verificatie:

Na het starten zou je moeten zien:
```
INFO  AI chat service initialized
INFO  AI service initialized successfully
INFO  Starting API server on :8080
```

Test vervolgens de chat endpoint:
```bash
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "Hallo, wat kan je voor me doen?"}'
```

Of gebruik de frontend chat modal - deze zou nu moeten werken! 🚀