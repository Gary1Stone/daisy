package cmd

import (
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/util"

	"github.com/gofiber/fiber/v2"
)

func GetRegistration(c *fiber.Ctx) error {
	ip := c.IP()
	ips := c.IPs()
	if len(ips) > 0 {
		ip = ips[0]
	}
	apicode := util.GetRandomPassword()
	db.SetApiCode(apicode, ip)

	return c.Render("registration", fiber.Map{
		"user":    "",
		"apicode": apicode,
	})
}
