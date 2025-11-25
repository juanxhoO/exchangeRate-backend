package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/exchanger"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/middlewares"
	"github.com/gin-gonic/gin"
)

func ExchangerRoutes(router *gin.RouterGroup, controller exchanger.IExchangerController) {
	u := router.Group("/exchanger")
	u.Use(middlewares.AuthJWTMiddleware())
	{
		u.GET("/:id", controller.GetExchangersById)
		u.POST("/", controller.NewExchanger)
		u.GET("/", controller.GetAllExchangers)
		u.PUT("/:id", controller.UpdateExchanger)
		u.DELETE("/:id", controller.DeleteExchanger)
	}
}
