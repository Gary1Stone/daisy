package cmd

import (
	"html/template"

	"github.com/gbsto/daisy/comstat"
	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/reports"
	"github.com/gofiber/fiber/v2"
)

func GetReports(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	//Render the page
	return c.Render("reports", addNavigationIcons(fiber.Map{
		"title":    template.HTML(icons.GetIcon("reports") + " Reports"),
		"fullName": user.Fullname,
		"isAdmin":  user.IsAdmin,
	}))
}

func PostReports(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}
	type reportsForm struct {
		Task    string `json:"task"`
		DevType string `json:"devtype"`
	}
	var request reportsForm
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusOK).SendString("ERROR: ")
	}
	// curUid, err := strconv.Atoi(fmt.Sprint(c.Locals("curUid")))
	// if err != nil || curUid < 1 {
	// 	return c.Status(fiber.StatusOK).Redirect("index.html")
	// }

	switch request.Task {
	case "DASHBOARD_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetDashboardReport())
	case "ISSUES_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetIssuesReport(user.Uid, request.DevType))
	case "DEVICES_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetDeviceReport(user.Uid, user.Tzoff, request.DevType))
	case "LAST_SEEN_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetLastSeenReport(user.Uid, user.Tzoff, true))
	case "BACKUP_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetLastSeenReport(user.Uid, user.Tzoff, false))
	case "USERS_REPORT":
		return c.Status(fiber.StatusOK).SendString(reports.GetUsersReport(user.Uid))
	case "TRACKED_SOFTWARE":
		return c.Status(fiber.StatusOK).SendString(reports.TrackedSoftware())
	case "USERS_COMPUTERS":
		return c.Status(fiber.StatusOK).SendString(reports.UsersAssignedDevices())
	case "NETWORK_GAPS":
		return c.Status(fiber.StatusOK).SendString(reports.NetworkGaps())
	case "COMSTAT":
		return c.Status(fiber.StatusOK).SendString(comstat.ComstatTable(user.Uid, user.Tzoff))
	}
	return c.Status(fiber.StatusOK).SendString("ERROR: Unknown report")
}
