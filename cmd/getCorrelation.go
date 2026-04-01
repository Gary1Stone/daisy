package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetCorrelation(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Admin.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	filter := db.MacCorrelationFilter{
		Jaccard:   70,    // Jaccard number x100
		Pearson:   85,    // Pearsons number x100
		Fixed:     true,  // false=dont care, true=Fixed MACs only
		Random:    false, // false=dont care, true=Random MACs only
		Hostnames: false, // false=Ignore Hostnames, true=Hostnames must match
		Jsign:     true,  // false=less than, true=greater than
		Psign:     true,  // false=less than, true=greater than
	}

	return c.Render("correlation", fiber.Map{
		"title":        template.HTML("<span class='mif-profile icon'></span>&nbsp;Device"),
		"fullName":     user.Fullname,
		"isAdmin":      user.IsAdmin,
		"isReadonly":   user.Permissions.Admin.Update,
		"isDisabled":   user.Permissions.Admin.Update,
		"cmd_one":      template.HTML(ctrls.MakeAdminHelpButton()),
		"cmd_two":      template.HTML(ctrls.MakeSearchBtn()),
		"affinityList": template.HTML(ctrls.BuildMacCorrelationTable(filter)),
	})
}

func PostCorrelation(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}
	if !user.Permissions.Admin.Update {
		return c.Status(fiber.StatusUnauthorized).Redirect("home.html")
	}
	var filter db.MacCorrelationFilter
	if err := c.BodyParser(&filter); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
	}
	return c.Status(fiber.StatusOK).SendString(ctrls.BuildMacCorrelationTable(filter))
}
