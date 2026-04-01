package cmd

import (
	"html/template"
	"log"
	"strconv"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetDevices(c *fiber.Ctx) error {

	// Read incoming requst cookie to get curUid
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Device.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	// Read the database to get the user's previous filter settings
	filter, err := db.GetDeviceFilter(user.Uid)
	if err != nil {
		log.Println(err)
	}
	filter.Page = 0 // Reset the page to 0

	return c.Render("devices", fiber.Map{
		"title":            template.HTML("<span class='mif-devices icon'></span>&nbsp;Devices"),
		"fullName":         user.Fullname,
		"isAdmin":          user.IsAdmin,
		"cmd_one":          template.HTML(ctrls.MakeAddButton(user.Permissions.Device.Create)),
		"cmd_two":          template.HTML(ctrls.MakeSearchBtn()),
		"cmd_three":        template.HTML(ctrls.MakeSeeButton()),
		"typeCtrl":         template.HTML(ctrls.BuildDropList("TYPESEARCH", filter.DevType, "", true, false)),
		"siteSearchCtrl":   template.HTML(ctrls.BuildDropList("SITESEARCH", filter.Site, "", true, false)),
		"officeSearchCtrl": template.HTML(ctrls.BuildDropList("OFFICESEARCH", filter.Office, filter.Site, true, false)),
		"groupSearchCtrl":  template.HTML(ctrls.BuildDropList("GROUPSEARCH", strconv.Itoa(filter.Gid), "", true, false)),
		"userSearchCtrl":   template.HTML(ctrls.BuildDropList("USERSEARCH", strconv.Itoa(filter.Uid), strconv.Itoa(filter.Gid), true, false)),
		"cards":            template.HTML(ctrls.DeviceCards(user.Uid, &filter, false)),
	})
}
