package cmd

import (
	"html/template"
	"log"

	"github.com/gbsto/daisy/aliases"
	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/macs"
	"github.com/gbsto/daisy/online"
	"github.com/gbsto/daisy/svg"

	"github.com/gofiber/fiber/v2"
)

func GetDuplicates(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO admin update capababilty, send them home
	if !user.Permissions.Admin.Update {
		return c.Status(fiber.StatusUnauthorized).Redirect("home.html")
	}

	// First try the list of matching hostnames
	avoidSelect, mac1, mac2 := ctrls.BuildAvoidSelectCtrl()
	if mac1 == "" {
		avoidSelect = "No matching hostnames found."
	}

	chartData, macInfo, title, err := online.GetAvoidChartData(user.Tzoff, mac1, mac2)
	if err != nil {
		log.Println(err)
	}

	return c.Render("duplicates", addNavigationIcons(fiber.Map{
		"title":       template.HTML(icons.GetIcon("devices") + " Network"),
		"fullName":    user.Fullname,
		"isAdmin":     user.IsAdmin,
		"isReadonly":  !user.Permissions.Device.Update,
		"isDisabled":  !user.Permissions.Device.Update,
		"cmd_one":     template.HTML(ctrls.MakeButton(ctrls.BtnHelp, true)),
		"avoidSelect": template.HTML(avoidSelect),
		"avoidChart":  template.HTML(title + svg.BuildOnlineChart(chartData, macInfo, svg.NameWithDate)),
	}))
}

func PostDuplicates(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}
	// If NO admin update capababilty, send them home
	if !user.Permissions.Admin.Update {
		return c.Status(fiber.StatusUnauthorized).Redirect("home.html")
	}

	type Mids struct {
		Mac1 string `json:"mac1"`
		Mac2 string `json:"mac2"`
	}
	recvd := new(Mids)

	if err := c.BodyParser(recvd); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
	}

	chartData, macInfo, title, err := online.GetAvoidChartData(user.Tzoff, recvd.Mac1, recvd.Mac2)
	if err != nil {
		log.Println(err)
	}
	reply := title + svg.BuildOnlineChart(chartData, macInfo, svg.NameWithDate)

	return c.Status(fiber.StatusOK).SendString(reply)
}

func PostDuplicatesJoin(c *fiber.Ctx) error {
	type Mids struct {
		Mac1     string `json:"mac1"`
		Mac2     string `json:"mac2"`
		IsSame   bool   `json:"isSame"`
		IsIgnore bool   `json:"isIgnore"`
	}
	recvd := new(Mids)
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO admin update capababilty, send them home
	if !user.Permissions.Admin.Update {
		return c.Status(fiber.StatusUnauthorized).Redirect("home.html")
	}

	if err := c.BodyParser(recvd); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
	}

	if recvd.IsSame {
		var a aliases.Aliases
		err = a.AddAliasPair(recvd.Mac1, recvd.Mac2)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
		}
	} else if recvd.IsIgnore {
		err = macs.SetIsIgnoreDevices(recvd.Mac1, recvd.Mac2)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
		}
	} else {
		err = macs.SetIsSolitaryDevices(recvd.Mac1, recvd.Mac2)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
		}
	}

	return c.Status(fiber.StatusOK).SendString("ok")
}
