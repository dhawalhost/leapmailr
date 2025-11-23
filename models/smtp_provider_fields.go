package models

// Common field builders to reduce duplication in smtp_providers.go

// FromEmailField creates a standard "from_email" field
func FromEmailField(placeholder, description string) SMTPProviderField {
	if placeholder == "" {
		placeholder = "noreply@yourdomain.com"
	}
	if description == "" {
		description = "The email address to send from"
	}
	return SMTPProviderField{
		Key:         "from_email",
		Label:       "From Email Address",
		Type:        "email",
		Required:    true,
		Placeholder: placeholder,
		Description: description,
	}
}

// UsernameField creates a standard "username" field
func UsernameField(label, placeholder, description string) SMTPProviderField {
	if label == "" {
		label = "Username"
	}
	return SMTPProviderField{
		Key:         "username",
		Label:       label,
		Type:        "text",
		Required:    true,
		Placeholder: placeholder,
		Description: description,
	}
}

// EmailUsernameField creates a username field that expects an email
func EmailUsernameField(label, placeholder, description string) SMTPProviderField {
	if label == "" {
		label = "Email Address"
	}
	return SMTPProviderField{
		Key:         "username",
		Label:       label,
		Type:        "email",
		Required:    true,
		Placeholder: placeholder,
		Description: description,
	}
}

// PasswordField creates a standard "password" field
func PasswordField(label, description string) SMTPProviderField {
	if label == "" {
		label = "Password"
	}
	return SMTPProviderField{
		Key:         "password",
		Label:       label,
		Type:        "password",
		Required:    true,
		Description: description,
	}
}

// APIKeyPasswordField creates a password field for API keys
func APIKeyPasswordField(label, placeholder, description string) SMTPProviderField {
	if label == "" {
		label = "API Key"
	}
	return SMTPProviderField{
		Key:         "password",
		Label:       label,
		Type:        "password",
		Required:    true,
		Placeholder: placeholder,
		Description: description,
	}
}

// TextFieldWithDefault creates a text field with a default value
func TextFieldWithDefault(key, label, placeholder, description, defaultValue string) SMTPProviderField {
	return SMTPProviderField{
		Key:         key,
		Label:       label,
		Type:        "text",
		Required:    true,
		Placeholder: placeholder,
		Description: description,
		Default:     defaultValue,
	}
}

// NumberField creates a number input field
func NumberField(key, label, placeholder, description, defaultValue string) SMTPProviderField {
	return SMTPProviderField{
		Key:         key,
		Label:       label,
		Type:        "number",
		Required:    true,
		Placeholder: placeholder,
		Description: description,
		Default:     defaultValue,
	}
}

// StandardSMTPFields returns the most common field set for SMTP providers
// (from_email, username as email, password)
func StandardSMTPFields(emailPlaceholder, emailDescription string) []SMTPProviderField {
	if emailPlaceholder == "" {
		emailPlaceholder = "your-email@example.com"
	}
	if emailDescription == "" {
		emailDescription = "Your email address (same as username)"
	}

	return []SMTPProviderField{
		FromEmailField(emailPlaceholder, emailDescription),
		EmailUsernameField("Email Address", emailPlaceholder, "Your email address"),
		PasswordField("Password", "Your account password"),
	}
}

// AppPasswordSMTPFields returns fields for SMTP providers requiring app passwords
func AppPasswordSMTPFields(emailPlaceholder, providerName string) []SMTPProviderField {
	return []SMTPProviderField{
		FromEmailField(emailPlaceholder, "Your "+providerName+" email address (same as username)"),
		EmailUsernameField(providerName+" Email", emailPlaceholder, "Your "+providerName+" email address"),
		PasswordField("App Password", "Use an App Password for security"),
	}
}
