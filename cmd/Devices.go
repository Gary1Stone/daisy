package cmd

import (
	"html/template"
	"log"
	"strconv"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/svg"

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

	return c.Render("devices", addNavigationIcons(fiber.Map{
		"title":            template.HTML(svg.GetIcon("devices") + " Devices"),
		"fullName":         user.Fullname,
		"isAdmin":          user.IsAdmin,
		"cmd_one":          template.HTML(ctrls.MakeButton(ctrls.BtnNew, user.Permissions.Device.Create)),
		"cmd_two":          template.HTML(ctrls.MakeButton(ctrls.BtnFilter, user.Permissions.Device.Read)),
		"cmd_three":        template.HTML(ctrls.MakeButton(ctrls.BtnSeen, user.Permissions.Device.Read)),
		"typeCtrl":         template.HTML(ctrls.BuildDropList("TYPESEARCH", filter.DevType, "", true, false)),
		"siteSearchCtrl":   template.HTML(ctrls.BuildDropList("SITESEARCH", filter.Site, "", true, false)),
		"officeSearchCtrl": template.HTML(ctrls.BuildDropList("OFFICESEARCH", filter.Office, filter.Site, true, false)),
		"groupSearchCtrl":  template.HTML(ctrls.BuildDropList("GROUPSEARCH", strconv.Itoa(filter.Gid), "", true, false)),
		"userSearchCtrl":   template.HTML(ctrls.BuildDropList("USERSEARCH", strconv.Itoa(filter.Uid), strconv.Itoa(filter.Gid), true, false)),
		"cards":            template.HTML(ctrls.DeviceCards(user.Uid, &filter, false)),
		"devicesIcon":      template.HTML(svg.GetIcon("devices")),
		"groupIcon":        template.HTML(svg.GetIcon("group")),
		"siteIcon":         template.HTML(svg.GetIcon("site")),
		"officeIcon":       template.HTML(svg.GetIcon("office")),
		"searchIcon":       template.HTML(svg.GetIcon("search")),
	}))
}

func PostDevices(c *fiber.Ctx) error {
	// Read incoming requst cookie to get curUid
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	filter := new(db.DeviceFilter)
	if err := c.BodyParser(filter); err != nil {
		return c.Status(fiber.StatusOK).SendString("ERROR")
	}

	//Do processing and saves
	response := ""
	switch filter.Task {
	case "get_next_page", "get_first_page": // Devices page search for devices
		err := filter.SetDeviceFilter(user.Uid) // Remember the device filter settings
		if err != nil {
			log.Println(err)
		}
		response = ctrls.DeviceCards(user.Uid, filter, false)
	case "search_for_devices": // wizard search for devices
		err := filter.SetDeviceFilter(user.Uid) // Remember the device filter settings
		if err != nil {
			log.Println(err)
		}
		response = ctrls.DeviceCards(user.Uid, filter, true)
	}
	return c.Status(fiber.StatusOK).SendString(response)
}
