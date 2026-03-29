package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

// Build list of user profiles
func GetProfiles(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	var filter db.ProfileFilter
	filter.Init()
	filter.Uid = 0 // get all profiles

	// If no read capababilty, send them home
	if !user.Permissions.Profile.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	return c.Render("profiles", fiber.Map{
		"title":         template.HTML("<span class='mif-profile icon'></span>&nbsp;Profiles"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"cmd_one":       template.HTML(ctrls.MakeAddButton(user.Permissions.Profile.Create)),
		"profilesTable": template.HTML(ctrls.ProfilesTable(user.Uid, filter)),
	})
}
