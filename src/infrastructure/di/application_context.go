package di

import (
	"sync"

	authUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/auth"
	currencyUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/currency"
	emailUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/email"
	exchangerUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/exchanger"
	userUseCase "github.com/gbrayhan/microservices-go/src/application/usecases/user"
	logger "github.com/gbrayhan/microservices-go/src/infrastructure/logger"
	infraMailer "github.com/gbrayhan/microservices-go/src/infrastructure/mailer"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/currency"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/exchanger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/token"
	"github.com/gbrayhan/microservices-go/src/infrastructure/repository/psql/user"
	authController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/auth"
	currencyController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/currency"
	exchangerController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/exchanger"
	userController "github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/user"
	"github.com/gbrayhan/microservices-go/src/infrastructure/security"
	"gorm.io/gorm"
)

// ApplicationContext holds all application dependencies and services
type ApplicationContext struct {
	DB                  *gorm.DB
	Logger              *logger.Logger
	AuthController      authController.IAuthController
	UserController      userController.IUserController
	CurrencyController  currencyController.ICurrencyController
	ExchangerController exchangerController.IExchangerController
	JWTService          security.IJWTService
	UserRepository      user.UserRepositoryInterface
	TokenRepository     token.TokenRepositoryInterface
	AuthUseCase         authUseCase.IAuthUseCase
	UserUseCase         userUseCase.IUserUseCase
	EmailUseCase        emailUseCase.IEmailUseCase
	CurrencyUseCase     currencyUseCase.ICurrencyUseCase
}

var (
	loggerInstance *logger.Logger
	loggerOnce     sync.Once
)

func GetLogger() *logger.Logger {
	loggerOnce.Do(func() {
		loggerInstance, _ = logger.NewLogger()
	})
	return loggerInstance
}

// SetupDependencies creates a new application context with all dependencies
func SetupDependencies(loggerInstance *logger.Logger) (*ApplicationContext, error) {
	// Initialize database with logger
	db, err := psql.InitPSQLDB(loggerInstance)
	if err != nil {
		return nil, err
	}

	// Initialize mailer with logger
	goMailer := infraMailer.NewGoMailer()
	emailUC := emailUseCase.NewEmailUseCase(goMailer, loggerInstance)

	// Initialize JWT service (manages its own configuration)
	jwtService := security.NewJWTService()
	apiService := security.NewAPIService()

	// Initialize repositories with logger
	userRepo := user.NewUserRepository(db, loggerInstance)
	currencyRepo := currency.NewCurrencyRepository(db, loggerInstance)
	exchangerRepo := exchanger.NewExchangerRepository(db, loggerInstance)
	tokenRepo := token.NewTokenRepository(db, loggerInstance)

	// Initialize use cases with logger
	authUC := authUseCase.NewAuthUseCase(userRepo, tokenRepo, jwtService, loggerInstance, goMailer)
	userUC := userUseCase.NewUserUseCase(userRepo, loggerInstance)
	exchangerUC := exchangerUseCase.NewExchangerUseCase(exchangerRepo, apiService, loggerInstance)
	currencyUC := currencyUseCase.NewCurrencyUseCase(currencyRepo, exchangerRepo, apiService, loggerInstance)

	// Initialize controllers with logger
	authController := authController.NewAuthController(authUC, loggerInstance)
	userController := userController.NewUserController(userUC, loggerInstance)
	currencyController := currencyController.NewCurrencyController(currencyUC, loggerInstance)
	exchangerController := exchangerController.NewExchangerController(exchangerUC, loggerInstance)

	return &ApplicationContext{
		DB:                  db,
		Logger:              loggerInstance,
		AuthController:      authController,
		UserController:      userController,
		CurrencyController:  currencyController,
		ExchangerController: exchangerController,
		JWTService:          jwtService,
		UserRepository:      userRepo,
		TokenRepository:     tokenRepo,
		AuthUseCase:         authUC,
		EmailUseCase:        emailUC,
		UserUseCase:         userUC,
		CurrencyUseCase:     currencyUC,
	}, nil
}

// NewTestApplicationContext creates an application context for testing with mocked dependencies
func NewTestApplicationContext(
	mockUserRepo user.UserRepositoryInterface,
	mockTokenRepo token.TokenRepositoryInterface,
	mockJWTService security.IJWTService,
	loggerInstance *logger.Logger,
) *ApplicationContext {
	// Initialize use cases with mocked repositories and logger
	authUC := authUseCase.NewAuthUseCase(mockUserRepo, mockTokenRepo, mockJWTService, loggerInstance, infraMailer.NewGoMailer())
	userUC := userUseCase.NewUserUseCase(mockUserRepo, loggerInstance)

	// Initialize controllers with logger
	authController := authController.NewAuthController(authUC, loggerInstance)
	userController := userController.NewUserController(userUC, loggerInstance)

	return &ApplicationContext{
		Logger:          loggerInstance,
		AuthController:  authController,
		UserController:  userController,
		JWTService:      mockJWTService,
		UserRepository:  mockUserRepo,
		TokenRepository: mockTokenRepo,
		AuthUseCase:     authUC,
		UserUseCase:     userUC,
	}
}
