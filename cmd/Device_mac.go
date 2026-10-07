package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
	"github.com/gofiber/fiber/v2"
)

func GetDevice_Mac(c *fiber.Ctx) error {

	// Read incoming requst cookie to get curUid
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Network.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	// Read the site from the URL, or default to 0
	site := c.Query("site", "")
	if site == "" {
		site = db.GetDefaultSite()
	}

	return c.Render("device_mac", addNavigationIcons(fiber.Map{
		"title":         template.HTML(svg.GetIcon("network") + " Device / MAC association"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"deviceList":   template.HTML(ctrls.BuildD2MDeviceList(site)),
		"site":          site,
	}))
}


// Return the default values for the select list of a single node in the tree, given its Mid (MacId)
func PostFetchD2MDevice(c *fiber.Ctx) error {
	type request struct {
		Cid  int    `json:"cid"`  // Device ID of the device (Computer ID)
		Site string `json:"site"` // Device site
	}
	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusOK).SendString("malformed request body")
	}
	node, err := db.GetNode(req.Cid, req.Site)
	if err != nil {
		return c.Status(fiber.StatusOK).SendString("error retrieving node")
	}
	return c.Status(fiber.StatusOK).JSON(node)
}

