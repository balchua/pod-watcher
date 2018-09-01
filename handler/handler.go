package handler

import (
	"crypto/tls"

	"github.com/Sirupsen/logrus"
	"github.com/go-mail/mail"
)

// MailHandler object
type MailHandler struct {
	m *mail.Dialer
}

//Mail is the message to send out.
type Mail struct {
	m *mail.Message
}

//NewMail contructs a mail message
func NewMail(from string, to string, subject string, body string) Mail {
	m := mail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	return Mail{
		m: m,
	}

}

//New returns a new mailHandler
func NewMailHandler(smtp string, port int, user string, password string) MailHandler {

	dialer := mail.NewDialer(smtp, port, user, password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return MailHandler{
		m: dialer,
	}
}

// Notify do the actual sending of emails.
func (h *MailHandler) Notify(mail Mail) {
	//TODO send the mail from here.
	// Send the email to Bob, Cora and Dan.
	if err := h.m.DialAndSend(mail.m); err != nil {
		logrus.Error("Unable to send mail.")
	} else {
		logrus.Info("Mail successfully sent.")
	}
}
