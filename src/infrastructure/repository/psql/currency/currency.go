package currency

import (
	"encoding/json"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
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
	SearchPaginated(filters domain.DataFilters) (*domainCurrency.SearchResultCurrency, error)
}

type Repository struct {
	DB     *gorm.DB
	Logger *logger.Logger
}

func NewCurrencyRepository(db *gorm.DB, loggerInstance *logger.Logger) CurrencyRepositoryInterface {
	return &Repository{DB: db, Logger: loggerInstance}
}

func (r *Repository) GetAll() (*[]domainCurrency.Currency, error) {
	var currencies []Currency
	if err := r.DB.Find(&currencies).Error; err != nil {
		r.Logger.Error("Error getting all currencies", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}
	r.Logger.Info("Successfully retrieved all currencies", zap.Int("count", len(currencies)))
	return arrayToDomainMapper(&currencies), nil
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
func (r *Repository) SearchPaginated(filters domain.DataFilters) (*domainCurrency.SearchResultCurrency, error) {
	query := r.DB.Model(&Currency{})

	r.Logger.Info("Building search query with filters", zap.Any("filters", filters))
	// Apply like filters
	for field, values := range filters.LikeFilters {
		if len(values) > 0 {
			for _, value := range values {
				if value != "" {
					column := ColumnsUserMapping[field]
					if column != "" {
						query = query.Where(column+" ILIKE ?", "%"+value+"%")
					}
				}
			}
		}
	}

	// Apply exact matches
	for field, values := range filters.Matches {
		if len(values) > 0 {
			column := ColumnsUserMapping[field]
			if column != "" {
				query = query.Where(column+" IN ?", values)
			}
		}
	}

	// Apply date range filters
	for _, dateFilter := range filters.DateRangeFilters {
		column := ColumnsUserMapping[dateFilter.Field]
		if column != "" {
			if dateFilter.Start != nil {
				query = query.Where(column+" >= ?", dateFilter.Start)
			}
			if dateFilter.End != nil {
				query = query.Where(column+" <= ?", dateFilter.End)
			}
		}
	}

	// Apply sorting
	if len(filters.SortBy) > 0 && filters.SortDirection.IsValid() {
		for _, sortField := range filters.SortBy {
			column := ColumnsUserMapping[sortField]
			if column != "" {
				query = query.Order(column + " " + string(filters.SortDirection))
			}
		}
	}

	// Count total records
	var total int64
	clonedQuery := query
	clonedQuery.Count(&total)

	// Apply pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 10
	}
	offset := (filters.Page - 1) * filters.PageSize

	var users []Currency
	r.Logger.Info("Searching users", zap.Any("users", users))
	if err := query.Offset(offset).Limit(filters.PageSize).Find(&users).Error; err != nil {
		r.Logger.Error("Error searching users", zap.Error(err))
		return nil, domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
	}

	totalPages := int((total + int64(filters.PageSize) - 1) / int64(filters.PageSize))

	result := &domainCurrency.SearchResultCurrency{
		Data:       arrayToDomainMapper(&users),
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}

	r.Logger.Info("Successfully searched users",
		zap.Int64("total", total),
		zap.Int("page", filters.Page),
		zap.Int("pageSize", filters.PageSize))

	return result, nil
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
