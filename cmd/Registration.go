package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/register"
	"github.com/gbsto/daisy/util"

	"github.com/gofiber/fiber/v2"
)

func GetRegistration(c *fiber.Ctx) error {
	var r register.RegisterCode
	r.Ip = c.IP()
	ips := c.IPs()
	if len(ips) > 0 {
		r.Ip = ips[0]
	}
	r.Apicode = util.GetRandomPassword()
	r.SetApiCode()

	return c.Render("registration", fiber.Map{
		"user":     "",
		"apicode":  r.Apicode,
		"userIcon": template.HTML(icons.GetIcon("user")),
	})
}
