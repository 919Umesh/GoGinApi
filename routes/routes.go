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
			users.GET("/:id", controllers.GetUsersByID)
			users.POST("/", controllers.CreateUser)
			users.PATCH("/:id", controllers.UpdateUser)
			users.DELETE("/:id", controllers.DeleteUser)
		}
		products := api.Group("/products")
		{
			products.GET("/", controllers.GetProducts)
			products.GET("/:id", controllers.GetProductByID)
			products.POST("/", controllers.CreateProduct)
			products.PATCH("/:id", controllers.UpdateProduct)
			products.DELETE("/:id", controllers.DeleteProduct)
		}
		venues := api.Group("/venues")
		{
			venues.GET("/", controllers.GetVenues)
			venues.POST("/", controllers.CreateVenue)
		}
	}

	return r
}
