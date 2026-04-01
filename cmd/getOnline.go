package cmd

import (
	"html/template"
	"strings"
	"time"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetOnline(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Profile.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	// Read the date parameter that was passed in the URL
	// If not present, default to today's date in YYYYMMDD format
	dateParam := c.Query("date")
	//remove dashes for searching the database
	if dateParam == "" {
		dateParam = time.Now().Format("2006-01-02")
	}
	searchDate := strings.ReplaceAll(dateParam, "-", "")

	// set the date picker limits
	minDate, maxDate := db.MinMaxHistoryDate(user.Uid)

	return c.Render("online", fiber.Map{
		"title":         template.HTML("<span class='mif-user icon'></span>&nbsp;Online"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"isReadonly":    !user.Permissions.Profile.Update,
		"isDisabled":    !user.Permissions.Profile.Update,
		"cmd_one":       template.HTML(ctrls.MakeAdminHelpButton()),
		"minDate":       minDate,
		"maxDate":       maxDate,
		"dateParam":     dateParam,
		"onlineDevices": template.HTML(ctrls.GetOnlineDevices(user.Tzoff, searchDate)),
		"kindCtrl":      template.HTML(ctrls.BuildDropList("KIND", "", "", true, false)),
		"officeCtrl":    template.HTML(ctrls.BuildDropList("OFFICE", "", "", true, false)),
		"offices":       template.HTML(db.BuildFieldList("OFFICE")),
	})
}
