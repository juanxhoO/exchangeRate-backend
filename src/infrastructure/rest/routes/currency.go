package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/currency"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/middlewares"
	"github.com/gin-gonic/gin"
)

func CurrencyRoutes(router *gin.RouterGroup, controller currency.ICurrencyController) {
	u := router.Group("/currency")
	{
		u.GET("/:id", controller.GetCurrenciesByID)
	}

	u.Use(middlewares.AuthJWTMiddleware())
	{
		u.GET("/", controller.GetAllCurrencies)
		u.DELETE("/:id", controller.DeleteCurrency)
		u.PUT("/rates", controller.UpdateExchanges)
	}
}
