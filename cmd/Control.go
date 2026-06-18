package cmd

import (
	"html/template"
	"log"

	"github.com/gbsto/daisy/reports"
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

	var opts svg.SparklineOptions
	opts.Warning = 10 // If more than 10 hits in a 15 minute interval, highlight red
	opts.Duration = svg.Day
	attacksDay := svg.BuildAttackChart(&opts)
	maxAttacksDay := opts.MaxValue
	opts.Duration = svg.Week
	attacksWeek := svg.BuildAttackChart(&opts)
	maxAttacksWeek := opts.MaxValue
	opts.Duration = svg.Month
	attacksMonth := svg.BuildAttackChart(&opts)
	maxAttacksMonth := opts.MaxValue

	var login svg.SparklineOptions
	login.Warning = 250 // If more than 250 logins in a 15 minute interval, highlight red
	login.Duration = svg.Day
	loginsDay := svg.BuildLoginsChart(&login)
	maxLoginsDay := login.MaxValue
	login.Duration = svg.Week
	loginsWeek := svg.BuildLoginsChart(&login)
	maxLoginsWeek := login.MaxValue
	login.Duration = svg.Month
	loginsMonth := svg.BuildLoginsChart(&login)
	maxLoginsMonth := login.MaxValue

	var hits svg.SparklineOptions
	hits.Warning = 1000 // If more than 1000 hits in a 15 minute interval, highlight red
	opts.Duration = svg.Day
	hitsDay := svg.BuildHitsChart(&hits)
	maxHitsDay := hits.MaxValue
	opts.Duration = svg.Week
	hitsWeek := svg.BuildHitsChart(&hits)
	maxHitsWeek := hits.MaxValue
	hits.Duration = svg.Month
	hitsMonth := svg.BuildHitsChart(&hits)
	maxHitsMonth := hits.MaxValue

	// The device meta data is used in most tables on this page,
	// So we get it here to significantly reduce the number of database hits
	devInfo, err := db.GetDeviceMeta()
	if err != nil {
		log.Println(err)
	}

	// Render the page
	return c.Render("control", addNavigationIcons(fiber.Map{
		"title":           template.HTML(svg.GetIcon("control") + " Control"),
		"fullName":        user.Fullname,
		"isAdmin":         user.IsAdmin,
		"cmd_one":         template.HTML(ctrls.MakeButton(ctrls.BtnSave, user.Permissions.Admin.Update)),
		"cmd_two":         template.HTML(ctrls.MakeButton(ctrls.BtnNew, user.Permissions.Admin.Create)),
		"cmd_three":       template.HTML(ctrls.MakeButton(ctrls.BtnDelete, user.Permissions.Admin.Delete)),
		"attacksDay":      template.HTML(attacksDay),
		"attacksWeek":     template.HTML(attacksWeek),
		"attacksMonth":    template.HTML(attacksMonth),
		"maxAttacksDay":   maxAttacksDay,
		"maxAttacksWeek":  maxAttacksWeek,
		"maxAttacksMonth": maxAttacksMonth,
		"dashboard":       template.HTML(reports.GetDeviceCounts()),
		"loginsDay":       template.HTML(loginsDay),
		"loginsWeek":      template.HTML(loginsWeek),
		"loginsMonth":     template.HTML(loginsMonth),
		"maxLoginsDay":    maxLoginsDay,
		"maxLoginsWeek":   maxLoginsWeek,
		"maxLoginsMonth":  maxLoginsMonth,
		"hitsDay":         template.HTML(hitsDay),
		"hitsWeek":        template.HTML(hitsWeek),
		"hitsMonth":       template.HTML(hitsMonth),
		"maxHitsDay":      maxHitsDay,
		"maxHitsWeek":     maxHitsWeek,
		"maxHitsMonth":    maxHitsMonth,
		"LastSeenDevices": template.HTML(reports.LastSeenDevices(user.Uid, devInfo)),
		"checkins":        template.HTML(reports.Checkins(user.Uid, devInfo)),
		"backups":         template.HTML(reports.Backups(user.Uid, devInfo)),
		"drivespace":      template.HTML(reports.Drivespace(user.Uid, devInfo)),
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
