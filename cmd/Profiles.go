package cmd

import (
	"html/template"
	"log"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/profiles"

	"github.com/gofiber/fiber/v2"
)

// Build list of user profiles
func GetProfiles(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	var filter profiles.Profile
	filter.Init()
	filter.Uid = 0 // get all profiles

	// If no read capababilty, send them home
	if !user.Permissions.Profile.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	return c.Render("profiles", addNavigationIcons(fiber.Map{
		"title":         template.HTML(icons.GetIcon("profiles") + " Profiles"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"cmd_one":       template.HTML(ctrls.MakeButton(ctrls.BtnNew, user.Permissions.Profile.Create)),
		"profilesTable": template.HTML(ctrls.ProfilesTable(user.Uid, filter)),
		"bellIcon":      template.HTML(icons.GetIcon("bell")),
	}))
}

func PostProfiles(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	var filter profiles.Profile
	if err := c.BodyParser(&filter); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}

	return c.Status(fiber.StatusOK).SendString(ctrls.ProfilesTable(user.Uid, filter))
}
