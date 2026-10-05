package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/svg"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetNetwork(c *fiber.Ctx) error {
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
	// If the count is greater than 0, then we need to show the site select list, otherwise just use a hidden field for the site

	return c.Render("network", addNavigationIcons(fiber.Map{
		"title":        template.HTML(svg.GetIcon("network") + " Network"),
		"fullName":     user.Fullname,
		"isAdmin":      user.IsAdmin,
		"isReadonly":   !user.Permissions.Admin.Update,
		"isDisabled":   !user.Permissions.Admin.Update,
		"cmd_one":      template.HTML(ctrls.MakeButton(ctrls.BtnOnline, true)),
		"cmd_two":      template.HTML(ctrls.MakeButton(ctrls.BtnDuplicate, true)),
		"cmd_three":    template.HTML(ctrls.MakeButton(ctrls.BtnCorrelate, true)),
		"cmd_four":     template.HTML(ctrls.MakeButton(ctrls.BtnHelp, true)),
		"networkImage": "/images/wknc-network.png",
		"officeIcon":   template.HTML(svg.GetIcon("office")),
		"parentIcon":   template.HTML(svg.GetIcon("parent")),
		"historyIcon":  template.HTML(svg.GetIcon("history")),
		"cloneIcon":    template.HTML(svg.GetIcon("clone")),
		"equalsIcon":   template.HTML(svg.GetIcon("equals")),
		"tree":         template.HTML(ctrls.BuildTreeView(site)),
		"siteCtrl":     template.HTML(ctrls.BuildDropList("SITE", site, "", false, !user.Permissions.Network.Read)),
		"parentCtrl":   template.HTML(ctrls.BuildParentSelect(0, site, !user.Permissions.Network.Update)),
		"kindCtrl":     template.HTML(ctrls.BuildDropList("KIND", "", "", true, !user.Permissions.Network.Update)),
		"officeCtrl":   template.HTML(ctrls.BuildDropList("OFFICE", "", site, true, !user.Permissions.Network.Update)),
		"site":         site,
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

	return c.Status(fiber.StatusOK).SendString(ctrls.BuildTreeView(req.Site))
}
