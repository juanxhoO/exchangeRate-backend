package currency

import (
	currencyDomain "github.com/gbrayhan/microservices-go/src/domain/currency"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/currency"
	"go.uber.org/zap"
)

type ICurrencyUseCase interface {
	GetAll() (*[]currencyDomain.Currency, error)
	GetByID(id int) (*currencyDomain.Currency, error)
	Create(newUser *currencyDomain.Currency) (*currencyDomain.Currency, error)
	Delete(id int) error
	Update(id int, userMap map[string]interface{}) (*currencyDomain.Currency, error)
}

type CurrencyUseCase struct {
	userRepository currency.CurrencyRepositoryInterface
	Logger         *logger.Logger
}

func NewCurrencyUseCase(userRepository currency.CurrencyRepositoryInterface, logger *logger.Logger) ICurrencyUseCase {
	return &CurrencyUseCase{
		userRepository: userRepository,
		Logger:         logger,
	}
}

func (s *CurrencyUseCase) GetAll() (*[]currencyDomain.Currency, error) {
	s.Logger.Info("Getting all users")
	return s.userRepository.GetAll()
}

func (s *CurrencyUseCase) GetByID(id int) (*currencyDomain.Currency, error) {
	s.Logger.Info("Getting user by ID", zap.Int("id", id))
	return s.userRepository.GetByID(id)
}

func (s *CurrencyUseCase) Create(newUser *currencyDomain.Currency) (*currencyDomain.Currency, error) {
	s.Logger.Info("Creating new user", zap.String("email", newUser.Name))
	newUser.Status = true
	return s.userRepository.Create(newUser)
}

func (s *CurrencyUseCase) Delete(id int) error {
	s.Logger.Info("Deleting user", zap.Int("id", id))
	return s.userRepository.Delete(id)
}

func (s *CurrencyUseCase) Update(id int, userMap map[string]interface{}) (*currencyDomain.Currency, error) {
	s.Logger.Info("Updating user", zap.Int("id", id))
	return s.userRepository.Update(id, userMap)
}
