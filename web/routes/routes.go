package routes

import (
	"github.com/gbsto/daisy/web/cmd"
	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App) {
	app.Get("/", cmd.GetIndex)
}
