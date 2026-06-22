package cmd

import (
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/svg"
	"github.com/gbsto/daisy/util"

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

	return c.Render("online", addNavigationIcons(fiber.Map{
		"title":         template.HTML(svg.GetIcon("user") + " Online"),
		"fullName":      user.Fullname,
		"isAdmin":       user.IsAdmin,
		"isReadonly":    !user.Permissions.Profile.Update,
		"isDisabled":    !user.Permissions.Profile.Update,
		"cmd_two":       template.HTML(ctrls.MakeButton(ctrls.BtnHelp, true)),
		"minDate":       minDate,
		"maxDate":       maxDate,
		"dateParam":     dateParam,
		"onlineDevices": template.HTML(ctrls.GetOnlineDevices(user.Tzoff, searchDate)),
		"kindCtrl":      template.HTML(ctrls.BuildDropList("KIND", "", "", true, false)),
		"officeCtrl":    template.HTML(ctrls.BuildDropList("OFFICE", "", "", true, false)),
		"offices":       template.HTML(db.BuildFieldList("OFFICE")),
	}))
}

func PostOnlineGetMac(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	reply := struct {
		Success bool   `json:"success"` // true = no errors
		Msg     string `json:"msg"`     // Error message
	}{
		Success: false,
		Msg:     "CRITICAL SERVER ERROR!",
	}

	type Mid struct {
		Mid int `json:"Mid"`
	}

	recvd := new(Mid) // The new() allocates HEAP to create the variable/struct, therefore must use address operator(*) in functions

	if err := c.BodyParser(recvd); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(reply)
	}

	// Check user permissions
	var perm util.Permissions
	perm.GetPermissions(fmt.Sprint(c.Locals("permissions")))
	if !perm.Device.Read {
		return c.Status(fiber.StatusUnauthorized).Redirect("home.html")
	}

	// Get the device information
	macInfo, err := db.GetMacInfoByMid(user.Tzoff, recvd.Mid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(reply)
	}

	type MacInfoView struct {
		db.MacInfo
		Summary string `json:"Summary"` // Text summary for display

	}
	view := MacInfoView{
		MacInfo: macInfo,
		Summary: buildSummary(macInfo),
	}

	return c.Status(fiber.StatusOK).JSON(view)
}

func buildSummary(mac db.MacInfo) string {
	summary := ""
	if mac.Intruder {
		summary += "<p><b>Intruder:</b> Yes</p>"
	}
	summary += "<p><b title='Media Access Control address'>MAC:</b> " + mac.Mac + "</p>"
	summary += "<p><b>Hostname:</b> " + mac.Hostname + "</p>"
	summary += "<p><b>IP:</b> " + mac.Ip + "</p>"
	summary += "<p><b>Site:</b> " + db.GetCodeDescription("SITE", mac.Site) + "</p>"
	summary += "<p><b>OS:</b> " + mac.Os + "</p>"
	summary += "<p><b>NIC Vendor:</b> " + mac.Vendor + "</p>"
	summary += "<p><b>First &sol; Last Seen:</b> " + mac.Firstseen + " &sol; " + mac.Lastseen + "</p>"
	return summary
}

func PostOnlineSetMac(c *fiber.Ctx) error {
	_, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	recvd := new(db.MacInfo) // The new() allocates HEAP to create the variable/struct, therefore must use address operator(*) in functions
	if err := c.BodyParser(recvd); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("CRITICAL SERVER ERROR!")
	}

	// Check user permissions
	var perm util.Permissions
	perm.GetPermissions(fmt.Sprint(c.Locals("permissions")))

	if !perm.Device.Update {
		log.Println("ERROR: You do not have permission to update device records.")
		return c.Status(fiber.StatusUnauthorized).Redirect("home.html")
	}

	// Do processing and saves
	err = db.UpdateMac(*recvd)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).SendString("CRITICAL SERVER ERROR!")
	}

	return c.Status(fiber.StatusOK).SendString("ok")
}
