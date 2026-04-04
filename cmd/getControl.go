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
	return c.Render("control", fiber.Map{
		"title":           template.HTML("<span class='mif-traff icon'></span>&nbsp;Control"),
		"fullName":        user.Fullname,
		"isAdmin":         user.IsAdmin,
		"cmd_one":         template.HTML(ctrls.MakeSaveButton(false)),
		"cmd_two":         template.HTML(ctrls.MakeAddButton(false)),
		"cmd_three":       template.HTML(ctrls.MakeDeleteButton(false)),
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
	})
}
