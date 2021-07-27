package mail

import (
	"crypto/tls"
	"fmt"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

// Mail encapsulates the dependency
type Mail struct {
	*mail.SMTPClient
	Enabled    bool
	DomainName string
}

// Config represents a configuration to connect to an SMTP server
type Config struct {
	Host       string
	Port       int
	Username   string
	Password   string
	DomainName string
}

// NewMailer creates a new SMTP client
func NewMailer(config Config) (*Mail, error) {
	smtpServer := mail.SMTPServer{
		Host:           config.Host,
		Port:           config.Port,
		Username:       config.Username,
		Password:       config.Password,
		Encryption:     mail.EncryptionTLS,
		Authentication: mail.AuthPlain,
		ConnectTimeout: 10 * time.Second,
		SendTimeout:    10 * time.Second,
		TLSConfig:      &tls.Config{InsecureSkipVerify: true},
	}

	smtpClient, err := smtpServer.Connect()
	if err != nil {
		return &Mail{nil, false, config.DomainName}, err
	}
	return &Mail{smtpClient, true, config.DomainName}, err
}

// SendEmail sends a plaintext email
func (m *Mail) SendEmail(recipient, subject, code string) error {
	body := fmt.Sprintf(`<html><body><a href=https://%s/reset?code=%s>Reset password</a> <p>(Link valid for 1 hour)</p></body></html>`, m.DomainName, code)
	email := mail.NewMSG()
	email.SetFrom("YSTV Security <no-reply@ystv.co.uk>").AddTo(recipient).SetSubject(subject)
	email.SetBody(mail.TextHTML, body)
	return email.Send(m.SMTPClient)
}
