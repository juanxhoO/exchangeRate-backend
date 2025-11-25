package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/currency"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/middlewares"
	"github.com/gin-gonic/gin"
)

func CurrencyRoutes(router *gin.RouterGroup, controller currency.ICurrencyController) {
	u := router.Group("/currency")
	{
		u.POST("/", controller.NewCurrency)

		u.GET("/:id", controller.GetCurrenciesByID)
	}

	u.Use(middlewares.AuthJWTMiddleware())
	{
		u.GET("/", controller.GetAllCurrencies)
		u.PUT("/:id", controller.UpdateCurency)
		u.DELETE("/:id", controller.DeleteCurrency)
	}
}
