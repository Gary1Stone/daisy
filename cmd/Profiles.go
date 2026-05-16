package cmd

import (
	"html/template"
	"log"

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

	return c.Render("profiles", addNavigationIcons(fiber.Map{
		"title":         template.HTML(svg.GetIcon("profiles") + " Profiles"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"cmd_one":       template.HTML(ctrls.MakeButton(ctrls.BtnNew, user.Permissions.Profile.Create)),
		"cmd_two":       template.HTML(ctrls.MakeButton(ctrls.BtnSearch, user.Permissions.Profile.Read)),
		"profilesTable": template.HTML(ctrls.ProfilesTable(user.Uid, filter)),
		"bellIcon":      template.HTML(svg.GetIcon("bell")),
	}))
}

func PostProfiles(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	var filter db.ProfileFilter
	if err := c.BodyParser(&filter); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}

	return c.Status(fiber.StatusOK).SendString(ctrls.ProfilesTable(user.Uid, filter))
}
