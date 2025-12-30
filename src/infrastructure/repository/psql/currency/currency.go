package currency

import (
	"encoding/json"
	"time"

	domainCurrency "github.com/gbrayhan/microservices-go/src/domain/currency"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Currency struct {
	ID        int       `gorm:"primaryKey"`
	Name      string    `gorm:"column:currency_name"`
	Rate      float64   `gorm:"column:rate"`
	Code      string    `gorm:"column:code;unique"`
	Status    bool      `gorm:"column:status"`
	CreatedAt time.Time `gorm:"autoCreateTime:mili"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:mili"`
}

func (Currency) TableName() string {
	return "currencies"
}

var ColumnsUserMapping = map[string]string{
	"id":        "id",
	"name":      "currency_name",
	"rate":      "rate",
	"code":      "code",
	"status":    "status",
	"createdAt": "created_at",
	"updatedAt": "updated_at",
}

// UserRepositoryInterface defines the interface for user repository operations
type CurrencyRepositoryInterface interface {
	GetAll() (*[]domainCurrency.Currency, error)
	Create(currencyDomain *domainCurrency.Currency) (*domainCurrency.Currency, error)
	GetByID(id int) (*domainCurrency.Currency, error)
	Update(id int, currencyMap map[string]interface{}) (*domainCurrency.Currency, error)
	Delete(id int) error
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewCurrencyRepository(db *gorm.DB, loggerInstance *logger.Logger) CurrencyRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]domainCurrency.Currency, error) {
	var users []Currency
	if err := r.DB.Find(&users).Error; err != nil {
		r.Logger.Error("Error getting all users", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved all users", zap.Int("count", len(users)))
	return arrayToDomainMapper(&users), nil
}

func (r *Repository) Create(currencyDomain *domainCurrency.Currency) (*domainCurrency.Currency, error) {
	r.Logger.Info("Creating new user", zap.String("code", currencyDomain.Code))
	userRepository := fromDomainMapper(currencyDomain)
	txDb := r.DB.Create(userRepository)
	err := txDb.Error
	if err != nil {
		r.Logger.Error("Error creating user", zap.Error(err), zap.String("code", currencyDomain.Code))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		errUnmarshal := json.Unmarshal(byteErr, &newError)
		if errUnmarshal != nil {
			return &domainCurrency.Currency{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			err = domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			return &domainCurrency.Currency{}, err
		default:
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	r.Logger.Info("Successfully created user", zap.String("code", currencyDomain.Code), zap.Int("id", userRepository.ID))
	return userRepository.toDomainMapper(), err
}

func (r *Repository) GetByID(id int) (*domainCurrency.Currency, error) {
	var user Currency
	err := r.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("User not found", zap.Int("id", id))
			err = domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		} else {
			r.Logger.Error("Error getting user by ID", zap.Error(err), zap.Int("id", id))
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
		return &domainCurrency.Currency{}, err
	}
	r.Logger.Info("Successfully retrieved user by ID", zap.Int("id", id))
	return user.toDomainMapper(), nil
}

func (r *Repository) Update(id int, userMap map[string]interface{}) (*domainCurrency.Currency, error) {
	var userObj Currency
	userObj.ID = id

	// Map JSON field names to DB column names
	updateData := make(map[string]interface{})
	for k, v := range userMap {
		if column, ok := ColumnsUserMapping[k]; ok {
			updateData[column] = v
		} else {
			updateData[k] = v
		}
	}

	err := r.DB.Model(&userObj).
		Select("currency_name", "code", "rate", "status").
		Updates(updateData).Error
	if err != nil {
		r.Logger.Error("Error updating currency", zap.Error(err), zap.Int("id", id))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		errUnmarshal := json.Unmarshal(byteErr, &newError)
		if errUnmarshal != nil {
			return &domainCurrency.Currency{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			return &domainCurrency.Currency{}, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
		default:
			return &domainCurrency.Currency{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	if err := r.DB.Where("id = ?", id).First(&userObj).Error; err != nil {
		r.Logger.Error("Error retrieving updated currency", zap.Error(err), zap.Int("id", id))
		return &domainCurrency.Currency{}, err
	}
	r.Logger.Info("Successfully updated currency", zap.Int("id", id))
	return userObj.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	tx := r.DB.Delete(&Currency{}, id)
	if tx.Error != nil {
		r.Logger.Error("Error deleting user", zap.Error(tx.Error), zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	if tx.RowsAffected == 0 {
		r.Logger.Warn("User not found for deletion", zap.Int("id", id))
		return domainErrors.NewAppErrorWithType(domainErrors.NotFound)
	}
	r.Logger.Info("Successfully deleted user", zap.Int("id", id))
	return nil
}

// Mappers
func (u *Currency) toDomainMapper() *domainCurrency.Currency {
	return &domainCurrency.Currency{
		ID:        u.ID,
		Name:      u.Name,
		Code:      u.Code,
		Rate:      u.Rate,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func fromDomainMapper(u *domainCurrency.Currency) *Currency {
	return &Currency{
		ID:        u.ID,
		Name:      u.Name,
		Code:      u.Code,
		Rate:      u.Rate,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func arrayToDomainMapper(users *[]Currency) *[]domainCurrency.Currency {
	usersDomain := make([]domainCurrency.Currency, len(*users))
	for i, user := range *users {
		usersDomain[i] = *user.toDomainMapper()
	}
	return &usersDomain
}
