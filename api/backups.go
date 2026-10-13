package api

import (
	"log"

	"github.com/gbsto/daisy/backups"
	"github.com/gofiber/fiber/v2"
)

type Backs struct {
	ApiKey  string               `json:"api_key"`
	Source  string               `json:"source"`
	Backups []backups.BackupInfo `json:"backups"`
}

// Store where the backups are kept
func PostBackupsApi(c *fiber.Ctx) error {
	back := new(Backs) // Allocate on heap (address of)
	resp := "CRITICAL SERVER ERROR!"
	if err := c.BodyParser(back); err != nil {
		log.Println("API: parser ", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Validate the API key
	if back.ApiKey != "Barbara_Anne_Little_Stone" {
		log.Println("API: Invalid API Key")
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Ensure the list of backups has data in it
	if len(back.Backups) == 0 {
		log.Println("API: Empty backup list")
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Save the backup list to the backups table
	var b backups.BackupInfo
	err := b.SaveBackups(back.Backups, back.Source)
	if err != nil {
		log.Println("API: Failure adding software", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	return c.Status(fiber.StatusOK).SendString("Okay")
}
