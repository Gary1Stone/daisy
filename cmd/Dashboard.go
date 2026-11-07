package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/comstat"
	"github.com/gbsto/daisy/dashboard"
	"github.com/gbsto/daisy/icons"
	"github.com/gofiber/fiber/v2"
)

func GetDashboard(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	return c.Render("dashboard", addNavigationIcons(fiber.Map{
		"title":          template.HTML(icons.GetIcon("about") + " Dashboard"),
		"fullName":       user.Fullname,
		"isAdmin":        user.IsAdmin,
		"cmd_one":        "",
		"cmd_two":        "",
		"cmd_three":      "",
		"kpis":           template.HTML(dashboard.BuildKPIs()),
		"computerStatus": template.HTML(comstat.ComstatBoxes(user.Uid, user.Tzoff)),
	}))
}
