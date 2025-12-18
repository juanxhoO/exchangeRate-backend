package currency

import (
	"time"
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

type ICurrencyService interface {
	GetAll() (*[]Currency, error)
	GetByID(id int) (*Currency, error)
	Delete(id int) error
	UpdateExchanges() (any, error)
}
