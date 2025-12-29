package currency

import (
	"errors"
	"reflect"
	"testing"

	currencyDomain "github.com/gbrayhan/microservices-go/src/domain/currency"
	exchangerDomain "github.com/gbrayhan/microservices-go/src/domain/exchanger"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	security "github.com/gbrayhan/microservices-go/src/infrastructure/security"
)

type mockAPIService struct{}

type mockExchangerService struct {
	getAllFn  func() (*[]exchangerDomain.Exchanger, error)
	getByIDFn func(id int) (*exchangerDomain.Exchanger, error)
	createFn  func(u *exchangerDomain.Exchanger) (*exchangerDomain.Exchanger, error)
	deleteFn  func(id int) error
	updateFn  func(id int, m map[string]interface{}) (*exchangerDomain.Exchanger, error)
}

type mockUserService struct {
	getAllFn  func() (*[]currencyDomain.Currency, error)
	getByIDFn func(id int) (*currencyDomain.Currency, error)
	createFn  func(u *currencyDomain.Currency) (*currencyDomain.Currency, error)
	deleteFn  func(id int) error
	updateFn  func(id int, m map[string]interface{}) (*currencyDomain.Currency, error)
}

func (m *mockExchangerService) GetAll() (*[]exchangerDomain.Exchanger, error) {
	return m.getAllFn()
}
func (m *mockExchangerService) GetByID(id int) (*exchangerDomain.Exchanger, error) {
	return m.getByIDFn(id)
}
func (m *mockExchangerService) Create(newUser *exchangerDomain.Exchanger) (*exchangerDomain.Exchanger, error) {
	return m.createFn(newUser)
}
func (m *mockExchangerService) Delete(id int) error {
	return m.deleteFn(id)
}
func (m *mockExchangerService) Update(id int, userMap map[string]interface{}) (*exchangerDomain.Exchanger, error) {
	return m.updateFn(id, userMap)
}
func (m *mockAPIService) EncryptApiKey(value string) (string, error) {
	return "encrypted-api-key", nil
}

func (m *mockAPIService) DecryptApiKey(value string) (string, error) {
	return "decrypted-api-key", nil
}

func (m *mockAPIService) GenerateApiKey(length int) (string, error) {
	return "generated-api-key", nil
}
func (m *mockUserService) GetAll() (*[]currencyDomain.Currency, error) {
	return m.getAllFn()
}
func (m *mockUserService) GetByID(id int) (*currencyDomain.Currency, error) {
	return m.getByIDFn(id)
}
func (m *mockUserService) Create(newUser *currencyDomain.Currency) (*currencyDomain.Currency, error) {
	return m.createFn(newUser)
}
func (m *mockUserService) Delete(id int) error {
	return m.deleteFn(id)
}
func (m *mockUserService) Update(id int, userMap map[string]interface{}) (*currencyDomain.Currency, error) {
	return m.updateFn(id, userMap)
}

func setupLogger(t *testing.T) *logger.Logger {
	loggerInstance, err := logger.NewLogger()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return loggerInstance
}

func TestUserUseCase(t *testing.T) {

	mockRepo := &mockUserService{}
	mockRepoExchanger := &mockExchangerService{}
	logger := setupLogger(t)
	apiService := &mockAPIService{}
	useCase := NewCurrencyUseCase(mockRepo, mockRepoExchanger, apiService, logger)

	t.Run("Test GetAll", func(t *testing.T) {
		mockRepo.getAllFn = func() (*[]currencyDomain.Currency, error) {
			return &[]currencyDomain.Currency{{ID: 1}}, nil
		}
		us, err := useCase.GetAll()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(*us) != 1 {
			t.Error("expected 1 user from GetAll")
		}
	})

	t.Run("Test GetByID", func(t *testing.T) {
		mockRepo.getByIDFn = func(id int) (*currencyDomain.Currency, error) {
			if id == 999 {
				return nil, errors.New("not found")
			}
			return &currencyDomain.Currency{ID: id}, nil
		}
		_, err := useCase.GetByID(999)
		if err == nil {
			t.Error("expected error, got nil")
		}
		u, err := useCase.GetByID(10)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if u.ID != 10 {
			t.Errorf("expected user ID=10, got %d", u.ID)
		}
	})

	t.Run("Test Delete", func(t *testing.T) {
		mockRepo.deleteFn = func(id int) error {
			if id == 101 {
				return nil
			}
			return errors.New("cannot delete")
		}
		err := useCase.Delete(999)
		if err == nil {
			t.Error("expected error for cannot delete")
		}
		err = useCase.Delete(101)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestNewUserUseCase(t *testing.T) {
	mockRepo := &mockUserService{}
	mockRepoExchanger := &mockExchangerService{}
	loggerInstance := setupLogger(t)
	apiService := &security.APIService{}
	useCase := NewCurrencyUseCase(mockRepo, mockRepoExchanger, apiService, loggerInstance)
	if reflect.TypeOf(useCase).String() != "*currency.CurrencyUseCase" {
		t.Error("expected *currency.CurrencyUseCase type")
	}
}
