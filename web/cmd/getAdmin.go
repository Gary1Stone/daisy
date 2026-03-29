package cmd

import (
	"encoding/json"
	"html/template"
	"log"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetAdmin(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
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
	return c.Render("admin", fiber.Map{
		"title":           template.HTML("<span class='mif-tools icon'></span>&nbsp;Admin"),
		"fullName":        user.Fullname,
		"isAdmin":         user.IsAdmin,
		"cmd_one":         template.HTML(ctrls.MakeAdminSelectButton(user.Permissions.Admin.Read)),
		"cmd_two":         template.HTML(ctrls.MakeAdminSaveButton(user.Permissions.Admin.Update)),
		"cmd_three":       template.HTML(ctrls.MakeAdminHelpButton()),
		"deviceTypesJson": string(deviceTypesJson),
	})
}
