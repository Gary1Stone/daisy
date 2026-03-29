package cmd

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/web/wizards"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type wizFormData struct {
	Task      string `json:"task"`
	Cid       int    `json:"cid"`
	Sid       int    `json:"sid"`
	Gid       int    `json:"gid"`
	Uid       int    `json:"uid"`
	Site      string `json:"site"`
	Office    string `json:"office"`
	Location  string `json:"location"`
	Notes     string `json:"notes"`
	Installed int64  `json:"installed"`
	Impact    int    `json:"impact"`
	Trouble   int    `json:"trouble"`
	Wizard    string `json:"wizard"`
	Type      string `json:"type"`
}

func PostWizard(c *fiber.Ctx) error {
	//	start := time.Now()

	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	recvd := new(wizFormData)
	reply := "Okay"
	if err := c.BodyParser(recvd); err != nil {
		return c.Status(fiber.StatusOK).SendString("Server Error")
	}

	switch recvd.Task {
	case wizards.Sighting:
		sighting(user.Uid, recvd)
	case wizards.Claiming:
		claiming(user.Uid, recvd)
	case wizards.Using:
		using(user.Uid, recvd)
	case wizards.Giving:
		giving(user.Uid, recvd)
	case wizards.Broken:
		broken(user.Uid, recvd)
	case wizards.Lost:
		lost(user.Uid, recvd)
	case wizards.Died:
		died(user.Uid, recvd)
	case wizards.Backup:
		backup(user.Uid, recvd)
	case wizards.Install:
		install(user.Uid, recvd)
	case wizards.Remove:
		remove(user.Uid, recvd)
	case wizards.Request:
		request(user.Uid, recvd)
	case wizards.Care:
		care(user.Uid, recvd)
	case "get_location_from_device":
		reply = ctrls.LocationCtrl(user.Uid, recvd.Cid)
	case "get_site_control":
		// If we have a computer, use the computer's assigned site
		if recvd.Cid > 0 {
			site, _ := db.GetDeviceAssignedSiteOffice(user.Uid, recvd.Cid)
			if len(site) > 0 {
				recvd.Site = site
			}
		}
		reply = ctrls.BuildDropList("SITE", recvd.Site, "", false, false)
	case "get_office_control":
		// If the computer is in the selected site (recvd.Site), set the default office
		if recvd.Cid > 0 {
			site, office := db.GetDeviceAssignedSiteOffice(user.Uid, recvd.Cid)
			if recvd.Site == site {
				recvd.Office = office
			}
		}
		reply = ctrls.BuildDropList("OFFICE", recvd.Office, recvd.Site, true, false)
	case "get_group_control":
		withBlank := true
		if recvd.Wizard == "broken" || recvd.Wizard == "died" || recvd.Wizard == "care" || recvd.Wizard == "request" {
			withBlank = false // not optional, user must pick
		}
		defaultToCurUser := false
		if recvd.Wizard == "claim" {
			defaultToCurUser = true
		}
		if recvd.Gid == 0 {
			recvd.Gid = db.FindGroupID(user.Uid, recvd.Cid, defaultToCurUser)
		}
		reply = ctrls.BuildDropList("GROUP", strconv.Itoa(recvd.Gid), "", withBlank, false)
	case "get_person_control":
		defaultUid := 0
		if recvd.Wizard == "claim" {
			defaultUid = user.Uid
		}
		includeBlankOption := true
		if recvd.Wizard == "broken" || recvd.Wizard == "care" || recvd.Wizard == "request" || recvd.Wizard == "claim" {
			includeBlankOption = false // Broken and care needs someone to look into it.
		}
		// If there is a group, and no user, but a device, lookup that device's assigned user (if any)
		if recvd.Gid > 0 && recvd.Uid == 0 && recvd.Cid > 0 {
			defaultUid, _ = db.GetDeviceAssignedUserGroup(user.Uid, recvd.Cid)
		}
		reply = ctrls.BuildDropList("USER", strconv.Itoa(defaultUid), strconv.Itoa(recvd.Gid), includeBlankOption, false)
	case "get_impact_control":
		reply = ctrls.BuildDropList("IMPACT", "-1", "", false, false)
	case "get_trouble_control":
		reply = ctrls.BuildDropList("TROUBLE", "-1", recvd.Type, false, false)
	case "get_office_search_control":
		reply = ctrls.BuildDropList("OFFICESEARCH", "", recvd.Site, true, false)
	case "get_user_search_control":
		reply = ctrls.BuildDropList("USERSEARCH", strconv.Itoa(recvd.Uid), strconv.Itoa(recvd.Gid), true, false)
	}
	//	log.Printf("PostWizard took %s", time.Since(start))
	return c.Status(fiber.StatusOK).SendString(reply)
}

func StatusSelect() string {
	var ctrl strings.Builder
	ctrl.WriteString("<select id='status' data-role='select' data-filter='false' >")
	ctrl.WriteString("<option value='1' selected>Open</option>")
	ctrl.WriteString("<option value='0'>Closed</option>")
	ctrl.WriteString("</select>")
	return ctrl.String()
}

// Notifictions for the actions:
// SIGHTING (Alert) --> CID y, SID n, UID y (who seen with it, becomes owner), INFORM Y (current Owner)
// USING            --> CID y, SID n, UID n, INFORM n --> Closed on open
// CLAIMING (Alert) --> CID y, SID n, UID n, INFORM y (previous owner)
// GIVING (Alert)   --> CID y, SID n, UID n, INFORM y (new owner)
// BROKEN (Ticket)  --> CID y, SID n, UID y (who will fix), INFORM y (current owner) -> NEEDS WORK LOG TO BE CLOSED! (Wait until closed to notify INFORM)
// LOST (Ticket)    --> CID y, SID n, UID y (who will fix), INFORM Y (current owner) -> NEEDS WORK LOG TO BE CLOSED! (Wait until closed to notify INFORM)
// DIED (Ticket)    --> CID y, SID n, UID y (who will fix), INFORM Y (current owner) -> NEEDS WORK LOG TO BE CLOSED! (Wait until closed to notify INFORM)
// CARE (Ticket)    --> CID y, SID n, UID y (who will fix), INFORM Y (current owner) -> NEEDS WORK LOG TO BE CLOSED! (Wait until closed to notify INFORM)
// BACKUP (Alert)   --> CID y, SID n, UID n, INFORM n
// INSTALL (Alert)  --> CID y, SID y, UID n, INFORM y (device current owner)
// REMOVE (Alert)   --> CID y, SID y, UID n, INFORM y (device current owner)
// REQUEST (Ticket) --> CID y, SID y, UID y, INFORM y (person requesting the software) -> NEEDS WORK LOG TO BE CLOSED! (Wait until closed to notify INFORM)

// Initialize and return a new db.Action with common fields
func newAction(curUid int, wiz *wizFormData) db.Action {
	var act db.Action
	act.Active = 1
	act.Action = wiz.Task
	act.Originator = curUid
	act.Cid = wiz.Cid
	act.Sid = wiz.Sid
	act.Report = wiz.Notes
	act.Trouble = wiz.Trouble
	act.Impact = wiz.Impact
	return act
}

// Set the person to look at the ticket, return user's full name and group id
func assignUser(curUid, uid, gid int, act *db.Action) (string, int) {
	act.Uid = uid
	// If gid missing, look up the user's group
	if uid > 0 && gid == 0 {
		gid = db.GetGid(uid)
	}
	act.Gid = gid
	// If closed, and there is a group or user, mark as acknowledged
	if act.Active == 0 && (gid > 0 || uid > 0) {
		act.Uid_ack = curUid
	}
	return setProfileColor(uid, colors.Alert)
}

// Set a profile (notification) color, return user's name and group id
func setProfileColor(uid int, color string) (string, int) {
	if uid < 1 {
		return "", 0
	}
	return db.SetProfileColor(uid, color)
}

// Save the alerts (max 2 alerts per action, so confusing optimization not needed)
func saveAlerts(curUid, aid, active int, alerts []db.Alert) {
	for _, alert := range alerts {
		alert.Aid = aid
		if active == 0 {
			alert.Ack = curUid
		}
		db.AddAlert(&alert)
	}
}

// Update device record to say device was seen
func saveSeenDevice(cid int) {
	db.SetAuditCheckin(cid)
}

// Update device record and returns note comment and previous assigned Uid
func updateDevice(curUid int, wiz *wizFormData) (string, int) {
	if wiz.Cid < 1 {
		return "", 0
	}
	dev, err := db.GetDevice(curUid, wiz.Cid)
	if err != nil {
		log.Println(err)
	}

	// Remember previous owner user ID
	devUid := dev.Uid

	// Build Note message
	note := cases.Title(language.English).String(wiz.Task)
	note += " <a href='device.html?cid=" + strconv.Itoa(dev.Cid) + "' >" + dev.Name + "</a>"

	// Device ownership changes
	if wiz.Task == "CLAIMING" || wiz.Task == "GIVING" || wiz.Task == "SIGHTING" {
		dev.Uid = wiz.Uid
		if wiz.Uid > 0 {
			dev.Gid = db.GetGid(wiz.Uid) // The user's Group
		} else {
			dev.Gid = wiz.Gid // New group assigned, no user selected
		}
	}

	// Device location changed
	if len(wiz.Site) > 0 {
		dev.Site = wiz.Site
		note += " at " + db.GetCodeDescription("SITE", wiz.Site)
	}
	if len(wiz.Office) > 0 {
		dev.Office = wiz.Office
		note += ": " + db.GetCodeDescription("OFFICE", wiz.Office)
	}
	if len(wiz.Location) > 0 {
		dev.Location = wiz.Location
		note += ": " + wiz.Location
	}
	if !db.SetDevice(curUid, &dev) {
		log.Println("ERROR: Cannot save device record")
	}
	return note, devUid
}

// Set the user's profile color and return an alert for them
func createAlert(uid int, msg string, act *db.Action) db.Alert {
	wait := 0
	colour := "info"
	switch act.Action {
	case "BROKEN", "LOST", "DIED", "CARE", "REQUEST":
		colour = colors.Alert
		wait = 1
	case "CLAIMING", "GIVING":
		colour = colors.Warning
	case "SIGHTING", "USING", "BACKUP", "INSTALL", "REMOVE":
		colour = colors.Yellow
	}
	fullname, gid := setProfileColor(uid, colour)
	act.Notes += msg + fullname + ". "
	return db.Alert{Uid: uid, Gid: gid, Wait: wait}
}

// ***********************************************************************************
// Handle wizards for each action type
// ***********************************************************************************

// User is reporting seeing a computer/device
// Update computer record if site/office/location/Owner(UID) are nonzero
// Panels displayed to the user are:
// "DeviceSelect", "SiteSelect", "OfficeSelect", "LocationText","UserSelect", "GroupSelect", "NotesText"
// SIGHTING (Alert) --> CID y, SID n, UID y (who seen with it, becomes owner), INFORM Y (current Owner)
func sighting(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)

	// Assume this action is closed unless there are alerts.
	// Active=0 Triggers closed_by=curUid and closedtime=now(), and fills cid_ack and sid_ack
	act.Active = 0

	// Remember any alerts until after AID is assigned and then they can be saved
	var alerts []db.Alert

	// If self assigned, ack the computer flag
	if act.Cid > 0 && wiz.Uid == curUid {
		act.Cid_ack = curUid
	}

	// No ticket (follow up) is needed for this task
	assignUser(curUid, 0, 0, &act)

	// Assign the new user to the device, and adjust the location/Site/Office if necessary
	// Get the previous device owner's ID (devUid)
	note, devUid := updateDevice(curUid, wiz)
	act.Notes = note + ". "

	// Notify new owner if different from current and previous owners
	if wiz.Uid > 0 && wiz.Uid != curUid && devUid != wiz.Uid {
		act.Active = 1
		alerts = append(alerts, createAlert(wiz.Uid, "Assigned to: ", &act))
	}

	// Let the previous device owner know about the change
	if devUid > 0 && devUid != curUid && devUid != wiz.Uid {
		act.Active = 1
		alerts = append(alerts, createAlert(devUid, "Reassigned from: ", &act))
	}
	act.Notes += wiz.Notes

	// Add the new action
	if err := act.AddAction(curUid); err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
	saveSeenDevice(act.Cid)
}

// User is self-assigning a computer to themself
// "DeviceSelect", "SiteSelect", "OfficeSelect", "LocationText", "NotesText"
// CLAIMING (Alert) --> CID y, SID n, UID n, INFORM y (previous owner)
func claiming(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)

	// Save any alerts until after AID is assigned
	var alerts []db.Alert

	// Self assigned, no need to ack the computer
	act.Cid_ack = curUid
	assignUser(curUid, 0, 0, &act)

	// Assigning device to current user. No user select list is shown, so wiz.Uid is 0
	wiz.Uid = curUid

	// Assign the new owner to the device
	note, devUid := updateDevice(curUid, wiz)
	act.Notes = note

	// Fetch previous owner record and update color
	if devUid > 0 && devUid != curUid {
		alerts = append(alerts, createAlert(devUid, " from: ", &act))

		// Close the action, nobody needs notification, no work to be done
	} else {
		act.Active = 0
	}
	act.Notes += ". " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
	saveSeenDevice(act.Cid)
}

// User is saying they used this computer, not taking ownership (claim) it
// "DeviceSelect", "NotesText"
// USING --> CID y, SID n, UID n, INFORM n
func using(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)

	// No followup required, close on open
	act.Active = 0
	assignUser(curUid, 0, 0, &act)

	// Not reassigning device
	wiz.Uid = 0
	act.Notes, _ = updateDevice(curUid, wiz)
	act.Notes += ". " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveSeenDevice(act.Cid)
}

// Gave (assigning) device to a user
// "DeviceSelect", "SiteSelect", "OfficeSelect", "LocationText", "GroupSelect", "UserSelect", "NotesText"
// GIVING (Alert)   --> CID y, SID n, UID n, INFORM y (new owner)
func giving(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)
	var alerts []db.Alert

	assignUser(curUid, 0, 0, &act)

	// wiz.Uid is becoming the new owner
	note, devUid := updateDevice(curUid, wiz)
	act.Notes = note

	// The person being given the computer needs to be informed
	if wiz.Uid > 0 && wiz.Uid != curUid && wiz.Uid != devUid {
		alerts = append(alerts, createAlert(wiz.Uid, " to: ", &act))
	} else if wiz.Uid == 0 {
		act.Notes += " to unassigned "
	}
	act.Notes += ". " + wiz.Notes

	// Set the profile color if taking it from somebody else
	if devUid > 0 && devUid != curUid && devUid != wiz.Uid {
		alerts = append(alerts, createAlert(devUid, " The device assigned to: ", &act))
	}
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
	saveSeenDevice(act.Cid)
}

// User is saying a device has broken
// "DeviceSelect", "SiteSelect", "OfficeSelect", "LocationText", "GroupSelect", "UserSelect", "NotesText", "ImpactSelect"
func broken(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)
	assignedFullName, _ := assignUser(curUid, wiz.Uid, wiz.Gid, &act)
	var alerts []db.Alert

	// Stop from assigning device to this user
	wiz.Uid = 0
	wiz.Gid = 0

	// Get device info, and update location info if any
	note, devUid := updateDevice(curUid, wiz)
	act.Notes = note
	act.Notes += " with impact: " + db.GetCodeDescription("IMPACT", wiz.Impact)
	act.Notes += " tasked to " + assignedFullName

	// The current owner needs to be informed it is being worked on
	if devUid > 0 {
		alerts = append(alerts, createAlert(devUid, " used by ", &act))
	} else {
		alerts = append(alerts, createAlert(curUid, " reported by ", &act))
	}
	act.Notes += ". " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
	saveSeenDevice(act.Cid)
}

// Lost device
// "DeviceSelect", "GroupSelect", "UserSelect", "NotesText", "ImpactSelect"
// LOST (Ticket) --> CID y, SID n, UID y (who will fix), INFORM Y (current owner) -> NEEDS WORK LOG TO BE CLOSED!
func lost(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)
	fullname, _ := assignUser(curUid, wiz.Uid, wiz.Gid, &act)
	var alerts []db.Alert

	// Do Not assign device to this user
	wiz.Uid = 0
	wiz.Gid = 0

	// Find current device owner
	note, devUid := updateDevice(curUid, wiz)
	act.Notes = note
	act.Notes += " with impact: " + db.GetCodeDescription("IMPACT", wiz.Impact)
	act.Notes += " tasked to " + fullname

	// The current owner needs to be informed it is lost
	if devUid > 0 && devUid != curUid {
		alerts = append(alerts, createAlert(devUid, " used by ", &act))
	}
	act.Notes += ". " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
}

// The Device died/dead/kaput/beyond repair
// "DeviceSelect", "GroupSelect", "UserSelect", "NotesText", "ImpactSelect"
// DIED (Ticket) --> CID y, SID n, UID y (who will fix), INFORM Y (current owner) -> NEEDS WORK LOG TO BE CLOSED!
func died(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)
	fullname, _ := assignUser(curUid, wiz.Uid, wiz.Gid, &act)
	var alerts []db.Alert

	// Do Not assign device to this user
	wiz.Uid = 0
	wiz.Gid = 0

	// Find current device owner
	note, devUid := updateDevice(curUid, wiz)
	act.Notes = note
	act.Notes += " with impact: " + db.GetCodeDescription("IMPACT", wiz.Impact)
	act.Notes += " tasked to " + fullname

	// The current owner needs to be informed it died
	if devUid > 0 && devUid != curUid {
		alerts = append(alerts, createAlert(devUid, " used by ", &act))
	}
	act.Notes += ". " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
	saveSeenDevice(act.Cid)
}

// Device was backed up
// DeviceSelect", "NotesText"
// BACKUP (Alert)   --> CID y, SID n, UID n, INFORM n
func backup(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)
	act.Active = 0 // Task needs no follow-up
	assignUser(curUid, 0, 0, &act)
	wiz.Uid = 0
	wiz.Gid = 0
	act.Cid_ack = curUid
	act.Notes, _ = updateDevice(curUid, wiz)
	act.Notes += ". Saved on " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
}

// Installed software package
// "DeviceSelect", "SoftwareSelect", "DateChoose", "NotesText"
// INSTALL (Alert) --> CID y, SID y, UID n, INFORM y (device current owner)
func install(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)
	var alerts []db.Alert

	act.OpenedInt = wiz.Installed // When installed, in seconds since 1970 GMT
	if wiz.Sid > 0 {
		software, _ := db.GetSoftware(curUid, wiz.Sid)
		act.Notes = "<a href='software.html?sid=" + strconv.Itoa(act.Sid) + "' >" + software.Name + "</a> was "
	}
	wiz.Uid = 0
	wiz.Gid = 0
	note, devUid := updateDevice(curUid, wiz)
	act.Notes += note

	// Notify the current owner if not the same as the person installing
	if devUid > 0 && devUid != curUid {
		alerts = append(alerts, createAlert(devUid, " used by ", &act))

		//If no previous owner, close the alert
	} else {
		act.Active = 0
		act.Sid_ack = curUid
		act.Cid_ack = curUid
	}
	act.Notes += ". " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
}

// Remove software so license can be reused maybe
// "DeviceSelect", "SoftwareSelect", "DateChoose", "NotesText"
// REMOVE (Alert)  --> CID y, SID y, UID n, INFORM y (device current owner)
func remove(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)
	var alerts []db.Alert

	act.OpenedInt = wiz.Installed
	if wiz.Sid > 0 {
		software, _ := db.GetSoftware(curUid, wiz.Sid)
		act.Notes = "<a href='software.html?sid=" + strconv.Itoa(act.Sid) + "' >" + software.Name + "</a> was "
	}
	wiz.Uid = 0
	wiz.Gid = 0
	note, devUid := updateDevice(curUid, wiz)
	act.Notes += note

	// If not the current owner, inform them
	if devUid > 0 && devUid != curUid {
		alerts = append(alerts, createAlert(devUid, " used by ", &act))
	} else { // Close the alert
		act.Active = 0
		act.Sid_ack = curUid
		act.Cid_ack = curUid
	}
	act.Notes += ". " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
}

// Request software
// "DeviceSelect", "SoftwareSelect", "GroupSelect", "NotesText"
// REQUEST (Ticket) --> CID y, SID y, UID y, INFORM y (person requesting the software) -> NEEDS WORK LOG TO BE CLOSED!
func request(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)
	var alerts []db.Alert

	if wiz.Sid > 0 {
		software, _ := db.GetSoftware(curUid, wiz.Sid)
		act.Notes = "<a href='software.html?sid=" + strconv.Itoa(act.Sid) + "' >" + software.Name + "</a> "
	}
	assignedFullName, _ := assignUser(curUid, wiz.Uid, wiz.Gid, &act)
	wiz.Uid = 0
	wiz.Gid = 0
	note, _ := updateDevice(curUid, wiz)
	act.Notes += " software was requested for " + note + ", approval by " + assignedFullName

	alerts = append(alerts, createAlert(curUid, " reported by", &act))

	act.Notes += ". " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
}

// Attention/Care is needed on a computer to deal with an issue
// "DeviceSelect", "SiteSelect", "OfficeSelect", "LocationText", "GroupSelect" "UserSelect", "NotesText", "ImpactSelect"
// CARE (Ticket) --> CID y, SID n, UID y (who will fix), INFORM Y (current owner) -> NEEDS WORK LOG TO BE CLOSED!
func care(curUid int, wiz *wizFormData) {
	act := newAction(curUid, wiz)
	var alerts []db.Alert

	fullName, _ := assignUser(curUid, wiz.Uid, wiz.Gid, &act)
	wiz.Uid = 0
	wiz.Gid = 0
	note, devUid := updateDevice(curUid, wiz)
	act.Notes += note
	act.Notes += " with impact: " + db.GetCodeDescription("IMPACT", wiz.Impact)
	act.Notes += " tasked to " + fullName

	if devUid > 0 && devUid != curUid {
		alerts = append(alerts, createAlert(devUid, " used by", &act))
	} else {
		alerts = append(alerts, createAlert(curUid, " reported by", &act))
	}
	act.Notes += ". " + wiz.Notes
	err := act.AddAction(curUid)
	if err != nil {
		log.Println(err)
	}
	saveAlerts(curUid, act.Aid, act.Active, alerts)
}
