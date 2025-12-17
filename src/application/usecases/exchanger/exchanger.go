package exchanger

import (
	exchangerDomain "github.com/gbrayhan/microservices-go/src/domain/exchanger"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/exchanger"
	"go.uber.org/zap"
)

type IExchangerUseCase interface {
	GetAll() (*[]exchangerDomain.Exchanger, error)
	GetByID(id int) (*exchangerDomain.Exchanger, error)
	Create(newUser *exchangerDomain.Exchanger) (*exchangerDomain.Exchanger, error)
	Delete(id int) error
	Update(id int, userMap map[string]interface{}) (*exchangerDomain.Exchanger, error)
}

type ExchangerUseCase struct {
	exchangerRepository exchanger.ExchangerRepositoryInterface
	Logger              *logger.Logger
}

func NewExchangerUseCase(exchangerRepository exchanger.ExchangerRepositoryInterface, logger *logger.Logger) IExchangerUseCase {
	return &ExchangerUseCase{
		exchangerRepository: exchangerRepository,
		Logger:              logger,
	}
}

func (s *ExchangerUseCase) GetAll() (*[]exchangerDomain.Exchanger, error) {
	s.Logger.Info("Getting all users")
	return s.exchangerRepository.GetAll()
}

func (s *ExchangerUseCase) GetByID(id int) (*exchangerDomain.Exchanger, error) {
	s.Logger.Info("Getting user by ID", zap.Int("id", id))
	return s.exchangerRepository.GetByID(id)
}

func (s *ExchangerUseCase) Create(newUser *exchangerDomain.Exchanger) (*exchangerDomain.Exchanger, error) {
	s.Logger.Info("Creating new user", zap.String("email", newUser.Name))
	newUser.IsActive = true

	return s.exchangerRepository.Create(newUser)
}

func (s *ExchangerUseCase) Delete(id int) error {
	s.Logger.Info("Deleting user", zap.Int("id", id))
	return s.exchangerRepository.Delete(id)
}

func (s *ExchangerUseCase) Update(id int, userMap map[string]interface{}) (*exchangerDomain.Exchanger, error) {
	s.Logger.Info("Updating user", zap.Int("id", id))
	return s.exchangerRepository.Update(id, userMap)
}
