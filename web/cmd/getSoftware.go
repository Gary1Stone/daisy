package cmd

import (
	"html/template"
	"log"
	"math"
	"strconv"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetSoftware(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	sid, err2 := strconv.Atoi(c.Query("sid", "0"))
	if err2 != nil || sid == 0 {
		sid = math.MaxInt
	}
	if sid == 0 {
		sid = math.MaxInt //Prevent getting all the records (sid=0 means get all records)
	}

	// If no read capababilty, send them home
	if !user.Permissions.Profile.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	isReadonly := !user.Permissions.Software.Update
	isDisabled := !user.Permissions.Software.Update

	// Get the software record
	software, err := db.GetSoftware(user.Uid, sid)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}
	//Last Updated by name
	lun := ""
	if len(software.Last_updated_time) > 0 {
		lun = "<p title='Last Updated'>"
		lun += software.Fullname
		lun += " at "
		lun += software.Last_updated_time
		lun += "</p>"
	}
	//Edit inventory
	edit := "Software Title"
	if user.IsAdmin {
		edit = "<a href='#' onclick='popDialog();'>Software Title</a>"
	}

	return c.Render("software", fiber.Map{
		"title":          template.HTML("<span class='mif-apps icon'></span>Software"),
		"fullName":       user.Fullname,
		"isAdmin":        user.IsAdmin,
		"sid":            software.Sid,
		"name":           software.Name,
		"licenses":       software.Licenses,
		"active":         software.Active,
		"isReadonly":     isReadonly,
		"isDisabled":     isDisabled,
		"reuseable":      software.Reuseable,
		"license_key":    software.License_key,
		"product":        software.Product,
		"source":         software.Source,
		"link":           software.Link,
		"inv_name":       software.Inv_name,
		"pre_installed":  software.Pre_installed,
		"free":           software.Free,
		"edit":           template.HTML(edit),
		"notes":          software.Notes,
		"lastupdated":    template.HTML(lun),
		"cmd_one":        template.HTML(ctrls.MakeSaveButton(user.Permissions.Software.Update)),
		"cmd_two":        template.HTML(ctrls.MakeAddButton(user.Permissions.Software.Create)),
		"cmd_three":      template.HTML(ctrls.MakeDeleteButton(user.Permissions.Software.Delete)),
		"actionlog":      template.HTML(ctrls.BuildSoftwareLog(user.Uid, sid, 0)),
		"purchased":      software.Purchased,
		"installed_list": template.HTML(ctrls.BuildInstalledList(user.Uid, sid)),
	})
}
