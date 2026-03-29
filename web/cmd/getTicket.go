package cmd

import (
	"html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/util"

	"github.com/gofiber/fiber/v2"
)

// Note: TODO: add automatic ack if assigned user looks at this ticket.

func GetTicket(c *fiber.Ctx) error {

	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// CRUD Create Read Update Delete
	// If NO Read capababilty, send them home
	// Create and Delete are handled at the control level
	isReadonly := true
	if !user.Permissions.Ticket.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	} else if user.Permissions.Ticket.Update {
		isReadonly = false
	}

	aid, err := strconv.Atoi(c.Query("aid", "0"))
	if err != nil {
		log.Println(err)
		aid = 0
	}
	action, err := db.GetAction(user.Uid, aid)
	if err != nil {
		log.Println(err)
	}

	// What to do if CID = 0 (No Device)
	dev, err := db.GetDevice(user.Uid, action.Cid)
	if err != nil {
		log.Println(err)
	}
	status := "Open"
	if action.Active == 0 {
		status = "Closed"
	}

	isClosed := false
	if action.Active == 0 {
		isReadonly = true
		isClosed = true
	}

	// For OPENED Date: (to make it easier to read)
	// Parse the Localtime string to time.Time
	// Then format the time.Time to the desired format
	t, err := time.Parse("2006-01-02 15:04", action.Localtime)
	if err != nil {
		log.Println("Error parsing date:", err)
	} else {
		action.Localtime = t.Format("2-January-2006")
		action.Localtime += " (" + util.CalcDuration(action.OpenedInt, action.ClosedInt) + ")"
	}

	if action.Inform_gid == 0 {
		action.Inform_gid = db.FindGroupID(0, 0, false)
	}

	// If there is a group, and no user, but a device, lookup that device's assigned user (if any)
	if action.Gid > 0 && action.Uid == 0 && action.Cid > 0 {
		action.Uid, _ = db.GetDeviceAssignedUserGroup(user.Uid, action.Cid)
	}

	if action.Gid == 0 {
		action.Gid = db.FindGroupID(user.Uid, action.Gid, true)
	}

	return c.Render("ticket", fiber.Map{
		"title":         template.HTML("<span class='mif-news icon'></span>&nbsp;Ticket"),
		"fullName":      user.Fullname,
		"cmd_one":       template.HTML(ctrls.BuildRouteButton(user.Permissions.Ticket.Update)),
		"isAdmin":       user.IsAdmin,
		"openedGMT":     action.OpenedInt,
		"closedGMT":     action.ClosedInt,
		"active":        strconv.Itoa(action.Active),
		"cid":           strconv.Itoa(action.Cid),
		"image":         template.HTML(action.Image),
		"deviceCtrl":    template.HTML(ctrls.BuildDeviceCtrl(dev)),
		"deviceName":    dev.Name,
		"deviceIcon":    template.HTML(action.DeviceIcon),
		"softwareCtrl":  template.HTML(ctrls.BuildDropList("SOFTWARE", strconv.Itoa(action.Sid), "", true, isClosed)),
		"troubleCtrl":   template.HTML(ctrls.BuildDropList("TROUBLE", strconv.Itoa(action.Trouble), dev.Type, false, isClosed)),
		"report":        template.HTML(ctrls.BuildReportCtrl(action.Report, isClosed)),
		"notes":         template.HTML(removePrefixSuffixAnchor(action.Action, action.Report, action.Notes)),
		"impactCtrl":    template.HTML(ctrls.BuildDropList("IMPACT", strconv.Itoa(action.Impact), "", false, isClosed)),
		"status":        status,
		"aid":           strconv.Itoa(action.Aid),
		"originator":    action.OriginatorName,
		"email":         action.OriginatorEmail,
		"opened":        action.Localtime,
		"assignedGroup": template.HTML(ctrls.BuildDropList("GROUP", strconv.Itoa(action.Gid), "", false, isClosed)),
		"assignedUser":  template.HTML(ctrls.BuildDropList("USER", strconv.Itoa(action.Uid), strconv.Itoa(action.Gid), true, isClosed)),
		"informGroup":   template.HTML(ctrls.BuildDropList("GROUPINFORM", strconv.Itoa(action.Inform_gid), "", false, isClosed)),
		"informUser":    template.HTML(ctrls.BuildDropList("USERINFORM", strconv.Itoa(action.Inform), strconv.Itoa(action.Inform_gid), false, isClosed)),
		"ctrlCidAck":    template.HTML(ctrls.BuildAckCheckbox(action.Cid_ack, isReadonly, "Device", "cid_ack")),
		"ctrlSidAck":    template.HTML(ctrls.BuildAckCheckbox(action.Sid_ack, isReadonly, "Software", "sid_ack")),
		"wlog":          template.HTML(ctrls.BuildWorklog(user.Uid, action.Aid)),
		"gid":           strconv.Itoa(action.Gid),
		"uid":           strconv.Itoa(action.Uid),
		"assignGroup":   action.AssignedGroupName,
		"assignUser":    action.AssignedUserName,
		"informList":    template.HTML(ctrls.BuildInformList(aid)),
	})
}

// Removes the specified prefix and suffix from the input string,
// and then removes any HTML anchor tags.
func removePrefixSuffixAnchor(prefix, suffix, input string) string {
	if len(input) > len(prefix)+len(suffix) {
		input = input[len(prefix)+1:]
		input = strings.TrimSuffix(input, suffix)
	}
	re := regexp.MustCompile(`<a[^>]*>(.*?)</a>`)
	return re.ReplaceAllString(input, "$1")
}
