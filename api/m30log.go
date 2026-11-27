package api

import (
	"log"
	"os"
	"strconv"

	"github.com/gbsto/daisy/m30"
	"github.com/gofiber/fiber/v2"
)

// Store where the log summaries are kept
func PostM30Api(c *fiber.Ctx) error {

	msg := new(m30.M30msgs) // Allocate on heap (address of)
	if err := c.BodyParser(msg); err != nil {
		log.Println("API: parser ", err)
		return c.Status(fiber.StatusBadRequest).SendString("Server error parsing message.")
	}

	// Validate the API key TODO: move to .env or use the master key
	if msg.ApiKey != os.Getenv("ISAW_KEY") {
		log.Println("API: Invalid API Key")
		return c.Status(fiber.StatusBadRequest).SendString("Bad API key")
	}

	switch msg.Cmd {

	case "HWM":
		i, err := msg.GetM30Hwm()
		if err != nil {
			log.Println("API: Failure getting taps high-water marks", err)
			return c.Status(fiber.StatusBadRequest).SendString("Error getting high-water marks")
		}
		return c.Status(fiber.StatusOK).SendString(strconv.Itoa(i))

	case "SAVE":
		err := msg.SaveM30()
		if err != nil {
			log.Println("API: Failure adding taps", err)
			return c.Status(fiber.StatusBadRequest).SendString("Error saving records")
		}

	}

	return c.Status(fiber.StatusOK).SendString("OKAY")

}
