package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetSoftwares(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If no read capababilty, send them home
	if !user.Permissions.Profile.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	var filter db.SoftwareFilter
	filter.Init()
	//Render the page
	return c.Render("softwares", fiber.Map{
		"title":          "Software",
		"fullName":       user.Fullname,
		"isAdmin":        user.IsAdmin,
		"cmd_one":        template.HTML(ctrls.MakeAddButton(user.Permissions.Software.Create)),
		"softwaresTable": template.HTML(ctrls.SoftwaresTable(user.Uid, filter)),
	})
}
