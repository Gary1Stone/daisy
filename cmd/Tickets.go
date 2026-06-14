package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/svg"

	"github.com/gofiber/fiber/v2"
)

func GetTickets(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If no read capababilty, send them home
	if !user.Permissions.Profile.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}
	isAdmin := user.Permissions.Admin.Read

	//Render the page
	return c.Render("tickets", addNavigationIcons(fiber.Map{
		"title":        template.HTML(svg.GetIcon("ticket") + " Tickets"),
		"fullName":     user.Fullname,
		"isAdmin":      isAdmin,
		"cmd_one":      template.HTML(ctrls.MakeAddButton(user.Permissions.Ticket.Create)),
		"ticketsTable": template.HTML(ctrls.BuildTicketsTable(user.Uid)),
	}))
}
