package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/umesh/ginapi/controllers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Static("/uploads", "./uploads")

	public := r.Group("/api")
	{
		public.POST("/login", controllers.Login)
		public.POST("/users", controllers.CreateUser)
	}

	protected := r.Group("/api")
	protected.Use(controllers.AuthMiddleware())
	{
		users := protected.Group("/users")
		{
			users.GET("/", controllers.GetUsers)
			users.GET("/:id", controllers.GetUsersByID)
			users.PATCH("/:id", controllers.UpdateUser)
		}

		products := protected.Group("/products")
		{
			products.GET("/", controllers.GetProducts)
			products.GET("/:id", controllers.GetProductByID)
			products.POST("/", controllers.CreateProduct)
			products.PATCH("/:id", controllers.UpdateProduct)
			products.DELETE("/:id", controllers.DeleteProduct)
		}

		venues := protected.Group("/venues")
		{
			venues.GET("/", controllers.GetVenues)
			venues.POST("/", controllers.CreateVenue)
		}
	}

	return r
}
