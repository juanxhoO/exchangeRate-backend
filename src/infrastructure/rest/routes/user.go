package routes

import (
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/controllers/user"
	"github.com/gbrayhan/microservices-go/src/infrastructure/rest/middlewares"
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.RouterGroup, controller user.IUserController) {
	u := router.Group("/user")
	{
		u.POST("/", controller.NewUser)

		u.GET("/:id", controller.GetUsersByID)
	}

	u.Use(middlewares.AuthJWTMiddleware())
	{
		u.GET("/", controller.GetAllUsers)
		u.PATCH("/:id", controller.UpdateUser)
		u.DELETE("/:id", controller.DeleteUser)
		u.GET("/search", controller.SearchPaginated)
		u.GET("/search-property", controller.SearchByProperty)
	}
}
