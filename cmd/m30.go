package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/m30"
	"github.com/gofiber/fiber/v2"
)

func GetM30(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	return c.Render("m30", addNavigationIcons(fiber.Map{
		"title":     template.HTML(icons.GetIcon("router") + " M30 Status"),
		"fullName":  user.Fullname,
		"isAdmin":   user.IsAdmin,
		"cmd_one":   "",
		"cmd_two":   "",
		"cmd_three": "",
		"m30status": template.HTML(m30.BuildM30Report(user.Tzoff)),
	}))
}
