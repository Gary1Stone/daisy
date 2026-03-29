package cmd

import (
	"log"
	"strconv"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func PostDevice(c *fiber.Ctx) error {
	// Read incoming requst cookie to get curUid
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// The new() allocates HEAP to create the variable/struct, therefore must use address operator(*) in functions
	recvd := new(db.Device)
	reply := struct {
		Success bool   `json:"success"`
		Cid     int    `json:"cid"`
		Msg     string `json:"msg"`
	}{
		Success: false,
		Cid:     0,
		Msg:     "ERROR: Processing Error",
	}

	if err := c.BodyParser(recvd); err != nil {
		return c.Status(fiber.StatusOK).JSON(reply)
	}

	if !user.Permissions.Device.Read {
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to view device records."
		return c.Status(fiber.StatusOK).JSON(reply)
	}

	isReadonly := true
	if user.Permissions.Device.Update {
		isReadonly = false
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
		return c.Status(fiber.StatusOK).SendString(ctrls.BuildDropList("OFFICE", recvd.Office, recvd.Site, true, isReadonly))
	case "get_person_control":
		return c.Status(fiber.StatusOK).SendString(ctrls.BuildDropList("USER", strconv.Itoa(recvd.Uid), strconv.Itoa(recvd.Gid), true, isReadonly))
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
