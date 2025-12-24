package exchanger

import (
	"time"
)

type Exchanger struct {
	ID        int
	Name      string
	ApiKey    string
	Url       string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type IExchangerService interface {
	GetAll() (*[]Exchanger, error)
	GetByID(id int) (*Exchanger, error)
	Create(newUser *Exchanger) (*Exchanger, error)
	Delete(id int) error
	Update(id int, userMap map[string]interface{}) (*Exchanger, error)
}
