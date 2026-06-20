package cmd

import (
	"html/template"
	"log"
	"math"
	"strconv"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/svg"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func GetDevice(c *fiber.Ctx) error {
	//	start := time.Now()

	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// Read the cid (computer ID) from the URL, or default to 0
	cid, err2 := strconv.Atoi(c.Query("cid", "0"))
	if err2 != nil || cid == 0 {
		cid = math.MaxInt
	}

	// If NO Read capababilty, send them home
	if !user.Permissions.Device.Read {
		return c.Status(fiber.StatusOK).Redirect("home.html")
	}

	// Get device record
	device, err := db.GetDevice(user.Uid, cid)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}

	// Guess at the default for the device's assigned Group
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

	//	log.Printf("GetDevice took %s", time.Since(start))
	// Render the page
	return c.Render("device", addNavigationIcons(fiber.Map{
		"title":             template.HTML(svg.GetIcon("devices") + " Devices"),
		"fullName":          user.Fullname,
		"isAdmin":           user.IsAdmin,
		"cmd_one":           template.HTML(ctrls.MakeButton(ctrls.BtnSave, user.Permissions.Device.Update)),
		"cmd_two":           template.HTML(ctrls.MakeButton(ctrls.BtnNew, user.Permissions.Device.Create)),
		"cmd_three":         template.HTML(ctrls.MakeButton(ctrls.BtnDelete, deleteAble)),
		"imageCtrl":         template.HTML(ctrls.MakeImageCtrl(&device, user.Permissions.Device.Update)),
		"isReadonly":        !user.Permissions.Device.Update,
		"isDisabled":        !user.Permissions.Device.Update,
		"name":              device.Name,
		"model":             device.Model,
		"active":            device.Active,
		"ethernet":          device.Ethernet,
		"wifi":              device.Wifi,
		"usb":               device.Usb,
		"cd":                device.Cd,
		"serial_number":     device.Serial_number,
		"typeCtrl":          template.HTML(ctrls.BuildDropList("TYPE", device.Type, "", true, false)),
		"makeCtrl":          template.HTML(ctrls.BuildDropList("MAKE", device.Make, "", true, false)),
		"statusCtrl":        template.HTML(ctrls.BuildDropList("STATUS", device.Status, "", false, false)),
		"groupCtrl":         template.HTML(ctrls.BuildDropList("GROUP", groupAssigned, "", true, false)),
		"osCtrl":            template.HTML(ctrls.BuildDropList("OS", device.Os, "", true, false)),
		"siteCtrl":          template.HTML(ctrls.BuildDropList("SITE", device.Site, "", false, false)),
		"coresCtrl":         template.HTML(ctrls.BuildDropList("CORES", strconv.Itoa(device.Cores), "", true, false)),
		"driveTypeCtrl":     template.HTML(ctrls.BuildDropList("DRIVETYPE", device.Drivetype, "", false, false)),
		"years":             template.HTML(ctrls.BuildYearsSelect(device.Year)),
		"uidCtrl":           template.HTML(ctrls.BuildDropList("USER", strconv.Itoa(device.Uid), strconv.Itoa(device.Gid), true, !user.Permissions.Device.Update)),
		"officeCtrl":        template.HTML(ctrls.BuildDropList("OFFICE", device.Office, device.Site, true, !user.Permissions.Device.Update)),
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
	}))
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

func PostDevice(c *fiber.Ctx) error {
	// Read incoming requst cookie to get curUid
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	reply := struct {
		Success bool   `json:"success"`
		Cid     int    `json:"cid"`
		Msg     string `json:"msg"`
	}{
		Success: false,
		Cid:     0,
		Msg:     "ERROR: Processing Error",
	}

	recvd := new(db.Device) // The new() allocates HEAP to create the variable/struct, therefore must use address operator(*) in functions
	if err := c.BodyParser(recvd); err != nil {
		return c.Status(fiber.StatusOK).JSON(reply)
	}

	if !user.Permissions.Device.Read {
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to view device records."
		return c.Status(fiber.StatusOK).JSON(reply)
	}

	//Do processing and saves
	switch recvd.Task {
	case "get_asset_id":
		var err error
		reply.Cid = recvd.Cid
		reply.Success = true
		reply.Msg, err = db.GetNextAssetId(recvd.Cid, recvd.Type)
		if err != nil {
			reply.Success = false
			reply.Msg = err.Error()
		}
	case "unique":
		reply.Success = db.IsUniqueDevice(recvd.Cid, recvd.Name)
		reply.Cid = recvd.Cid
		reply.Msg = "The name " + recvd.Name + " is available."
		if !reply.Success {
			reply.Msg = "ERROR: The name " + recvd.Name + " is already used."
		}
	case "get_office_control":
		return c.Status(fiber.StatusOK).SendString(ctrls.BuildDropList("OFFICE", recvd.Office, recvd.Site, true, !user.Permissions.Device.Update))
	case "get_person_control":
		return c.Status(fiber.StatusOK).SendString(ctrls.BuildDropList("USER", strconv.Itoa(recvd.Uid), strconv.Itoa(recvd.Gid), true, !user.Permissions.Device.Update))
	case "getactionlog":
		if recvd.Aid > 0 { //set_accept_cid_by_aid
			//	db.AckAlert(curUid, 0, recvd.Aid, true, true, false)
			db.AckAction(user.Uid, recvd.Aid, true, true, false)
		}
		return c.Status(fiber.StatusOK).SendString(ctrls.BuildDeviceLog(user.Uid, recvd.Cid, recvd.ShowHistory))
	case "delete":
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to delete device records."
		if user.Permissions.Device.Delete {
			reply.Success = db.MarkDeviceAsDeleted(user.Uid, recvd.Cid)
			reply.Cid = recvd.Cid
			reply.Msg = "The device " + recvd.Name + " was deleted."
			if !reply.Success {
				reply.Msg = "ERROR: The device " + recvd.Name + " was NOT deleted."
			}
		}
	case "save":
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to update device records."
		if user.Permissions.Device.Update {
			dto, err := db.GetDevice(user.Uid, recvd.Cid)
			if err != nil {
				log.Println(err)
			} else {
				//Add action log entry for newly assigned person
				if dto.Uid != recvd.Uid {
					var wiz wizFormData
					wiz.Cid = recvd.Cid
					wiz.Gid = recvd.Gid
					wiz.Uid = recvd.Uid
					wiz.Site = recvd.Site
					wiz.Office = recvd.Office
					wiz.Location = recvd.Location
					wiz.Sid = 0
					wiz.Notes = ""
					wiz.Installed = 0
					wiz.Impact = 0
					// If the current user is assigning the device to themselves:
					if recvd.Uid == user.Uid {
						wiz.Task = "CLAIMING"
						claiming(user.Uid, &wiz)
					} else {
						wiz.Task = "GIVING"
						giving(user.Uid, &wiz)
					}
				}
				//Update and save record
				dto.Name = recvd.Name
				dto.Type = recvd.Type
				dto.Site = recvd.Site
				dto.Office = recvd.Office
				dto.Location = recvd.Location
				dto.Year = recvd.Year
				dto.Make = recvd.Make
				dto.Model = recvd.Model
				dto.Cpu = recvd.Cpu
				dto.Cores = recvd.Cores
				dto.Ram = recvd.Ram
				dto.Drivetype = recvd.Drivetype
				dto.Drivesize = recvd.Drivesize
				dto.Notes = recvd.Notes
				dto.Gpu = recvd.Gpu
				dto.Cd = recvd.Cd
				dto.Wifi = recvd.Wifi
				dto.Ethernet = recvd.Ethernet
				dto.Usb = recvd.Usb
				dto.Active = recvd.Active
				dto.Last_updated_by = user.Uid
				dto.Image = recvd.Image
				dto.Color = recvd.Color
				dto.Speed = recvd.Speed
				dto.Uid = recvd.Uid
				dto.Status = recvd.Status
				dto.Os = recvd.Os
				dto.Serial_number = recvd.Serial_number
				dto.Gid = recvd.Gid
				reply.Success = db.SetDevice(user.Uid, &dto)
				reply.Cid = dto.Cid
				reply.Msg = "The device record was saved"
				if !reply.Success {
					reply.Msg = "ERROR: The device record was NOT saved"
				}
			}
		}
	case "add":
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to create devices."
		if user.Permissions.Device.Create {
			recvd.Active = 1
			recvd.Last_updated_by = user.Uid
			recvd.Color = colors.Light
			reply.Success = db.AddDevice(recvd)
			reply.Cid = recvd.Cid
			reply.Msg = "The device record was created"
			if !reply.Success {
				reply.Msg = "ERROR: The device record was NOT created"
			}
		}
	}
	return c.Status(fiber.StatusOK).JSON(reply)
}
