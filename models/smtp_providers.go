package models

// SMTPProvider represents a pre-configured SMTP provider
type SMTPProvider struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	Host         string              `json:"host"`
	Port         int                 `json:"port"`
	UseTLS       bool                `json:"use_tls"`
	UseSSL       bool                `json:"use_ssl"`
	AuthRequired bool                `json:"auth_required"`
	Fields       []SMTPProviderField `json:"fields"`
	HelpURL      string              `json:"help_url,omitempty"`
	Logo         string              `json:"logo,omitempty"`
	Category     string              `json:"category"` // "transactional", "smtp", "api"
}

// SMTPProviderField represents a configuration field for a provider
type SMTPProviderField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"` // "text", "password", "number", "email"
	Required    bool   `json:"required"`
	Placeholder string `json:"placeholder,omitempty"`
	Description string `json:"description,omitempty"`
	Default     string `json:"default,omitempty"`
}

// GetSMTPProviders returns all available SMTP providers
func GetSMTPProviders() []SMTPProvider {
	return []SMTPProvider{
		// Transactional Email Services
		{
			ID:           "sendgrid",
			Name:         "SendGrid",
			Description:  "Twilio SendGrid - Reliable email delivery at scale",
			Host:         "smtp.sendgrid.net",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "transactional",
			HelpURL:      "https://docs.sendgrid.com/for-developers/sending-email/integrating-with-the-smtp-api",
			Logo:         "/providers/sendgrid.svg",
			Fields: []SMTPProviderField{
				FromEmailField("noreply@yourdomain.com", "The email address to send from (must be verified in SendGrid)"),
				FixedUsernameField("API Username", "apikey", "Use 'apikey' as the username"),
				APIKeyPasswordField("API Key", "SG.xxxxxxxxxxxxx", "Your SendGrid API Key"),
			},
		},
		{
			ID:           "mailgun",
			Name:         "Mailgun",
			Description:  "Powerful email API for developers",
			Host:         "smtp.mailgun.org",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "transactional",
			HelpURL:      "https://documentation.mailgun.com/en/latest/user_manual.html#sending-via-smtp",
			Logo:         "/providers/mailgun.svg",
			Fields: []SMTPProviderField{
				FromEmailField("noreply@yourdomain.com", "The email address to send from"),
				UsernameField("SMTP Username", "postmaster@yourdomain.com", "Your Mailgun SMTP username"),
				PasswordField("SMTP Password", "Your Mailgun SMTP password"),
			},
		},
		{
			ID:           "postmark",
			Name:         "Postmark",
			Description:  "Fast and reliable transactional email service",
			Host:         "smtp.postmarkapp.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "transactional",
			HelpURL:      "https://postmarkapp.com/developer/user-guide/send-email-with-smtp",
			Logo:         "/providers/postmark.svg",
			Fields: []SMTPProviderField{
				FromEmailField("noreply@yourdomain.com", "The email address to send from (must be verified in Postmark)"),
				APITokenUsernameField("Server API Token", "Use your Server API Token as both username and password"),
				APITokenPasswordField("Server API Token (repeat)", "Same as username - your Server API Token"),
			},
		},
		{
			ID:           "ses",
			Name:         "Amazon SES",
			Description:  "Amazon Simple Email Service - Scalable email platform",
			Host:         "email-smtp.us-east-1.amazonaws.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "transactional",
			HelpURL:      "https://docs.aws.amazon.com/ses/latest/dg/send-email-smtp.html",
			Logo:         "/providers/aws-ses.svg",
			Fields: []SMTPProviderField{
				FromEmailField("noreply@yourdomain.com", "The email address to send from (must be verified in AWS SES)"),
				RegionField("AWS Region", "us-east-1", "AWS region for your SES service"),
				UsernameField("SMTP Username", "", "Your AWS SES SMTP username"),
				PasswordField("SMTP Password", "Your AWS SES SMTP password"),
			},
		},
		{
			ID:           "sparkpost",
			Name:         "SparkPost",
			Description:  "High-performance email delivery service",
			Host:         "smtp.sparkpostmail.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "transactional",
			HelpURL:      "https://developers.sparkpost.com/api/smtp/",
			Logo:         "/providers/sparkpost.svg",
			Fields: []SMTPProviderField{
				FromEmailField("noreply@yourdomain.com", "The email address to send from"),
				FixedUsernameField("SMTP Username", "SMTP_Injection", "Use 'SMTP_Injection' as username"),
				APIKeyPasswordField("API Key", "", "Your SparkPost API Key"),
			},
		},
		{
			ID:           "brevo",
			Name:         "Brevo (Sendinblue)",
			Description:  "All-in-one marketing platform with email",
			Host:         "smtp-relay.brevo.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "transactional",
			HelpURL:      "https://help.brevo.com/hc/en-us/articles/209467485",
			Logo:         "/providers/brevo.svg",
			Fields: []SMTPProviderField{
				FromEmailField("noreply@yourdomain.com", "The email address to send from"),
				EmailUsernameField("Login Email", "your@email.com", "Your Brevo account email"),
				PasswordField("SMTP Key", "Your Brevo SMTP key (not account password)"),
			},
		},
		{
			ID:           "mailjet",
			Name:         "Mailjet",
			Description:  "Email delivery and marketing platform",
			Host:         "in-v3.mailjet.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "transactional",
			HelpURL:      "https://dev.mailjet.com/smtp-relay/configuration/",
			Logo:         "/providers/mailjet.svg",
			Fields: []SMTPProviderField{
				FromEmailField("noreply@yourdomain.com", "The email address to send from"),
				UsernameField("API Key", "", "Your Mailjet API Key"),
				PasswordField("Secret Key", "Your Mailjet Secret Key"),
			},
		},
		{
			ID:           "elastic-email",
			Name:         "Elastic Email",
			Description:  "Cost-effective email delivery platform",
			Host:         "smtp.elasticemail.com",
			Port:         2525,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "transactional",
			HelpURL:      "https://elasticemail.com/resources/usage/sending-emails-via-smtp/",
			Logo:         "/providers/elastic-email.svg",
			Fields: []SMTPProviderField{
				FromEmailField("noreply@yourdomain.com", "The email address to send from"),
				EmailUsernameField("Account Email", "your@email.com", "Your Elastic Email account email"),
				PasswordField("SMTP Password", "Your Elastic Email SMTP password or API key"),
			},
		},

		// Popular Email Providers (Personal/Business)
		{
			ID:           "gmail",
			Name:         "Gmail",
			Description:  "Google Gmail SMTP server",
			Host:         "smtp.gmail.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			HelpURL:      "https://support.google.com/mail/answer/7126229",
			Logo:         "/providers/gmail.svg",
			Fields:       AppPasswordSMTPFields("your-email@gmail.com", "Gmail"),
		},
		{
			ID:           "outlook",
			Name:         "Outlook.com / Office 365",
			Description:  "Microsoft Outlook and Office 365 SMTP",
			Host:         "smtp.office365.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			HelpURL:      "https://support.microsoft.com/en-us/office/pop-imap-and-smtp-settings-8361e398-8af4-4e97-b147-6c6c4ac95353",
			Logo:         "/providers/outlook.svg",
			Fields:       StandardSMTPFields("your-email@outlook.com", "Your Outlook email address (same as username)"),
		},
		{
			ID:           "yahoo",
			Name:         "Yahoo Mail",
			Description:  "Yahoo Mail SMTP server",
			Host:         "smtp.mail.yahoo.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			HelpURL:      "https://help.yahoo.com/kb/SLN4075.html",
			Logo:         "/providers/yahoo.svg",
			Fields:       AppPasswordSMTPFields("your-email@yahoo.com", "Yahoo"),
		},
		{
			ID:           "zoho",
			Name:         "Zoho Mail",
			Description:  "Zoho Mail SMTP server",
			Host:         "smtp.zoho.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			HelpURL:      "https://www.zoho.com/mail/help/zoho-smtp.html",
			Logo:         "/providers/zoho.svg",
			Fields:       StandardSMTPFields("your-email@zoho.com", "Your Zoho email address (same as username)"),
		},
		{
			ID:           "icloud",
			Name:         "iCloud Mail",
			Description:  "Apple iCloud SMTP server",
			Host:         "smtp.mail.me.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			HelpURL:      "https://support.apple.com/en-us/HT202304",
			Logo:         "/providers/icloud.svg",
			Fields:       AppPasswordSMTPFields("your-email@icloud.com", "iCloud"),
		},
		{
			ID:           "fastmail",
			Name:         "Fastmail",
			Description:  "Fastmail SMTP server",
			Host:         "smtp.fastmail.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			HelpURL:      "https://www.fastmail.help/hc/en-us/articles/1500000278342",
			Logo:         "/providers/fastmail.svg",
			Fields:       AppPasswordSMTPFields("your-email@fastmail.com", "Fastmail"),
		},
		{
			ID:           "protonmail",
			Name:         "ProtonMail Bridge",
			Description:  "ProtonMail SMTP (requires Bridge)",
			Host:         "127.0.0.1",
			Port:         1025,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			HelpURL:      "https://proton.me/support/bridge",
			Logo:         "/providers/protonmail.svg",
			Fields: []SMTPProviderField{
				FromEmailField("your-email@protonmail.com", "Your ProtonMail email address (same as username)"),
				EmailUsernameField("ProtonMail Address", "your-email@protonmail.com", "Your ProtonMail email address"),
				PasswordField("Bridge Password", "Password from ProtonMail Bridge (not your account password)"),
			},
		},

		// Business/Hosting Providers
		{
			ID:           "godaddy",
			Name:         "GoDaddy",
			Description:  "GoDaddy Email Hosting SMTP",
			Host:         "smtpout.secureserver.net",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			HelpURL:      "https://www.godaddy.com/help/server-and-port-settings-for-workspace-email-6949",
			Logo:         "/providers/godaddy.svg",
			Fields:       StandardSMTPFields("your-email@yourdomain.com", "Your full email address (same as username)"),
		},
		{
			ID:           "namecheap",
			Name:         "Namecheap",
			Description:  "Namecheap Private Email SMTP",
			Host:         "mail.privateemail.com",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			HelpURL:      "https://www.namecheap.com/support/knowledgebase/article.aspx/1090/",
			Logo:         "/providers/namecheap.svg",
			Fields:       StandardSMTPFields("your-email@yourdomain.com", "Your full email address (same as username)"),
		},
		{
			ID:           "custom",
			Name:         "Custom SMTP Server",
			Description:  "Configure your own SMTP server",
			Host:         "",
			Port:         587,
			UseTLS:       true,
			UseSSL:       false,
			AuthRequired: true,
			Category:     "smtp",
			Logo:         "/providers/custom.svg",
			Fields: []SMTPProviderField{
				FromEmailField("noreply@yourdomain.com", "The email address to send from"),
				{
					Key:         "host",
					Label:       "SMTP Host",
					Type:        "text",
					Required:    true,
					Placeholder: "smtp.example.com",
					Description: "Your SMTP server hostname",
				},
				{
					Key:         "port",
					Label:       "SMTP Port",
					Type:        "number",
					Required:    true,
					Placeholder: "587",
					Description: "SMTP port (usually 587, 465, or 25)",
					Default:     "587",
				},
				UsernameField("Username", "", "SMTP authentication username"),
				PasswordField("Password", "SMTP authentication password"),
			},
		},
	}
}

// GetSMTPProviderByID returns a specific provider by ID
func GetSMTPProviderByID(id string) *SMTPProvider {
	providers := GetSMTPProviders()
	for _, provider := range providers {
		if provider.ID == id {
			return &provider
		}
	}
	return nil
}

// GetSMTPProviderCategories returns available categories
func GetSMTPProviderCategories() []map[string]string {
	return []map[string]string{
		{
			"id":          "transactional",
			"name":        "Transactional Services",
			"description": "Professional email delivery services with APIs and analytics",
		},
		{
			"id":          "smtp",
			"name":        "SMTP Providers",
			"description": "Traditional email providers and hosting services",
		},
		{
			"id":          "api",
			"name":        "API-Based",
			"description": "HTTP API-based email services",
		},
	}
}
