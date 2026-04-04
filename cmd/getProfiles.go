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
		"title":         template.HTML(svg.GetIcon("user.svg") + "&nbsp;Profiles"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"cmd_one":       template.HTML(ctrls.MakeAddButton(user.Permissions.Profile.Create)),
		"cmd_two":       template.HTML(ctrls.MakeSearchBtn()),
		"profilesTable": template.HTML(ctrls.ProfilesTable(user.Uid, filter)),
		"userIcon":      template.HTML(svg.GetIcon("user.svg")),
		"homeIcon":      template.HTML(svg.GetIcon("home.svg")),
		"ticketsIcon":   template.HTML(svg.GetIcon("ticket.svg")),
		"devicesIcon":   template.HTML(svg.GetIcon("devices-pc.svg")),
		"softwaresIcon": template.HTML(svg.GetIcon("binary.svg")),
		"profilesIcon":  template.HTML(svg.GetIcon("id.svg")),
		"reportsIcon":   template.HTML(svg.GetIcon("report.svg")),
		"controlIcon":   template.HTML(svg.GetIcon("steering-wheel.svg")),
		"networkIcon":   template.HTML(svg.GetIcon("tournament.svg")),
		"adminIcon":     template.HTML(svg.GetIcon("settings.svg")),
		"aboutIcon":     template.HTML(svg.GetIcon("info-hexagon.svg")),
		"exitIcon":      template.HTML(svg.GetIcon("door-exit.svg")),
		"wizardIcon":    template.HTML(svg.GetIcon("wand.svg")),
		"alertIcon":     template.HTML(svg.GetIcon("bell.svg")),
	})
}
