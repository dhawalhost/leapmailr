# Backend Auto-Reply Support - Template-Based Approach

## Overview
The backend supports **template-based auto-reply configuration**. All auto-replies must be templates themselves, ensuring reusability, consistency, and professional design. Users can attach auto-reply templates to their email templates or override them when sending emails.

## Architecture

### Auto-Reply Templates
- Auto-reply messages are **regular templates** stored in the database
- Can be reused across multiple email templates
- Support all template features (variables, HTML/text content, versioning)
- Can be created, edited, and managed like any other template

### Template Attachment
- Email templates can reference an auto-reply template via `AutoReplyTemplateID`
- When `AutoReplyEnabled` is true and `AutoReplyTemplateID` is set, the auto-reply is sent automatically
- Auto-reply templates are loaded and processed using the same template engine

### Request Override
- When sending emails, users can override the template's auto-reply
- Set `auto_reply_enabled: true` and `auto_reply_template_id` in the email request
- This allows per-send customization while maintaining template structure

## Changes Made

### 1. **Models (`models/mail.go`)**

#### Template Model (Unchanged)
```go
AutoReplyEnabled    bool       `json:"auto_reply_enabled" gorm:"default:false"`
AutoReplyTemplateID *uuid.UUID `json:"auto_reply_template_id,omitempty" gorm:"type:uuid"`
```

#### EmailRequest Model
Added request-level override fields:
```go
AutoReplyEnabled    bool       `json:"auto_reply_enabled,omitempty"`
AutoReplyTemplateID *uuid.UUID `json:"auto_reply_template_id,omitempty"`
```

### 2. **Email Service (`service/email.go`)**

#### Updated `SendEmail` Method
Supports auto-reply with priority: Request override > Template configuration

```go
// Send auto-reply if enabled (template-based or request override)
autoReplyTemplateID := emailTemplate.AutoReplyTemplateID
if req.AutoReplyEnabled && req.AutoReplyTemplateID != nil {
    // Request-level override
    autoReplyTemplateID = req.AutoReplyTemplateID
}

if emailTemplate.AutoReplyEnabled && autoReplyTemplateID != nil {
    log.Printf("Auto-reply enabled, sending auto-reply using template %s", *autoReplyTemplateID)
    go s.sendAutoReply(emailService, *autoReplyTemplateID, req.ToEmail, req.ToName, userID)
}
```

#### Updated `sendAutoReply` Method
Now accepts `autoReplyTemplateID` directly instead of extracting from original template:
```go
func (s *EmailService) sendAutoReply(service models.EmailService, autoReplyTemplateID uuid.UUID, toEmail, toName string, userID uuid.UUID)
```

## Auto-Reply Workflows

### Workflow 1: Create Template with Auto-Reply
```json
POST /api/templates
{
  "name": "Welcome Email",
  "subject": "Welcome to LeapMailr!",
  "html_content": "<p>Welcome {{name}}!</p>",
  "auto_reply_enabled": true,
  "auto_reply_template_id": "uuid-of-auto-reply-template"
}
```

### Workflow 2: Send Email (Uses Template's Auto-Reply)
```json
POST /api/emails/send
{
  "template_id": "uuid-of-welcome-template",
  "to_email": "user@example.com",
  "to_name": "John Doe",
  "service_id": "uuid-of-service"
}
```
→ Email sent + Auto-reply template automatically triggered

### Workflow 3: Send Email with Override Auto-Reply
```json
POST /api/emails/send
{
  "template_id": "uuid-of-welcome-template",
  "to_email": "user@example.com",
  "auto_reply_enabled": true,
  "auto_reply_template_id": "uuid-of-different-auto-reply-template"
}
```
→ Overrides template's auto-reply with specified one

### Workflow 4: Create Auto-Reply Template First
```json
POST /api/templates
{
  "name": "Auto-Reply: Thank You",
  "subject": "Thank you for contacting us",
  "html_content": "<p>Thank you {{name}}, we'll respond within 24 hours.</p>",
  "category": "auto_reply"
}
```
→ Returns template ID to use in other templates

## Frontend Integration

### Template Creation Page
```typescript
// User creates auto-reply template first
const autoReplyTemplate = await templateAPI.create({
  name: "Auto-Reply: Thank You",
  subject: "Thanks for your email!",
  html_content: "<p>We received your message.</p>"
});

// Then creates main template with auto-reply attached
const mainTemplate = await templateAPI.create({
  name: "Contact Form Response",
  subject: "RE: Your Inquiry",
  html_content: "<p>Response content</p>",
  auto_reply_enabled: true,
  auto_reply_template_id: autoReplyTemplate.id
});
```

### Email Sending Page
```typescript
// Load template (includes auto_reply_template_id if configured)
const template = await templateAPI.get(templateId);

// Load the auto-reply template details
if (template.auto_reply_enabled && template.auto_reply_template_id) {
  const autoReplyTemplate = await templateAPI.get(template.auto_reply_template_id);
  // Display auto-reply preview in UI
}

// Send with template's auto-reply
await emailAPI.send({
  template_id: template.id,
  to_email: "user@example.com"
});

// OR override with different auto-reply
await emailAPI.send({
  template_id: template.id,
  to_email: "user@example.com",
  auto_reply_enabled: true,
  auto_reply_template_id: otherAutoReplyTemplateId
});
```

## Database Schema

No migration needed - existing schema already supports this:
```sql
CREATE TABLE templates (
    ...
    auto_reply_enabled BOOLEAN DEFAULT FALSE,
    auto_reply_template_id UUID REFERENCES templates(id),
    ...
);
```

## Benefits

✅ **Reusability**: Create auto-reply templates once, use across multiple email templates
✅ **Consistency**: All auto-replies follow template structure (HTML, text, variables)
✅ **Management**: Edit auto-reply templates in one place, changes reflect everywhere
✅ **Versioning**: Auto-reply templates support versioning like regular templates
✅ **Flexibility**: Override per-send or use template defaults
✅ **Professionalism**: Encourages well-designed auto-reply messages
✅ **Variables**: Auto-reply templates can use template variables (name, email, etc.)

## Priority System

1. **Request Override** (highest): `req.AutoReplyTemplateID` if `req.AutoReplyEnabled` is true
2. **Template Default**: `template.AutoReplyTemplateID` if `template.AutoReplyEnabled` is true
3. **None**: No auto-reply sent

## Testing

### Test Case 1: Create Auto-Reply Template
```bash
curl -X POST http://localhost:8080/api/templates \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "name": "Auto-Reply: Thank You",
    "subject": "Thank you for your email",
    "html_content": "<p>We received your message and will respond soon.</p>",
    "category": "auto_reply"
  }'
```

### Test Case 2: Attach Auto-Reply to Template
```bash
curl -X POST http://localhost:8080/api/templates \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "name": "Contact Form",
    "subject": "Contact Request Received",
    "html_content": "<p>Thank you for contacting us.</p>",
    "auto_reply_enabled": true,
    "auto_reply_template_id": "<auto-reply-template-id>"
  }'
```

### Test Case 3: Send with Auto-Reply
```bash
curl -X POST http://localhost:8080/api/emails/send \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "template_id": "<template-id>",
    "to_email": "test@example.com",
    "to_name": "Test User"
  }'
```

## Backwards Compatibility
✅ Fully backwards compatible - existing auto-reply logic unchanged
✅ No breaking changes to existing APIs
✅ Works with existing templates that have auto-reply configured

## Next Steps
1. Update frontend to allow selecting auto-reply templates
2. Add "Create Auto-Reply Template" shortcut in template creation flow
3. Display auto-reply preview when selecting templates
4. Add filter to show only auto-reply templates in selector
5. Consider adding auto-reply template gallery
