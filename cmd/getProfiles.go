package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/svg"

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
		"title":         template.HTML(svg.GetIcon("profiles") + " Profiles"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"cmd_one":       template.HTML(ctrls.MakeAddButton(user.Permissions.Profile.Create)),
		"cmd_two":       template.HTML(ctrls.MakeSearchBtn()),
		"profilesTable": template.HTML(ctrls.ProfilesTable(user.Uid, filter)),
		"menu":          template.HTML(svg.GetIcon("menu")),
		"home":          template.HTML(svg.GetIcon("home")),
		"ticket":        template.HTML(svg.GetIcon("ticket")),
		"devices":       template.HTML(svg.GetIcon("devices")),
		"software":      template.HTML(svg.GetIcon("software")),
		"profiles":      template.HTML(svg.GetIcon("profiles")),
		"reports":       template.HTML(svg.GetIcon("reports")),
		"control":       template.HTML(svg.GetIcon("control")),
		"network":       template.HTML(svg.GetIcon("network")),
		"settings":      template.HTML(svg.GetIcon("settings")),
		"about":         template.HTML(svg.GetIcon("about")),
		"logout":        template.HTML(svg.GetIcon("logout")),
		"user":          template.HTML(svg.GetIcon("user")),
		"person_add":        template.HTML(svg.GetIcon("person_add")),
		"bell":          template.HTML(svg.GetIcon("bell")),
	})
}
