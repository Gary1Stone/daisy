package cmd

import (
	"html/template"
	"log"
	"math"
	"strconv"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"

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

	//get profile record
	profile, err := db.GetProfile(user.Uid, uid)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}

	//Last Updated by name
	lun := ""
	if len(profile.Last_updated_time) > 0 {
		lun = "<p title='Last Updated'>Last Updated by: "
		lun += profile.Lun
		lun += " at "
		lun += profile.Last_updated_time
		lun += "</p>"
	}

	//When the record was created by an external registration process
	//the inital color is amber, and the lastUpdatedName is system.
	//We Resets the color to light grey when the user saves the record
	if profile.Color == "alert" && profile.Last_updated_by == db.SYS_PROFILE.Uid {
		profile.Color = colors.Light
	}

	//Create the delete button
	deleteButton := ""
	if profile.Deleteable {
		deleteButton = ctrls.MakeDeleteButton(user.Permissions.Profile.Delete)
	}

	ipBanned := ""
	//Check if user was banned by their IP address
	if db.CheckUsersLastIpBanned(profile.Uid) {
		ipBanned = "<span id='bttn'><button class='button alert' onclick='resetBanned(" + strconv.Itoa(profile.Uid) + ");'>Reset Banned</button></span>"
	}
	//WARNING - a fiber.Map{ cannot handle emtpy strings.
	return c.Render("profile", fiber.Map{
		"title":            template.HTML("<span class='mif-user icon'></span>&nbsp;Profile"),
		"fullName":         user.Fullname,
		"isAdmin":          user.IsAdmin,
		"isReadonly":       !user.Permissions.Profile.Update,
		"isDisabled":       !user.Permissions.Profile.Update,
		"uid":              profile.Uid,
		"user":             profile.User,
		"first":            profile.First,
		"last":             profile.Last,
		"pwd_reset":        profile.Pwd_reset,
		"colour":           profile.Color,
		"GroupOptions":     template.HTML(ctrls.BuildDropList("GROUP", strconv.Itoa(profile.Gid), "", false, false)),
		"FenceOptions":     template.HTML(ctrls.BuildDropList("GEOFENCE", profile.Geo_fence, "", true, false)),
		"RadiusOptions":    template.HTML(ctrls.BuildRadiusOptions(profile.Geo_radius)),
		"lastUpdated":      template.HTML(lun),
		"cmd_one":          template.HTML(ctrls.MakeSaveButton(user.Permissions.Profile.Update)),
		"cmd_two":          template.HTML(ctrls.MakeAddButton(user.Permissions.Profile.Create)),
		"cmd_three":        template.HTML(deleteButton),
		"curUid":           user.Uid,
		"assigned_devices": template.HTML(ctrls.BuildAssignedDevices(user.Uid, uid)),
		"banned":           template.HTML(ipBanned),
		"chkActive":        template.HTML(ctrls.BuildActiveCheckbox(profile.Active, !user.Permissions.Profile.Update)),
		"chkNotify":        template.HTML(ctrls.BuildNotifyCheckbox(profile.Notify, !user.Permissions.Profile.Update)),
		"userAlerts":       template.HTML(ctrls.GetAlertTable(uid)),
		"loginHistory":     template.HTML(ctrls.GetProfileLogins(user.Uid, uid)),
	})
}
