package smtp

import (
	"context"
	"crypto/tls"
	"os"

	mail "github.com/go-mail/mail/v2"
)

type EmailSender struct {
	dialer *mail.Dialer
	from   string
}

func NewEmailSender() *EmailSender {
	host := os.Getenv("SMTP_HOST")
	port := 587
	if v := os.Getenv("SMTP_PORT"); v != "" {
	}
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	dialer := mail.NewDialer(host, port, user, pass)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: false}

	return &EmailSender{
		dialer: dialer,
		from:   user,
	}
}

func (s *EmailSender) Send(ctx context.Context, to, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	return s.dialer.DialAndSend(m)
}
