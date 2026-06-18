package cmd

import (
	"github.com/gbsto/daisy/svg"
	"github.com/gofiber/fiber/v2"
)

func PostIcon(c *fiber.Ctx) error {
	// Standardize authentication check for consistency across endpoints
	if _, err := extractUserInfo(c); err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	reply := struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
	}{
		Success: false,
		Msg:     "Invalid Request",
	}

	var input struct {
		Icon string `json:"icon"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(reply)
	}

	svgIcon := svg.GetIcon(input.Icon)
	if svgIcon == "" {
		reply.Msg = "Icon not found"
		return c.Status(fiber.StatusOK).JSON(reply)
	}

	reply.Success = true
	reply.Msg = svgIcon
	return c.Status(fiber.StatusOK).JSON(reply)
}
