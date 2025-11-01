package database

import (
	"encoding/json"
	"log"

	"github.com/dhawalhost/leapmailr/models"
	"github.com/google/uuid"
)

// DefaultTemplate represents a pre-built template definition
type DefaultTemplate struct {
	Name         string
	Description  string
	Category     string
	Subject      string
	HTMLContent  string
	TextContent  string
	Variables    []string
	PreviewImage string
}

// GetDefaultTemplates returns all pre-built templates
func GetDefaultTemplates() []DefaultTemplate {
	return []DefaultTemplate{
		// Contact Form Templates
		{
			Name:        "Simple Contact Form",
			Description: "Basic contact form response with name, email, and message fields",
			Category:    "contact_form",
			Subject:     "New Contact Form Submission",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4F46E5; color: white; padding: 20px; border-radius: 5px 5px 0 0; }
        .content { background: #f9fafb; padding: 30px; border: 1px solid #e5e7eb; }
        .field { margin-bottom: 15px; }
        .label { font-weight: bold; color: #6B7280; }
        .value { color: #111827; margin-top: 5px; }
        .footer { text-align: center; padding: 20px; color: #6B7280; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin: 0;">New Contact Form Submission</h1>
        </div>
        <div class="content">
            <div class="field">
                <div class="label">Name:</div>
                <div class="value">{{.name}}</div>
            </div>
            <div class="field">
                <div class="label">Email:</div>
                <div class="value">{{.email}}</div>
            </div>
            <div class="field">
                <div class="label">Message:</div>
                <div class="value">{{.message}}</div>
            </div>
        </div>
        <div class="footer">
            Sent via LeapMailr
        </div>
    </div>
</body>
</html>`,
			TextContent: `New Contact Form Submission

Name: {{.name}}
Email: {{.email}}
Message: {{.message}}

---
Sent via LeapMailr`,
			Variables:    []string{"name", "email", "message"},
			PreviewImage: "/templates/previews/simple-contact-form.png",
		},

		{
			Name:        "Business Contact Form",
			Description: "Professional contact form with company details and phone number",
			Category:    "contact_form",
			Subject:     "New Business Inquiry from {{.company}}",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; background: #f4f4f4; }
        .container { max-width: 650px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .content { padding: 40px; }
        .info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 30px; }
        .info-item { padding: 15px; background: #f8f9fa; border-left: 4px solid #667eea; border-radius: 4px; }
        .label { font-size: 12px; color: #6c757d; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 5px; }
        .value { font-size: 16px; color: #212529; font-weight: 500; }
        .message-section { background: #fff3cd; border-left: 4px solid #ffc107; padding: 20px; margin-top: 30px; border-radius: 4px; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #6c757d; font-size: 13px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📧 New Business Inquiry</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">From {{.company}}</p>
        </div>
        <div class="content">
            <div class="info-grid">
                <div class="info-item">
                    <div class="label">Contact Person</div>
                    <div class="value">{{.name}}</div>
                </div>
                <div class="info-item">
                    <div class="label">Company</div>
                    <div class="value">{{.company}}</div>
                </div>
                <div class="info-item">
                    <div class="label">Email Address</div>
                    <div class="value">{{.email}}</div>
                </div>
                <div class="info-item">
                    <div class="label">Phone Number</div>
                    <div class="value">{{.phone}}</div>
                </div>
            </div>
            <div class="message-section">
                <div class="label">Message</div>
                <div style="margin-top: 10px; line-height: 1.8;">{{.message}}</div>
            </div>
        </div>
        <div class="footer">
            Powered by LeapMailr | Secure Email Delivery
        </div>
    </div>
</body>
</html>`,
			TextContent: `New Business Inquiry

Contact Person: {{.name}}
Company: {{.company}}
Email: {{.email}}
Phone: {{.phone}}

Message:
{{.message}}

---
Powered by LeapMailr`,
			Variables:    []string{"name", "company", "email", "phone", "message"},
			PreviewImage: "/templates/previews/business-contact-form.png",
		},

		// Welcome Email Templates
		{
			Name:        "Welcome Email - Simple",
			Description: "Clean and simple welcome email for new users",
			Category:    "transactional",
			Subject:     "Welcome to {{.app_name}}, {{.name}}!",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { text-align: center; padding: 40px 20px; }
        .logo { font-size: 32px; font-weight: bold; color: #4F46E5; }
        .content { background: white; padding: 40px; border-radius: 8px; border: 1px solid #e5e7eb; }
        .greeting { font-size: 24px; font-weight: bold; margin-bottom: 20px; }
        .button { display: inline-block; padding: 14px 32px; background: #4F46E5; color: white; text-decoration: none; border-radius: 6px; font-weight: 500; margin: 20px 0; }
        .footer { text-align: center; padding: 30px 20px; color: #6B7280; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">{{.app_name}}</div>
        </div>
        <div class="content">
            <div class="greeting">Welcome, {{.name}}! 🎉</div>
            <p>Thank you for joining {{.app_name}}. We're excited to have you on board!</p>
            <p>You can now access all the features and start exploring what we have to offer.</p>
            <center>
                <a href="{{.dashboard_url}}" class="button">Get Started</a>
            </center>
            <p>If you have any questions, feel free to reach out to our support team.</p>
            <p>Best regards,<br>The {{.app_name}} Team</p>
        </div>
        <div class="footer">
            © {{.year}} {{.app_name}}. All rights reserved.
        </div>
    </div>
</body>
</html>`,
			TextContent: `Welcome to {{.app_name}}, {{.name}}!

Thank you for joining {{.app_name}}. We're excited to have you on board!

You can now access all the features and start exploring what we have to offer.

Get Started: {{.dashboard_url}}

If you have any questions, feel free to reach out to our support team.

Best regards,
The {{.app_name}} Team

---
© {{.year}} {{.app_name}}. All rights reserved.`,
			Variables:    []string{"name", "app_name", "dashboard_url", "year"},
			PreviewImage: "/templates/previews/welcome-simple.png",
		},

		// Newsletter Templates
		{
			Name:        "Newsletter - Modern",
			Description: "Modern newsletter template with featured content section",
			Category:    "newsletter",
			Subject:     "{{.newsletter_title}} - {{.month}} Edition",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; margin: 0; padding: 0; }
        .container { max-width: 680px; margin: 0 auto; background: white; }
        .header { background: #1a1a1a; color: white; padding: 40px 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 28px; }
        .header p { margin: 10px 0 0 0; opacity: 0.8; }
        .content { padding: 40px 30px; }
        .featured { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 8px; margin-bottom: 30px; }
        .featured h2 { margin: 0 0 15px 0; }
        .article { margin-bottom: 30px; padding-bottom: 30px; border-bottom: 1px solid #e5e7eb; }
        .article h3 { color: #4F46E5; margin: 0 0 10px 0; }
        .read-more { color: #4F46E5; text-decoration: none; font-weight: 500; }
        .footer { background: #f9fafb; padding: 30px; text-align: center; color: #6B7280; font-size: 14px; }
        .social { margin: 20px 0; }
        .social a { display: inline-block; margin: 0 10px; color: #4F46E5; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.newsletter_title}}</h1>
            <p>{{.month}} Edition | Issue #{{.issue_number}}</p>
        </div>
        <div class="content">
            <div class="featured">
                <h2>🌟 Featured Story</h2>
                <p>{{.featured_content}}</p>
            </div>
            <div class="article">
                <h3>{{.article1_title}}</h3>
                <p>{{.article1_excerpt}}</p>
                <a href="{{.article1_url}}" class="read-more">Read more →</a>
            </div>
            <div class="article">
                <h3>{{.article2_title}}</h3>
                <p>{{.article2_excerpt}}</p>
                <a href="{{.article2_url}}" class="read-more">Read more →</a>
            </div>
        </div>
        <div class="footer">
            <div class="social">
                <a href="{{.twitter_url}}">Twitter</a> | 
                <a href="{{.linkedin_url}}">LinkedIn</a> | 
                <a href="{{.website_url}}">Website</a>
            </div>
            <p>You're receiving this because you subscribed to {{.newsletter_title}}</p>
            <p><a href="{{.unsubscribe_url}}" style="color: #6B7280;">Unsubscribe</a></p>
        </div>
    </div>
</body>
</html>`,
			TextContent: `{{.newsletter_title}} - {{.month}} Edition
Issue #{{.issue_number}}

FEATURED STORY
{{.featured_content}}

---

{{.article1_title}}
{{.article1_excerpt}}
Read more: {{.article1_url}}

---

{{.article2_title}}
{{.article2_excerpt}}
Read more: {{.article2_url}}

---

Connect with us:
Twitter: {{.twitter_url}}
LinkedIn: {{.linkedin_url}}
Website: {{.website_url}}

You're receiving this because you subscribed to {{.newsletter_title}}
Unsubscribe: {{.unsubscribe_url}}`,
			Variables: []string{
				"newsletter_title", "month", "issue_number", "featured_content",
				"article1_title", "article1_excerpt", "article1_url",
				"article2_title", "article2_excerpt", "article2_url",
				"twitter_url", "linkedin_url", "website_url", "unsubscribe_url",
			},
			PreviewImage: "/templates/previews/newsletter-modern.png",
		},

		// Notification Templates
		{
			Name:        "Order Confirmation",
			Description: "E-commerce order confirmation email with order details",
			Category:    "transactional",
			Subject:     "Order Confirmation #{{.order_number}}",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f9fafb; }
        .container { max-width: 600px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; }
        .header { background: #10B981; color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 26px; }
        .checkmark { font-size: 48px; margin-bottom: 10px; }
        .content { padding: 40px 30px; }
        .order-number { font-size: 18px; font-weight: bold; color: #10B981; margin-bottom: 20px; }
        .details { background: #f9fafb; padding: 20px; border-radius: 6px; margin: 20px 0; }
        .detail-row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #e5e7eb; }
        .detail-row:last-child { border-bottom: none; }
        .label { color: #6B7280; }
        .value { font-weight: 500; }
        .total { font-size: 20px; font-weight: bold; color: #10B981; }
        .button { display: inline-block; padding: 14px 28px; background: #10B981; color: white; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { background: #f9fafb; padding: 20px 30px; text-align: center; color: #6B7280; font-size: 13px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="checkmark">✓</div>
            <h1>Order Confirmed!</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">Thank you for your order</p>
        </div>
        <div class="content">
            <div class="order-number">Order #{{.order_number}}</div>
            <p>Hi {{.customer_name}},</p>
            <p>We've received your order and will send you another email when it ships.</p>
            <div class="details">
                <div class="detail-row">
                    <span class="label">Order Date:</span>
                    <span class="value">{{.order_date}}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Shipping Address:</span>
                    <span class="value">{{.shipping_address}}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Payment Method:</span>
                    <span class="value">{{.payment_method}}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Total Amount:</span>
                    <span class="total">${{.total_amount}}</span>
                </div>
            </div>
            <center>
                <a href="{{.order_url}}" class="button">View Order Details</a>
            </center>
            <p>If you have any questions, please don't hesitate to contact us.</p>
        </div>
        <div class="footer">
            <p>Need help? Contact us at {{.support_email}}</p>
            <p>© {{.year}} {{.company_name}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
			TextContent: `✓ ORDER CONFIRMED!

Order #{{.order_number}}

Hi {{.customer_name}},

We've received your order and will send you another email when it ships.

Order Details:
- Order Date: {{.order_date}}
- Shipping Address: {{.shipping_address}}
- Payment Method: {{.payment_method}}
- Total Amount: ${{.total_amount}}

View Order: {{.order_url}}

If you have any questions, please don't hesitate to contact us.

Need help? Contact us at {{.support_email}}

© {{.year}} {{.company_name}}. All rights reserved.`,
			Variables: []string{
				"order_number", "customer_name", "order_date", "shipping_address",
				"payment_method", "total_amount", "order_url", "support_email",
				"company_name", "year",
			},
			PreviewImage: "/templates/previews/order-confirmation.png",
		},

		// Password Reset
		{
			Name:        "Password Reset",
			Description: "Secure password reset email with time-limited link",
			Category:    "transactional",
			Subject:     "Reset Your Password - {{.app_name}}",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; background: #f4f4f4; }
        .container { max-width: 600px; margin: 20px auto; background: white; border-radius: 8px; overflow: hidden; }
        .header { background: #EF4444; color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .icon { font-size: 40px; margin-bottom: 10px; }
        .content { padding: 40px 30px; }
        .warning { background: #FEF3C7; border-left: 4px solid #F59E0B; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .button { display: inline-block; padding: 14px 32px; background: #EF4444; color: white; text-decoration: none; border-radius: 6px; margin: 20px 0; font-weight: 500; }
        .expires { color: #6B7280; font-size: 14px; margin-top: 20px; }
        .footer { background: #f9fafb; padding: 20px 30px; text-align: center; color: #6B7280; font-size: 13px; }
        .security-note { background: #DBEAFE; border-left: 4px solid #3B82F6; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="icon">🔐</div>
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <p>Hi {{.name}},</p>
            <p>We received a request to reset your password for your {{.app_name}} account.</p>
            <div class="warning">
                ⚠️ <strong>Action Required:</strong> Click the button below to reset your password.
            </div>
            <center>
                <a href="{{.reset_url}}" class="button">Reset Password</a>
            </center>
            <p class="expires">This link will expire in {{.expiry_hours}} hours.</p>
            <div class="security-note">
                🛡️ <strong>Security Note:</strong> If you didn't request this password reset, please ignore this email or contact support if you have concerns.
            </div>
            <p>For security reasons, this link can only be used once.</p>
        </div>
        <div class="footer">
            <p>If the button doesn't work, copy and paste this URL into your browser:</p>
            <p style="word-break: break-all; font-size: 12px;">{{.reset_url}}</p>
            <p style="margin-top: 20px;">© {{.year}} {{.app_name}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
			TextContent: `PASSWORD RESET REQUEST

Hi {{.name}},

We received a request to reset your password for your {{.app_name}} account.

Click the link below to reset your password:
{{.reset_url}}

This link will expire in {{.expiry_hours}} hours.

SECURITY NOTE: If you didn't request this password reset, please ignore this email or contact support if you have concerns.

For security reasons, this link can only be used once.

© {{.year}} {{.app_name}}. All rights reserved.`,
			Variables:    []string{"name", "app_name", "reset_url", "expiry_hours", "year"},
			PreviewImage: "/templates/previews/password-reset.png",
		},

		// Team Invitation
		{
			Name:        "Team Invitation",
			Description: "Invite team members to join your workspace or organization",
			Category:    "notification",
			Subject:     "{{.inviter_name}} invited you to join {{.workspace_name}}",
			HTMLContent: `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f9fafb; }
        .container { max-width: 600px; margin: 20px auto; background: white; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%); color: white; padding: 40px 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 26px; }
        .avatar { width: 80px; height: 80px; border-radius: 50%; background: white; color: #6366F1; display: inline-flex; align-items: center; justify-content: center; font-size: 32px; font-weight: bold; margin-bottom: 15px; }
        .content { padding: 40px 30px; }
        .invitation-box { background: #F3F4F6; padding: 25px; border-radius: 8px; margin: 25px 0; text-align: center; }
        .workspace { font-size: 22px; font-weight: bold; color: #6366F1; margin-bottom: 10px; }
        .role { display: inline-block; background: #DBEAFE; color: #1E40AF; padding: 6px 12px; border-radius: 20px; font-size: 14px; font-weight: 500; }
        .button { display: inline-block; padding: 16px 40px; background: #6366F1; color: white; text-decoration: none; border-radius: 8px; margin: 20px 0; font-weight: 600; font-size: 16px; }
        .features { margin: 30px 0; }
        .feature { padding: 12px 0; border-bottom: 1px solid #E5E7EB; }
        .feature:last-child { border-bottom: none; }
        .feature::before { content: "✓"; color: #10B981; font-weight: bold; margin-right: 10px; }
        .footer { background: #F9FAFB; padding: 20px 30px; text-align: center; color: #6B7280; font-size: 13px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="avatar">{{.inviter_initial}}</div>
            <h1>You've Been Invited!</h1>
            <p style="margin: 10px 0 0 0; opacity: 0.9;">{{.inviter_name}} wants you on their team</p>
        </div>
        <div class="content">
            <p>Hi {{.invitee_name}},</p>
            <p><strong>{{.inviter_name}}</strong> ({{.inviter_email}}) has invited you to join:</p>
            <div class="invitation-box">
                <div class="workspace">{{.workspace_name}}</div>
                <p style="color: #6B7280; margin: 10px 0;">as a <span class="role">{{.role}}</span></p>
            </div>
            <div class="features">
                <div class="feature">Collaborate with your team in real-time</div>
                <div class="feature">Access shared projects and resources</div>
                <div class="feature">Stay connected and productive</div>
            </div>
            <center>
                <a href="{{.invitation_url}}" class="button">Accept Invitation</a>
            </center>
            <p style="color: #6B7280; font-size: 14px; margin-top: 20px;">This invitation expires in {{.expiry_days}} days.</p>
        </div>
        <div class="footer">
            <p>If you weren't expecting this invitation, you can safely ignore this email.</p>
            <p>© {{.year}} {{.app_name}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
			TextContent: `YOU'VE BEEN INVITED!

Hi {{.invitee_name}},

{{.inviter_name}} ({{.inviter_email}}) has invited you to join:

{{.workspace_name}}
Role: {{.role}}

What you'll get:
✓ Collaborate with your team in real-time
✓ Access shared projects and resources
✓ Stay connected and productive

Accept invitation: {{.invitation_url}}

This invitation expires in {{.expiry_days}} days.

If you weren't expecting this invitation, you can safely ignore this email.

© {{.year}} {{.app_name}}. All rights reserved.`,
			Variables: []string{
				"inviter_name", "inviter_email", "inviter_initial", "invitee_name",
				"workspace_name", "role", "invitation_url", "expiry_days", "app_name", "year",
			},
			PreviewImage: "/templates/previews/team-invitation.png",
		},
	}
}

// SeedDefaultTemplates creates default templates in the database
func SeedDefaultTemplates() error {
	templates := GetDefaultTemplates()

	for _, tmpl := range templates {
		variablesJSON, err := json.Marshal(tmpl.Variables)
		if err != nil {
			log.Printf("Error marshaling variables for template %s: %v", tmpl.Name, err)
			continue
		}

		template := models.Template{
			ID:           uuid.New(),
			Name:         tmpl.Name,
			Description:  tmpl.Description,
			Category:     tmpl.Category,
			Subject:      tmpl.Subject,
			HTMLContent:  tmpl.HTMLContent,
			TextContent:  tmpl.TextContent,
			Variables:    string(variablesJSON),
			IsDefault:    true,
			IsPublic:     true,
			IsActive:     true,
			PreviewImage: tmpl.PreviewImage,
			Version:      1,
		}

		// Check if template already exists
		var existing models.Template
		result := DB.Where("name = ? AND is_default = ?", template.Name, true).First(&existing)

		if result.Error == nil {
			// Template exists, update it
			log.Printf("Updating default template: %s", template.Name)
			DB.Model(&existing).Updates(template)
		} else {
			// Template doesn't exist, create it
			log.Printf("Creating default template: %s", template.Name)
			if err := DB.Create(&template).Error; err != nil {
				log.Printf("Error creating template %s: %v", template.Name, err)
				continue
			}
		}
	}

	log.Println("Default templates seeded successfully")
	return nil
}
