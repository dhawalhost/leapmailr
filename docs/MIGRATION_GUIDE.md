# Migration Guide: New Features Update

This guide helps existing LeapMailr users adopt the new EmailJS parity features.

---

## Overview

LeapMailr now includes 5 new features:
1. ✅ CAPTCHA Verification
2. ✅ Suppressions List
3. ✅ Auto-Reply Feature
4. ✅ Enhanced API Key Management
5. ✅ Contact Management

**Database Changes**: All features use automatic migrations. No manual database changes required.

---

## For Existing Users

### Step 1: Update Your Codebase

```bash
cd /path/to/leapmailr
git pull origin main
go mod tidy
go build
```

### Step 2: Run Database Migrations

Database migrations run automatically on server start. New tables created:
- `captcha_configs`
- `suppressions`
- `auto_reply_configs`
- `auto_reply_logs`
- `api_key_pairs`
- `api_key_usage_logs`
- `contacts`
- `contact_lists`

### Step 3: Restart Your Server

```bash
# Stop existing server
pkill leapmailr

# Start new server
./leapmailr
```

---

## Migration Scenarios

### Scenario 1: Add CAPTCHA to Existing Forms

**Before** (No spam protection):
```javascript
fetch('/api/v1/email/send-form', {
  method: 'POST',
  headers: {
    'X-API-Key': 'legacy-key'
  },
  body: JSON.stringify({
    template_id: 'contact-form',
    template_params: { ... }
  })
});
```

**After** (With CAPTCHA):

1. **Configure CAPTCHA in Dashboard**:
   - Go to `/dashboard/settings/captcha`
   - Add reCAPTCHA v2 or hCaptcha configuration
   - Activate it

2. **Update Frontend HTML**:
```html
<!-- Add CAPTCHA widget -->
<form id="contact-form">
  <input type="email" name="email" required>
  <textarea name="message"></textarea>
  
  <!-- Add this -->
  <div class="g-recaptcha" data-sitekey="YOUR_SITE_KEY"></div>
  
  <button type="submit">Send</button>
</form>

<!-- Add script -->
<script src="https://www.google.com/recaptcha/api.js"></script>
```

3. **Update JavaScript**:
```javascript
const captchaResponse = grecaptcha.getResponse();

fetch('/api/v1/email/send-form', {
  method: 'POST',
  headers: {
    'X-API-Key': 'legacy-key'
  },
  body: JSON.stringify({
    template_id: 'contact-form',
    captcha_response: captchaResponse, // Add this
    template_params: { ... }
  })
});
```

---

### Scenario 2: Migrate from Legacy API Keys to Key Pairs

**Benefits**:
- More secure (public + private key)
- Rate limiting per key
- Usage tracking
- Key rotation support

**Migration Steps**:

1. **Generate New Key Pair**:
   - Login to dashboard
   - Go to `/dashboard/settings/api-keys`
   - Click "Generate New Key Pair"
   - Copy both `pk_live_xxxxx` and `sk_live_xxxxx`
   - **Important**: Private key shown only once!

2. **Update Frontend Code**:

**Before**:
```javascript
headers: {
  'X-API-Key': 'legacy-key'
}
```

**After**:
```javascript
headers: {
  'X-Public-Key': 'pk_live_xxxxx',
  'X-Private-Key': 'sk_live_xxxxx'
}
```

3. **Set Environment Variables** (Recommended):
```javascript
// .env.local
NEXT_PUBLIC_LEAPMAILR_PUBLIC_KEY=pk_live_xxxxx
LEAPMAILR_PRIVATE_KEY=sk_live_xxxxx

// Usage
headers: {
  'X-Public-Key': process.env.NEXT_PUBLIC_LEAPMAILR_PUBLIC_KEY,
  'X-Private-Key': process.env.LEAPMAILR_PRIVATE_KEY
}
```

4. **Test New Keys**:
   - Send a test email
   - Check usage stats in dashboard

5. **Deactivate Legacy Keys** (After Testing):
   - Monitor usage for 1-2 weeks
   - Once confident, remove legacy keys

---

### Scenario 3: Enable Auto-Replies for Contact Forms

**Use Case**: Send confirmation emails when users submit contact forms.

**Steps**:

1. **Create Auto-Reply Template**:
   - Go to `/dashboard/templates`
   - Create new template:
     - Name: "Contact Form Confirmation"
     - Subject: `Thanks for contacting us, {{name}}!`
     - Body: 
       ```
       Hi {{name}},
       
       We received your message about "{{subject}}".
       We'll respond to {{email}} within 24 hours.
       
       Best regards,
       Support Team
       ```

2. **Configure Auto-Reply**:
   - Go to `/dashboard/settings/auto-reply`
   - Click "Create Auto-Reply"
   - Select template created above
   - Configure:
     - From Email: `noreply@yourdomain.com`
     - From Name: `Your Company`
     - Reply-To: `support@yourdomain.com`
     - Triggers: Check "Form Submissions"
     - Delay: 0 seconds (immediate)
   - Activate

3. **No Code Changes Required**:
   - Auto-replies trigger automatically on form submissions
   - Variables replaced from `template_params`

4. **Test**:
   - Submit a test form
   - Check recipient inbox for confirmation email

---

### Scenario 4: Set Up Bounce/Complaint Handling

**Benefits**:
- Prevent sending to invalid emails
- Improve deliverability
- Automatic cleanup via webhooks

**Steps**:

1. **Configure Webhooks** (SendGrid Example):
   
   **SendGrid Dashboard**:
   - Go to Settings → Mail Settings → Event Webhook
   - Enable Event Webhook
   - HTTP POST URL: `https://yourdomain.com/api/v1/webhooks/sendgrid`
   - Select Events:
     - ✅ Bounce
     - ✅ Spam Report
     - ✅ Unsubscribe
   - Save

2. **Test Webhook**:
   - SendGrid provides webhook testing tool
   - Send test bounce event
   - Check `/dashboard/settings/suppressions` for new entry

3. **Manual Suppressions** (Optional):
   - Go to `/dashboard/settings/suppressions`
   - Add known bounce/complaint emails
   - Bulk upload CSV if you have a list

4. **Monitor**:
   - Check suppression list regularly
   - Review bounce reasons
   - Clean up false positives

---

### Scenario 5: Start Collecting Contacts

**Benefit**: Build your mailing list automatically from form submissions.

**Steps**:

1. **Enable Contact Collection** (Already Default):
   - Contact collection is **enabled by default**
   - All form submissions automatically create/update contacts

2. **Customize Contact Collection**:

**Disable for specific forms**:
```javascript
fetch('/api/v1/email/send-form', {
  method: 'POST',
  body: JSON.stringify({
    template_id: 'transactional-email',
    template_params: { ... },
    collect_contact: false // Disable contact collection
  })
});
```

**Provide additional contact info**:
```javascript
template_params: {
  to_email: 'support@company.com',
  from_name: 'John Doe',      // → contact.name
  email: 'john@example.com',  // → contact.email
  phone: '+1234567890',       // → contact.phone
  company: 'Acme Corp',       // → contact.company
  subject: 'Product Inquiry', // → contact.metadata
  message: 'I want...'        // → contact.metadata
}
```

3. **View Contacts**:
   - Go to `/dashboard/contacts`
   - See all collected contacts
   - Search, filter by source/tags/subscription
   - Export to CSV

4. **Import Existing Contacts** (Optional):
   - Prepare CSV file:
     ```csv
     email,name,phone,company,tags
     john@example.com,John Doe,+1234567890,Acme Corp,"customer,vip"
     ```
   - Go to `/dashboard/contacts`
   - Click "Import"
   - Upload CSV

5. **Export Contacts**:
   - Apply filters (e.g., only subscribed)
   - Click "Export"
   - Download CSV

---

## Backward Compatibility

### ✅ All Existing Features Work Unchanged

1. **Email Sending**: No changes to `/send-email` or `/send-form`
2. **Templates**: Existing templates work as-is
3. **Email Services**: No configuration changes needed
4. **Authentication**: Legacy API keys still work

### New Optional Parameters

These are **optional** - existing code continues to work:

**`/send-form` endpoint**:
```javascript
{
  "template_id": "required",
  "template_params": "required",
  "captcha_response": "optional", // new
  "collect_contact": "optional"   // new, defaults to true
}
```

---

## Common Issues & Solutions

### Issue 1: CAPTCHA Verification Fails

**Symptoms**: Error "CAPTCHA verification failed"

**Solutions**:
1. Check CAPTCHA is configured and active
2. Verify `site_key` matches frontend
3. Ensure `secret_key` is correct
4. Check domain whitelist includes your domain
5. Test CAPTCHA token is not expired (2 min timeout)

### Issue 2: Auto-Replies Not Sending

**Symptoms**: No confirmation emails received

**Checklist**:
- ✅ Auto-reply config is active
- ✅ Template exists and is valid
- ✅ Trigger includes "form" (for form submissions)
- ✅ Email service is configured
- ✅ From email is verified with provider
- ✅ Check auto-reply logs: `/dashboard/settings/auto-reply` → Logs

**Debug**:
```bash
# Check logs
tail -f /var/log/leapmailr.log | grep auto-reply
```

### Issue 3: Emails Still Sending to Suppressed Addresses

**Symptoms**: Bounced emails not being blocked

**Solutions**:
1. Verify email is in suppression list
2. Check suppression reason matches
3. Ensure webhooks are configured correctly
4. Test webhook manually
5. Check email service logs for delivery attempts

### Issue 4: API Key Rate Limit Hit

**Symptoms**: Error "Rate limit exceeded"

**Solutions**:
1. Check current rate limit: `/dashboard/settings/api-keys`
2. Increase rate limit for production keys
3. Implement client-side rate limiting
4. Use multiple API keys for different apps
5. Cache responses where possible

### Issue 5: Contact Duplicates

**Symptoms**: Multiple contacts with same email

**This shouldn't happen** - contacts are deduplicated by email.

**Debug**:
1. Check database for duplicates:
   ```sql
   SELECT email, COUNT(*) 
   FROM contacts 
   WHERE user_id = 'your-user-id' 
   GROUP BY email 
   HAVING COUNT(*) > 1;
   ```

2. If duplicates exist, manually clean:
   ```sql
   -- Keep most recent, delete others
   DELETE FROM contacts 
   WHERE id NOT IN (
     SELECT id FROM (
       SELECT id, ROW_NUMBER() OVER (PARTITION BY email ORDER BY created_at DESC) as rn
       FROM contacts
     ) t WHERE rn = 1
   );
   ```

---

## Performance Optimization

### 1. Database Indexes (Already Added)

Indexes created automatically:
- `idx_user_email_unique` on contacts(user_id, email)
- `idx_email` on suppressions(email)
- `idx_public_key` on api_key_pairs(public_key)

### 2. Recommended: Add GIN Index for JSONB

For faster metadata/tags queries:

```sql
-- Add to PostgreSQL
CREATE INDEX idx_contacts_metadata ON contacts USING GIN (metadata);
CREATE INDEX idx_contacts_tags ON contacts USING GIN (tags);
CREATE INDEX idx_suppression_metadata ON suppressions USING GIN (metadata);
```

### 3. Rate Limiting Best Practices

**Production Setup**:
```go
// Adjust in config/config.go
RateLimitPerMinute: 1000 // for production keys
RateLimitPerMinute: 50   // for development keys
```

**Client-Side Throttling**:
```javascript
// Implement client-side queue
class EmailQueue {
  private queue: Array<() => Promise<void>> = [];
  private processing = false;
  private rateLimitPerSecond = 10;

  async add(fn: () => Promise<void>) {
    this.queue.push(fn);
    if (!this.processing) this.process();
  }

  private async process() {
    this.processing = true;
    while (this.queue.length > 0) {
      const fn = this.queue.shift()!;
      await fn();
      await new Promise(resolve => setTimeout(resolve, 1000 / this.rateLimitPerSecond));
    }
    this.processing = false;
  }
}
```

---

## Security Checklist

### ✅ Before Going to Production

1. **API Keys**:
   - [ ] Generate separate keys for production/staging
   - [ ] Store private keys in environment variables
   - [ ] Never commit keys to git
   - [ ] Set appropriate rate limits
   - [ ] Enable expiration for temporary keys

2. **CAPTCHA**:
   - [ ] Use production CAPTCHA keys (not test keys)
   - [ ] Whitelist only production domains
   - [ ] Enable CAPTCHA on all public forms

3. **Webhooks**:
   - [ ] Configure all email provider webhooks
   - [ ] Use HTTPS endpoints only
   - [ ] Verify webhook signatures (future feature)

4. **Suppressions**:
   - [ ] Import existing bounce list
   - [ ] Set up automated webhook processing
   - [ ] Monitor suppression list growth

5. **Auto-Replies**:
   - [ ] Test with real email addresses
   - [ ] Verify SPF/DKIM records
   - [ ] Use verified sender domains
   - [ ] Include unsubscribe links

6. **Contacts**:
   - [ ] Review data retention policy
   - [ ] Export regular backups
   - [ ] Respect unsubscribe requests
   - [ ] Implement GDPR compliance (if EU users)

---

## Rollback Plan

If issues occur, you can temporarily disable new features:

### Disable CAPTCHA
```sql
UPDATE captcha_configs SET is_active = false WHERE user_id = 'your-user-id';
```

### Disable Auto-Replies
```sql
UPDATE auto_reply_configs SET is_active = false WHERE user_id = 'your-user-id';
```

### Disable Contact Collection
```javascript
// Add to all send-form requests
{
  collect_contact: false
}
```

### Revert to Legacy API Keys
```javascript
// Switch back to legacy auth
headers: {
  'X-API-Key': 'legacy-key'
}
```

---

## Testing Checklist

### Before Migration

- [ ] Backup database
- [ ] Test on staging environment
- [ ] Verify all existing features work
- [ ] Check email delivery rates

### After Migration

- [ ] Test form submission with CAPTCHA
- [ ] Verify auto-reply delivery
- [ ] Check contact collection
- [ ] Test suppression blocking
- [ ] Monitor API key usage
- [ ] Check rate limiting
- [ ] Verify webhook processing

### Load Testing

```bash
# Install apache bench
apt-get install apache2-utils

# Test API key rate limiting
ab -n 1000 -c 10 \
  -H "X-Public-Key: pk_live_xxxxx" \
  -H "X-Private-Key: sk_live_xxxxx" \
  -p payload.json \
  -T application/json \
  https://yourdomain.com/api/v1/email/send-form
```

---

## Support & Resources

### Documentation
- `/docs/IMPLEMENTATION_SUMMARY.md` - Feature overview
- `/docs/API_REFERENCE.md` - Complete API docs
- `/docs/FEATURE_ROADMAP.md` - Feature status
- `/docs/ARCHITECTURE.md` - System architecture

### Community
- GitHub Issues: Report bugs, request features
- Discussions: Ask questions, share tips

### Professional Support
- Email: support@leapmailr.com
- Priority support available for enterprise users

---

## What's Next?

Recommended next steps:

1. **Week 1**: Test all features in staging
2. **Week 2**: Migrate production forms one by one
3. **Week 3**: Monitor metrics and adjust rate limits
4. **Week 4**: Full production deployment

**Future Features** (Under Consideration):
- [ ] Email open/click tracking
- [ ] A/B testing for templates
- [ ] Advanced analytics dashboard
- [ ] Contact list segmentation
- [ ] Webhook signature verification
- [ ] Visual template editor

---

## Conclusion

The migration is designed to be **zero-downtime** and **backward compatible**. All existing functionality continues to work while new features are opt-in.

Take your time testing each feature before full production rollout. Start with CAPTCHA (easiest), then progress to contacts, auto-replies, and finally API key migration.

Questions? Check the docs or reach out to support!
