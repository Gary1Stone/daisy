package cmd

import (
	"encoding/json"
	"html/template"
	"log"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/icons"

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
		"title":           template.HTML(icons.GetIcon("settings") + " Admin"),
		"fullName":        user.Fullname,
		"isAdmin":         user.IsAdmin,
		"cmd_one":         template.HTML(ctrls.MakeButton(ctrls.BtnTables, user.Permissions.Admin.Read)),
		"cmd_two":         template.HTML(ctrls.MakeButton(ctrls.BtnSave, user.Permissions.Admin.Update)),
		"cmd_three":       template.HTML(ctrls.MakeButton(ctrls.BtnHelp, true)),
		"deviceTypesJson": string(deviceTypesJson),
		"siteIcon":        template.HTML(icons.GetIcon("site")),
		"officeIcon":      template.HTML(icons.GetIcon("office")),
		"groupIcon":       template.HTML(icons.GetIcon("group")),
		"impactIcon":      template.HTML(icons.GetIcon("hammer")),
		"statusIcon":      template.HTML(icons.GetIcon("status")),
		"makeIcon":        template.HTML(icons.GetIcon("factory")),
		"coresIcon":       template.HTML(icons.GetIcon("cores")),
		"drivetypeIcon":   template.HTML(icons.GetIcon("harddisk")),
		"osIcon":          template.HTML(icons.GetIcon("os")),
		"geofenceIcon":    template.HTML(icons.GetIcon("location")),
		"troubleIcon":     template.HTML(icons.GetIcon("news")),
		"typeIcon":        template.HTML(icons.GetIcon("troubles")),
		"kindsIcon":       template.HTML(icons.GetIcon("kinds")),
	}))
}

func PostAdmin(c *fiber.Ctx) error {
	// Read incoming requst cookie to get curUid
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	if !user.Permissions.Admin.Read {
		return c.Status(fiber.StatusOK).SendString("ERROR: You do not have permission to view admin records.")
	}

	//Get the data sent from the browser
	adminForm := struct {
		Task      string `json:"task"`
		Field     string `json:"field"`
		AdminData string `json:"adminData"`
	}{}

	if err := c.BodyParser(&adminForm); err != nil {
		return c.Status(fiber.StatusOK).SendString("ERROR: " + err.Error())
	}

	//If the user wanted to save anything
	if adminForm.Task == "save_table" {
		var items []db.ChoiceAdmin
		err := json.Unmarshal([]byte(adminForm.AdminData), &items)
		if err != nil {
			log.Println(err)
		}

		//Checking the user permissions Create(add), Delete, Update
		for _, item := range items {
			if (item.Add && !item.Delete) && user.Permissions.Admin.Create {
				if item.AddRecord() != nil {
					return c.Status(fiber.StatusOK).SendString("ERROR: ")
				}
			} else if item.Delete && user.Permissions.Admin.Delete {
				if item.DeleteRecord() != nil {
					return c.Status(fiber.StatusOK).SendString("ERROR: ")
				}
			} else if item.Update && user.Permissions.Admin.Update {
				if item.UpdateRecord() != nil {
					return c.Status(fiber.StatusOK).SendString("ERROR: ")
				}
			}
		}
	}

	//Always send back the selected admin data from the choices table
	items := db.GetChoicesAdmin(adminForm.Field)
	return c.Status(fiber.StatusOK).JSON(items)
}
