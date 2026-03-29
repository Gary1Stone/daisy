package cmd

import (
	"strconv"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

type droplistData struct {
	Task       string `json:"task"`
	IsTicket   bool   `json:"isTicket"`
	IsWizard   bool   `json:"isWizard"`
	Cid        int    `json:"cid"`
	Gid        int    `json:"gid"`
	Uid        int    `json:"uid"`
	Site       string `json:"site"`
	Office     string `json:"office"`
	Impact     int    `json:"impact"`
	Trouble    int    `json:"trouble"`
	Wizard     string `json:"wizard"`
	Type       string `json:"type"`
	Inform_gid int    `json:"inform_gid"`
	IsReadonly bool   `json:"isReadonly"`
}

func PostDroplist(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}
	recvd := new(droplistData)
	reply := "Okay"
	if err := c.BodyParser(recvd); err != nil {
		return c.Status(fiber.StatusOK).SendString("Server Error")
	}

	switch recvd.Task {

	// case "get_site_control_wizard":
	case "SITE":
		if recvd.IsWizard {
			// If we have a computer, use the computer's assigned site
			if recvd.Cid > 0 {
				site, _ := db.GetDeviceAssignedSiteOffice(user.Uid, recvd.Cid)
				if len(site) > 0 {
					recvd.Site = site
				}
			}
		}
		reply = ctrls.BuildDropList("SITE", recvd.Site, "", false, false)

	// case "get_office_control_wizard":
	// case "get_office_control_device":
	case "OFFICE":
		// If the computer is in the selected site (recvd.Site), set the default office
		if recvd.IsWizard {
			recvd.IsReadonly = false
			if recvd.Cid > 0 {
				site, office := db.GetDeviceAssignedSiteOffice(user.Uid, recvd.Cid)
				if recvd.Site == site {
					recvd.Office = office
				}
			}
		}
		reply = ctrls.BuildDropList("OFFICE", recvd.Office, recvd.Site, true, recvd.IsReadonly)

	//case "get_group_control_wizard":
	case "GROUP":
		withBlank := true
		if recvd.IsWizard {
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
		}
		reply = ctrls.BuildDropList("GROUP", strconv.Itoa(recvd.Gid), "", withBlank, false)

	// case "get_person_control_device":
	// case "get_person_control_ticket":
	// case "get_person_control_wizard":
	case "USER":
		withBlank := true
		if recvd.IsWizard {
			if recvd.Wizard == "claim" {
				recvd.Uid = user.Uid
			}
			if recvd.Wizard == "broken" || recvd.Wizard == "care" || recvd.Wizard == "request" || recvd.Wizard == "claim" {
				withBlank = false // Broken and care needs someone to look into it.
			}
		}
		if recvd.IsTicket || recvd.IsWizard {
			recvd.IsReadonly = false
			// If there is a group, and no user, but a device, lookup that device's assigned user (if any)
			if recvd.Gid > 0 && recvd.Uid == 0 && recvd.Cid > 0 {
				recvd.Uid, _ = db.GetDeviceAssignedUserGroup(user.Uid, recvd.Cid)
			}
		}
		reply = ctrls.BuildDropList("USER", strconv.Itoa(recvd.Uid), strconv.Itoa(recvd.Gid), withBlank, recvd.IsReadonly)

	case "IMPACT":
		reply = ctrls.BuildDropList("IMPACT", "-1", "", false, false)

	case "TROUBLE":
		reply = ctrls.BuildDropList("TROUBLE", "-1", recvd.Type, false, false)

	case "OFFICESEARCH":
		if len(recvd.Wizard) > 0 {
			recvd.Office = ""
		}
		reply = ctrls.BuildDropList("OFFICESEARCH", recvd.Office, recvd.Site, true, false)

	// case "get_user_search_control_wizard":
	// case "get_user_control_devices":
	case "USERSEARCH":
		reply = ctrls.BuildDropList("USERSEARCH", strconv.Itoa(recvd.Uid), strconv.Itoa(recvd.Gid), true, false)

	// case "get_inform_person_control_ticket":
	case "USERINFORM":
		if recvd.IsTicket {
			// If there is a group, and no user, but a device, lookup that device's assigned user (if any)
			if recvd.Gid > 0 && recvd.Uid == 0 && recvd.Cid > 0 {
				recvd.Uid, _ = db.GetDeviceAssignedUserGroup(user.Uid, recvd.Cid)
			}
		}
		reply = ctrls.BuildDropList("USERINFORM", strconv.Itoa(recvd.Uid), strconv.Itoa(recvd.Inform_gid), true, false)

	}
	return c.Status(fiber.StatusOK).SendString(reply)
}
