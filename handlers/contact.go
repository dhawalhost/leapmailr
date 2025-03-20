package handlers

import (
	"fmt"
	"net/http"

	"github.com/dhawalhost/leapmailr/config"
	"github.com/dhawalhost/leapmailr/models"
	"github.com/dhawalhost/leapmailr/service"
	"github.com/dhawalhost/leapmailr/utils"

	"github.com/gin-gonic/gin"
)

// HandleContactForm handles the contact form submission
func HandleContactForm(c *gin.Context) {
	var form models.ContactForm
	conf := config.GetConfig()
	if err := c.BindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	smtpServer := models.SMTPDetails{
		Server: conf.SMTPServer,
		Port:   conf.SMTPPort,
		Email:  conf.SMTPMail,
		Secret: conf.SMTPSecret,
	}

	// Sender and recipient details
	sender := models.Sender{Name: conf.CompanyName, Email: conf.DefaultSenderMail}
	// Validate email format

	if !utils.IsValidEmail(form.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address"})
		return
	}
	recipient := models.Recipient{Name: form.Name, Email: form.Email}

	// Sending to Person who contacted
	err := service.SendReply(sender, recipient, smtpServer)
	if err != nil {
		fmt.Println("Failed to Send Mail", err)
	}

	// Sending Form details to the company mail
	recipient.Name = conf.CompanyName
	recipient.Email = conf.ContactMail
	err = service.SubmitForm(sender, recipient, form, smtpServer)
	if err != nil {
		fmt.Println("Failed to Send Mail", err)
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{
		"status":  "Form submitted successfully",
		"name":    form.Name,
		"email":   form.Email,
		"subject": form.Subject,
		"message": form.Message,
	})
}
