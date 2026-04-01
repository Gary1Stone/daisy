package cmd

import (
	"log"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func PostProfiles(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	var filter db.ProfileFilter
	if err := c.BodyParser(&filter); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}

	return c.Status(fiber.StatusOK).SendString(ctrls.ProfilesTable(user.Uid, filter))
}
