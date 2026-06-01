package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/svg"
	"github.com/gofiber/fiber/v2"
)

func GetAbout(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	return c.Render("about", addNavigationIcons(fiber.Map{
		"title":     template.HTML(svg.GetIcon("about") + " About"),
		"fullName":  user.Fullname,
		"isAdmin":   user.IsAdmin,
		"cmd_one":   "",
		"cmd_two":   "",
		"cmd_three": "",
	}))
}

func GetBanned(c *fiber.Ctx) error {
	return c.Render("banned", fiber.Map{
		"warning": "blocked",
	})
}

func GetCaptcha(c *fiber.Ctx) error {
	return c.Render("captcha", fiber.Map{
		"user": " ",
	})
}
