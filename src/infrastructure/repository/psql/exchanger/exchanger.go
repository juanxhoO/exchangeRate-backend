package exchanger

import (
	"encoding/json"
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainExchanger "github.com/gbrayhan/microservices-go/src/domain/exchanger"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Exchanger struct {
	ID        int       `gorm:"primaryKey"`
	Name      string    `gorm:"column:user_name;unique"`
	ApiKey    string    `gorm:"unique"`
	IsActive  bool      `gorm:"unique"`
	CreatedAt time.Time `gorm:"autoCreateTime:mili"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:mili"`
}

func (Exchanger) TableName() string {
	return "exchangers"
}

var ColumnsUserMapping = map[string]string{
	"id":           "id",
	"userName":     "user_name",
	"email":        "email",
	"firstName":    "first_name",
	"lastName":     "last_name",
	"status":       "status",
	"hashPassword": "hash_password",
	"createdAt":    "created_at",
	"updatedAt":    "updated_at",
}

// UserRepositoryInterface defines the interface for user repository operations
type ExchangerRepositoryInterface interface {
	GetAll() (*[]domainExchanger.Exchanger, error)
	Create(userDomain *domainExchanger.Exchanger) (*domainExchanger.Exchanger, error)
	GetByID(id int) (*domainExchanger.Exchanger, error)
	Update(id int, userMap map[string]interface{}) (*domainExchanger.Exchanger, error)
	Delete(id int) error
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewUserRepository(db *gorm.DB, loggerInstance *logger.Logger) ExchangerRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]domainExchanger.Exchanger, error) {
	var users []Exchanger
	if err := r.DB.Find(&users).Error; err != nil {
		r.Logger.Error("Error getting all users", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved all users", zap.Int("count", len(users)))
	return arrayToDomainMapper(&users), nil
}

func (r *Repository) Create(userDomain *domainExchanger.Exchanger) (*domainExchanger.Exchanger, error) {
	r.Logger.Info("Creating new user", zap.String("email", userDomain.Name))
	userRepository := fromDomainMapper(userDomain)
	txDb := r.DB.Create(userRepository)
	err := txDb.Error
	if err != nil {
		r.Logger.Error("Error creating user", zap.Error(err), zap.String("email", userDomain.Name))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		errUnmarshal := json.Unmarshal(byteErr, &newError)
		if errUnmarshal != nil {
			return &domainExchanger.Exchanger{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			err = domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
			return &domainExchanger.Exchanger{}, err
		default:
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	r.Logger.Info("Successfully created user", zap.String("email", userDomain.Name), zap.Int("id", userRepository.ID))
	return userRepository.toDomainMapper(), err
}

func (r *Repository) GetByID(id int) (*domainExchanger.Exchanger, error) {
	var user Exchanger
	err := r.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("User not found", zap.Int("id", id))
			err = domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		} else {
			r.Logger.Error("Error getting user by ID", zap.Error(err), zap.Int("id", id))
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
		return &domainExchanger.Exchanger{}, err
	}
	r.Logger.Info("Successfully retrieved user by ID", zap.Int("id", id))
	return user.toDomainMapper(), nil
}

func (r *Repository) GetByEmail(email string) (*domainExchanger.Exchanger, error) {
	var user Exchanger
	err := r.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.Logger.Warn("User not found", zap.String("email", email))
			err = domainErrors.NewAppErrorWithType(domainErrors.NotFound)
		} else {
			r.Logger.Error("Error getting user by email", zap.Error(err), zap.String("email", email))
			err = domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
		return &domainExchanger.Exchanger{}, err
	}
	r.Logger.Info("Successfully retrieved user by email", zap.String("email", email))
	return user.toDomainMapper(), nil
}

func (r *Repository) Update(id int, userMap map[string]interface{}) (*domainExchanger.Exchanger, error) {
	var userObj Exchanger
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
		Select("user_name", "email", "first_name", "last_name", "status", "role").
		Updates(updateData).Error
	if err != nil {
		r.Logger.Error("Error updating user", zap.Error(err), zap.Int("id", id))
		byteErr, _ := json.Marshal(err)
		var newError domainErrors.GormErr
		errUnmarshal := json.Unmarshal(byteErr, &newError)
		if errUnmarshal != nil {
			return &domainExchanger.Exchanger{}, errUnmarshal
		}
		switch newError.Number {
		case 1062:
			return &domainExchanger.Exchanger{}, domainErrors.NewAppErrorWithType(domainErrors.ResourceAlreadyExists)
		default:
			return &domainExchanger.Exchanger{}, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		}
	}
	if err := r.DB.Where("id = ?", id).First(&userObj).Error; err != nil {
		r.Logger.Error("Error retrieving updated user", zap.Error(err), zap.Int("id", id))
		return &domainExchanger.Exchanger{}, err
	}
	r.Logger.Info("Successfully updated user", zap.Int("id", id))
	return userObj.toDomainMapper(), nil
}

func (r *Repository) Delete(id int) error {
	tx := r.DB.Delete(&Exchanger{}, id)
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
func (u *Exchanger) toDomainMapper() *domainExchanger.Exchanger {
	return &domainExchanger.Exchanger{
		ID:        u.ID,
		Name:      u.Name,
		IsActive:  u.IsActive,
		ApiKey:    u.ApiKey,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func fromDomainMapper(u *domainExchanger.Exchanger) *Exchanger {
	return &Exchanger{
		ID:        u.ID,
		Name:      u.Name,
		IsActive:  u.IsActive,
		ApiKey:    u.ApiKey,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func arrayToDomainMapper(users *[]Exchanger) *[]domainExchanger.Exchanger {
	usersDomain := make([]domainExchanger.Exchanger, len(*users))
	for i, user := range *users {
		usersDomain[i] = *user.toDomainMapper()
	}
	return &usersDomain
}
