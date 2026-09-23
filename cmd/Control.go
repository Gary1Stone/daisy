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
		"title":           template.HTML(svg.GetIcon("control") + " Control"),
		"fullName":        user.Fullname,
		"isAdmin":         user.IsAdmin,
		"attacksDay":      template.HTML(svg.GraphCache.GetGraph(0)),
		"attacksWeek":     template.HTML(svg.GraphCache.GetGraph(1)),
		"attacksMonth":    template.HTML(svg.GraphCache.GetGraph(2)),
		"maxAttacksDay":   svg.GraphCache.GetMax(0),
		"maxAttacksWeek":  svg.GraphCache.GetMax(1),
		"maxAttacksMonth": svg.GraphCache.GetMax(2),
		"loginsDay":       template.HTML(svg.GraphCache.GetGraph(3)),
		"loginsWeek":      template.HTML(svg.GraphCache.GetGraph(4)),
		"loginsMonth":     template.HTML(svg.GraphCache.GetGraph(5)),
		"maxLoginsDay":    svg.GraphCache.GetMax(3),
		"maxLoginsWeek":   svg.GraphCache.GetMax(4),
		"maxLoginsMonth":  svg.GraphCache.GetMax(5),
		"hitsDay":         template.HTML(svg.GraphCache.GetGraph(6)),
		"hitsWeek":        template.HTML(svg.GraphCache.GetGraph(7)),
		"hitsMonth":       template.HTML(svg.GraphCache.GetGraph(8)),
		"maxHitsDay":      svg.GraphCache.GetMax(6),
		"maxHitsWeek":     svg.GraphCache.GetMax(7),
		"maxHitsMonth":    svg.GraphCache.GetMax(8),
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
