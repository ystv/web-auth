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
	Repo interface {
		AddDefaults(defaults Defaults)
		CheckSendable(item Mail) error
		SendMail(item Mail) error
		SendErrorFatalMail(item Mail) error
		parseHeader(item Mail) (to, from string, cc, bcc []string)
	}

	// Mailer encapsulates the dependency
	Mailer struct {
		*mail.SMTPClient
		Defaults   Defaults
		Enabled    bool
		DomainName string
	}

	Defaults struct {
		DefaultTo   string
		DefaultCC   []string
		DefaultBCC  []string
		DefaultFrom string
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

var _ Repo = &Mailer{}

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
		return &Mailer{nil, Defaults{}, false, config.DomainName}, err
	}
	return &Mailer{smtpClient, Defaults{}, true, config.DomainName}, err
}

// AddDefaults adds the default recipients
func (m *Mailer) AddDefaults(defaults Defaults) {
	m.Defaults.DefaultTo = defaults.DefaultTo
	m.Defaults.DefaultCC = defaults.DefaultCC
	m.Defaults.DefaultBCC = defaults.DefaultBCC
	m.Defaults.DefaultFrom = defaults.DefaultFrom
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
	to, from, cc, bcc := m.parseHeader(item)
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

// SendErrorFatalMail sends a standard template error fatal email
func (m *Mailer) SendErrorFatalMail(item Mail) error {
	err := m.CheckSendable(item)
	if err != nil {
		return err
	}
	to, from, cc, bcc := m.parseHeader(item)
	errorTemplate := template.New("Fatal Error Template")
	errorTemplate = template.Must(errorTemplate.Parse("<body><p style=\"color: red;\">A <b>FATAL ERROR</b> OCCURRED!<br><br><code>{{.}}</code></p><br><br>We apologise for the inconvenience.</body>"))
	body := bytes.Buffer{}
	err = errorTemplate.Execute(&body, item.Error)
	if err != nil {
		return fmt.Errorf("failed to exec tpl: %w", err)
	}
	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject("FATAL ERROR - YSTV STV")
	if len(cc) != 0 {
		email.AddCc(cc...)
	}
	if len(bcc) != 0 {
		email.AddBcc(bcc...)
	}
	email.SetBody(mail.TextHTML, body.String())
	if email.Error != nil {
		return fmt.Errorf("failed to set mail data: %w", email.Error)
	}
	return email.Send(m.SMTPClient)
}

func (m *Mailer) parseHeader(item Mail) (to, from string, cc, bcc []string) {
	if len(item.To) > 0 {
		to = item.To
	} else {
		to = m.Defaults.DefaultTo
	}

	if len(item.Cc) > 0 {
		cc = item.Cc
	} else {
		cc = m.Defaults.DefaultCC
	}

	if len(item.Bcc) > 0 {
		bcc = item.Bcc
	} else {
		bcc = m.Defaults.DefaultBCC
	}

	if len(item.From) > 0 {
		from = item.From
	} else {
		from = m.Defaults.DefaultFrom
	}
	return
}
