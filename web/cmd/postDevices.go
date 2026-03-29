package cmd

import (
	"log"
	"strconv"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func PostDevices(c *fiber.Ctx) error {
	// Read incoming requst cookie to get curUid
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	filter := new(db.DeviceFilter)
	if err := c.BodyParser(filter); err != nil {
		return c.Status(fiber.StatusOK).SendString("ERROR")
	}

	//Do processing and saves
	response := ""
	switch filter.Task {
	case "get_office_control":
		response = ctrls.BuildDropList("OFFICESEARCH", filter.Office, filter.Site, true, false)
	case "get_user_control":
		response = ctrls.BuildDropList("USERSEARCH", strconv.Itoa(filter.Uid), strconv.Itoa(filter.Gid), true, false)
	case "get_next_page", "get_first_page": // Devices page search for devices
		err := filter.SetDeviceFilter(user.Uid) // Remember the device filter settings
		if err != nil {
			log.Println(err)
		}
		response = ctrls.DeviceCards(user.Uid, filter, false)
	case "search_for_devices": // wizard search for devices
		err := filter.SetDeviceFilter(user.Uid) // Remember the device filter settings
		if err != nil {
			log.Println(err)
		}
		response = ctrls.DeviceCards(user.Uid, filter, true)
	}
	return c.Status(fiber.StatusOK).SendString(response)
}
