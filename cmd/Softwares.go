package cmd

import (
	"encoding/json"
	"html/template"
	"log"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/inventory"
	"github.com/gbsto/daisy/software"

	"github.com/gofiber/fiber/v2"
)

func GetSoftwares(c *fiber.Ctx) error {

	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If no read capababilty, send them home
	if !user.Permissions.Software.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	var unwanted software.Unwanted
	var other software.Other
	var filter software.SoftwareFilter
	filter.Init()

	//Render the page
	return c.Render("softwares", addNavigationIcons(fiber.Map{
		"title":            template.HTML(icons.GetIcon("software") + " Software"),
		"fullName":         user.Fullname,
		"isAdmin":          user.IsAdmin,
		"cmd_one":          template.HTML(ctrls.MakeButton(ctrls.BtnNew, user.Permissions.Software.Create)),
		"trackedSoftware":  template.HTML(ctrls.SoftwaresTable(user.Uid, user.Tzoff, filter)),
		"otherSoftware":    template.HTML(other.BuildOtherTable()),
		"unwantedSoftware": template.HTML(unwanted.BuildUnwantedEditTable()),
	}))
}

func PostSoftwares(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	var filter software.SoftwareFilter
	if err := c.BodyParser(&filter); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}

	return c.Status(fiber.StatusOK).SendString(ctrls.SoftwaresTable(user.Uid, user.Tzoff, filter))
}

func PostPreInstalled(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}
	if !user.Permissions.Device.Update {
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}
	var data []inventory.PreInstalled
	if err := json.Unmarshal(c.Body(), &data); err != nil {
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}
	return c.Status(fiber.StatusOK).SendString(inventory.SetPreInstalled(data))
}

func PostUnwantedAdd(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	if !user.Permissions.Admin.Update {
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}

	var rcvd software.Unwanted
	if err := json.Unmarshal(c.Body(), &rcvd); err != nil {
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}
	rcvd.Add()
	return c.Status(fiber.StatusOK).SendString(rcvd.BuildUnwantedTable())
}

func PostUnwantedDelete(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	if !user.Permissions.Admin.Update {
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}

	var rcvd software.Unwanted
	if err := json.Unmarshal(c.Body(), &rcvd); err != nil {
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}
	rcvd.Delete()
	return c.Status(fiber.StatusOK).SendString(rcvd.BuildUnwantedTable())
}
