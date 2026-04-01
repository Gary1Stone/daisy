package cmd

import (
	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

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
