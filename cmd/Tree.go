package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
	"github.com/gofiber/fiber/v2"
)

func GetTree(c *fiber.Ctx) error {

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

	return c.Render("tree", addNavigationIcons(fiber.Map{
		"title":      template.HTML(svg.GetIcon("network") + " Edit Network"),
		"fullName":   user.Fullname,
		"isAdmin":    user.IsAdmin,
		"officeIcon": template.HTML(svg.GetIcon("office")),
		"parentIcon": template.HTML(svg.GetIcon("parent")),
		"tree":       template.HTML(ctrls.BuildTreeView(site)),
		"parentCtrl": template.HTML(ctrls.BuildParentSelect(0, site, !user.Permissions.Network.Update)),
		"kindCtrl":   template.HTML(ctrls.BuildDropList("KIND", "", "", true, !user.Permissions.Network.Update)),
		"officeCtrl": template.HTML(ctrls.BuildDropList("OFFICE", "", site, true, !user.Permissions.Network.Update)),
		"site":       site,
	}))
}

// Return the default values for the select list of a single node in the tree, given its Mid (MacId)
func PostTreeShow(c *fiber.Ctx) error {
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

// Save the updated values for a single node in the tree, given its Mid (MacId)
func PostTreeUpdate(c *fiber.Ctx) error {
	var req db.Node
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusOK).SendString("malformed request body")
	}

	// Update both the device AND the MAC record in the database
	err := db.SetTreeParent(req.Cid, req.Parent, req.Kind, req.Office)

	if err != nil {
		return c.Status(fiber.StatusOK).SendString("error updating device")
	}
	return c.Status(fiber.StatusOK).SendString("OK")
}
