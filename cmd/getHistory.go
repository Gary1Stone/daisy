package cmd

import (
	"html/template"
	"strconv"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetHistory(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Device.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	// Read the mac parameter that was passed in the URL
	midParam := c.Query("mid")
	if midParam == "" {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}
	// convert midParam to integer
	mid, err := strconv.Atoi(midParam)
	if err != nil {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	macInfo, err := db.GetMacInfoByMid(user.Tzoff, mid)
	if err != nil {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	return c.Render("history", fiber.Map{
		"title":         template.HTML("<span class='mif-devices icon'></span>&nbsp;History"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"isReadonly":    !user.Permissions.Device.Update,
		"isDisabled":    !user.Permissions.Device.Update,
		"cmd_one":       template.HTML(ctrls.MakeAdminHelpButton()),
		"midCtrl":       template.HTML(ctrls.BuildDropList("MID", midParam, "", false, false)),
		"deviceHistory": template.HTML(ctrls.GetOnlineDeviceHistory(user.Tzoff, macInfo.Mac)),
	})
}
