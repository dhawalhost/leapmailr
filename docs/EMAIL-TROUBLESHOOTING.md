# Email Not Arriving - Troubleshooting Guide

## Quick Diagnosis

### Step 1: Check Backend Logs
With the new logging added, when you send an email, you should see:
```
Attempting to send email to user@example.com via service MyService
SMTP Config - Host: smtp.gmail.com, Port: 587, Username: me@gmail.com, UseTLS: true, UseSSL: false
Connecting to smtp.gmail.com:587
Starting TLS negotiation
Authenticating as me@gmail.com
Authentication successful
Setting sender: noreply@mydomain.com
Setting recipient: user@example.com
Composed message: 1234 bytes
✓ Email sent successfully to user@example.com
Email sent successfully to user@example.com
```

**If you DON'T see these logs** → Email isn't being sent at all
**If you DO see these logs** → Email was sent but might be blocked/filtered

### Step 2: Check Email Service Configuration
```bash
# Get your service ID from the list
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/email-services

# Check the configuration (passwords masked)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/email-services/SERVICE_ID/config
```

Verify:
- ✅ `host` is correct for your provider
- ✅ `port` matches your provider's requirements
- ✅ `from_email` is valid and verified with your provider
- ✅ `use_tls` / `use_ssl` settings match your provider

### Step 3: Check Email Logs
```bash
# Get recent email history
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/emails

# Check specific email status
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/emails/EMAIL_ID
```

Look for:
- `"status": "failed"` → Check `error_message` field
- `"status": "sent"` → Email was sent, see common issues below

## Common Issues When Status is "sent" but Email Doesn't Arrive

### 1. From Email Not Verified/Authorized

**Gmail/Yahoo/Personal Providers:**
- The `from_email` MUST match the authenticated account
- Example: If you log in as `myapp@gmail.com`, you MUST send from `myapp@gmail.com`

**Fix:**
```json
{
  "from_email": "myapp@gmail.com",  // Same as username
  "username": "myapp@gmail.com"
}
```

**Transactional Providers (SendGrid, Mailgun, etc.):**
- You MUST verify your sending domain first
- Go to your provider's dashboard and verify the domain
- Only send from verified domains/addresses

### 2. Email Going to Spam

**Check:**
1. Recipient's spam folder
2. Your email might be marked as spam because:
   - No SPF/DKIM/DMARC records set up
   - Sending from unverified domain
   - Content triggers spam filters

**Fix:**
- Set up SPF/DKIM records for your domain
- Use a verified sending domain
- Avoid spam trigger words in subject/content

### 3. Gmail App Passwords

If using Gmail, you CANNOT use your regular password:

**Steps:**
1. Enable 2-Factor Authentication on your Google account
2. Go to https://myaccount.google.com/apppasswords
3. Generate an "App Password" for "Mail"
4. Use this 16-character password (not your regular password)

**Configuration:**
```json
{
  "host": "smtp.gmail.com",
  "port": 587,
  "use_tls": true,
  "use_ssl": false,
  "username": "yourmail@gmail.com",
  "password": "abcd efgh ijkl mnop",  // App password
  "from_email": "yourmail@gmail.com"
}
```

### 4. Provider-Specific Settings

**Gmail:**
```json
{
  "host": "smtp.gmail.com",
  "port": 587,
  "use_tls": true,
  "use_ssl": false
}
```

**Outlook/Office 365:**
```json
{
  "host": "smtp.office365.com",
  "port": 587,
  "use_tls": true,
  "use_ssl": false
}
```

**Yahoo:**
```json
{
  "host": "smtp.mail.yahoo.com",
  "port": 587,
  "use_tls": true,
  "use_ssl": false
}
```

### 5. Rate Limiting

Some providers have sending limits:
- **Gmail (free):** 500 emails/day, 100-150/hour
- **Yahoo:** Similar limits
- **Transactional services:** Check your plan

If you hit limits, emails may be queued or rejected silently.

### 6. Recipient Address Issues

- Check recipient email address is valid
- Check recipient's mailbox isn't full
- Try sending to a different email address to isolate the issue

## Testing Steps

### 1. Send to Yourself First
```bash
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "TEMPLATE_ID",
    "to_email": "YOUR_OWN_EMAIL",
    "template_params": {"name": "Test"}
  }' \
  http://localhost:8080/api/v1/email/send
```

### 2. Check Multiple Inboxes
Test sending to:
- Gmail account
- Outlook/Hotmail account
- Yahoo account
- Custom domain email

This helps identify if it's a provider-specific issue.

### 3. Monitor Backend Logs
Watch the terminal where your Go backend is running. You should see detailed SMTP logs.

### 4. Test Email Service
```bash
curl -X POST -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"to_email": "test@example.com"}' \
  http://localhost:8080/api/v1/email-services/SERVICE_ID/test
```

## Provider-Specific Verification

### SendGrid
1. Verify sender identity: https://app.sendgrid.com/settings/sender_auth
2. Check activity: https://app.sendgrid.com/email_activity
3. Ensure API key has "Mail Send" permission

### Mailgun
1. Verify domain: https://app.mailgun.com/app/sending/domains
2. Check logs: https://app.mailgun.com/app/logs
3. Ensure domain is out of sandbox mode for production

### Amazon SES
1. Check if in sandbox mode (can only send to verified addresses)
2. Request production access
3. Verify sending domain

## Still Not Working?

### Enable Debug Mode
Check the backend logs for detailed SMTP conversation:
- Connection attempts
- Authentication
- Sender/recipient acceptance
- Data transfer

### Common Error Messages

**"SMTP authentication failed"**
→ Wrong username or password

**"failed to set sender"**
→ From email not allowed/verified

**"failed to set recipient"**
→ Recipient address rejected by server

**"failed to connect"**
→ Wrong host/port or firewall blocking connection

**"failed to start TLS"**
→ TLS/SSL configuration mismatch

### Contact Provider Support

If emails show "sent" but never arrive:
1. Check your provider's email activity/logs
2. Look for bounce messages in your inbox
3. Contact provider support with:
   - Timestamp of sent email
   - Recipient address
   - Email ID from your logs

## Need Help?

Share:
1. Backend logs (with passwords removed)
2. Email service configuration (use the `/config` endpoint)
3. Email status from `/emails/EMAIL_ID`
4. Provider you're using
5. Whether you see the email in spam folder
