package models

type Sender struct {
	Name  string
	Email string
}
type Recipient struct {
	Name    string
	Email   string
	Subject string
	Message string
}

type SMTPDetails struct {
	Server string
	Port   int
	Email  string
	Secret string
}
