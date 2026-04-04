package cmd

import (
	"fmt"
	"strconv"

	"github.com/gbsto/daisy/reports"

	"github.com/gofiber/fiber/v2"
)

func PostReports(c *fiber.Ctx) error {
	type reportsForm struct {
		Task    string `json:"task"`
		DevType string `json:"devtype"`
	}
	var request reportsForm
	curUid, err := strconv.Atoi(fmt.Sprint(c.Locals("curUid")))
	if err != nil || curUid < 1 {
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusOK).SendString("ERROR: ")
	}
	switch request.Task {
	case "DASHBOARD_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetDashboardReport())
	case "ISSUES_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetIssuesReport(curUid, request.DevType))
	case "DEVICES_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetDeviceReport(curUid, request.DevType))
	case "LAST_SEEN_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetLastSeenReport(curUid, true))
	case "BACKUP_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetLastSeenReport(curUid, false))
	case "USERS_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetUsersReport(curUid))
	case "TRACKED_SOFTWARE":
		return c.Status(fiber.StatusOK).SendString(reports.TrackedSoftware())
	case "OTHER_SOFTWARE":
		return c.Status(fiber.StatusOK).SendString(reports.OtherSoftware())
	case "USERS_COMPUTERS":
		return c.Status(fiber.StatusOK).SendString(reports.UsersAssignedDevices())
	case "NETWORK_GAPS":
		return c.Status(fiber.StatusOK).SendString(reports.NetworkGaps())
	}
	return c.Status(fiber.StatusOK).SendString("ERROR: Unknown report")
}
