package api

import (
	"encoding/json"
	"log"

	"github.com/gbsto/daisy/db"
	"github.com/gofiber/fiber/v2"
)

func PostiSawApi(c *fiber.Ctx) error {
	response := "CRITICAL SERVER ERROR!"

	type ISawInfo struct {
		ApiKey  string `json:"apikey"`
		Command string `json:"command"`
		Data    string `json:"data"`
		Pin     string `json:"pin"`
		Email   string `json:"email"`
	}

	request := new(ISawInfo) // Allocate on heap (address of)
	if err := c.BodyParser(request); err != nil {
		log.Println("API: parser ", err)
		return c.Status(fiber.StatusOK).SendString(response)
	}

	// Validate the API key
	if request.ApiKey != "it was a dark and stormy night" {
		log.Println("API: Invalid API Key")
		return c.Status(fiber.StatusOK).SendString(response)
	}

	switch request.Command {
	case "GET_PROFILES":
		profiles, err := db.GetProfiles()
		if err != nil {
			log.Println("API: Failure getting profiles", err)
			return c.Status(fiber.StatusOK).SendString(response)
		}
		reply, err := json.Marshal(profiles)
		if err != nil {
			log.Println("API: Failure marshalling profiles", err)
			return c.Status(fiber.StatusOK).SendString(response)
		}
		return c.Status(fiber.StatusOK).SendString(string(reply))
	case "GET_COMPUTERS":
		computers, err := db.GetComputers()
		if err != nil {
			log.Println("API: Failure getting profiles", err)
			return c.Status(fiber.StatusOK).SendString(response)
		}
		reply, err := json.Marshal(computers)
		if err != nil {
			log.Println("API: Failure marshalling profiles", err)
			return c.Status(fiber.StatusOK).SendString(response)
		}
		return c.Status(fiber.StatusOK).SendString(string(reply))
	case "GET_CHOICES":
		choices, err := db.GetApiChoices()
		if err != nil {
			log.Println("API: Failure getting choices", err)
			return c.Status(fiber.StatusOK).SendString(response)
		}
		reply, err := json.Marshal(choices)
		if err != nil {
			log.Println("API: Failure marshalling choices", err)
			return c.Status(fiber.StatusOK).SendString(response)
		}
		return c.Status(fiber.StatusOK).SendString(string(reply))
	case "SET_PIN":
		err := db.EmailNewPin(request.Email, request.Pin)
		if err != nil {
			log.Println(err)
			return c.Status(fiber.StatusOK).SendString("Error")
		}
		return c.Status(fiber.StatusOK).SendString("Okay")
	}
	return c.Status(fiber.StatusOK).SendString(response)
}
