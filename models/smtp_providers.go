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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "noreply@yourdomain.com",
					Description: "The email address to send from (must be verified in SendGrid)",
				},
				{
					Key:         "username",
					Label:       "API Username",
					Type:        "text",
					Required:    true,
					Placeholder: "apikey",
					Description: "Use 'apikey' as the username",
					Default:     "apikey",
				},
				{
					Key:         "password",
					Label:       "API Key",
					Type:        "password",
					Required:    true,
					Placeholder: "SG.xxxxxxxxxxxxx",
					Description: "Your SendGrid API Key",
				},
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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "noreply@yourdomain.com",
					Description: "The email address to send from",
				},
				{
					Key:         "username",
					Label:       "SMTP Username",
					Type:        "text",
					Required:    true,
					Placeholder: "postmaster@yourdomain.com",
					Description: "Your Mailgun SMTP username",
				},
				{
					Key:         "password",
					Label:       "SMTP Password",
					Type:        "password",
					Required:    true,
					Description: "Your Mailgun SMTP password",
				},
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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "noreply@yourdomain.com",
					Description: "The email address to send from (must be verified in Postmark)",
				},
				{
					Key:         "username",
					Label:       "Server API Token",
					Type:        "password",
					Required:    true,
					Description: "Use your Server API Token as both username and password",
				},
				{
					Key:         "password",
					Label:       "Server API Token (repeat)",
					Type:        "password",
					Required:    true,
					Description: "Same as username - your Server API Token",
				},
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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "noreply@yourdomain.com",
					Description: "The email address to send from (must be verified in AWS SES)",
				},
				{
					Key:         "region",
					Label:       "AWS Region",
					Type:        "text",
					Required:    true,
					Placeholder: "us-east-1",
					Description: "AWS region for your SES service",
					Default:     "us-east-1",
				},
				{
					Key:         "username",
					Label:       "SMTP Username",
					Type:        "text",
					Required:    true,
					Description: "Your AWS SES SMTP username",
				},
				{
					Key:         "password",
					Label:       "SMTP Password",
					Type:        "password",
					Required:    true,
					Description: "Your AWS SES SMTP password",
				},
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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "noreply@yourdomain.com",
					Description: "The email address to send from",
				},
				{
					Key:         "username",
					Label:       "SMTP Username",
					Type:        "text",
					Required:    true,
					Placeholder: "SMTP_Injection",
					Description: "Use 'SMTP_Injection' as username",
					Default:     "SMTP_Injection",
				},
				{
					Key:         "password",
					Label:       "API Key",
					Type:        "password",
					Required:    true,
					Description: "Your SparkPost API Key",
				},
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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "noreply@yourdomain.com",
					Description: "The email address to send from",
				},
				{
					Key:         "username",
					Label:       "Login Email",
					Type:        "email",
					Required:    true,
					Description: "Your Brevo account email",
				},
				{
					Key:         "password",
					Label:       "SMTP Key",
					Type:        "password",
					Required:    true,
					Description: "Your Brevo SMTP key (not account password)",
				},
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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "noreply@yourdomain.com",
					Description: "The email address to send from",
				},
				{
					Key:         "username",
					Label:       "API Key",
					Type:        "text",
					Required:    true,
					Description: "Your Mailjet API Key",
				},
				{
					Key:         "password",
					Label:       "Secret Key",
					Type:        "password",
					Required:    true,
					Description: "Your Mailjet Secret Key",
				},
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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "noreply@yourdomain.com",
					Description: "The email address to send from",
				},
				{
					Key:         "username",
					Label:       "Account Email",
					Type:        "email",
					Required:    true,
					Description: "Your Elastic Email account email",
				},
				{
					Key:         "password",
					Label:       "SMTP Password",
					Type:        "password",
					Required:    true,
					Description: "Your Elastic Email SMTP password or API key",
				},
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
			Fields: []SMTPProviderField{
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "your-email@icloud.com",
					Description: "Your iCloud email address (same as username)",
				},
				{
					Key:         "username",
					Label:       "iCloud Email",
					Type:        "email",
					Required:    true,
					Placeholder: "your-email@icloud.com",
					Description: "Your iCloud email address",
				},
				{
					Key:         "password",
					Label:       "App-Specific Password",
					Type:        "password",
					Required:    true,
					Description: "Generate an app-specific password from Apple ID settings",
				},
			},
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
			Fields: []SMTPProviderField{
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "your-email@fastmail.com",
					Description: "Your Fastmail email address (same as username)",
				},
				{
					Key:         "username",
					Label:       "Fastmail Email",
					Type:        "email",
					Required:    true,
					Description: "Your Fastmail email address",
				},
				{
					Key:         "password",
					Label:       "App Password",
					Type:        "password",
					Required:    true,
					Description: "Use an app password for SMTP",
				},
			},
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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "your-email@protonmail.com",
					Description: "Your ProtonMail email address (same as username)",
				},
				{
					Key:         "username",
					Label:       "ProtonMail Address",
					Type:        "email",
					Required:    true,
					Description: "Your ProtonMail email address",
				},
				{
					Key:         "password",
					Label:       "Bridge Password",
					Type:        "password",
					Required:    true,
					Description: "Password from ProtonMail Bridge (not your account password)",
				},
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
				{
					Key:         "from_email",
					Label:       "From Email Address",
					Type:        "email",
					Required:    true,
					Placeholder: "noreply@yourdomain.com",
					Description: "The email address to send from",
				},
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
				{
					Key:         "username",
					Label:       "Username",
					Type:        "text",
					Required:    true,
					Description: "SMTP authentication username",
				},
				{
					Key:         "password",
					Label:       "Password",
					Type:        "password",
					Required:    true,
					Description: "SMTP authentication password",
				},
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
