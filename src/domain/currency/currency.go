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

type SearchResultUser struct {
	Data       *[]Currency
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

type ICurrencyService interface {
	GetAll() (*[]Currency, error)
	GetByID(id int) (*Currency, error)
	Create(newUser *Currency) (*Currency, error)
	Delete(id int) error
	Update(id int, userMap map[string]interface{}) (*Currency, error)
}
