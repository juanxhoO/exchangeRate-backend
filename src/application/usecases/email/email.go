package email

import (
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/security"
	"go.uber.org/zap"
)

type IEmailUseCase interface {
	SendEmail(to string, subject string, body string) error
}

type EmailUseCase struct {
	apiService security.IAPIService
	Logger     *logger.Logger
}

func NewEmailUseCase(apiService security.IAPIService, logger *logger.Logger) IEmailUseCase {
	return &EmailUseCase{
		apiService: apiService,
		Logger:     logger,
	}
}

func (s *EmailUseCase) SendEmail(to string, subject string, body string) error {
	s.Logger.Info("Sending email", zap.String("to", to))
	return nil
}
