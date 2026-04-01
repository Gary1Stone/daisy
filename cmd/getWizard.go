package cmd

import (
	"html/template"
	"log"
	"strconv"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/devices"
	"github.com/gbsto/daisy/web/wizards"

	"github.com/gofiber/fiber/v2"
)

func GetWizard(c *fiber.Ctx) error {

	// Read incoming requst cookie to get curUid
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Device.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	// Read the database to get the user's previous filter settings
	filter, err := db.GetDeviceFilter(user.Uid)
	if err != nil {
		log.Println(err)
	}
	filter.Page = 0 // Reset the page to 0

	wizkey := c.Query("wizkey", "sighting")
	title := ctrls.BuildWizardTitle(wizkey)
	cid, err := strconv.Atoi(c.Query("cid", "0"))
	if err != nil {
		cid = 0
	}
	defaultType := devices.Desktop
	devPic := "images/missing-sm.jpg"
	devDetails := ""
	site := ""
	office := ""

	// Handle software install/remove/request
	if wizkey == wizards.Install || wizkey == wizards.Remove || wizkey == wizards.Request {
		defaultType = "SOFTWARE"
	}

	// Install and Remove needs software package, so no blank option in their select list
	swBlank := true
	if wizkey == wizards.Install || wizkey == wizards.Remove {
		swBlank = false
	}
	// Broken, Died, Care needs an assigned group, so no blank option in their select list
	grpBlank := true
	userGroup := ""
	if wizkey == wizards.Broken || wizkey == wizards.Died || wizkey == wizards.Care || wizkey == wizards.Request {
		grpBlank = false
		userGroup = db.GetDefaultGroup()
	}

	if cid > 0 {
		device, err := db.GetDevice(user.Uid, cid)
		if err != nil {
			cid = 0
		}
		defaultType = device.Type
		devPic = "images/" + device.Small_image
		devDetails = "<p><span class='" + device.Icon + " icon'></span>&nbsp;"
		devDetails += device.Name + "</p><p>"
		devDetails += device.Make_usr + " (" + strconv.Itoa(device.Year) + ")</p><p>"
		if len(device.Model) > 25 {
			device.Model = device.Model[0:25]
		}
		devDetails += device.Model
		site, office = db.GetDeviceAssignedSiteOffice(user.Uid, cid)
	}

	return c.Render("wizard", fiber.Map{
		"title":            template.HTML(title),
		"fullName":         user.Fullname,
		"cmd_one":          template.HTML(ctrls.MakeSearchBtn()),
		"isAdmin":          user.IsAdmin,
		"curUid":           user.Uid,
		"wizKey":           wizkey,
		"cid":              strconv.Itoa(cid),
		"devPic":           devPic,
		"devDetails":       template.HTML(devDetails),
		"typeCtrl":         template.HTML(ctrls.BuildDropList("TYPESEARCH", filter.DevType, wizkey, true, false)),
		"siteSearchCtrl":   template.HTML(ctrls.BuildDropList("SITESEARCH", filter.Site, "", true, false)),
		"officeSearchCtrl": template.HTML(ctrls.BuildDropList("OFFICESEARCH", filter.Office, filter.Site, true, false)),
		"groupSearchCtrl":  template.HTML(ctrls.BuildDropList("GROUPSEARCH", strconv.Itoa(filter.Gid), "", true, false)),
		"userSearchCtrl":   template.HTML(ctrls.BuildDropList("USERSEARCH", strconv.Itoa(filter.Uid), "0", true, false)),
		"softwareCtrl":     template.HTML(ctrls.BuildDropList("SOFTWARE", "", "", swBlank, false)),
		"groupCtrl":        template.HTML(ctrls.BuildDropList("GROUP", "0", "", grpBlank, false)),
		"userCtrl":         template.HTML(ctrls.BuildDropList("USER", "0", userGroup, false, false)),
		"impactCtrl":       template.HTML(ctrls.BuildDropList("IMPACT", "-1", "", false, false)),
		"troubleCtrl":      template.HTML(ctrls.BuildDropList("TROUBLE", "0", defaultType, false, false)),
		"siteCtrl":         template.HTML(ctrls.BuildDropList("SITE", site, "", false, false)),
		"officeCtrl":       template.HTML(ctrls.BuildDropList("OFFICE", office, site, false, false)),
		"locationCtrl":     template.HTML(ctrls.LocationCtrl(user.Uid, cid)),
		"cards":            template.HTML(ctrls.DeviceCards(user.Uid, &filter, true)),
	})
}
