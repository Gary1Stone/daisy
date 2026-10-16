package cmd

import (
	"html/template"
	"strconv"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/device_mac"
	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/tree"
	"github.com/gofiber/fiber/v2"
)

func GetDevice_Mac(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Device.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	// Read the cid (computer ID) from the URL, or default to 0
	cid, err2 := strconv.Atoi(c.Query("cid", "0"))
	if err2 != nil {
		cid = 0
	}
	site := ""
	name := ""
	model := ""
	kindName := ""
	officeName := ""
	siteName := ""
	deviceIcon := ""
	associated := ""
	if cid > 0 {
		device, err := device_mac.GetDevice(cid)
		if err == nil {
			site = device.Site
			name = device.Name
			model = device.Model
			kindName = device.KindName
			officeName = device.OfficeName
			siteName = device.SiteName
			associated = ctrls.BuildMacs(user.Uid, cid)
			deviceIcon = icons.GetIcon(device.Icon)
		}
	}

	return c.Render("device_mac", addNavigationIcons(fiber.Map{
		"title":      template.HTML(icons.GetIcon("devices") + " Device / MAC association"),
		"fullName":   user.Fullname,
		"isAdmin":    user.IsAdmin,
		"cid":        cid,
		"site":       site,
		"associated": template.HTML(associated),
		"searchIcon": template.HTML(icons.GetIcon("search")),
		"deviceIcon": template.HTML(deviceIcon),
		"name":       name,
		"model":      model,
		"kindName":   kindName,
		"officeName": officeName,
		"siteName":   siteName,
		"macTable":   "one two three",
	}))
}

// Return the default values for the select list of a single node in the tree, given its Mid (MacId)
func PostGetDeviceList(c *fiber.Ctx) error {

	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	if !user.Permissions.Device.Update {
		return c.Status(fiber.StatusOK).SendString("ERROR: You do not have permission to update device records.")
	}

	recvd := struct {
		Cid  int    `json:"cid"`  // ID of the device (Computer ID)
		Site string `json:"site"` // Device site
	}{}

	if err := c.BodyParser(&recvd); err != nil {
		return c.Status(fiber.StatusOK).SendString("ERROR: server error.")
	}
	if len(recvd.Site) == 0 {
		recvd.Site = tree.GetDefaultSite()
	}

	return c.Status(fiber.StatusOK).SendString(device_mac.BuildDeviceList(recvd.Site))
}
