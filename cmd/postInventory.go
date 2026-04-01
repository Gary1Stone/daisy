package cmd

import (
	"github.com/gbsto/daisy/ctrls"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

func PostInventory(c *fiber.Ctx) error {
	invForm := struct {
		Task string `json:"task"`
	}{}
	reply := struct {
		Success    bool     `json:"success"`    // true = no errors
		Msg        string   `json:"msg"`        // Error message
		Inv_table  string   `json:"inv_table"`  // HTML table of all the software packages in inventory
		Used_names []string `json:"used_names"` // List of all the in use software names
	}{
		Success:    false,
		Msg:        "Server Error",
		Inv_table:  "",
		Used_names: []string{"", ""},
	}
	if err := c.BodyParser(&invForm); err != nil {
		return c.Status(fiber.StatusOK).JSON(reply)
	}
	if invForm.Task == "get_software_inventory" {
		reply.Inv_table = ctrls.BuildInventoryList()
		reply.Used_names, _ = db.GetUsedInventoryNames()
		reply.Success = true
		reply.Msg = "Okay"
	}
	return c.Status(fiber.StatusOK).JSON(reply)
}
