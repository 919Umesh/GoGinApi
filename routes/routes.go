package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/umesh/ginapi/controllers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		users := api.Group("/users")
		{
			users.GET("/", controllers.GetUsers)
			users.POST("/", controllers.CreateUser)
			users.PATCH("/:id", controllers.UpdateUser)
		}
	}

	return r
}
