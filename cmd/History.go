package cmd

import (
	"html/template"
	"log"
	"strconv"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/macs"
	"github.com/gbsto/daisy/online"
	"github.com/gbsto/daisy/svg"

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

	mac, err := macs.GetMacInfoByMid(user.Tzoff, mid)
	if err != nil {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	chartData, macInfo, err := online.GetOnlineDeviceHistoryChartData(user.Tzoff, mac.Mac)
	if err != nil {
		log.Println(err)
	}

	return c.Render("history", addNavigationIcons(fiber.Map{
		"title":         template.HTML(icons.GetIcon("devices") + " History"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"isReadonly":    !user.Permissions.Device.Update,
		"isDisabled":    !user.Permissions.Device.Update,
		"cmd_one":       template.HTML(ctrls.MakeButton(ctrls.BtnHelp, true)),
		"midCtrl":       template.HTML(ctrls.BuildDropList("MID", midParam, "", false, false)),
		"deviceHistory": template.HTML(svg.BuildOnlineChart(chartData, macInfo, svg.DateOnly)),
	}))
}
