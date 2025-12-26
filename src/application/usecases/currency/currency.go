package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

type NormalizedRate struct {
	Provider string
	Base     string
	Currency string
	Rate     float64
}

func normalizeExchange(
	provider string,
	base string,
	raw map[string]float64,
) []NormalizedRate {
	rates := make([]NormalizedRate, 0, len(raw))

	for currency, rate := range raw {
		rates = append(rates, NormalizedRate{
			Provider: provider,
			Base:     base,
			Currency: currency,
			Rate:     rate,
		})
	}

	return rates
}

func (s *CurrencyUseCase) UpdateExchanges() (any, error) {
	s.Logger.Info("Updating Exchanges in service")
	exchangers, err := s.exchangeRepository.GetAll()
	if err != nil {
		return nil, err
	}

	allRates := []NormalizedRate{}

	for _, exchanger := range *exchangers {
		//first unhash the api key
		decodedApiKey, err := s.apiService.DecryptApiKey(exchanger.ApiKey)
		if err != nil {
			return nil, err
		}
		//now i have to Fetch every exchanger and also unhash de api key to send
		data, err := s.fetchExchangeData(exchanger.Url, decodedApiKey)
		if err != nil {
			return nil, err
		}

		normalizedRate := normalizeExchange(exchanger.Name, "USD", data)
		allRates = append(allRates, normalizedRate...)
		s.Logger.Info("Exchange rates fetched", zap.Any("rates", allRates))
	}
	return nil, nil
}

type AggregatedRate struct {
	Base     string
	Currency string
	Rate     float64
	Sources  int
}

func aggregateRates(rates []NormalizedRate) map[string]AggregatedRate {
	acc := make(map[string]struct {
		sum      float64
		count    int
		base     string
		currency string
	})

	for _, r := range rates {
		key := r.Base + "_" + r.Currency

		v := acc[key]
		v.sum += r.Rate
		v.count++
		v.base = r.Base
		v.currency = r.Currency
		acc[key] = v
	}

	result := make(map[string]AggregatedRate)

	for key, v := range acc {
		result[key] = AggregatedRate{
			Base:     v.base,
			Currency: v.currency,
			Rate:     v.sum / float64(v.count),
			Sources:  v.count,
		}
	}
	return result
}

type ExchangeResponse struct {
	Data map[string]float64 `json:"data"`
}

func (s *CurrencyUseCase) fetchExchangeData(
	url string,
	apiKey string,
) (map[string]float64, error) {

	s.Logger.Info("Fetching exchange data", zap.String("url", url+"?apikey="+apiKey))
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
