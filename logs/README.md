# Logs Directory

This directory contains application logs for LeapMailR.

## Log Files

- `application.log` - Main application logs (JSON format)
- Rotated files: `application-YYYY-MM-DD.log.gz`

## Configuration

Logs are automatically rotated based on:
- **Max Size**: 100MB per file
- **Max Backups**: 10 files
- **Max Age**: 30 days
- **Compression**: Enabled (gzip)

## Log Levels

- `DEBUG` - Detailed debugging information
- `INFO` - General informational messages
- `WARN` - Warning messages
- `ERROR` - Error messages with stack traces

## Environment Variables

- `LOG_LEVEL` - Set log level (debug, info, warn, error)
  - Default: `info`
  - Production: `warn` or `error`
  - Development: `debug`

## Centralized Logging (GAP-SEC-009)

Logs are output to both:
1. **File** - Rotated log files in this directory
2. **stdout** - For container environments (ELK Stack, Grafana Loki)

## Log Format

Logs use structured JSON format with fields:
```json
{
  "timestamp": "2025-11-04T10:30:45.123Z",
  "level": "info",
  "logger": "http",
  "caller": "middleware/logger.go:45",
  "message": "Request completed",
  "correlation_id": "uuid-here",
  "request_id": "uuid-here",
  "method": "POST",
  "path": "/api/v1/auth/login",
  "status": 200,
  "latency": "0.045s",
  "client_ip": "192.168.1.1",
  "user_id": "user-uuid"
}
```

## Security

- Sensitive data is automatically redacted
- Passwords, tokens, API keys are never logged
- Email addresses are partially masked
- Stack traces included for errors

## Monitoring Integration

Logs can be shipped to:
- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Grafana Loki**
- **Splunk**
- **Datadog**

Configure log forwarders to read from stdout or this directory.
