# Quick Reference Guide

Fast lookup reference for LeapMailr features.

---

## 🔐 Authentication

```javascript
// Method 1: JWT (Dashboard)
headers: { 'Authorization': 'Bearer <jwt_token>' }

// Method 2: API Key Pair (Recommended for SDK)
headers: {
  'X-Public-Key': 'pk_live_xxxxx',
  'X-Private-Key': 'sk_live_xxxxx'
}

// Method 3: Legacy API Key
headers: { 'X-API-Key': '<legacy_key>' }
```

---

## 📧 Send Email

### Basic Send
```javascript
POST /api/v1/email/send
{
  "to_email": "user@example.com",
  "template_id": "uuid",
  "template_params": { "name": "John" }
}
```

### Form Send (with all features)
```javascript
POST /api/v1/email/send-form
{
  "template_id": "uuid",
  "template_params": {
    "to_email": "user@example.com",
    "name": "John Doe",
    "message": "Hello"
  },
  "captcha_response": "token",    // Optional
  "collect_contact": true         // Optional (default: true)
}
```

---

## 🛡️ CAPTCHA

### Setup
1. Dashboard → Settings → CAPTCHA
2. Add config (reCAPTCHA v2 or hCaptcha)
3. Copy site key to frontend

### Frontend
```html
<div class="g-recaptcha" data-sitekey="YOUR_SITE_KEY"></div>
<script src="https://www.google.com/recaptcha/api.js"></script>
```

```javascript
const token = grecaptcha.getResponse();
// Pass token in captcha_response field
```

---

## 🚫 Suppressions

### Add Single
```bash
POST /api/v1/suppressions
{
  "email": "bounce@example.com",
  "reason": "bounce"  # bounce|complaint|unsubscribe|manual
}
```

### Add Bulk
```bash
POST /api/v1/suppressions/bulk
{
  "emails": ["email1@example.com", "email2@example.com"],
  "reason": "unsubscribe"
}
```

### Check
```bash
GET /api/v1/suppressions/check?email=test@example.com
```

### Webhook URLs
- SendGrid: `https://yourdomain.com/api/v1/webhooks/sendgrid`
- Mailgun: `https://yourdomain.com/api/v1/webhooks/mailgun`

---

## 🔁 Auto-Reply

### Create Config
```javascript
POST /api/v1/autoreplies
{
  "template_id": "uuid",
  "from_email": "noreply@example.com",
  "subject": "Thanks {{name}}!",
  "triggers": ["form", "api"],
  "delay_seconds": 0,
  "is_active": true
}
```

### Variables
Template: `Hi {{name}}, we got your {{subject}}`
Variables from `template_params` are auto-replaced

---

## 🔑 API Keys

### Generate
```bash
POST /api/v1/api-keys
{
  "name": "Production",
  "rate_limit": 100,
  "permissions": ["send_email", "send_form"]
}

Response:
{
  "public_key": "pk_live_xxxxx",
  "private_key": "sk_live_xxxxx"  # Save this! Only shown once
}
```

### Usage
```javascript
headers: {
  'X-Public-Key': 'pk_live_xxxxx',
  'X-Private-Key': 'sk_live_xxxxx'
}
```

### Rotate
```bash
POST /api/v1/api-keys/:id/rotate
```

---

## 👥 Contacts

### List
```bash
GET /api/v1/contacts?search=john&source=form&is_subscribed=true
```

### Create
```bash
POST /api/v1/contacts
{
  "email": "user@example.com",
  "name": "John Doe",
  "tags": ["customer", "vip"],
  "metadata": { "source": "landing-page" }
}
```

### Import CSV
```bash
POST /api/v1/contacts/import
Content-Type: multipart/form-data

File format:
email,name,phone,company,tags
john@example.com,John Doe,+1234567890,Acme,"tag1,tag2"
```

### Export CSV
```bash
GET /api/v1/contacts/export?is_subscribed=true
```

### Statistics
```bash
GET /api/v1/contacts/stats
```

---

## 📊 Common Queries

### Get All Templates
```bash
GET /api/v1/templates
```

### Get Email Services
```bash
GET /api/v1/email-services
```

### Get Email History
```bash
GET /api/v1/emails?limit=100&status=sent
```

### Get Auto-Reply Logs
```bash
GET /api/v1/autoreplies/logs?auto_reply_id=uuid
```

---

## ⚡ Error Codes

| Code | Meaning | Common Cause |
|------|---------|--------------|
| 400 | Bad Request | Missing required fields |
| 401 | Unauthorized | Invalid/missing auth |
| 403 | Forbidden | Rate limit exceeded |
| 404 | Not Found | Resource doesn't exist |
| 422 | Unprocessable | Email suppressed/invalid |
| 429 | Too Many Requests | Rate limit hit |
| 500 | Server Error | Internal error |

---

## 🎯 Frontend Routes

| Page | Route |
|------|-------|
| Dashboard | `/dashboard` |
| Send Email | `/dashboard/send` |
| Templates | `/dashboard/templates` |
| Services | `/dashboard/services` |
| Analytics | `/dashboard/analytics` |
| CAPTCHA | `/dashboard/settings/captcha` |
| Suppressions | `/dashboard/settings/suppressions` |
| Auto-Reply | `/dashboard/settings/auto-reply` |
| API Keys | `/dashboard/settings/api-keys` |
| Contacts | `/dashboard/contacts` |

---

## 🔧 Environment Variables

### Backend (.env)
```bash
PORT=8080
DATABASE_URL=postgresql://user:pass@localhost:5432/leapmailr
JWT_SECRET=your-secret-key
ENV_MODE=development  # or "release"
```

### Frontend (.env.local)
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_LEAPMAILR_PUBLIC_KEY=pk_live_xxxxx
LEAPMAILR_PRIVATE_KEY=sk_live_xxxxx
```

---

## 🚀 Quick Start Commands

### Backend
```bash
cd leapmailr
go build
./leapmailr
```

### Frontend
```bash
cd leapmailr-ui
npm install
npm run dev      # Development
npm run build    # Production build
npm start        # Production serve
```

### Database
```bash
# Create database
createdb leapmailr

# Migrations run automatically on server start

# Backup
pg_dump leapmailr > backup.sql

# Restore
psql leapmailr < backup.sql
```

---

## 📱 SDK Example

```javascript
// Initialize
const leapmailr = {
  publicKey: 'pk_live_xxxxx',
  privateKey: 'sk_live_xxxxx',
  apiUrl: 'https://api.leapmailr.com'
};

// Send email
async function sendEmail(templateId, params) {
  const response = await fetch(`${leapmailr.apiUrl}/api/v1/email/send-form`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Public-Key': leapmailr.publicKey,
      'X-Private-Key': leapmailr.privateKey
    },
    body: JSON.stringify({
      template_id: templateId,
      template_params: params
    })
  });
  
  return await response.json();
}

// Usage
await sendEmail('contact-form', {
  to_email: 'support@company.com',
  name: 'John Doe',
  email: 'john@example.com',
  message: 'Hello!'
});
```

---

## 🐞 Troubleshooting

### CAPTCHA Not Working
- Check site key matches frontend
- Verify secret key is correct
- Ensure config is active
- Check domain whitelist

### Auto-Reply Not Sending
- Verify config is active
- Check trigger includes "form"
- Ensure template exists
- Check logs: GET /autoreplies/logs

### Email Suppressed
- Check: GET /suppressions/check?email=xxx
- Remove: DELETE /suppressions/:id
- Review bounce reason in metadata

### Rate Limit Hit
- Check limit: GET /api-keys/:id
- Increase limit: PUT /api-keys/:id
- Wait for reset (shown in response header)

### Contact Not Created
- Check collect_contact flag (default: true)
- Verify email in template_params
- Check logs for errors

---

## 📚 Documentation Links

- [Implementation Summary](./IMPLEMENTATION_SUMMARY.md)
- [API Reference](./API_REFERENCE.md)
- [Migration Guide](./MIGRATION_GUIDE.md)
- [Feature Roadmap](./FEATURE_ROADMAP.md)
- [Complete Review](./COMPLETE_REVIEW.md)

---

## 🔗 Useful Links

**Provider Docs**:
- [SendGrid Webhooks](https://docs.sendgrid.com/for-developers/tracking-events/event)
- [Mailgun Webhooks](https://documentation.mailgun.com/en/latest/user_manual.html#webhooks)
- [Google reCAPTCHA](https://developers.google.com/recaptcha/docs/display)
- [hCaptcha](https://docs.hcaptcha.com/)

**LeapMailr**:
- GitHub: [github.com/dhawalhost/leapmailr](https://github.com/dhawalhost/leapmailr)
- Support: support@leapmailr.com

---

*Keep this guide handy for quick reference!*
