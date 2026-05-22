package smtp

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"strconv"

	mail "github.com/go-mail/mail/v2"
)

type EmailSender struct {
	dialer *mail.Dialer
	from   string
	mock   bool
}

func NewEmailSender() *EmailSender {
	host := os.Getenv("SMTP_HOST")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	if host == "" || user == "" || pass == "" || user == "your-email@gmail.com" {
		log.Println("SMTP not configured — email sending is MOCKED")
		return &EmailSender{mock: true}
	}

	port := 587
	if v := os.Getenv("SMTP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	dialer := mail.NewDialer(host, port, user, pass)
	dialer.TLSConfig = &tls.Config{
		ServerName: host,
	}

	return &EmailSender{
		dialer: dialer,
		from:   user,
	}
}

func (s *EmailSender) Send(ctx context.Context, to, subject, body string) error {
	if s.mock {
		log.Printf("MOCK EMAIL SENT: To=%s | Subject=%s | Body=%s", to, subject, body)
		return nil
	}

	m := mail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	return s.dialer.DialAndSend(m)
}
