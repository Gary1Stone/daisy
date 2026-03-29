package cmd

import (
	"fmt"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/util"

	"github.com/gofiber/fiber/v2"
)

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
