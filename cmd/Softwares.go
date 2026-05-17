package cmd

import (
	"encoding/json"
	"html/template"
	"log"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/svg"

	"github.com/gbsto/daisy/db"

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

	var filter db.SoftwareFilter
	filter.Init()

	//Render the page
	return c.Render("softwares", addNavigationIcons(fiber.Map{
		"title":          template.HTML(svg.GetIcon("software") + " Software"),
		"fullName":       user.Fullname,
		"isAdmin":        user.IsAdmin,
		"cmd_one":        template.HTML(ctrls.MakeButton(ctrls.BtnNew, user.Permissions.Software.Create)),
		"softwaresTable": template.HTML(ctrls.SoftwaresTable(user.Uid, filter)),
	}))
}

func PostSoftwares(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	var filter db.SoftwareFilter
	if err := c.BodyParser(&filter); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}

	return c.Status(fiber.StatusOK).SendString(ctrls.SoftwaresTable(user.Uid, filter))
}

func PostPreInstalled(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}
	if !user.Permissions.Device.Update {
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}
	var data []db.PreInstalled
	if err := json.Unmarshal(c.Body(), &data); err != nil {
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}
	return c.Status(fiber.StatusOK).SendString(db.SetPreInstalled(data))
}
