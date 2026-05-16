package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/svg"
	"github.com/gofiber/fiber/v2"
)

func GetReports(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	//Render the page
	return c.Render("reports", addNavigationIcons(fiber.Map{
		"title":    template.HTML(svg.GetIcon("reports") + " Reports"),
		"fullName": user.Fullname,
		"isAdmin":  user.IsAdmin,
	}))
}
