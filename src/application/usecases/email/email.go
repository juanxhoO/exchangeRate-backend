package email

import (
	domain "github.com/gbrayhan/microservices-go/src/domain/ports"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
)

type IEmailUseCase interface {
	SendEmail(to string, subject string, body string) error
}

type EmailUseCase struct {
	mailer domain.IMailer
	Logger *logger.Logger
}

func NewEmailUseCase(mailer domain.IMailer, logger *logger.Logger) IEmailUseCase {
	return &EmailUseCase{
		mailer: mailer,
		Logger: logger,
	}
}

func (s *EmailUseCase) SendEmail(to string, subject string, body string) error {
		
	s.Logger.Info("Sending email", zap.String("to", to), zap.String("subject", subject), zap.String("body", body))

	return s.mailer.Send(domain.EmailMessage{
			To:      []string{to},
			Subject: subject,
			Body:    body,
		})
}
