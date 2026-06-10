package mailer

import (
	"os"
	"strconv"

	domainMailer "github.com/gbrayhan/microservices-go/src/domain/ports"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

type GoMailer struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewGoMailer() domainMailer.IMailer {
	// Load .env if present so SMTP configuration can come from a local file.
	_ = godotenv.Load()

	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	port := 587
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	username := os.Getenv("SMTP_USER")
	if username == "" {
		username = os.Getenv("SMTP_USERNAME")
	}
	password := os.Getenv("SMTP_PASS")
	if password == "" {
		password = os.Getenv("SMTP_PASSWORD")
	}

	return &GoMailer{Host: host, Port: port, Username: username, Password: password}
}

func (m *GoMailer) Send(msg domainMailer.EmailMessage) error {
	gm := gomail.NewMessage()
	gm.SetHeader("From", "testing@gmail.com")
	gm.SetHeader("To", msg.To...)
	gm.SetHeader("Subject", msg.Subject)

	contentType := "text/plain"
	if msg.IsHTML {
		contentType = "text/html"
	}
	gm.SetBody(contentType, msg.Body)

	d := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)
	return d.DialAndSend(gm)
}
