# NieuwsScraper v2.0 - Quick Reference Card

**⚡ Snelle Referentie voor Operations & Monitoring**

---

## 🏥 Health Checks

### Check System Status
```bash
curl http://localhost:8080/health
```

**Expected:** `"status": "healthy"`  
**Action if degraded:** Check component details in response

### Quick Health Check
```bash
curl http://localhost:8080/health/live
```

**Expected:** `"status": "alive"`  
**Action if not responding:** Service is down, restart needed

### Readiness Check
```bash
curl http://localhost:8080/health/ready
```

**Expected:** `"status": "ready"`  
**Action if not ready:** Check database/Redis connection

---

## 📊 Key Metrics

### Cache Performance
```bash
curl http://localhost:8080/health/metrics | jq '.data'
```

**Key Indicators:**
- `cache_hit_rate` > 40% ✅
- `cache_hit_rate` < 20% ⚠️
- `cache_hit_rate` < 10% 🚨

### Processing Stats
```bash
curl http://localhost:8080/api/v1/ai/processor/stats
```

**Key Indicators:**
- `is_running: true` ✅
- `consecutive_errors: 0` ✅
- `consecutive_errors > 3` ⚠️
- `consecutive_errors > 10` 🚨

### Database Performance
```sql
-- Connection pool usage
SELECT 
    numbackends as connections,
    (SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'idle') as idle,
    (SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'active') as active
FROM pg_stat_database WHERE datname = 'nieuws_scraper';
```

**Healthy:** connections < 20, idle > 5, active < 15

---

## 🔄 Common Operations

### Refresh Materialized View (Manual)
```powershell
.\scripts\refresh-materialized-views.ps1
```

**When:** If trending topics seem stale  
**Frequency:** Automated every 10 minutes  
**Duration:** ~5-10 seconds

### Trigger AI Processing
```bash
curl -X POST http://localhost:8080/api/v1/ai/process/trigger \
  -H "X-API-Key: your-key"
```

**When:** Need immediate processing  
**Expected:** Process batch in 5-15 seconds

### Clear Cache
```bash
redis-cli FLUSHDB
```

**When:** After major data changes  
**Impact:** Temporary performance hit until cache rebuilds

### Check Circuit Breakers
```bash
curl http://localhost:8080/api/v1/scraper/stats | jq '.data.circuit_breakers'
```

**Status:**
- `"state": "closed"` - Normal ✅
- `"state": "half-open"` - Recovering ⚠️
- `"state": "open"` - Blocked 🚨

---

## 🚨 Alerts & Responses

### Critical Alerts

#### Database Unavailable
```
Symptom: /health returns "unhealthy"
Action: 
  1. Check PostgreSQL status
  2. Verify connection string
  3. Check firewall rules
```

#### High Error Rate (>10%)
```
Symptom: consecutive_errors > 10
Action:
  1. Check OpenAI API status
  2. Review error logs
  3. System will auto-recover with backoff
```

#### All Circuit Breakers Open
```
Symptom: All sources showing "state": "open"
Action:
  1. Check source availability
  2. Review robots.txt compliance
  3. Wait 5 minutes for auto-recovery
```

### Warning Alerts

#### Low Cache Hit Rate (<20%)
```
Action:
  1. Verify Redis is running
  2. Check cache configuration
  3. Monitor for cache evictions
```

#### Slow API Responses (>500ms)
```
Action:
  1. Check database performance
  2. Refresh materialized view
  3. Check cache availability
```

---

## 📈 Performance Optimization Tips

### Improve Cache Hit Rate
1. ✅ Increase Redis memory
2. ✅ Extend cache TTL (5→10 min)
3. ✅ Pre-warm cache on deployment

### Reduce API Costs
1. ✅ Monitor OpenAI calls in logs
2. ✅ Check cache is working
3. ✅ Verify no duplicate requests

### Speed Up Processing
1. ✅ Increase worker count (4→8)
2. ✅ Reduce batch size for faster feedback
3. ✅ Optimize database queries

### Improve Database Performance
1. ✅ Run VACUUM ANALYZE weekly
2. ✅ Monitor slow queries
3. ✅ Refresh materialized view more frequently

---

## 🎯 Target Metrics (Production)

### Must Achieve
- ✅ API response time p95 < 200ms
- ✅ Cache hit rate > 40%
- ✅ Success rate > 99%
- ✅ Error rate < 2%

### Should Achieve
- ✅ Database query time < 100ms
- ✅ Processing > 50 articles/min
- ✅ OpenAI cost < $400/month
- ✅ Uptime > 99.5%

### Nice to Have
- 🎯 Cache hit rate > 60%
- 🎯 Processing > 80 articles/min
- 🎯 OpenAI cost < $300/month
- 🎯 Uptime > 99.9%

---

## 🔍 Quick Diagnostics

### System Running Slow?
```bash
# 1. Check cache
curl /health/metrics | jq '.data.cache_hit_rate'

# 2. Check database
psql -c "SELECT COUNT(*) FROM pg_stat_activity;"

# 3. Check materialized view
psql -c "SELECT COUNT(*) FROM mv_trending_keywords;"
```

### High Costs?
```bash
# 1. Check cache hit rate
curl /health/metrics | jq '.data.cache_*'

# 2. Check for duplicate processing
tail -f logs/app.log | grep "Cache MISS"

# 3. Review OpenAI usage
# Check OpenAI dashboard
```

### Processing Failing?
```bash
# 1. Check processor status
curl /api/v1/ai/processor/stats

# 2. Check consecutive errors
# If > 5, system is in backoff mode (auto-recovery)

# 3. Check OpenAI API status
curl https://status.openai.com/api/v2/status.json
```

---

## 🛠️ Maintenance Schedule

### Every 10 Minutes (Automated)
- ✅ Refresh materialized views
- ✅ AI processor runs (if queue not empty)

### Every Hour
- ✅ Check health endpoint
- ✅ Monitor error logs

### Daily
- ✅ Review OpenAI costs
- ✅ Check cache hit rates
- ✅ Verify all circuit breakers closed

### Weekly
- ✅ Database VACUUM ANALYZE
- ✅ Review slow queries
- ✅ Cost trend analysis
- ✅ Performance benchmarks

### Monthly
- ✅ Capacity planning
- ✅ Security updates
- ✅ Documentation updates
- ✅ Team review

---

## 📞 Emergency Contacts

### Service Down
1. Check `/health/live` endpoint
2. Review application logs
3. Restart service if needed
4. Check database connectivity

### High Costs
1. Check cache hit rate
2. Verify no infinite loops
3. Check for duplicate requests
4. Review OpenAI API usage

### Data Issues
1. Check database integrity
2. Verify migrations ran correctly
3. Review scraping logs
4. Check duplicate detection

---

## 🎓 Training Checklist

### For New Team Members
- [ ] Read [`README.md`](README.md:1)
- [ ] Review [`DEPLOYMENT_GUIDE.md`](DEPLOYMENT_GUIDE.md:1)
- [ ] Study [`EXECUTIVE_SUMMARY.md`](EXECUTIVE_SUMMARY.md:1)
- [ ] Practice health check commands
- [ ] Test rollback procedure
- [ ] Review monitoring dashboard

### For Operations
- [ ] Setup monitoring alerts
- [ ] Configure Task Scheduler
- [ ] Test backup procedures
- [ ] Practice troubleshooting scenarios
- [ ] Document runbooks

---

## 📚 Documentation Index

**For Management:**
- [`EXECUTIVE_SUMMARY.md`](EXECUTIVE_SUMMARY.md:1) - ROI & business impact

**For Deployment:**
- [`DEPLOYMENT_GUIDE.md`](DEPLOYMENT_GUIDE.md:1) - Step-by-step deployment
- [`CHANGELOG_v2.0.md`](CHANGELOG_v2.0.md:1) - What's new in v2.0

**For Developers:**
- [`AGENT_OPTIMIZATIONS.md`](AGENT_OPTIMIZATIONS.md:1) - Technical details
- [`FINAL_IMPLEMENTATION_REPORT.md`](FINAL_IMPLEMENTATION_REPORT.md:1) - Complete report

**For Operations:**
- [`QUICK_REFERENCE.md`](QUICK_REFERENCE.md:1) - This document
- Health endpoints: `/health`, `/health/metrics`

---

**Last Updated:** 2025-10-28  
**Version:** 2.0  
**Status:** Production Ready ✅

**Print this card and keep it handy! 📋**