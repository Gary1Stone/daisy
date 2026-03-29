package cmd

import (
	"encoding/json"
	"log"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func PostAdmin(c *fiber.Ctx) error {
	adminForm := struct {
		Task      string `json:"task"`
		Field     string `json:"field"`
		AdminData string `json:"adminData"`
	}{}

	//Get the data sent from the browser
	if err := c.BodyParser(&adminForm); err != nil {
		return c.Status(fiber.StatusOK).SendString("ERROR: ")
	}

	// Read incoming requst cookie to get curUid
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	if !user.Permissions.Admin.Read {
		return c.Status(fiber.StatusOK).SendString("ERROR: You do not have permission to view admin records.")
	}

	//If the user wanted to save anything
	if adminForm.Task == "save_table" {
		var items []db.ChoiceAdmin
		err := json.Unmarshal([]byte(adminForm.AdminData), &items)
		if err != nil {
			log.Println(err)
		}

		//Checking the user permissions Create(add), Delete, Update
		for _, item := range items {
			if (item.Add && !item.Delete) && user.Permissions.Admin.Create {
				if item.AddRecord() != nil {
					return c.Status(fiber.StatusOK).SendString("ERROR: ")
				}
			} else if item.Delete && user.Permissions.Admin.Delete {
				if item.DeleteRecord() != nil {
					return c.Status(fiber.StatusOK).SendString("ERROR: ")
				}
			} else if item.Update && user.Permissions.Admin.Update {
				if item.UpdateRecord() != nil {
					return c.Status(fiber.StatusOK).SendString("ERROR: ")
				}
			}
		}
	}

	//Always send back the selected admin data from the choices table
	items := db.GetChoicesAdmin(adminForm.Field)
	return c.Status(fiber.StatusOK).JSON(items)
}
