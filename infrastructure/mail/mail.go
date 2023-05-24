package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type (
	// Mailer encapsulates the dependency
	Mailer struct {
		*mail.SMTPClient
		Enabled    bool
		DomainName string
	}

	// Config represents a configuration to connect to an SMTP server
	Config struct {
		Host       string
		Port       int
		Username   string
		Password   string
		DomainName string
	}

	// Mail represents an email to be sent
	Mail struct {
		Subject string
		To      string
		Cc      []string
		Bcc     []string
		From    string
		Error   error
		Tpl     *template.Template
		TplData interface{}
	}
)

// NewMailer creates a new SMTP client
func NewMailer(config Config) (*Mailer, error) {
	smtpServer := mail.SMTPServer{
		Host:           config.Host,
		Port:           config.Port,
		Username:       config.Username,
		Password:       config.Password,
		Encryption:     mail.EncryptionSTARTTLS,
		Authentication: mail.AuthLogin,
		ConnectTimeout: 10 * time.Second,
		SendTimeout:    10 * time.Second,
		TLSConfig:      &tls.Config{InsecureSkipVerify: true},
	}

	smtpClient, err := smtpServer.Connect()
	if err != nil {
		return &Mailer{nil, false, config.DomainName}, err
	}
	return &Mailer{smtpClient, true, config.DomainName}, err
}

// SendResetEmail sends a plaintext email
func (m *Mailer) SendResetEmail(recipient, subject, code string) error {
	body := fmt.Sprintf(`<html><body><a href=https://auth.%s/reset?code=%s>Reset password</a> <p>(Link valid for 1 hour)</p></body></html>`, m.DomainName, code)
	email := mail.NewMSG()
	email.SetFrom("YSTV Security <no-reply@ystv.co.uk>").AddTo(recipient).SetSubject(subject)
	email.SetBody(mail.TextHTML, body)
	return email.Send(m.SMTPClient)
}

// CheckSendable verifies that the email can be sent
func (m *Mailer) CheckSendable(item Mail) error {
	if item.To == "" {
		return fmt.Errorf("no To field is set")
	}
	return nil
}

// SendMail sends a template email
func (m *Mailer) SendMail(item Mail) error {
	err := m.CheckSendable(item)
	if err != nil {
		return err
	}
	var (
		to, from string
		cc, bcc  []string
	)
	to = item.To
	cc = item.Cc
	bcc = item.Bcc
	from = item.From
	body := bytes.Buffer{}
	err = item.Tpl.Execute(&body, item.TplData)
	if err != nil {
		return fmt.Errorf("failed to exec tpl: %w", err)
	}
	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(item.Subject)
	if len(item.Cc) != 0 {
		email.AddCc(cc...)
	}
	if len(item.Bcc) != 0 {
		email.AddBcc(bcc...)
	}
	email.SetBody(mail.TextHTML, body.String())
	if email.Error != nil {
		return fmt.Errorf("failed to set mail data: %w", email.Error)
	}
	return email.Send(m.SMTPClient)
}
