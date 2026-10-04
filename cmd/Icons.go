package cmd

import (
	"strconv"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
	"github.com/gofiber/fiber/v2"
)

func PostIcon(c *fiber.Ctx) error {
	if _, err := extractUserInfo(c); err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	var request struct {
		Ctrl     string `json:"ctrl"`
		Selected string `json:"selected"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Error: Malformed request")
	}

	iconName := ""
	switch request.Ctrl {
	case "parent":
		// Get the cid for the parent, look up the device kind, get the icon name from icon table, get the iconSVG from cache
		// First convert "selected" to an integer
		cid, err := strconv.Atoi(request.Selected)
		if err != nil {
			iconName = ""
		} else {
			iconName = db.GetKindUsingParentCid(cid)
		}
	case "kind":
		iconName = request.Selected
	default:

	}

	return c.Status(fiber.StatusOK).SendString(svg.GetIcon(iconName))
}
