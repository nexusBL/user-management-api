package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nexusBL/user-management-api/internal/handler"
)

func Register(app *fiber.App, userHandler *handler.UserHandler) {
	app.Post("/users", userHandler.CreateUser)
	app.Get("/users/:id", userHandler.GetUserByID)
	app.Put("/users/:id", userHandler.UpdateUser)
	app.Delete("/users/:id", userHandler.DeleteUser)
	app.Get("/users", userHandler.ListUsers)
}
