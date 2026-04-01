package cmd

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func PostTicket(c *fiber.Ctx) error {
	response := "Critical Server Error!"

	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	recvd := new(db.Ticket)
	if err := c.BodyParser(recvd); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).SendString(response)
	}

	switch recvd.Task {
	case "get_person_control":
		// If there is a group, and no user, but a device, lookup that device's assigned user (if any)
		if recvd.Gid > 0 && recvd.Uid == 0 && recvd.Cid > 0 {
			recvd.Uid, _ = db.GetDeviceAssignedUserGroup(user.Uid, recvd.Cid)
		}
		response = ctrls.BuildDropList("USER", strconv.Itoa(recvd.Uid), strconv.Itoa(recvd.Gid), true, false)
	case "get_inform_person_control":
		// If there is a group, and no user, but a device, lookup that device's assigned user (if any)
		if recvd.Gid > 0 && recvd.Uid == 0 && recvd.Cid > 0 {
			recvd.Uid, _ = db.GetDeviceAssignedUserGroup(user.Uid, recvd.Cid)
		}
		response = ctrls.BuildDropList("USERINFORM", strconv.Itoa(recvd.Uid), strconv.Itoa(recvd.Inform_gid), true, false)
	case "route_ticket":
		recvd.Cmd = "ROUTE"
		if recvd.OldGid != recvd.Gid || recvd.OldUid != recvd.Uid {
			recvd.Log += " Reassigned from: " + recvd.OldGroup + ", " + recvd.OldUser
			recvd.Log = strings.TrimSpace(recvd.Log)
		}
		err := db.SetTicket(user.Uid, recvd)
		if err != nil {
			log.Println(err)
		}
		response = ctrls.BuildWorklog(user.Uid, recvd.Aid)
	case "add_log":
		err := db.AddLog(user.Uid, recvd)
		if err != nil {
			log.Println(err)
		}
		if recvd.Cmd == "CLOSED" {
			db.AckAction(user.Uid, recvd.Aid, true, true, true)
			db.AckAlert(user.Uid, user.Uid, recvd.Aid, true, true, false)
			// TODO: Close ticket or mark as closed and notify user ???????????????????
			// Concept: have a notify at close flag in Alerts table DONE

		}
		response = ctrls.BuildWorklog(user.Uid, recvd.Aid)
	}
	return c.Status(fiber.StatusOK).SendString(response)
}
