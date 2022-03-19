package routes

import (
	"github.com/gofiber/fiber/v2"
	"ambassador/src/controllers"
	"ambassador/src/middlewares"
)

func Setup(app *fiber.App) {
	api := app.Group("api")
	admin := api.Group("admin")

	admin.Post("/register", controllers.Register)
	admin.Post("/login", controllers.Login)
	
	// Auth middleware
	adminAuthenticated := admin.Use(middlewares.IsAuthenticated)
	adminAuthenticated.Get("/user", controllers.User)
	adminAuthenticated.Post("/logout", controllers.Logout)
	adminAuthenticated.Put("/users/info", controllers.UpdateInfo)
	adminAuthenticated.Put("/users/password", controllers.UpdatePassword)
	adminAuthenticated.Get("/ambassadors", controllers.Ambassadors)
}