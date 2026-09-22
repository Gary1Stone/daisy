package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/svg"

	"github.com/gofiber/fiber/v2"
)

func GetComstat(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If no read capababilty, send them home
	if !user.Permissions.Device.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	//Render the page
	return c.Render("comstat", addNavigationIcons(fiber.Map{
		"title":          template.HTML(svg.GetIcon("devices") + " Computer Status"),
		"fullName":       user.Fullname,
		"isAdmin":        user.IsAdmin,
		"cmd_one":        template.HTML(ctrls.MakeButton(ctrls.BtnNew, user.Permissions.Device.Create)),
		"comstatTable": template.HTML(ctrls.ComstatTable(user.Uid)),
	}))
}
