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
	if !user.Permissions.Admin.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	online, offline := db.GetCurrentOnOffCounts()

	var hits svg.SparklineOptions
	hits.Warning = 40 // If more than 40 hits in a day, highlight red

	return c.Render("network", fiber.Map{
		"title":         template.HTML("<span class='mif-devices icon'></span>&nbsp;Network"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"isReadonly":    !user.Permissions.Admin.Update,
		"isDisabled":    !user.Permissions.Admin.Update,
		"cmd_one":       template.HTML(ctrls.MakeAdminHelpButton()),
		"site":          "WKNC",
		"networkImage":  "/images/wknc-network.png",
		"midCtrl":       template.HTML(ctrls.BuildDropList("MID", "", "", false, false)),
		"onlineCount":   online,
		"offlineCount":  offline,
		"avoidanceList": template.HTML(ctrls.BuildAvoidListCtrl()),
		"onlineSpark":   template.HTML(svg.BuildNetworkLoadChart(&hits)),
		"maxHitsMonth":  hits.MaxValue,
	})
}
