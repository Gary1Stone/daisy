package cmd

import (
	"log"

	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func PostSoftware(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	recvd := new(db.Software)
	reply := struct {
		Success bool   `json:"success"`
		Sid     int    `json:"sid"`
		Msg     string `json:"msg"`
	}{
		Success: false,
		Sid:     0,
		Msg:     "ERROR: Processing Error",
	}
	if err := c.BodyParser(recvd); err != nil {
		return c.Status(fiber.StatusOK).JSON(reply)
	}

	if !user.Permissions.Software.Read {
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to view software records."
		return c.Status(fiber.StatusOK).JSON(reply)
	}

	// Do processing and saves
	switch recvd.Task {
	case "getactionlog":
		if recvd.Aid > 0 {
			db.AckAction(user.Uid, recvd.Aid, true, false, true)
		}
		return c.Status(fiber.StatusOK).SendString(ctrls.BuildSoftwareLog(user.Uid, recvd.Sid, recvd.Showhistory))
	case "unique":
		reply.Success = recvd.IsUnique()
		reply.Sid = recvd.Sid
		reply.Msg = "The software name " + recvd.Name + " is available."
		if !reply.Success {
			reply.Msg = "ERROR: The software name " + recvd.Name + " is already used."
		}
	case "delete":
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to delete a software record."
		if user.Permissions.Software.Delete {
			reply.Success = recvd.DeleteRecord(user.Uid)
			reply.Sid = recvd.Sid
			reply.Msg = "The software package " + recvd.Name + " was deleted."
			if !reply.Success {
				reply.Msg = "ERROR: The software package " + recvd.Name + " was not deleted."
			}
		}
	case "save":
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to update a software record."
		if user.Permissions.Software.Update {
			dto, err := db.GetSoftware(user.Uid, recvd.Sid)
			if err != nil {
				log.Println(err)
			} else {
				fullScanMatchFlag := false
				if dto.Inv_name != recvd.Inv_name {
					fullScanMatchFlag = true
				}
				dto.Name = recvd.Name
				dto.Licenses = recvd.Licenses
				dto.License_key = recvd.License_key
				dto.Product = recvd.Product
				dto.Source = recvd.Source
				dto.Link = recvd.Link
				dto.Notes = recvd.Notes
				dto.Active = recvd.Active
				dto.Reuseable = recvd.Reuseable
				dto.Purchased = recvd.Purchased
				dto.Inv_name = recvd.Inv_name
				dto.Pre_installed = recvd.Pre_installed
				dto.Free = recvd.Free
				dto.Last_updated_by = user.Uid
				reply.Success = dto.UpdateRecord(user.Uid)
				reply.Sid = dto.Sid
				reply.Msg = "The software record was saved"
				if !reply.Success {
					reply.Msg = "ERROR: The software record was NOT saved"
				}
				if fullScanMatchFlag {
					err := db.MatchSoftwareToInventory(dto.Sid)
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	case "add":
		reply.Success = false
		reply.Msg = "ERROR: You do not have permission to create a software record."
		if user.Permissions.Software.Create {
			recvd.Active = 1
			recvd.Last_updated_by = user.Uid
			recvd.Color = "light"
			reply.Success = recvd.AddRecord() //			 db.AddSoftware(recvd)
			reply.Sid = recvd.Sid
			reply.Msg = "The software record was created"
			if !reply.Success {
				reply.Msg = "ERROR: The software record was NOT created"
			}
			if len(recvd.Inv_name) > 0 {
				err := db.MatchSoftwareToInventory(recvd.Sid)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
	return c.Status(fiber.StatusOK).JSON(reply)
}
