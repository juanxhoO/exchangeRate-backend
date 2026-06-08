package currency

import (
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
)

type Currency struct {
	ID        int
	Name      string
	Code      string
	Rate      float64
	Status    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SearchResultCurrency struct {
	Data       *[]Currency
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

type ICurrencyService interface {
	GetAll() (*[]Currency, error)
	GetByID(id int) (*Currency, error)
	Delete(id int) error
	UpdateExchanges() (any, error)
	SearchPaginated(filters domain.DataFilters) (*SearchResultCurrency, error)
}
