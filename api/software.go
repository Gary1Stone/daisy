package api

import (
	"database/sql"
	"log"

	"github.com/gbsto/daisy/db"
	"github.com/gofiber/fiber/v2"
)

func PostSoftwareApi(c *fiber.Ctx) error {
	sysInfo := new(db.SysInfo) // Allocate on heap (address of)
	resp := "CRITICAL SERVER ERROR!"
	if err := c.BodyParser(sysInfo); err != nil {
		log.Println("API: parser ", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Validate the API key
	if sysInfo.ApiKey != "3278328732ty8y80987301ye689621emncnkj" {
		log.Println("API: Invalid API Key")
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Check the hostname of the computer is in the database
	// If this computer is not registered, automatically add it
	cid, err := db.GetCidByName(sysInfo.Hostname)
	if err == sql.ErrNoRows {
		cid, err = db.ApiNewComputer(sysInfo)
		if err != nil {
			log.Println("API: hostname not registered")
			return c.Status(fiber.StatusOK).SendString(resp)
		}
	} else if err != nil {
		log.Println("API: hostname not registered")
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Save the device info and track the location
	err = db.UpdateSysInfo(cid, sysInfo)
	if err != nil {
		log.Println("API: Failure adding device information", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Ensure the software list has data in it
	if len(sysInfo.SoftwareList) == 0 {
		log.Println("API: Empty software list")
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Save the software list to the inventory table
	// After adding the Operating System name to it
	// Simplifies queries later on
	sysInfo.SoftwareList = append(sysInfo.SoftwareList, sysInfo.OS_Name)
	err = db.AddSoftwareList(cid, sysInfo)
	if err != nil {
		log.Println("API: Failure adding software", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Add the mac addresses
	err = db.SetMacAddress(cid, sysInfo)
	if err != nil {
		log.Println("API: Failure adding mac address information", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Add the Drive Information
	err = db.SetDiskInfo(cid, sysInfo.Disk_Info)
	if err != nil {
		log.Println("API: Failure adding drive information", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	return c.Status(fiber.StatusOK).SendString("Okay")
}
