package service

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/models"
)

var (
	contact_us_reply_template = "./templates/contact_us_reply_template.html"
	contact_us_template       = "./templates/contact_us_template.html"
)

func SubmitForm(sender models.Sender, recipient models.Recipient, form models.ContactForm, smtpServer models.SMTPDetails) (err error) {
	sb := strings.Builder{}
	subject := "Contact Us Submission from "
	sb.WriteString(subject)
	sb.WriteString(form.Name)
	subject = sb.String()

	htmlTemplate, err := os.ReadFile(contact_us_template)
	if err != nil {
		fmt.Println("Error reading HTML file:", err)
		return
	}

	data := models.ContactUsData{
		Name:    form.Name,
		Email:   form.Email,
		Message: form.Message,
		Logo:    config.GetConfig().LogoURL,
	}

	tpl, err := template.New("emailTemplate").Parse(string(htmlTemplate))
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	var tplBuffer bytes.Buffer
	if err = tpl.Execute(&tplBuffer, data); err != nil {
		fmt.Println("Error executing template:", err)
		return
	}
	htmlContent := tplBuffer.String()

	auth := smtp.CRAMMD5Auth(smtpServer.Email, smtpServer.Secret)

	smtpAddr := fmt.Sprintf("%s:%d", smtpServer.Server, smtpServer.Port)
	client, err := smtp.Dial(smtpAddr)
	if err != nil {
		fmt.Println("Failed to connect to the SMTP server:", err)
		return
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		fmt.Println("Authentication error:", err)
		return
	}

	if err = client.Mail(sender.Email); err != nil {
		fmt.Println("Error setting sender:", err)
		return
	}
	if err = client.Rcpt(recipient.Email); err != nil {
		fmt.Println("Error setting recipient:", err)
		return
	}

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", sender.Name, sender.Email)
	headers["To"] = recipient.Email
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=utf-8"

	var emailBuffer bytes.Buffer
	for key, value := range headers {
		emailBuffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString(htmlContent)

	w, err := client.Data()
	if err != nil {
		fmt.Println("Error preparing data:", err)
		return
	}
	defer w.Close()

	_, err = w.Write(emailBuffer.Bytes())
	if err != nil {
		fmt.Println("Error writing message:", err)
		return
	}

	fmt.Println("Email sent successfully!")
	return
}

func SendReply(sender models.Sender, recipient models.Recipient, smtpServer models.SMTPDetails) (err error) {
	subject := "Thank you for Contacting Us!"

	htmlTemplate, err := os.ReadFile(contact_us_reply_template)
	if err != nil {
		fmt.Println("Error reading HTML file:", err)
		return
	}

	data := models.ContactReplyData{
		RecipientName: recipient.Name,
		SenderName:    sender.Name,
		Logo:          config.GetConfig().LogoURL,
		MailTo:        config.GetConfig().ContactMail,
	}

	tpl, err := template.New("emailTemplate").Parse(string(htmlTemplate))
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	var tplBuffer bytes.Buffer
	if err = tpl.Execute(&tplBuffer, data); err != nil {
		fmt.Println("Error executing template:", err)
		return
	}
	htmlContent := tplBuffer.String()

	auth := smtp.CRAMMD5Auth(smtpServer.Email, smtpServer.Secret)

	smtpAddr := fmt.Sprintf("%s:%d", smtpServer.Server, smtpServer.Port)
	client, err := smtp.Dial(smtpAddr)
	if err != nil {
		fmt.Println("Failed to connect to the SMTP server:", err)
		return
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		fmt.Println("Authentication error:", err)
		return
	}

	if err = client.Mail(sender.Email); err != nil {
		fmt.Println("Error setting sender:", err)
		return
	}
	if err = client.Rcpt(recipient.Email); err != nil {
		fmt.Println("Error setting recipient:", err)
		return
	}

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", sender.Name, sender.Email)
	headers["To"] = recipient.Email
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=utf-8"

	var emailBuffer bytes.Buffer
	for key, value := range headers {
		emailBuffer.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	emailBuffer.WriteString("\r\n")
	emailBuffer.WriteString(htmlContent)

	w, err := client.Data()
	if err != nil {
		fmt.Println("Error preparing data:", err)
		return
	}
	defer w.Close()

	_, err = w.Write(emailBuffer.Bytes())
	if err != nil {
		fmt.Println("Error writing message:", err)
		return
	}

	fmt.Println("Email sent successfully!")
	return
}
