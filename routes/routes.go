package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/umesh/ginapi/controllers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Static("/uploads", "./uploads")

	//Public route to generate and authenticate the token
	public := r.Group("/api")
	{
		public.POST("/login", controllers.Login)
		public.POST("/users", controllers.CreateUser)
	}

	//Protected Route only accessible with the jwt token
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
			products.GET("/search", controllers.SearchProducts)
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
		orders := protected.Group("/orders")
		{
			orders.POST("/", controllers.CreateOrder)
			orders.GET("/", controllers.GetUserOrders)
			orders.GET("/:id", controllers.GetOrderByID)
		}
	}

	return r
}
