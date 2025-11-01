# LeapMailr API Reference

Complete API documentation for all implemented features.

---

## Authentication

LeapMailr supports **flexible authentication** to work with both frontend and backend applications:

### 1. JWT Token (Dashboard/Admin APIs)
For authenticated dashboard users.

```http
Authorization: Bearer <jwt_token>
```

**Use Case:** Dashboard operations, admin management

---

### 2. Public + Private Key Pair (Backend/Server-Side)
**Enhanced Security Mode** - Requires both keys for full API access.

```http
X-Public-Key: pk_live_xxxxx
X-Private-Key: sk_live_xxxxx
```

**Features:**
- ✅ Full API access
- ✅ Custom email content
- ✅ Higher rate limits
- ✅ All endpoints available
- ✅ Server-to-server communication

**Use Case:** Backend applications, server-side scripts, secure integrations

**Security:** Private key should **NEVER** be exposed in frontend code.

---

### 3. Public Key Only (Frontend/Client-Side)
**Basic Security Mode** - Only public key required, safe for frontend use.

```http
X-Public-Key: pk_live_xxxxx
```

**Features:**
- ✅ Safe for client-side JavaScript
- ✅ Template-based email sending
- ✅ CAPTCHA protection available
- ✅ Auto-reply triggering
- ✅ Contact collection
- ⚠️ Cannot send custom HTML/content
- ⚠️ Standard rate limits apply

**Use Case:** Frontend applications, React/Vue/Angular apps, contact forms, client-side SDKs

**Security:** Public key can be safely exposed in frontend code. Your SMTP credentials remain hidden on the server.

**How it works:**
1. Users can only trigger predefined email templates (not send custom content)
2. All email content is controlled by templates you configure in the dashboard
3. Similar to EmailJS - prevents spam while allowing frontend usage
4. Rate limiting prevents abuse

---

### 4. Legacy API Key (Backward Compatibility)
```http
X-API-Key: <legacy_key>
```

**Note:** This method is maintained for backward compatibility. New integrations should use Public/Private key pairs.

---

## Authentication Examples

### Frontend Usage (React/Vue/Angular)
```javascript
// Safe to use in frontend - only public key exposed
const response = await fetch('https://api.leapmailr.com/api/v1/email/send-form', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Public-Key': 'pk_live_xxxxx' // Only public key - safe to expose
  },
  body: JSON.stringify({
    template_id: 'template_uuid',
    template_params: {
      to_email: 'user@example.com',
      name: 'John Doe',
      message: 'Hello from frontend!'
    }
  })
});
```

### Backend Usage (Node.js/Python/Go)
```javascript
// Backend - can use both keys for full access
const response = await fetch('https://api.leapmailr.com/api/v1/email/send', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Public-Key': 'pk_live_xxxxx',
    'X-Private-Key': 'sk_live_xxxxx' // Private key - keep secret!
  },
  body: JSON.stringify({
    to_email: 'user@example.com',
    to_name: 'John Doe',
    subject: 'Custom Subject',
    html: '<h1>Custom HTML content</h1>',
    from: 'noreply@yourdomain.com'
  })
});
```

---

## Core Email APIs

### Send Email
```http
POST /api/v1/email/send
Authorization: Bearer <jwt_token>

{
  "to_email": "recipient@example.com",
  "to_name": "Recipient Name",
  "template_id": "uuid",
  "template_params": {
    "variable1": "value1",
    "variable2": "value2"
  },
  "service_id": "uuid" // optional
}
```

### Send Form (EmailJS Compatible)
```http
POST /api/v1/email/send-form
X-Public-Key: pk_live_xxxxx
X-Private-Key: sk_live_xxxxx

{
  "template_id": "uuid",
  "template_params": {
    "to_email": "recipient@example.com",
    "name": "John Doe",
    "message": "Hello World"
  },
  "captcha_response": "captcha_token", // optional
  "collect_contact": true // optional, default true
}
```

### Send Bulk Email
```http
POST /api/v1/email/send-bulk
Authorization: Bearer <jwt_token>

{
  "recipients": [
    {"email": "user1@example.com", "name": "User 1"},
    {"email": "user2@example.com", "name": "User 2"}
  ],
  "template_id": "uuid",
  "template_params": {
    "company": "Acme Corp"
  }
}
```

---

## CAPTCHA APIs

### Create CAPTCHA Config
```http
POST /api/v1/captcha
Authorization: Bearer <jwt_token>

{
  "provider": "recaptcha_v2", // or "hcaptcha"
  "site_key": "your_site_key",
  "secret_key": "your_secret_key",
  "domains": ["example.com", "www.example.com"],
  "is_active": true
}
```

### List CAPTCHA Configs
```http
GET /api/v1/captcha
Authorization: Bearer <jwt_token>

Response:
{
  "captcha_configs": [
    {
      "id": "uuid",
      "provider": "recaptcha_v2",
      "site_key": "6Lc...",
      "domains": ["example.com"],
      "is_active": true
    }
  ]
}
```

### Update CAPTCHA Config
```http
PUT /api/v1/captcha/:id
Authorization: Bearer <jwt_token>

{
  "is_active": false
}
```

### Delete CAPTCHA Config
```http
DELETE /api/v1/captcha/:id
Authorization: Bearer <jwt_token>
```

---

## Suppression List APIs

### Add Suppression
```http
POST /api/v1/suppressions
Authorization: Bearer <jwt_token>

{
  "email": "bounced@example.com",
  "reason": "bounce" // "bounce", "complaint", "unsubscribe", "manual"
}
```

### Add Bulk Suppressions
```http
POST /api/v1/suppressions/bulk
Authorization: Bearer <jwt_token>

{
  "emails": ["email1@example.com", "email2@example.com"],
  "reason": "unsubscribe"
}
```

### List Suppressions
```http
GET /api/v1/suppressions?reason=bounce&source=webhook&search=example
Authorization: Bearer <jwt_token>

Response:
{
  "suppressions": [
    {
      "id": "uuid",
      "email": "bounced@example.com",
      "reason": "bounce",
      "source": "webhook",
      "metadata": {"bounce_type": "hard"},
      "created_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total": 1
}
```

### Check Suppression
```http
GET /api/v1/suppressions/check?email=test@example.com
Authorization: Bearer <jwt_token>

Response:
{
  "is_suppressed": true,
  "reason": "bounce"
}
```

### Delete Suppression
```http
DELETE /api/v1/suppressions/:id
Authorization: Bearer <jwt_token>
```

---

## Webhook APIs (Public)

### SendGrid Webhook
```http
POST /api/v1/webhooks/sendgrid

[
  {
    "email": "bounce@example.com",
    "event": "bounce",
    "reason": "550 5.1.1 User unknown"
  }
]
```

### Mailgun Webhook
```http
POST /api/v1/webhooks/mailgun

{
  "event-data": {
    "event": "failed",
    "recipient": "bounce@example.com",
    "severity": "permanent"
  }
}
```

### Generic Webhook
```http
POST /api/v1/webhooks/generic

{
  "email": "bounce@example.com",
  "event": "bounce"
}
```

---

## Auto-Reply APIs

### Create Auto-Reply Config
```http
POST /api/v1/autoreplies
Authorization: Bearer <jwt_token>

{
  "email_service_id": "uuid", // Optional: specific service, omit for default
  "name": "Contact Form Auto-Reply",
  "subject": "Thanks for your message!",
  "body": "Hi {{name}},\n\nWe received your message...",
  "from_email": "noreply@example.com",
  "from_name": "Auto Reply",
  "reply_to": "support@example.com",
  "delay_seconds": 0,
  "trigger_on_form": true,
  "trigger_on_api": false,
  "include_variables": true,
  "is_active": true
}
```

### List Auto-Replies
```http
GET /api/v1/autoreplies
Authorization: Bearer <jwt_token>

Response:
{
  "autoreplies": [
    {
      "id": "uuid",
      "email_service_id": "uuid",
      "name": "Contact Form Auto-Reply",
      "subject": "Thanks!",
      "body": "Hi {{name}}...",
      "from_email": "noreply@example.com",
      "trigger_on_form": true,
      "trigger_on_api": false,
      "is_active": true
    }
  ]
}
```

### Update Auto-Reply
```http
PUT /api/v1/autoreplies/:id
Authorization: Bearer <jwt_token>

{
  "is_active": false,
  "delay_seconds": 300
}
```

### Test Auto-Reply
```http
POST /api/v1/autoreplies/:id/test
Authorization: Bearer <jwt_token>

{
  "recipient_email": "test@example.com",
  "variables": {
    "name": "John Doe",
    "subject": "Test Subject"
  }
}
```

### Get Auto-Reply Logs
```http
GET /api/v1/autoreplies/logs?auto_reply_id=uuid&limit=100
Authorization: Bearer <jwt_token>

Response:
{
  "logs": [
    {
      "id": "uuid",
      "auto_reply_id": "uuid",
      "recipient_email": "user@example.com",
      "status": "sent",
      "sent_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

### Delete Auto-Reply
```http
DELETE /api/v1/autoreplies/:id
Authorization: Bearer <jwt_token>
```

---

## API Key Management APIs

### Generate Key Pair
```http
POST /api/v1/api-keys
Authorization: Bearer <jwt_token>

{
  "name": "Production Website",
  "rate_limit": 100,
  "permissions": ["send_email", "send_form"],
  "expires_at": "2025-12-31T23:59:59Z" // optional
}

Response:
{
  "id": "uuid",
  "name": "Production Website",
  "public_key": "pk_live_abc123...",
  "private_key": "sk_live_xyz789...", // Only shown once!
  "rate_limit": 100,
  "is_active": true
}
```

### List Key Pairs
```http
GET /api/v1/api-keys
Authorization: Bearer <jwt_token>

Response:
{
  "api_keys": [
    {
      "id": "uuid",
      "name": "Production Website",
      "public_key": "pk_live_abc123...",
      "is_active": true,
      "rate_limit": 100,
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

### Get Key Pair
```http
GET /api/v1/api-keys/:id
Authorization: Bearer <jwt_token>

Response:
{
  "id": "uuid",
  "name": "Production Website",
  "public_key": "pk_live_abc123...",
  "is_active": true,
  "rate_limit": 100,
  "permissions": ["send_email"],
  "expires_at": null
}
```

### Update Key Pair
```http
PUT /api/v1/api-keys/:id
Authorization: Bearer <jwt_token>

{
  "name": "Updated Name",
  "rate_limit": 200,
  "permissions": ["send_email", "send_form", "send_bulk"]
}
```

### Revoke Key Pair
```http
POST /api/v1/api-keys/:id/revoke
Authorization: Bearer <jwt_token>

Response:
{
  "message": "API key pair revoked successfully"
}
```

### Rotate Key Pair
```http
POST /api/v1/api-keys/:id/rotate
Authorization: Bearer <jwt_token>

Response:
{
  "public_key": "pk_live_new123...",
  "private_key": "sk_live_new789..." // Only shown once!
}
```

### Get Key Usage
```http
GET /api/v1/api-keys/:id/usage?start_date=2024-01-01&end_date=2024-01-31
Authorization: Bearer <jwt_token>

Response:
{
  "total_requests": 1543,
  "successful_requests": 1520,
  "failed_requests": 23,
  "endpoints": {
    "/api/v1/email/send-form": 1200,
    "/api/v1/email/send": 343
  },
  "daily_usage": [
    {"date": "2024-01-15", "count": 52}
  ]
}
```

### Delete Key Pair
```http
DELETE /api/v1/api-keys/:id
Authorization: Bearer <jwt_token>
```

---

## Contact Management APIs

### Create Contact
```http
POST /api/v1/contacts
Authorization: Bearer <jwt_token>

{
  "email": "contact@example.com",
  "name": "John Doe",
  "phone": "+1234567890",
  "company": "Acme Corp",
  "source": "form",
  "tags": ["newsletter", "customer"],
  "metadata": {
    "referral_source": "google",
    "campaign": "summer-2024"
  }
}

Response:
{
  "id": "uuid",
  "email": "contact@example.com",
  "name": "John Doe",
  "submission_count": 1,
  "is_subscribed": true,
  "created_at": "2024-01-15T10:00:00Z"
}
```

### List Contacts
```http
GET /api/v1/contacts?search=john&source=form&is_subscribed=true&tags=newsletter&limit=50&offset=0
Authorization: Bearer <jwt_token>

Response:
{
  "contacts": [
    {
      "id": "uuid",
      "email": "contact@example.com",
      "name": "John Doe",
      "phone": "+1234567890",
      "company": "Acme Corp",
      "source": "form",
      "tags": ["newsletter", "customer"],
      "metadata": {"campaign": "summer-2024"},
      "is_subscribed": true,
      "submission_count": 3,
      "created_at": "2024-01-15T10:00:00Z"
    }
  ],
  "total": 1
}
```

### Get Contact
```http
GET /api/v1/contacts/:id
Authorization: Bearer <jwt_token>

Response:
{
  "id": "uuid",
  "email": "contact@example.com",
  "name": "John Doe",
  "phone": "+1234567890",
  "company": "Acme Corp",
  "source": "form",
  "tags": ["newsletter"],
  "metadata": {"campaign": "summer-2024"},
  "is_subscribed": true,
  "submission_count": 3
}
```

### Update Contact
```http
PUT /api/v1/contacts/:id
Authorization: Bearer <jwt_token>

{
  "name": "John Smith",
  "phone": "+0987654321",
  "is_subscribed": false,
  "tags": ["newsletter", "vip"]
}
```

### Delete Contact
```http
DELETE /api/v1/contacts/:id
Authorization: Bearer <jwt_token>
```

### Import Contacts (CSV)
```http
POST /api/v1/contacts/import
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data

Form Data:
- file: contacts.csv

CSV Format:
email,name,phone,company,tags
john@example.com,John Doe,+1234567890,Acme Corp,"tag1,tag2"
jane@example.com,Jane Smith,+0987654321,Corp Inc,"tag3"

Response:
{
  "imported": 2,
  "skipped": 0,
  "errors": []
}
```

### Get Contact Statistics
```http
GET /api/v1/contacts/stats
Authorization: Bearer <jwt_token>

Response:
{
  "total_contacts": 150,
  "total_subscribed": 120,
  "total_unsubscribed": 30,
  "recent_contacts": 15,
  "contacts_by_source": {
    "form": 100,
    "api": 30,
    "import": 20
  },
  "contacts_by_tag": {
    "newsletter": 80,
    "customer": 50,
    "vip": 20
  }
}
```

### Export Contacts (CSV)
```http
GET /api/v1/contacts/export?source=form&is_subscribed=true&tags=newsletter
Authorization: Bearer <jwt_token>

Response:
Content-Type: text/csv
Content-Disposition: attachment; filename="contacts_2024-01-15.csv"

email,name,phone,company,source,tags,is_subscribed,submission_count
john@example.com,John Doe,+1234567890,Acme Corp,form,"newsletter,customer",true,3
```

---

## Template Management APIs

### Create Template
```http
POST /api/v1/templates
Authorization: Bearer <jwt_token>

{
  "name": "Welcome Email",
  "subject": "Welcome {{name}}!",
  "html_body": "<h1>Hi {{name}}</h1><p>{{message}}</p>",
  "text_body": "Hi {{name}}\n\n{{message}}",
  "variables": ["name", "message"]
}
```

### List Templates
```http
GET /api/v1/templates
Authorization: Bearer <jwt_token>
```

### Get Template
```http
GET /api/v1/templates/:id
Authorization: Bearer <jwt_token>
```

### Update Template
```http
PUT /api/v1/templates/:id
Authorization: Bearer <jwt_token>
```

### Delete Template
```http
DELETE /api/v1/templates/:id
Authorization: Bearer <jwt_token>
```

### Test Template
```http
POST /api/v1/templates/:id/test
Authorization: Bearer <jwt_token>

{
  "to_email": "test@example.com",
  "template_params": {
    "name": "John",
    "message": "Test message"
  }
}
```

---

## Email Service Management APIs

### Create Email Service
```http
POST /api/v1/email-services
Authorization: Bearer <jwt_token>

{
  "name": "SendGrid Production",
  "provider": "sendgrid",
  "api_key": "SG.xxxxx",
  "from_email": "noreply@example.com",
  "from_name": "My App"
}
```

### List Email Services
```http
GET /api/v1/email-services
Authorization: Bearer <jwt_token>
```

### Set Default Service
```http
POST /api/v1/email-services/:id/default
Authorization: Bearer <jwt_token>
```

---

## Error Responses

### Common Error Codes

**400 Bad Request**
```json
{
  "error": "Invalid request parameters",
  "details": "email is required"
}
```

**401 Unauthorized**
```json
{
  "error": "Invalid or missing authentication credentials"
}
```

**403 Forbidden**
```json
{
  "error": "Rate limit exceeded for API key",
  "retry_after": 60
}
```

**404 Not Found**
```json
{
  "error": "Resource not found"
}
```

**422 Unprocessable Entity**
```json
{
  "error": "Email is in suppression list",
  "reason": "bounce"
}
```

**500 Internal Server Error**
```json
{
  "error": "Internal server error",
  "message": "Failed to send email"
}
```

---

## Rate Limiting

Rate limits are enforced per API key pair:

**Headers in Response**:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642252800
```

**When Limit Exceeded**:
```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1642252800

{
  "error": "Rate limit exceeded",
  "retry_after": 45
}
```

---

## Webhooks Configuration

### SendGrid Setup
1. Go to SendGrid Dashboard → Settings → Mail Settings → Event Webhook
2. Set HTTP POST URL: `https://your-domain.com/api/v1/webhooks/sendgrid`
3. Select Events: Bounce, Spam Report, Unsubscribe
4. Save

### Mailgun Setup
1. Go to Mailgun Dashboard → Webhooks
2. Add Webhook URL: `https://your-domain.com/api/v1/webhooks/mailgun`
3. Select Events: Bounced, Complained, Unsubscribed
4. Save

---

## SDK Integration Example

### JavaScript/TypeScript
```typescript
class LeapMailrSDK {
  private publicKey: string;
  private privateKey: string;
  private apiUrl: string;

  constructor(publicKey: string, privateKey: string, apiUrl: string) {
    this.publicKey = publicKey;
    this.privateKey = privateKey;
    this.apiUrl = apiUrl;
  }

  async sendForm(templateId: string, params: Record<string, any>) {
    const response = await fetch(`${this.apiUrl}/api/v1/email/send-form`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Public-Key': this.publicKey,
        'X-Private-Key': this.privateKey,
      },
      body: JSON.stringify({
        template_id: templateId,
        template_params: params,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error);
    }

    return await response.json();
  }
}

// Usage
const sdk = new LeapMailrSDK(
  'pk_live_xxxxx',
  'sk_live_xxxxx',
  'https://api.leapmailr.com'
);

await sdk.sendForm('contact-form-template', {
  to_email: 'support@company.com',
  name: 'John Doe',
  email: 'john@example.com',
  message: 'Hello!'
});
```

---

## Best Practices

1. **Never expose private keys in client-side code**
   - Use environment variables
   - Rotate keys regularly

2. **Use CAPTCHA for public forms**
   - Prevents spam and bot submissions

3. **Monitor suppression lists**
   - Configure webhooks for automatic updates
   - Remove hard bounces immediately

4. **Set appropriate rate limits**
   - Production: 100-1000 req/min
   - Development: 10-50 req/min

5. **Use auto-replies for user engagement**
   - Confirmation emails
   - Thank you messages
   - Next steps instructions

6. **Collect contacts ethically**
   - Include unsubscribe links
   - Respect subscription status
   - Export regularly for backups

7. **Monitor API key usage**
   - Track requests per endpoint
   - Identify unusual patterns
   - Set up alerts for rate limit violations

---

## Support

For issues, questions, or feature requests:
- GitHub: [github.com/dhawalhost/leapmailr](https://github.com/dhawalhost/leapmailr)
- Documentation: See `docs/` folder
- Email: support@leapmailr.com
