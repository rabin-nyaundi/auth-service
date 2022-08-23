package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// ↓↓↓
//go:embed "templates"
var templateFs embed.FS

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

/**
* Send email
* recipient, templatefile, data to be sent(interface)
 */

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFs, "templates/"+templateFile)

	if err != nil {
		fmt.Println("Error is here")
		return err
	}

	subject := new(bytes.Buffer)

	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)

	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)

	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)

	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", htmlBody.String())
	msg.AddAlternative("text/plain", plainBody.String())

	err = m.dialer.DialAndSend(msg)
	if err != nil {
		return err
	}

	return nil
}
