# LeapMailr Authentication Guide

Complete guide to understanding and implementing LeapMailr's flexible authentication system.

---

## Overview

LeapMailr provides **two authentication modes** designed for different use cases:

1. **Public Key Only** (Frontend/Client-Side) - Basic authentication
2. **Public + Private Key** (Backend/Server-Side) - Enhanced authentication

This dual-mode approach allows you to:
- ✅ Use LeapMailr safely from frontend applications
- ✅ Hide your SMTP provider details from end users
- ✅ Prevent spam while allowing public forms
- ✅ Have full API control from backend applications

---

## Why Two Authentication Modes?

### The Problem LeapMailr Solves

**Traditional Email Services:**
- ❌ Expose SMTP credentials in frontend code
- ❌ Require backend proxy for security
- ❌ Complex setup for simple contact forms
- ❌ No protection against spam/abuse

**LeapMailr's Solution:**
- ✅ Public key safe to expose in frontend
- ✅ SMTP credentials stay hidden on server
- ✅ Template-based approach prevents spam
- ✅ No backend required for contact forms
- ✅ Full control when you need it (backend mode)

---

## Authentication Mode Comparison

| Feature | Public Key Only | Public + Private Keys |
|---------|----------------|----------------------|
| **Security Level** | Basic | Enhanced |
| **Frontend Safe** | ✅ Yes | ❌ No (private key) |
| **Backend Use** | Limited | ✅ Recommended |
| **Custom HTML** | ❌ No | ✅ Yes |
| **Template-Based** | ✅ Yes | ✅ Yes |
| **Rate Limits** | Standard | Higher |
| **SMTP Hidden** | ✅ Yes | ✅ Yes |
| **Spam Protection** | ✅ Built-in | ✅ Built-in |
| **CAPTCHA Support** | ✅ Yes | ✅ Yes |
| **Auto-Reply** | ✅ Yes | ✅ Yes |
| **Contact Collection** | ✅ Yes | ✅ Yes |
| **Webhooks** | ✅ Yes | ✅ Yes |

---

## Mode 1: Public Key Only (Frontend)

### When to Use
- ✅ Contact forms on your website
- ✅ React, Vue, Angular applications
- ✅ Static sites (Gatsby, Next.js, etc.)
- ✅ Mobile apps (Flutter, React Native)
- ✅ Any client-side JavaScript

### How It Works

**Step 1:** Create email templates in the dashboard
```
Dashboard → Templates → Create New Template
```

**Step 2:** Get your public key
```
Dashboard → Settings → API Keys → Copy Public Key
```

**Step 3:** Use in frontend code
```javascript
// Safe to commit to Git - only public key exposed
const LEAPMAILR_PUBLIC_KEY = 'pk_live_xxxxx';

async function sendContactForm(formData) {
  const response = await fetch('https://api.leapmailr.com/api/v1/email/send-form', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Public-Key': LEAPMAILR_PUBLIC_KEY
    },
    body: JSON.stringify({
      template_id: 'your_template_id',
      template_params: {
        to_email: formData.email,
        name: formData.name,
        message: formData.message
      },
      captcha_response: formData.captchaToken, // Optional
      collect_contact: true // Save to contacts list
    })
  });
  
  return response.json();
}
```

### Security Features

**Template-Only Sending:**
- Users can **only** trigger templates you've predefined
- Cannot send arbitrary HTML or custom content
- Similar to EmailJS's security model

**Built-in Protection:**
- Rate limiting per public key
- CAPTCHA integration (reCAPTCHA/hCaptcha)
- Suppression list filtering
- Domain whitelisting

**What's Hidden:**
- Your SMTP server details
- SMTP username/password
- Email service provider
- Server infrastructure

### Example: React Contact Form

```jsx
import { useState } from 'react';

const LEAPMAILR_PUBLIC_KEY = 'pk_live_xxxxx';
const TEMPLATE_ID = 'template_uuid';

function ContactForm() {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    message: ''
  });
  const [status, setStatus] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setStatus('sending');

    try {
      const response = await fetch('https://api.leapmailr.com/api/v1/email/send-form', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Public-Key': LEAPMAILR_PUBLIC_KEY
        },
        body: JSON.stringify({
          template_id: TEMPLATE_ID,
          template_params: formData
        })
      });

      if (response.ok) {
        setStatus('success');
        setFormData({ name: '', email: '', message: '' });
      } else {
        setStatus('error');
      }
    } catch (error) {
      setStatus('error');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={formData.name}
        onChange={(e) => setFormData({...formData, name: e.target.value})}
        placeholder="Your Name"
        required
      />
      <input
        type="email"
        value={formData.email}
        onChange={(e) => setFormData({...formData, email: e.target.value})}
        placeholder="Your Email"
        required
      />
      <textarea
        value={formData.message}
        onChange={(e) => setFormData({...formData, message: e.target.value})}
        placeholder="Your Message"
        required
      />
      <button type="submit" disabled={status === 'sending'}>
        {status === 'sending' ? 'Sending...' : 'Send Message'}
      </button>
      {status === 'success' && <p>Message sent successfully!</p>}
      {status === 'error' && <p>Failed to send message. Please try again.</p>}
    </form>
  );
}
```

---

## Mode 2: Public + Private Keys (Backend)

### When to Use
- ✅ Server-side applications (Node.js, Python, Go, PHP)
- ✅ Sending custom email content
- ✅ Transactional emails
- ✅ Bulk email sending
- ✅ Automated workflows
- ✅ API integrations

### How It Works

**Step 1:** Generate API key pair in dashboard
```
Dashboard → Settings → API Keys → Generate New Key Pair
```

**Step 2:** Store keys securely
```bash
# .env file (NEVER commit to Git)
LEAPMAILR_PUBLIC_KEY=pk_live_xxxxx
LEAPMAILR_PRIVATE_KEY=sk_live_xxxxx
```

**Step 3:** Use in backend code
```javascript
// Node.js/Express example
const axios = require('axios');

async function sendTransactionalEmail(toEmail, orderData) {
  const response = await axios.post(
    'https://api.leapmailr.com/api/v1/email/send',
    {
      to_email: toEmail,
      to_name: orderData.customerName,
      subject: 'Order Confirmation',
      html: `
        <h1>Thank you for your order!</h1>
        <p>Order #${orderData.orderId}</p>
        <p>Total: $${orderData.total}</p>
      `,
      from: 'orders@yourdomain.com',
      reply_to: 'support@yourdomain.com'
    },
    {
      headers: {
        'Content-Type': 'application/json',
        'X-Public-Key': process.env.LEAPMAILR_PUBLIC_KEY,
        'X-Private-Key': process.env.LEAPMAILR_PRIVATE_KEY
      }
    }
  );
  
  return response.data;
}
```

### Enhanced Features

**Custom Content:**
- Send any HTML content
- Dynamic subject lines
- Custom from/reply-to addresses
- File attachments (coming soon)

**Higher Limits:**
- Increased rate limits
- Bulk sending support
- Priority delivery

**Full API Access:**
- All endpoints available
- Admin operations
- Webhook management
- Analytics access

### Example: Python Transactional Emails

```python
import os
import requests
from typing import Dict, Any

LEAPMAILR_PUBLIC_KEY = os.getenv('LEAPMAILR_PUBLIC_KEY')
LEAPMAILR_PRIVATE_KEY = os.getenv('LEAPMAILR_PRIVATE_KEY')
API_URL = 'https://api.leapmailr.com/api/v1'

def send_email(to_email: str, subject: str, html: str) -> Dict[str, Any]:
    """Send transactional email using LeapMailr"""
    
    response = requests.post(
        f'{API_URL}/email/send',
        headers={
            'Content-Type': 'application/json',
            'X-Public-Key': LEAPMAILR_PUBLIC_KEY,
            'X-Private-Key': LEAPMAILR_PRIVATE_KEY
        },
        json={
            'to_email': to_email,
            'subject': subject,
            'html': html,
            'from': 'noreply@yourdomain.com'
        }
    )
    
    response.raise_for_status()
    return response.json()

# Usage
send_email(
    to_email='user@example.com',
    subject='Welcome to Our Platform!',
    html='<h1>Welcome!</h1><p>Thanks for signing up.</p>'
)
```

---

## Best Practices

### For Frontend Applications

1. **Use Public Key Only**
   - Never include private key in frontend code
   - Safe to commit public key to version control
   - Use environment variables if you prefer

2. **Create Templates in Dashboard**
   - Design email templates before coding
   - Use template variables for dynamic content
   - Test templates in dashboard first

3. **Add CAPTCHA Protection**
   - Enable reCAPTCHA or hCaptcha
   - Add captcha_response to API calls
   - Prevents automated spam

4. **Enable Contact Collection**
   - Automatically save form submissions
   - Build your contact list
   - Export contacts via dashboard

5. **Monitor Rate Limits**
   - Track usage in dashboard
   - Upgrade plan if needed
   - Implement client-side throttling

### For Backend Applications

1. **Keep Private Key Secret**
   - Store in environment variables
   - Never commit to Git
   - Rotate keys periodically

2. **Use Environment Variables**
   ```bash
   # .env
   LEAPMAILR_PUBLIC_KEY=pk_live_xxxxx
   LEAPMAILR_PRIVATE_KEY=sk_live_xxxxx
   ```

3. **Implement Error Handling**
   ```javascript
   try {
     await sendEmail(data);
   } catch (error) {
     if (error.response?.status === 429) {
       // Rate limited - retry with backoff
     } else if (error.response?.status === 401) {
       // Invalid credentials - check keys
     } else {
       // Other errors
     }
   }
   ```

4. **Use Webhooks for Tracking**
   - Configure webhooks in dashboard
   - Get delivery notifications
   - Track bounces and complaints

5. **Monitor Usage**
   - Check API key usage stats
   - Set up alerts for high usage
   - Review analytics regularly

---

## Security Considerations

### Public Key Exposure
**Is it safe to expose my public key?**

✅ **Yes!** The public key is designed to be safely exposed in:
- Frontend JavaScript code
- Mobile apps
- Static websites
- Public GitHub repositories

**Why it's safe:**
- Only allows template-based sending
- Cannot send custom content
- Rate limited per key
- CAPTCHA protection available
- SMTP credentials remain hidden

### Private Key Protection
**Never expose your private key!**

❌ **Do NOT:**
- Include in frontend code
- Commit to version control
- Share via email/chat
- Log in application logs

✅ **Do:**
- Store in environment variables
- Use secret management services (AWS Secrets, Vault)
- Rotate keys regularly
- Revoke compromised keys immediately

### SMTP Credential Protection
**Your SMTP credentials are always protected:**
- Stored encrypted in database
- Never exposed via API
- Not included in responses
- Separate from API keys

---

## Migration Guide

### From EmailJS to LeapMailr

LeapMailr's public key mode is **100% compatible** with EmailJS patterns:

**EmailJS:**
```javascript
emailjs.send(
  'service_id',
  'template_id',
  { name: 'John', email: 'john@example.com' },
  'user_id'
);
```

**LeapMailr:**
```javascript
fetch('https://api.leapmailr.com/api/v1/email/send-form', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Public-Key': 'pk_live_xxxxx' // Similar to EmailJS user_id
  },
  body: JSON.stringify({
    template_id: 'template_uuid',
    template_params: {
      name: 'John',
      email: 'john@example.com'
    }
  })
});
```

### From SendGrid/Mailgun

If you're using SendGrid/Mailgun directly:

**Before (SendGrid):**
```javascript
// Exposes API key and email logic
sgMail.setApiKey(process.env.SENDGRID_API_KEY);
await sgMail.send({
  to: 'user@example.com',
  from: 'sender@yourdomain.com',
  subject: 'Welcome',
  html: '<h1>Welcome</h1>'
});
```

**After (LeapMailr Backend):**
```javascript
// Keys stay on server, SMTP provider hidden
await fetch('https://api.leapmailr.com/api/v1/email/send', {
  headers: {
    'X-Public-Key': process.env.LEAPMAILR_PUBLIC_KEY,
    'X-Private-Key': process.env.LEAPMAILR_PRIVATE_KEY
  },
  body: JSON.stringify({
    to_email: 'user@example.com',
    subject: 'Welcome',
    html: '<h1>Welcome</h1>'
  })
});
```

---

## Troubleshooting

### Common Issues

**1. "Invalid or expired public key"**
- Check key is copied correctly
- Verify key is active in dashboard
- Check for extra spaces/characters

**2. "Invalid or expired API key pair"**
- Both public AND private keys required
- Check both keys are correct
- Verify keys haven't been revoked

**3. "Rate limit exceeded"**
- Check rate limit in dashboard
- Implement request throttling
- Upgrade plan for higher limits

**4. "Template not found"**
- Verify template_id is correct
- Check template is active
- Ensure template belongs to your account

**5. "SMTP configuration not found"**
- Add email service in dashboard
- Configure SMTP settings
- Set default service

---

## FAQs

**Q: Can I use the same keys for both modes?**
A: Yes! The same public/private key pair works for both modes. Just use public key only for frontend, or both keys for backend.

**Q: How do I rotate my keys?**
A: Go to Dashboard → Settings → API Keys → Rotate. Old keys are revoked immediately.

**Q: What happens if my private key is leaked?**
A: Revoke it immediately in the dashboard and generate a new pair. Update your backend applications with the new keys.

**Q: Can I have multiple key pairs?**
A: Yes! Create different keys for different applications/environments (dev/staging/prod).

**Q: Is there a rate limit?**
A: Yes. Public key mode: 100 requests/min. Public+Private mode: 500 requests/min. Upgrade for higher limits.

**Q: Can I send to multiple recipients?**
A: Use the `/send-bulk` endpoint with public+private keys. Not available in public-key-only mode.

**Q: Do I need a backend to use LeapMailr?**
A: No! Public key mode works entirely from frontend. Only use backend mode if you need custom email content.

---

## Support

Need help? We're here for you:

- 📧 Email: support@leapmailr.com
- 💬 Discord: [LeapMailr Community](https://discord.gg/leapmailr)
- 📚 Docs: [docs.leapmailr.com](https://docs.leapmailr.com)
- 🐛 Issues: [GitHub Issues](https://github.com/dhawalhost/leapmailr/issues)

---

**Last Updated:** November 2025  
**Version:** 1.0
