package cmd

import (
	"fmt"
	"html/template"
	"log"
	"math"
	"strconv"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
	"github.com/gbsto/daisy/util"

	"github.com/gofiber/fiber/v2"
)

func GetProfile(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// Read the uid from the URL, or default to 0
	uid, err2 := strconv.Atoi(c.Query("uid", "0"))
	if err2 != nil || uid == 0 {
		uid = math.MaxInt //Prevent getting all the records (uid=0 means get all records)
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Profile.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	// get profile record
	profile, err := db.GetProfile(user.Uid, uid)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}

	// Last Updated by name
	lun := ""
	if len(profile.Last_updated_time) > 0 {
		lun = "<p title='Last Updated'>Last Updated by: "
		lun += profile.Lun
		lun += " at "
		lun += profile.Last_updated_time
		lun += "</p>"
	}

	// When the record was created by an external registration process
	// the inital color is amber, and the lastUpdatedName is system.
	// We Resets the color to light grey when the user saves the record
	if profile.Color == "alert" && profile.Last_updated_by == db.SYS_PROFILE.Uid {
		profile.Color = colors.Light
	}

	// Can the record be deleted?
	isDeleteable := false
	if profile.Deleteable && user.Permissions.Profile.Delete {
		isDeleteable = true
	}

	ipBanned := ""
	// Check if user was banned by their IP address
	if db.CheckUsersLastIpBanned(profile.Uid) {
		ipBanned = "<span id='bttn'><button type='button' class='button alert' onclick='resetBanned(" + strconv.Itoa(profile.Uid) + ");'>Reset Banned</button></span>"
	}

	return c.Render("profile", addNavigationIcons(fiber.Map{
		"title":            template.HTML(svg.GetIcon("profiles") + " Profile"),
		"fullName":         user.Fullname,
		"isAdmin":          user.IsAdmin,
		"isReadonly":       !user.Permissions.Profile.Update,
		"isDisabled":       !user.Permissions.Profile.Update,
		"uid":              profile.Uid,
		"userid":           profile.User,
		"first":            profile.First,
		"last":             profile.Last,
		"pwd_reset":        profile.Pwd_reset,
		"colour":           profile.Color,
		"GroupOptions":     template.HTML(ctrls.BuildDropList("GROUP", strconv.Itoa(profile.Gid), "", false, false)),
		"FenceOptions":     template.HTML(ctrls.BuildDropList("GEOFENCE", profile.Geo_fence, "", true, false)),
		"RadiusOptions":    template.HTML(ctrls.BuildRadiusOptions(profile.Geo_radius)),
		"lastUpdated":      template.HTML(lun),
		"save_button":      template.HTML(ctrls.MakeSaveButton(user.Permissions.Profile.Update)),
		"add_button":       template.HTML(ctrls.MakeAddButton(user.Permissions.Profile.Create)),
		"delete_button":    template.HTML(ctrls.MakeDeleteButton(isDeleteable)),
		"cancel_button":    template.HTML(ctrls.MakeCancelButton(user.Permissions.Profile.Read)),
		"curUid":           user.Uid,
		"assigned_devices": template.HTML(ctrls.BuildAssignedDevices(user.Uid, uid)),
		"banned":           template.HTML(ipBanned),
		"chkActive":        template.HTML(ctrls.BuildActiveCheckbox(profile.Active, !user.Permissions.Profile.Update)),
		"chkNotify":        template.HTML(ctrls.BuildNotifyCheckbox(profile.Notify, !user.Permissions.Profile.Update)),
		"userAlerts":       template.HTML(ctrls.GetAlertTable(uid)),
		"loginHistory":     template.HTML(ctrls.GetProfileLogins(user.Uid, uid)),
		"groupIcon":        template.HTML(svg.GetIcon("group")),
		"locationIcon":     template.HTML(svg.GetIcon("location")),
		"bellIcon":         template.HTML(svg.GetIcon("bell")),
	}))
}

func PostProfile(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	recvd := new(db.Profile) // The new() allocates HEAP to create the variable/struct, therefore must use address operator(*) in functions
	reply := struct {
		Success bool   `json:"success"`
		Uid     int    `json:"uid"`
		Msg     string `json:"msg"`
	}{
		Success: false,
		Uid:     0,
		Msg:     "ERROR: Processing Error",
	}

	if err := c.BodyParser(recvd); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).JSON(reply)
	}
	recvd.Fullname = recvd.First + " " + recvd.Last

	// Check user permissions
	var perm util.Permissions
	perm.GetPermissions(fmt.Sprint(c.Locals("permissions")))

	if !perm.Profile.Read {
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to view profile records."
		return c.Status(fiber.StatusOK).JSON(reply)
	}

	// Do processing and saves
	switch recvd.Task {
	case "unique":
		reply.Success = recvd.IsUnique()
		reply.Uid = recvd.Uid
		reply.Msg = "Good. The User ID " + recvd.User + " is available."
		if !reply.Success {
			reply.Msg = "ERROR: The User ID " + recvd.User + " is already used."
		}
	case "delete":
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to delete a profile."
		if perm.Profile.Delete {
			reply.Success = recvd.DeleteRecord(user.Uid)
			reply.Uid = recvd.Uid
			reply.Msg = "The User ID " + recvd.User + " was deleted."
			if !reply.Success {
				reply.Msg = "ERROR: The User " + recvd.User + " was not deleted. They have assigned devices."
			}
		}
	case "save":
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to save a profile."
		if perm.Profile.Update {
			reply.Success = false
			reply.Msg = "ERROR: The profile was NOT saved."
			usr, err := db.GetProfile(user.Uid, recvd.Uid)
			if err == nil {
				usr.User = recvd.User
				usr.First = recvd.First
				usr.Last = recvd.Last
				usr.Fullname = recvd.Fullname
				//If user's Gid changed, cascade groupID in devices and action tables
				if usr.Gid != recvd.Gid {
					recvd.CascadeGidChange()
				}
				usr.Gid = recvd.Gid
				usr.Geo_fence = recvd.Geo_fence
				usr.Geo_radius = recvd.Geo_radius
				usr.Pwd_reset = recvd.Pwd_reset
				usr.Active = recvd.Active
				usr.Notify = recvd.Notify
				reply.Success = true
				if err := usr.UpdateRecord(user.Uid); err != nil {
					reply.Success = false
				}
				reply.Uid = recvd.Uid
				reply.Msg = "The profile was saved."
				if !reply.Success {
					reply.Msg = "ERROR: The profile was NOT saved."
				}
			}
		}
	case "add":
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to create a profile."
		if perm.Profile.Create {
			recvd.Active = 1
			recvd.Last_updated_by = user.Uid
			reply.Success = recvd.AddRecord(user.Uid)
			reply.Uid = recvd.Uid
			reply.Msg = "The profile was created."
			if !reply.Success {
				reply.Msg = "ERROR: The profile was NOT created."
			}
		}
	case "unban":
		reply.Success = true
		reply.Uid = recvd.Uid
		reply.Msg = "Okay"
		err := db.UnBanUser(recvd.Uid)
		if err != nil {
			reply.Success = false
			reply.Msg = err.Error()
		}
	}
	return c.Status(fiber.StatusOK).JSON(reply)
}
