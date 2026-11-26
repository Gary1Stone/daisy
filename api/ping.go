package api

import (
	"log"
	"os"

	"github.com/gbsto/daisy/devices"
	"github.com/gofiber/fiber/v2"
)

func PostPingApi(c *fiber.Ctx) error {
	pingInfo := new(devices.PingInfo) // Allocate on heap (address of)
	resp := "CRITICAL SERVER ERROR!"
	if err := c.BodyParser(pingInfo); err != nil {
		log.Println("API: parser ", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Validate the API key
	if pingInfo.ApiKey != os.Getenv("AUDIT_KEY") {
		log.Println("API: Invalid API Key")
		return c.Status(fiber.StatusOK).SendString(resp)
	}

	// Add the check-in record
	err := pingInfo.AddRecord()
	if err != nil {
		log.Println("API: Failure adding check-in record", err)
		return c.Status(fiber.StatusOK).SendString(resp)
	}
	return c.Status(fiber.StatusOK).SendString("Okay")
}
