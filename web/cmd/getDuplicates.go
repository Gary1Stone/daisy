package cmd

import (
	"html/template"
	"log"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

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

	return c.Render("duplicates", fiber.Map{
		"title":       template.HTML("<span class='mif-devices icon'></span>&nbsp;Network"),
		"fullName":    user.Fullname,
		"isAdmin":     user.IsAdmin,
		"isReadonly":  !user.Permissions.Device.Update,
		"isDisabled":  !user.Permissions.Device.Update,
		"cmd_one":     template.HTML(ctrls.MakeAdminHelpButton()),
		"avoidSelect": template.HTML(avoidSelect),
		"avoidChart":  template.HTML(ctrls.GetAvoidChart(user.Tzoff, mac1, mac2)),
	})
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

	return c.Status(fiber.StatusOK).SendString(ctrls.GetAvoidChart(user.Tzoff, recvd.Mac1, recvd.Mac2))
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
		err = db.AddAliasPair(recvd.Mac1, recvd.Mac2)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
		}
	} else if recvd.IsIgnore {
		err = db.SetIsIgnoreDevices(recvd.Mac1, recvd.Mac2)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
		}
	} else {
		err = db.SetIsSolitaryDevices(recvd.Mac1, recvd.Mac2)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
		}
	}

	return c.Status(fiber.StatusOK).SendString("ok")
}
