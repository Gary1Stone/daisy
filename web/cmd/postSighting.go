package cmd

import (
	"log"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

type Sightings struct {
	Id        int    `json:"id"`        // Primary key, autonumber
	Email     string `json:"email"`     // Who saw it
	Eid       int    `json:"eid"`       // Who saw it, if they have a uid
	Timestamp int64  `json:"timestamp"` // When the sighting occured
	Cid       int    `json:"cid"`       // Computer they saw
	Site      string `json:"site"`      // What site the device was at
	Office    string `json:"office"`    // What office the device was in
	Location  string `json:"location"`  // Where the device was (in drawer, on desk)
	Gid       int    `json:"gid"`       // If a group had ownership
	Uid       int    `json:"uid"`       // If a user had ownership
	Comment   string `json:"comment"`   // Comment such as: a student has it, or is it broken
	Sent      bool   `json:"sent"`      // Was this sent to the daisy server
	Api_key   string `json:"api_key"`   // placeholder for the API Key to send to the daisy server
}

func PostSightingApi(c *fiber.Ctx) error {
	sighting := new(Sightings) // Allocate on heap (address of)
	resp := "CRITICAL SERVER ERROR!"
	if err := c.BodyParser(sighting); err != nil {
		log.Println("API: parser ", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Validate the API key
	if sighting.Api_key != "asbhfge78rg2tfb54vn1703tbgvjsyhfv1374tfgb5vjhdf80v43nv" {
		log.Println("API: Invalid API Key")
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Add the sighting record
	err := sighting.AddRecord()
	if err != nil {
		log.Println("API: Failure adding sighting", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}
	return c.Status(fiber.StatusOK).SendString("Okay")
}

func (s *Sightings) AddRecord() error {
	if s.Eid < 1 {
		s.Eid = db.SYS_PROFILE.Uid
	}
	recvd := new(wizFormData)
	recvd.Task = "SIGHTING"
	recvd.Cid = s.Cid
	recvd.Sid = 0
	recvd.Gid = s.Gid
	recvd.Uid = s.Uid
	recvd.Site = s.Site
	recvd.Office = s.Office
	recvd.Location = s.Location
	recvd.Notes = s.Comment
	recvd.Installed = 0
	recvd.Impact = 0
	recvd.Trouble = 0
	recvd.Wizard = "sighting" // wizkey
	recvd.Type = "LAPTOP"     // Device Type: DESKTOP, LAPTOP...
	sighting(s.Eid, recvd)
	return nil
}
