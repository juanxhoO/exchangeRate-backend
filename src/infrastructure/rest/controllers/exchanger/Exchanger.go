package exchanger

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	domainErrors "github.com/gbrayhan/microservices-go/src/domain/errors"
	domainExchanger "github.com/gbrayhan/microservices-go/src/domain/exchanger"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Structures
type NewExchangerRequest struct {
	Name     string `json:"name" binding:"required"`
	Url      string `json:"url" binding:"required"`
	ApiKey   string `json:"apiKey" binding:"required"`
	IsActive bool   `json:"isActive"`
}

type ResponseUser struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"isActive"`
	Url       string    `json:"url"`
	ApiKey    string    `json:"apiKey"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type IExchangerController interface {
	NewExchanger(ctx *gin.Context)
	GetAllExchangers(ctx *gin.Context)
	GetExchangersById(ctx *gin.Context)
	UpdateExchanger(ctx *gin.Context)
	DeleteExchanger(ctx *gin.Context)
}

type ExchangerController struct {
	exchangerService domainExchanger.IExchangerService
	Logger           *logger.Logger
}

func NewExchangerController(exchangerService domainExchanger.IExchangerService, loggerInstance *logger.Logger) IExchangerController {
	return &ExchangerController{exchangerService: exchangerService, Logger: loggerInstance}
}

func (c *ExchangerController) NewExchanger(ctx *gin.Context) {
	c.Logger.Info("Creating new Exchanger")
	var request NewExchangerRequest
	if err := controllers.BindJSON(ctx, &request); err != nil {
		c.Logger.Error("Error binding JSON for new exchanger", zap.Error(err))
		appError := domainErrors.NewAppError(err, domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	userModel, err := c.exchangerService.Create(toUsecaseMapper(&request))
	if err != nil {
		c.Logger.Error("Error creating exchanger", zap.Error(err), zap.String("name", request.Name))
		_ = ctx.Error(err)
		return
	}
	userResponse := domainToResponseMapper(userModel)
	c.Logger.Info("Exchanger created successfully", zap.String("name", request.Name), zap.Int("id", userModel.ID))
	ctx.JSON(http.StatusOK, userResponse)
}

func (c *ExchangerController) GetAllExchangers(ctx *gin.Context) {
	c.Logger.Info("Getting all users")
	users, err := c.exchangerService.GetAll()
	if err != nil {
		c.Logger.Error("Error getting all users", zap.Error(err))
		appError := domainErrors.NewAppErrorWithType(domainErrors.UnknownError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Successfully retrieved all users", zap.Int("count", len(*users)))
	ctx.JSON(http.StatusOK, arrayDomainToResponseMapper(users))
}

func (c *ExchangerController) GetExchangersById(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Error("Invalid user ID parameter", zap.Error(err), zap.String("id", ctx.Param("id")))
		appError := domainErrors.NewAppError(errors.New("user id is invalid"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Getting user by ID", zap.Int("id", userID))
	user, err := c.exchangerService.GetByID(userID)
	if err != nil {
		c.Logger.Error("Error getting user by ID", zap.Error(err), zap.Int("id", userID))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("Successfully retrieved user by ID", zap.Int("id", userID))
	ctx.JSON(http.StatusOK, domainToResponseMapper(user))
}

func (c *ExchangerController) UpdateExchanger(ctx *gin.Context) {

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

	userUpdated, err := c.exchangerService.Update(userID, requestMap)
	if err != nil {
		c.Logger.Error("Error updating user", zap.Error(err), zap.Int("id", userID))
		_ = ctx.Error(err)
		return
	}

	c.Logger.Info("User updated successfully", zap.Int("id", userID))
	ctx.JSON(http.StatusOK, domainToResponseMapper(userUpdated))
}

func (c *ExchangerController) DeleteExchanger(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		c.Logger.Error("Invalid user ID parameter for deletion", zap.Error(err), zap.String("id", ctx.Param("id")))
		appError := domainErrors.NewAppError(errors.New("param id is necessary"), domainErrors.ValidationError)
		_ = ctx.Error(appError)
		return
	}
	c.Logger.Info("Deleting user", zap.Int("id", userID))
	err = c.exchangerService.Delete(userID)
	if err != nil {
		c.Logger.Error("Error deleting user", zap.Error(err), zap.Int("id", userID))
		_ = ctx.Error(err)
		return
	}
	c.Logger.Info("User deleted successfully", zap.Int("id", userID))
	ctx.JSON(http.StatusOK, gin.H{"message": "resource deleted successfully"})
}

// Mappers
func domainToResponseMapper(domainExchanger *domainExchanger.Exchanger) *ResponseUser {
	return &ResponseUser{
		ID:        domainExchanger.ID,
		Name:      domainExchanger.Name,
		Url:       domainExchanger.Url,
		IsActive:  domainExchanger.IsActive,
		ApiKey:    domainExchanger.ApiKey,
		CreatedAt: domainExchanger.CreatedAt,
		UpdatedAt: domainExchanger.UpdatedAt,
	}
}

func arrayDomainToResponseMapper(users *[]domainExchanger.Exchanger) *[]ResponseUser {
	res := make([]ResponseUser, len(*users))
	for i, u := range *users {
		res[i] = *domainToResponseMapper(&u)
	}
	return &res
}

func toUsecaseMapper(req *NewExchangerRequest) *domainExchanger.Exchanger {
	return &domainExchanger.Exchanger{
		Name:     req.Name,
		Url:      req.Url,
		IsActive: req.IsActive,
		ApiKey:   req.ApiKey,
	}
}
