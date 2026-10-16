package cmd

import (
	"html/template"
	"strconv"

	"github.com/gbsto/daisy/devmacs"
	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/macs"
	"github.com/gbsto/daisy/tree"
	"github.com/gofiber/fiber/v2"
)

func GetDevMacs(c *fiber.Ctx) error {
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
	siteName := ""
	name := ""
	model := ""
	kindName := ""
	officeName := ""
	deviceIcon := ""
	associated := ""
	macTable := ""
	if cid > 0 {
		device, err := devmacs.GetDevice(cid)
		if err == nil {
			site = device.Site
			siteName = device.SiteName
			name = device.Name
			model = device.Model
			kindName = device.KindName
			officeName = device.OfficeName
			associated = devmacs.BuildDevMacs(cid, user.Tzoff, user.Permissions.Device.Update)
			deviceIcon = icons.GetIcon(device.Icon)
			macTable = devmacs.BuildUnmatchedMacs(cid, user.Tzoff, site)
		}
	}

	return c.Render("devmacs", addNavigationIcons(fiber.Map{
		"title":      template.HTML(icons.GetIcon("devices") + " Device / MAC association"),
		"fullName":   user.Fullname,
		"isAdmin":    user.IsAdmin,
		"cid":        cid,
		"site":       site,
		"siteName":   siteName,
		"associated": template.HTML(associated),
		"searchIcon": template.HTML(icons.GetIcon("search")),
		"deviceIcon": template.HTML(deviceIcon),
		"name":       name,
		"model":      model,
		"kindName":   kindName,
		"officeName": officeName,
		"macTable":   template.HTML(macTable),
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

	return c.Status(fiber.StatusOK).SendString(devmacs.BuildDeviceList(recvd.Site))
}

// Unassociate a mac to a device
func PostRemoveMac(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	if !user.Permissions.Device.Update {
		return c.Status(fiber.StatusOK).SendString("ERROR: You do not have permission to update device records.")
	}

	recvd := struct {
		Cid int `json:"cid"`
		Mid int `json:"mid"`
	}{}

	if err := c.BodyParser(&recvd); err != nil {
		return c.Status(fiber.StatusOK).SendString("ERROR: server error.")
	}

	// Do processing to remove the mac association
	var m macs.MacInfo
	m.Mid = recvd.Mid
	m.Unassociate()

	return c.Status(fiber.StatusOK).SendString(devmacs.BuildDevMacs(recvd.Cid, user.Tzoff, false))
}
