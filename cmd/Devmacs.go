package cmd

import (
	"html/template"
	"log"
	"strconv"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/devmacs"
	"github.com/gbsto/daisy/icons"
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
	photo := `<img src="images/missing.jpg" alt="photo" width="200px" >`
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
		var dev devmacs.DeviceInfo
		device, err := dev.GetDevice(cid)
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
			photo = `<img src="images/` + device.Image + `" alt="photo" width="200px" >`
		}
	}

	return c.Render("devmacs", addNavigationIcons(fiber.Map{
		"title":      template.HTML(icons.GetIcon("devices") + " Device / MAC association"),
		"fullName":   user.Fullname,
		"isAdmin":    user.IsAdmin,
		"cmd_one":    template.HTML(ctrls.MakeButton(ctrls.BtnHelp, true)),
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
		"photo":      template.HTML(photo),
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
	err = devmacs.Unassociate(recvd.Mid)
	if err != nil {
		log.Println(err)
	}

	return c.Status(fiber.StatusOK).SendString(devmacs.BuildDevMacs(recvd.Cid, user.Tzoff, false))
}

// Generate a device to mac association recomendation
func PostGetAssociations(c *fiber.Ctx) error {
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

	return c.Status(fiber.StatusOK).SendString(devmacs.BuildAssociatedMacs(recvd.Cid, recvd.Mid))
}

// Associate a MAC to a device
func PostMakeAssociation(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	if !user.Permissions.Device.Update {
		return c.Status(fiber.StatusOK).SendString("ERROR: You do not have permission to update device records.")
	}

	recvd := struct {
		Cid int `json:"cid"` // Computer ID
		Mid int `json:"mid"` // MAC ID
	}{}

	if err := c.BodyParser(&recvd); err != nil {
		return c.Status(fiber.StatusOK).SendString("ERROR: server error.")
	}

	// Do magic here, associate the mac, Add the cid to the macs table for the mid
	msg := "okay"
	err = devmacs.SetAssociatedMac(recvd.Cid, recvd.Mid)
	if err != nil {
		msg = "Server Error: Unable to associate the MAC to the device."
	}
	return c.Status(fiber.StatusOK).SendString(msg)
}
