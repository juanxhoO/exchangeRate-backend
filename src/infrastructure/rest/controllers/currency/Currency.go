package currency

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gbrayhan/microservices-go/src/domain"
	domainCurrency "github.com/gbrayhan/microservices-go/src/domain/currency"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/currency"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/user"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Structures
type NewCurrencyRequest struct {
	Name   string  `json:"user" binding:"required"`
	Code   string  `json:"email" binding:"required"`
	Status bool    `json:"status" binding:"required"`
	Rate   float64 `json:"firstName" binding:"required"`
}

type ResponseUser struct {
	ID        int       `json:"id"`
	Name      string    `json:"user"`
	Code      string    `json:"email"`
	Rate      float64   `json:"firstName"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type ICurrencyController interface {
	GetAllCurrencies(ctx *gin.Context)
	GetCurrenciesByID(ctx *gin.Context)
	DeleteCurrency(ctx *gin.Context)
	UpdateExchanges(ctx *gin.Context)
	SearchPaginated(ctx *gin.Context)
}

type CurrencyController struct {
	currencyService domainCurrency.ICurrencyService
	Logger          *logger.Logger
}

func NewCurrencyController(currencyService domainCurrency.ICurrencyService, loggerInstance *logger.Logger) ICurrencyController {
	return &CurrencyController{currencyService: currencyService, Logger: loggerInstance}
}

func (c *CurrencyController) GetAllCurrencies(ctx *gin.Context) {
	c.Logger.Info("Getting all users")
	users, err := c.currencyService.GetAll()
	if err != nil {
		c.Logger.Error("Error getting all users", zap.Error(err))
		appError := domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Successfully retrieved all users", zap.Int("count", len(*users)))
	ctx.JSON(http.StatusOK, arrayDomainToResponseMapper(users))
}

func (c *CurrencyController) UpdateExchanges(ctx *gin.Context) {
	c.Logger.Info("Updating Exchanges")
	_, err := c.currencyService.UpdateExchanges()
	if err != nil {
		c.Logger.Error("Error updating exchanges", zap.Error(err))
		appError := domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		_ = ctx.Error(appError)
		return
	}

	c.Logger.Info("Successfully updated exchanges")
	ctx.JSON(http.StatusOK, gin.H{"message": "Exchanges updated successfully"})
}

func (c *CurrencyController) GetCurrenciesByID(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Error("Invalid user ID parameter", zap.Error(err), zap.String("id", ctx.Param("id")))
		appError := domainErrors.NewAppError(errors.New("user id is invalid"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Getting user by ID", zap.Int("id", userID))
	user, err := c.currencyService.GetByID(userID)
	if err != nil {
		c.Logger.Error("Error getting user by ID", zap.Error(err), zap.Int("id", userID))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("Successfully retrieved user by ID", zap.Int("id", userID))
	ctx.JSON(http.StatusOK, domainToResponseMapper(user))
}

func (c *CurrencyController) DeleteCurrency(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Error("Invalid user ID parameter for deletion", zap.Error(err), zap.String("id", ctx.Param("id")))
		appError := domainErrors.NewAppError(errors.New("param id is necessary"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Deleting user", zap.Int("id", userID))
	err = c.currencyService.Delete(userID)
	if err != nil {
		c.Logger.Error("Error deleting user", zap.Error(err), zap.Int("id", userID))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("User deleted successfully", zap.Int("id", userID))
	ctx.JSON(http.StatusOK, gin.H{"message": "resource deleted successfully"})
}

func (c *CurrencyController) SearchPaginated(ctx *gin.Context) {
	c.Logger.Info("Searching currencies with pagination")

	// Parse query parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if pageSize < 1 {
		pageSize = 10
	}

	// Build filters
	filters := domain.DataFilters{
		Page:     page,
		PageSize: pageSize,
	}

	// Parse like filters
	likeFilters := make(map[string][]string)

	for field := range currency.ColumnsUserMapping {
		if values := ctx.QueryArray(field + "_like"); len(values) > 0 {
			c.Logger.Info("Using currency.ColumnsUserMapping", zap.Any("mapping", likeFilters))

			likeFilters[field] = values
		}
	}
	filters.LikeFilters = likeFilters

	// Parse exact matches
	matches := make(map[string][]string)
	for field := range user.ColumnsUserMapping {
		if values := ctx.QueryArray(field + "_match"); len(values) > 0 {
			matches[field] = values
		}
	}
	filters.Matches = matches

	// Parse date range filters
	var dateRanges []domain.DateRangeFilter
	for field := range user.ColumnsUserMapping {
		startStr := ctx.Query(field + "_start")
		endStr := ctx.Query(field + "_end")

		if startStr != "" || endStr != "" {
			dateRange := domain.DateRangeFilter{Field: field}

			if startStr != "" {
				if startTime, err := time.Parse(time.RFC3339, startStr); err == nil {
					dateRange.Start = &startTime
				}
			}

			if endStr != "" {
				if endTime, err := time.Parse(time.RFC3339, endStr); err == nil {
					dateRange.End = &endTime
				}
			}

			dateRanges = append(dateRanges, dateRange)
		}
	}
	filters.DateRangeFilters = dateRanges

	// Parse sorting
	sortBy := ctx.QueryArray("sortBy")
	if len(sortBy) > 0 {
		filters.SortBy = sortBy
	}

	sortDirection := domain.SortDirection(ctx.DefaultQuery("sortDirection", "asc"))
	if sortDirection.IsValid() {
		filters.SortDirection = sortDirection
	}

	result, err := c.currencyService.SearchPaginated(filters)
	if err != nil {
		c.Logger.Error("Error searching currencies", zap.Error(err))
		_ = ctx.Error(err)
		return
	}

	response := gin.H{
		"data":       arrayDomainToResponseMapper(result.Data),
		"total":      result.Total,
		"page":       result.Page,
		"pageSize":   result.PageSize,
		"totalPages": result.TotalPages,
		"filters":    filters,
	}

	c.Logger.Info("Successfully searched currencies",
		zap.Int64("total", result.Total),
		zap.Int("page", result.Page))
	ctx.JSON(http.StatusOK, response)
}

// Mappers
func domainToResponseMapper(domainUser *domainCurrency.Currency) *ResponseUser {
	return &ResponseUser{
		ID:        domainUser.ID,
		Name:      domainUser.Name,
		Code:      domainUser.Code,
		Rate:      domainUser.Rate,
		Status:    domainUser.Status,
		CreatedAt: domainUser.CreatedAt,
		UpdatedAt: domainUser.UpdatedAt,
	}
}

func arrayDomainToResponseMapper(users *[]domainCurrency.Currency) *[]ResponseUser {
	res := make([]ResponseUser, len(*users))
	for i, u := range *users {
		res[i] = *domainToResponseMapper(&u)
	}
	return &res
}

func toUsecaseMapper(req *NewCurrencyRequest) *domainCurrency.Currency {
	return &domainCurrency.Currency{
		Name:   req.Name,
		Code:   req.Code,
		Status: req.Status,
		Rate:   req.Rate,
	}
}
