# Default Email Templates Feature

## Overview

LeapMailr now includes a library of pre-built, professional email templates that users can use immediately or customize to their needs. This feature is similar to EmailJS's template system but with more variety and customization options.

---

## Available Default Templates

### 1. Contact Form Templates

#### Simple Contact Form
- **Category:** `contact_form`
- **Use Case:** Basic contact forms
- **Variables:** `name`, `email`, `message`
- **Features:**
  - Clean, minimal design
  - Mobile-responsive
  - Easy to customize

#### Business Contact Form
- **Category:** `contact_form`
- **Use Case:** Professional business inquiries
- **Variables:** `name`, `company`, `email`, `phone`, `message`
- **Features:**
  - Professional gradient header
  - Grid layout for contact info
  - Company branding support

### 2. Welcome Emails

#### Welcome Email - Simple
- **Category:** `transactional`
- **Use Case:** New user onboarding
- **Variables:** `name`, `app_name`, `dashboard_url`, `year`
- **Features:**
  - Warm, friendly design
  - Clear call-to-action button
  - Brand customizable

### 3. Newsletter Templates

#### Newsletter - Modern
- **Category:** `newsletter`
- **Use Case:** Monthly newsletters, content distribution
- **Variables:** 
  - `newsletter_title`, `month`, `issue_number`
  - `featured_content`
  - `article1_title`, `article1_excerpt`, `article1_url`
  - `article2_title`, `article2_excerpt`, `article2_url`
  - Social links
- **Features:**
  - Featured content section
  - Multiple article slots
  - Social media links
  - Unsubscribe link

### 4. Transactional Templates

#### Order Confirmation
- **Category:** `transactional`
- **Use Case:** E-commerce order confirmations
- **Variables:**
  - `order_number`, `customer_name`, `order_date`
  - `shipping_address`, `payment_method`, `total_amount`
  - `order_url`, `support_email`
- **Features:**
  - Clear order summary
  - Payment and shipping details
  - Track order button
  - Support contact info

#### Password Reset
- **Category:** `transactional`
- **Use Case:** Password recovery
- **Variables:** `name`, `app_name`, `reset_url`, `expiry_hours`, `year`
- **Features:**
  - Security-focused design
  - Clear warning messages
  - Time-limited link notice
  - Security best practices

### 5. Notification Templates

#### Team Invitation
- **Category:** `notification`
- **Use Case:** Inviting team members
- **Variables:**
  - `inviter_name`, `inviter_email`, `inviter_initial`
  - `invitee_name`, `workspace_name`, `role`
  - `invitation_url`, `expiry_days`
- **Features:**
  - Personal touch with inviter avatar
  - Role badge
  - Feature highlights
  - Expiration notice

---

## API Endpoints

### Get All Default Templates
```http
GET /api/v1/templates/defaults
Authorization: Bearer <jwt_token>

Query Parameters:
- category (optional): Filter by category (contact_form, newsletter, transactional, notification)

Response:
{
  "status": "success",
  "data": [
    {
      "id": "uuid",
      "name": "Simple Contact Form",
      "description": "Basic contact form response...",
      "category": "contact_form",
      "subject": "New Contact Form Submission",
      "html_content": "<!DOCTYPE html>...",
      "text_content": "New Contact Form...",
      "variables": ["name", "email", "message"],
      "preview_image": "/templates/previews/simple-contact-form.png",
      "is_default": true,
      "is_public": true
    }
  ],
  "count": 7
}
```

### Get Template Categories
```http
GET /api/v1/templates/categories
Authorization: Bearer <jwt_token>

Response:
{
  "status": "success",
  "data": [
    {
      "id": "contact_form",
      "name": "Contact Forms",
      "description": "Contact form templates for collecting user inquiries",
      "icon": "📧"
    },
    {
      "id": "transactional",
      "name": "Transactional",
      "description": "Order confirmations, password resets, account notifications",
      "icon": "🔔"
    },
    // ... more categories
  ]
}
```

### Clone a Default Template
```http
POST /api/v1/templates/:id/clone
Authorization: Bearer <jwt_token>

Request Body:
{
  "name": "My Custom Contact Form"
}

Response:
{
  "status": "success",
  "data": {
    "id": "new_uuid",
    "name": "My Custom Contact Form",
    "cloned_from": "original_template_id",
    "user_id": "your_user_id",
    // ... full template details
  }
}
```

---

## Usage Guide

### For Users

#### 1. Browse Default Templates
```javascript
// Frontend - Fetch all default templates
const response = await fetch('https://api.leapmailr.com/api/v1/templates/defaults', {
  headers: {
    'Authorization': `Bearer ${accessToken}`
  }
});
const { data: templates } = await response.json();
```

#### 2. Filter by Category
```javascript
// Get only contact form templates
const response = await fetch(
  'https://api.leapmailr.com/api/v1/templates/defaults?category=contact_form',
  {
    headers: { 'Authorization': `Bearer ${accessToken}` }
  }
);
```

#### 3. Clone a Template
```javascript
// Clone a default template to customize it
const response = await fetch(
  `https://api.leapmailr.com/api/v1/templates/${templateId}/clone`,
  {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      name: 'My Custom Welcome Email'
    })
  }
);

const { data: newTemplate } = await response.json();
// Now you can edit the cloned template
```

#### 4. Use a Template
```javascript
// Send an email using a template (default or custom)
const response = await fetch('https://api.leapmailr.com/api/v1/email/send-form', {
  method: 'POST',
  headers: {
    'X-Public-Key': 'pk_live_xxxxx',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    template_id: 'template_uuid',
    template_params: {
      name: 'John Doe',
      email: 'john@example.com',
      message: 'Hello, this is a test message'
    }
  })
});
```

---

## Database Schema Updates

### Template Model Fields Added:
```go
type Template struct {
    // ... existing fields ...
    Category       string     `json:"category"`         // contact_form, newsletter, etc.
    IsDefault      bool       `json:"is_default"`       // System default template
    IsPublic       bool       `json:"is_public"`        // Available to all users
    ClonedFrom     *uuid.UUID `json:"cloned_from"`      // Original template if cloned
    UsageCount     int64      `json:"usage_count"`      // Track usage
    PreviewImage   string     `json:"preview_image"`    // Preview image URL
}
```

---

## Seeding Default Templates

### Automatic Seeding
The default templates are automatically seeded when the application starts. The seeding function is located in `database/default_templates.go`.

### Manual Seeding
```go
// In your database initialization or migration
import "github.com/dhawalhost/leapmailr/database"

func main() {
    // ... database connection ...
    
    // Seed default templates
    if err := database.SeedDefaultTemplates(); err != nil {
        log.Printf("Failed to seed templates: %v", err)
    }
}
```

### Updating Default Templates
When you update a default template definition in the code:
1. The seeding function checks if the template exists
2. If it exists, it updates the content
3. If it doesn't exist, it creates a new one

This ensures default templates are always up-to-date.

---

## Customization Guide

### For Developers: Adding New Default Templates

1. **Edit `database/default_templates.go`**
2. **Add a new template to the `GetDefaultTemplates()` function:**

```go
{
    Name:        "Your Template Name",
    Description: "Description of your template",
    Category:    "contact_form", // or newsletter, transactional, notification
    Subject:     "Email Subject with {{.variables}}",
    HTMLContent: `<!DOCTYPE html>...your HTML...`,
    TextContent: `Plain text version...`,
    Variables:   []string{"var1", "var2", "var3"},
    PreviewImage: "/templates/previews/your-template.png",
},
```

3. **Restart the application** - templates will be auto-seeded

### Template Variables

Use Go template syntax for variables:
```html
<!-- Simple variable -->
<p>Hello {{.name}}!</p>

<!-- Conditional -->
{{if .company}}
  <p>Company: {{.company}}</p>
{{end}}

<!-- Loop -->
{{range .items}}
  <li>{{.name}}: {{.price}}</li>
{{end}}
```

---

## Benefits

### For Users
✅ **Quick Start** - Use professional templates immediately  
✅ **Customizable** - Clone and modify to fit your brand  
✅ **Time-Saving** - No need to design from scratch  
✅ **Best Practices** - Templates follow email design standards  
✅ **Mobile-Responsive** - All templates work on mobile devices  

### For LeapMailr
✅ **Lower Barrier to Entry** - Users can start immediately  
✅ **Showcase Features** - Demonstrates platform capabilities  
✅ **Competitive Advantage** - More templates than EmailJS  
✅ **User Engagement** - Encourages exploration and usage  

---

## Template Categories Explained

### Contact Forms
- Simple inquiries
- Business contact
- Support requests
- Lead generation

### Transactional
- Order confirmations
- Password resets
- Account notifications
- Payment receipts

### Newsletters
- Monthly updates
- Content distribution
- Product announcements
- Company news

### Notifications
- Team invitations
- System alerts
- Activity updates
- Reminder emails

---

## Future Enhancements

### Planned Features
- [ ] Template preview in dashboard
- [ ] Template marketplace
- [ ] User-submitted templates
- [ ] Template analytics (usage stats)
- [ ] A/B testing for templates
- [ ] More template categories
- [ ] Template import/export
- [ ] Visual template editor

---

## Migration from EmailJS

If you're migrating from EmailJS, you can:

1. **Browse Similar Templates** - We have equivalents to common EmailJS templates
2. **Clone and Customize** - Start with a default template
3. **Import Your Templates** - Copy your HTML/text content
4. **Use Same Variables** - Variable syntax is compatible

---

## Support

Need help with templates?
- 📧 Email: support@leapmailr.com
- 📚 Docs: docs.leapmailr.com/templates
- 💬 Discord: discord.gg/leapmailr

---

**Last Updated:** November 2025  
**Version:** 1.0
