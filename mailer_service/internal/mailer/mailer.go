package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

type Email struct {
	Recipient string
	File      string
	Data      any
}

func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// execute templates into a buffer - this also helps us identify if there are
	// any errors in our template definitions
	subjectBuffer := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subjectBuffer, "subject", data)
	if err != nil {
		return err
	}

	plainBodyBuffer := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBodyBuffer, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBodyBuffer := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBodyBuffer, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subjectBuffer.String())
	msg.SetBody("text/plain", plainBodyBuffer.String())
	msg.AddAlternative("text/html", htmlBodyBuffer.String())

	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
