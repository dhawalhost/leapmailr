package models

type ContactReplyData struct {
	RecipientName string
	SenderName    string
	Logo          string
	MailTo        string
}

type ContactUsData struct {
	Name    string
	Email   string
	Subject string
	Message string
	Logo    string
}

type ContactForm struct {
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Message string `json:"message" binding:"required"`
}
