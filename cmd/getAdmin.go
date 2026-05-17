package cmd

import (
	"encoding/json"
	"html/template"
	"log"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"

	"github.com/gofiber/fiber/v2"
)

func GetAdmin(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Admin.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	actionCodes, err := db.GetActionCodes(true)
	if err != nil {
		log.Println(err)
	}

	deviceTypesJson, err := json.Marshal(actionCodes)
	if err != nil {
		log.Println(err)
	}

	//Render the page
	return c.Render("admin", addNavigationIcons(fiber.Map{
		"title":           template.HTML(svg.GetIcon("settings") + " Admin"),
		"fullName":        user.Fullname,
		"isAdmin":         user.IsAdmin,
		"cmd_one":         template.HTML(ctrls.MakeButton(ctrls.BtnTables, user.Permissions.Admin.Read)),
		"cmd_two":         template.HTML(ctrls.MakeButton(ctrls.BtnSave, user.Permissions.Admin.Update)),
		"cmd_three":       template.HTML(ctrls.MakeButton(ctrls.BtnHelp, true)),
		"deviceTypesJson": string(deviceTypesJson),
		"siteIcon":        template.HTML(svg.GetIcon("site")),
		"officeIcon":      template.HTML(svg.GetIcon("office")),
		"groupIcon":       template.HTML(svg.GetIcon("group")),
		"impactIcon":      template.HTML(svg.GetIcon("hammer")),
		"statusIcon":      template.HTML(svg.GetIcon("status")),
		"makeIcon":        template.HTML(svg.GetIcon("factory")),
		"coresIcon":       template.HTML(svg.GetIcon("cores")),
		"drivetypeIcon":   template.HTML(svg.GetIcon("harddisk")),
		"osIcon":          template.HTML(svg.GetIcon("os")),
		"geofenceIcon":    template.HTML(svg.GetIcon("location")),
		"troubleIcon":     template.HTML(svg.GetIcon("news")),
		"typeIcon":        template.HTML(svg.GetIcon("troubles")),
		"kindsIcon":       template.HTML(svg.GetIcon("kinds")),
	}))
}
