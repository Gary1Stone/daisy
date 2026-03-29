package cmd

import (
	"html/template"
	"log"
	"math"
	"strconv"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetDevice(c *fiber.Ctx) error {
	//	start := time.Now()

	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Profile.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	// Prevent getting all the records (cid=0 means get all records)
	cid, err2 := strconv.Atoi(c.Query("cid", "0"))
	if err2 != nil || cid == 0 {
		cid = math.MaxInt
	}

	// CRUD Create Read Update Delete
	// Create and Delete are handled at the control level
	isReadonly := !user.Permissions.Device.Update
	isDisabled := !user.Permissions.Device.Update

	// Get device record
	device, err := db.GetDevice(user.Uid, cid)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}

	// Guess at the default for the assigned Group
	groupAssigned := guessAssignedGroupDefault(user.Uid, &device)
	if device.Gid == 0 {
		device.Gid, _ = strconv.Atoi(groupAssigned)
	}

	// Delete button enabled only if the device has no open action log entries. // maybe future: nor tracked software
	deleteAble := user.Permissions.Device.Delete
	if deleteAble {
		deleteAble = db.IsDeletableDevice(cid)
	}

	// page state: empty (cid=0), with record (cid>0); affects type control, if cid=0 then user must select, with blank included but not selectable

	// log.Printf("GetDevice took %s", time.Since(start))
	// Render the page
	return c.Render("device", fiber.Map{
		"title":             template.HTML("<span class='mif-profile icon'></span>&nbsp;Device"),
		"fullName":          user.Fullname,
		"isAdmin":           user.IsAdmin,
		"isReadonly":        isReadonly,
		"isDisabled":        isDisabled,
		"cmd_one":           template.HTML(ctrls.MakeSaveButton(user.Permissions.Device.Update)),
		"cmd_two":           template.HTML(ctrls.MakeAddButton(user.Permissions.Device.Create)),
		"cmd_three":         template.HTML(ctrls.MakeDeleteButton(deleteAble)),
		"imageCtrl":         template.HTML(ctrls.MakeImageCtrl(&device, user.Permissions.Device.Update)),
		"name":              device.Name,
		"model":             device.Model,
		"active":            device.Active,
		"ethernet":          device.Ethernet,
		"wifi":              device.Wifi,
		"usb":               device.Usb,
		"cd":                device.Cd,
		"serial_number":     device.Serial_number,
		"typeCtrl":          template.HTML(ctrls.BuildDropList("TYPE", device.Type, "", false, false)),
		"makeCtrl":          template.HTML(ctrls.BuildDropList("MAKE", device.Make, "", true, false)),
		"statusCtrl":        template.HTML(ctrls.BuildDropList("STATUS", device.Status, "", false, false)),
		"groupCtrl":         template.HTML(ctrls.BuildDropList("GROUP", groupAssigned, "", true, false)),
		"osCtrl":            template.HTML(ctrls.BuildDropList("OS", device.Os, "", true, false)),
		"siteCtrl":          template.HTML(ctrls.BuildDropList("SITE", device.Site, "", false, false)),
		"coresCtrl":         template.HTML(ctrls.BuildDropList("CORES", strconv.Itoa(device.Cores), "", true, false)),
		"driveTypeCtrl":     template.HTML(ctrls.BuildDropList("DRIVETYPE", device.Drivetype, "", false, false)),
		"years":             template.HTML(ctrls.BuildYearsSelect(device.Year)),
		"uidCtrl":           template.HTML(ctrls.BuildDropList("USER", strconv.Itoa(device.Uid), strconv.Itoa(device.Gid), true, isReadonly)),
		"officeCtrl":        template.HTML(ctrls.BuildDropList("OFFICE", device.Office, device.Site, true, isReadonly)),
		"locationCtrl":      device.Location,
		"ramCtrl":           device.Ram,
		"cpuCtrl":           device.Cpu,
		"drivesizeCtrl":     device.Drivesize,
		"gpuCtrl":           device.Gpu,
		"speed":             device.Speed,
		"cid":               device.Cid,
		"notesCtrl":         device.Notes,
		"last_updated_by":   device.Last_updated_by,
		"last_updated_time": device.Last_updated_time,
		"fullname":          device.Lun,
		"software":          template.HTML(ctrls.BuildSoftwareList(user.Uid, device.Cid)),
		"actionLog":         template.HTML(ctrls.BuildDeviceLog(user.Uid, cid, 0)),
		"backups":           template.HTML(ctrls.BuildBackups(user.Uid, cid)),
	})
}

func guessAssignedGroupDefault(curUid int, dev *db.Device) string {
	if dev.Gid > 0 {
		return strconv.Itoa(dev.Gid)
	}
	if dev.Uid == 0 {
		return "0"
	}
	assignedPerson, err := db.GetProfile(curUid, dev.Uid)
	if err != nil {
		log.Println(err)
		return "0"
	}
	return strconv.Itoa(assignedPerson.Gid)
}
