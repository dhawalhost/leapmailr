# New Relic APM Integration Guide

## Overview
New Relic APM (Application Performance Monitoring) has been integrated into LeapMailR to provide:
- Real-time application performance monitoring
- Distributed tracing across requests
- Error tracking and analysis
- Transaction performance metrics
- Log forwarding to New Relic

## Quick Start

### 1. Get Your New Relic License Key

1. Sign up for New Relic: https://newrelic.com/signup
2. Free tier includes:
   - 100 GB data ingest/month
   - 1 full platform user
   - Unlimited basic users
3. Get your license key from: **Account Settings → API Keys → License Key**

### 2. Configure Environment

Add to your `.env` file:
```bash
NR_LICENSE_KEY=your_license_key_here
```

Or set as environment variable:
```bash
export NR_LICENSE_KEY=your_license_key_here
```

### 3. Verify Integration

Run the application:
```bash
go run .
```

You should see in the logs:
```
INFO    Starting LeapMailR API Server
INFO    Initializing New Relic APM {"app_name": "leapmailr", "log_forwarding": true}
INFO    New Relic APM initialized successfully
INFO    New Relic APM integration enabled
```

### 4. View in New Relic

1. Log into New Relic One: https://one.newrelic.com
2. Navigate to **APM & Services**
3. Find your app: **leapmailr**
4. Wait 1-2 minutes for data to appear

## Features Enabled

### ✅ Transaction Tracking
Every HTTP request is tracked as a transaction:
- URL and method
- Response time
- Status code
- Throughput

### ✅ Distributed Tracing
Trace requests across:
- API handlers
- Database queries
- External service calls
- Email operations

### ✅ Error Tracking
Automatic capture of:
- Panics and errors
- Stack traces
- Error rates
- Failed transactions

### ✅ Log Forwarding
Application logs sent to New Relic:
- Structured JSON logs
- Correlation with transactions
- Search and filter capabilities

### ✅ Custom Metrics
Application-specific metrics:
- Email sent counts
- Authentication events
- Rate limit hits
- Business metrics

## Running Without New Relic

**The application works perfectly without New Relic configured!**

If `NR_LICENSE_KEY` is not set:
- Application starts normally
- Warning logged: "New Relic license key not configured - APM monitoring disabled"
- No-op middleware used (zero overhead)
- All functionality works as expected

## New Relic Dashboard

### Key Metrics to Monitor

1. **Web Transactions**
   - Response time percentiles (p50, p95, p99)
   - Throughput (requests per minute)
   - Error rate

2. **Transactions**
   - Slowest endpoints
   - Database query performance
   - External service calls

3. **Errors**
   - Error rate trends
   - Error classes and messages
   - Stack traces

4. **Databases**
   - Query performance
   - Slow queries
   - Database time

5. **Logs**
   - Application logs with context
   - Error logs
   - Correlation with transactions

## Advanced Configuration

### Custom Transaction Names

In your handlers, you can customize transaction names:
```go
txn := nrgin.Transaction(c)
if txn != nil {
    txn.SetName("CustomName")
}
```

### Record Custom Events

```go
import "github.com/newrelic/go-agent/v3/newrelic"

txn := nrgin.Transaction(c)
if txn != nil {
    txn.Application().RecordCustomEvent("UserAction", map[string]interface{}{
        "action": "email_sent",
        "user_id": userID,
        "count": 1,
    })
}
```

### Add Custom Attributes

```go
txn := nrgin.Transaction(c)
if txn != nil {
    txn.AddAttribute("userID", userID)
    txn.AddAttribute("emailService", serviceName)
}
```

### Notice Errors Manually

```go
txn := nrgin.Transaction(c)
if txn != nil {
    txn.NoticeError(err)
}
```

## Environment Variables

All New Relic configuration:

```bash
# Required
NR_LICENSE_KEY=your_license_key_here

# Optional (set in code, can override)
NEW_RELIC_APP_NAME=leapmailr                    # Default: leapmailr
NEW_RELIC_LOG_FORWARDING_ENABLED=true           # Default: true
NEW_RELIC_DISTRIBUTED_TRACING_ENABLED=true      # Default: true
```

## Monitoring Stack Integration

New Relic complements your existing monitoring:

### Works Alongside Prometheus
- **Prometheus**: Infrastructure metrics, custom business metrics
- **New Relic**: APM, distributed tracing, error tracking, logs

### Multi-Platform Strategy
You can run both simultaneously:
```
Application Metrics → Prometheus (port 9090)
                   → New Relic APM
                   
Application Logs   → JSON files (logs/)
                   → New Relic Logs
```

## Cost Management

### Free Tier Limits
- 100 GB data ingest/month
- 1 full platform user
- Unlimited basic users
- Data retention: 8 days

### Tips to Stay Within Free Tier
1. **Sample transactions** (for high-traffic apps)
2. **Filter logs** (send only errors/warnings)
3. **Monitor usage** in NR account settings
4. **Use custom sampling rates** if needed

### Upgrade When Needed
- Standard: $99/user/month (starts at $0/month with free tier)
- Pro: $349/user/month
- Enterprise: Custom pricing

## Troubleshooting

### "New Relic license key not configured"
**Solution**: Set `NR_LICENSE_KEY` environment variable

### No data in New Relic dashboard
**Checks**:
1. Wait 1-2 minutes for data to appear
2. Verify license key is correct
3. Check application logs for errors
4. Ensure firewall allows HTTPS to `collector.newrelic.com`

### License key invalid error
**Solution**: 
1. Verify license key from New Relic account
2. Ensure no extra spaces or newlines
3. Check it's a **License Key**, not an API Key

### High data usage
**Solutions**:
1. Enable transaction sampling
2. Filter logs (errors only)
3. Reduce trace sampling rate

## Migration from Monitoring Stack

If you want to switch from self-hosted Prometheus/Grafana:

### Option 1: Full Migration to New Relic
1. Enable New Relic (done ✅)
2. Test for 1-2 weeks alongside Prometheus
3. Create dashboards in New Relic
4. Migrate alerting rules
5. Deprecate Prometheus/Grafana

### Option 2: Hybrid Approach (Recommended)
1. Keep Prometheus for infrastructure metrics
2. Use New Relic for APM and logs
3. Best of both worlds

## Security Considerations

### License Key Protection
- Never commit license key to Git
- Use environment variables or secrets management
- Rotate keys periodically
- Use separate keys for dev/staging/production

### Data Privacy
- New Relic stores: metrics, traces, logs, errors
- Configure data retention policies
- Enable data obfuscation if needed
- Review New Relic security features

### Compliance
- SOC 2 Type II certified
- GDPR compliant
- HIPAA eligible (Enterprise tier)
- ISO 27001 certified

## Next Steps

1. **Create Dashboards**
   - Navigate to New Relic → Dashboards → Create Dashboard
   - Add widgets for key metrics
   - Share with team

2. **Set Up Alerts**
   - APM → Alert Conditions → Create Alert
   - Set thresholds (response time, error rate, throughput)
   - Configure notification channels (email, Slack, PagerDuty)

3. **Optimize Performance**
   - Use transaction traces to find slow queries
   - Identify N+1 query problems
   - Optimize database indexes

4. **Custom Instrumentation**
   - Add custom events for business metrics
   - Track specific user actions
   - Monitor background jobs

## Support

- **Documentation**: https://docs.newrelic.com/docs/apm/agents/go-agent/
- **Go Agent GitHub**: https://github.com/newrelic/go-agent
- **New Relic Support**: https://support.newrelic.com
- **Community Forum**: https://discuss.newrelic.com

## Summary

✅ New Relic APM is now integrated and ready to use
✅ Application works with or without New Relic configured
✅ Zero code changes needed to enable/disable
✅ Comprehensive monitoring across all layers
✅ Compatible with existing Prometheus monitoring

**To enable**: Just add your `NR_LICENSE_KEY` and restart the app!
