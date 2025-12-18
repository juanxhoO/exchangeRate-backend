package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"os"

	currencyDomain "github.com/gbrayhan/microservices-go/src/domain/currency"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/currency"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/exchanger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/security"
	"go.uber.org/zap"
)

type ICurrencyUseCase interface {
	GetAll() (*[]currencyDomain.Currency, error)
	GetByID(id int) (*currencyDomain.Currency, error)
	Delete(id int) error
	UpdateExchanges() (any, error)
}

type CurrencyUseCase struct {
	currencyRepository currency.CurrencyRepositoryInterface
	exchangeRepository exchanger.ExchangerRepositoryInterface
	apiService         security.IAPIService
	Logger             *logger.Logger
}

func NewCurrencyUseCase(currencyRepository currency.CurrencyRepositoryInterface, exchangeRepository exchanger.ExchangerRepositoryInterface, apiService security.IAPIService, logger *logger.Logger) ICurrencyUseCase {
	return &CurrencyUseCase{
		currencyRepository: currencyRepository,
		exchangeRepository: exchangeRepository,
		apiService:         apiService,
		Logger:             logger,
	}
}

func GetProviderRates(ctx context.Context) (map[string]float64, error) {
	req, _ := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://api.exchangerate-api.com/v4/latest/USD",
		nil,
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		Base  string             `json:"base"`
		Rates map[string]float64 `json:"rates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Rates, nil
}

func (s *CurrencyUseCase) GetAll() (*[]currencyDomain.Currency, error) {
	s.Logger.Info("Getting all users")
	return s.currencyRepository.GetAll()
}

func (s *CurrencyUseCase) GetByID(id int) (*currencyDomain.Currency, error) {
	s.Logger.Info("Getting user by ID", zap.Int("id", id))
	return s.currencyRepository.GetByID(id)
}

func (s *CurrencyUseCase) Delete(id int) error {
	s.Logger.Info("Deleting currency", zap.Int("id", id))
	return s.currencyRepository.Delete(id)
}

func (s *CurrencyUseCase) UpdateExchanges() (any, error) {
	s.Logger.Info("Updating Exchanges in service")
	exchangers, err := s.exchangeRepository.GetAll()
	if err != nil {
		return nil, err
	}
	//Fetch the data of each Exchanger and Unhash the api key
	for _, exchanger := range *exchangers {
		//first unhash the api key
		decodedApiKey, err := s.apiService.DecryptApiKey(exchanger.ApiKey, []byte(os.Getenv("SECRET_API_KEY_GENERATOR")))
		if err != nil {
			return nil, err
		}
		s.Logger.Info("Decoded Api Key", zap.String("apiKey", decodedApiKey))

		//now i have to Fetch every exchanger and also unhash de api key to send

		data, err := s.fetchExchangeData(exchanger.Url, decodedApiKey)
		if err != nil {
			return nil, err
		}

		s.Logger.Info(
			"Exchange rates fetched",
			zap.Any("rates", data),
		)
	}
	return nil, nil
}

type ExchangeResponse struct {
	Data map[string]float64 `json:"data"`
}

func (s *CurrencyUseCase) fetchExchangeData(
	url string,
	apiKey string,
) (map[string]float64, error) {

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		url+"?apikey="+apiKey,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("provider error %d: %s", resp.StatusCode, body)
	}

	var payload ExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload.Data, nil
}
