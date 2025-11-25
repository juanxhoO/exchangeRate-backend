package currency

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	domainCurrency "github.com/gbrayhan/microservices-go/src/domain/currency"
	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
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
	NewCurrency(ctx *gin.Context)
	GetAllCurrencies(ctx *gin.Context)
	GetCurrenciesByID(ctx *gin.Context)
	UpdateCurency(ctx *gin.Context)
	DeleteCurrency(ctx *gin.Context)
}

type CurrencyController struct {
	currencyService domainCurrency.ICurrencyService
	Logger          *logger.Logger
}

func NewCurrencyController(currencyService domainCurrency.ICurrencyService, loggerInstance *logger.Logger) ICurrencyController {
	return &CurrencyController{currencyService: currencyService, Logger: loggerInstance}
}

func (c *CurrencyController) NewCurrency(ctx *gin.Context) {
	c.Logger.Info("Creating new Currency")
	var request NewCurrencyRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Error binding JSON for new user", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	userModel, err := c.currencyService.Create(toUsecaseMapper(&request))
	if err != nil {
		c.Logger.Error("Error creating user", zap.Error(err), zap.String("code", request.Code))
		_ = ctx.Error(err)
		return
	}
	userResponse := domainToResponseMapper(userModel)
	c.Logger.Info("User created successfully", zap.String("code", request.Code), zap.Int("id", userModel.ID))
	ctx.JSON(http.StatusOK, userResponse)
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

func (c *CurrencyController) UpdateCurency(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Error("Invalid user ID parameter for update", zap.Error(err), zap.String("id", ctx.Param("id")))
		appError := domainErrors.NewAppError(errors.New("param id is necessary"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Updating user", zap.Int("id", userID))
	var requestMap map[string]any
	err = controllers.BindJSONMap(ctx, &requestMap)
	if err != nil {
		c.Logger.Error("Error binding JSON for user update", zap.Error(err), zap.Int("id", userID))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	err = updateValidation(requestMap)
	if err != nil {
		c.Logger.Error("Validation error for user update", zap.Error(err), zap.Int("id", userID))
		_ = ctx.Error(err)
		return
	}
	userUpdated, err := c.currencyService.Update(userID, requestMap)
	if err != nil {
		c.Logger.Error("Error updating user", zap.Error(err), zap.Int("id", userID))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("User updated successfully", zap.Int("id", userID))
	ctx.JSON(http.StatusOK, domainToResponseMapper(userUpdated))
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
