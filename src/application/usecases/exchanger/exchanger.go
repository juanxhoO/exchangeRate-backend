package exchanger

import (
	"os"

	exchangerDomain "github.com/gbrayhan/microservices-go/src/domain/exchanger"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/exchanger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/security"
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
	apiService          security.IAPIService
	Logger              *logger.Logger
}

func NewExchangerUseCase(exchangerRepository exchanger.ExchangerRepositoryInterface, apiService security.IAPIService, logger *logger.Logger) IExchangerUseCase {
	return &ExchangerUseCase{
		exchangerRepository: exchangerRepository,
		apiService:          apiService,
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

func (s *ExchangerUseCase) Create(newExchanger *exchangerDomain.Exchanger) (*exchangerDomain.Exchanger, error) {
	s.Logger.Info("Creating new user", zap.String("email", newExchanger.Name))
	//Encrypt  the apiKey
	key := os.Getenv("SECRET_API_KEY_GENERATOR")
	var err error
	newExchanger.ApiKey, err = s.apiService.EncryptApiKey(newExchanger.ApiKey, []byte(key))
	if err != nil {
		return nil, err
	}
	return s.exchangerRepository.Create(newExchanger)
}

func (s *ExchangerUseCase) Delete(id int) error {
	s.Logger.Info("Deleting user", zap.Int("id", id))
	return s.exchangerRepository.Delete(id)
}

func (s *ExchangerUseCase) Update(id int, userMap map[string]interface{}) (*exchangerDomain.Exchanger, error) {
	s.Logger.Info("Updating user", zap.Int("id", id))
	//Encrypt  the apiKey
	key := os.Getenv("SECRET_API_KEY_GENERATOR")
	var err error
	userMap["apiKey"], err = s.apiService.EncryptApiKey(userMap["apiKey"].(string), []byte(key))
	if err != nil {
		return nil, err
	}
	return s.exchangerRepository.Update(id, userMap)
}
