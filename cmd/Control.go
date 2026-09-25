package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/svg"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetControl(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}
	// If NO Read capababilty, send them home
	if !user.Permissions.Admin.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	return c.Render("control", addNavigationIcons(fiber.Map{
		"title":           template.HTML(svg.GetIcon("control") + " Daisy Server Control"),
		"fullName":        user.Fullname,
		"isAdmin":         user.IsAdmin,
		"attacksDay":      template.HTML(svg.GraphCache.GetGraph(svg.AttacksPerDay)),
		"attacksWeek":     template.HTML(svg.GraphCache.GetGraph(svg.AttacksPerWeek)),
		"attacksMonth":    template.HTML(svg.GraphCache.GetGraph(svg.AttacksPerMonth)),
		"maxAttacksDay":   svg.GraphCache.GetMax(svg.AttacksPerDay),
		"maxAttacksWeek":  svg.GraphCache.GetMax(svg.AttacksPerWeek),
		"maxAttacksMonth": svg.GraphCache.GetMax(svg.AttacksPerMonth),
		"loginsDay":       template.HTML(svg.GraphCache.GetGraph(svg.LoginsPerDay)),
		"loginsWeek":      template.HTML(svg.GraphCache.GetGraph(svg.LoginsPerWeek)),
		"loginsMonth":     template.HTML(svg.GraphCache.GetGraph(svg.LoginsPerMonth)),
		"maxLoginsDay":    svg.GraphCache.GetMax(svg.LoginsPerDay),
		"maxLoginsWeek":   svg.GraphCache.GetMax(svg.LoginsPerWeek),
		"maxLoginsMonth":  svg.GraphCache.GetMax(svg.LoginsPerMonth),
		"hitsDay":         template.HTML(svg.GraphCache.GetGraph(svg.HitsPerDay)),
		"hitsWeek":        template.HTML(svg.GraphCache.GetGraph(svg.HitsPerWeek)),
		"hitsMonth":       template.HTML(svg.GraphCache.GetGraph(svg.HitsPerMonth)),
		"maxHitsDay":      svg.GraphCache.GetMax(svg.HitsPerDay),
		"maxHitsWeek":     svg.GraphCache.GetMax(svg.HitsPerWeek),
		"maxHitsMonth":    svg.GraphCache.GetMax(svg.HitsPerMonth),
	}))
}

// Pop-up dialogs of the details for the graphs
func PostControl(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	recvd := struct {
		Task string `json:"task"`
		Id   int    `json:"id"`
	}{}

	if err := c.BodyParser(&recvd); err != nil {
		return c.Status(fiber.StatusOK).SendString("Server Error")
	}

	if !user.Permissions.Admin.Read {
		return c.Status(fiber.StatusOK).SendString("Permissions Error")
	}

	reply := ""

	//Do processing and saves
	switch recvd.Task {
	case "get_active_users":
		reply = ctrls.BuildActiveUsersTable(user.Uid)
	case "end_session":
		db.EndSession(recvd.Id)
		reply = ctrls.BuildActiveUsersTable(user.Uid)
	case "end_session_all":
		reply = ctrls.BuildActiveUsersTable(user.Uid)
	case "get_server_load":
		reply = ctrls.BuildActiveUsersTable(user.Uid)
	case "get_attacks":
		reply = ctrls.BuildAttacksTable(user.Uid, recvd.Id)
	}
	return c.Status(fiber.StatusOK).SendString(reply)
}
