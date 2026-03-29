package api

import (
	"encoding/json"
	"log"

	"github.com/gbsto/daisy/db"
	"github.com/gofiber/fiber/v2"
)

func PostOnlineApi(c *fiber.Ctx) error {
	response := "CRITICAL SERVER ERROR!"

	request := new(db.OnlineAPI) // Allocate on heap (address of)
	if err := c.BodyParser(request); err != nil {
		log.Println("API: parser ", err)
		return c.Status(fiber.StatusOK).SendString(response)
	}

	// Validate the API key
	if request.ApiKey != "weird science class Kevin!" {
		log.Println("API: Invalid API Key")
		return c.Status(fiber.StatusOK).SendString(response)
	}

	switch request.Command {
	case "GET_HIGH_WATER_MARKS":
		thedates := db.GetLastUpdated()
		reply, err := json.Marshal(thedates)
		if err != nil {
			log.Println("API: Failure marshalling dates", err)
			return c.Status(fiber.StatusOK).SendString(response)
		}
		return c.Status(fiber.StatusOK).SendString(string(reply)) // Notice that JSON can be sent in string format

	case "GET_NEW_ALIASES":
		newAliases, err := db.GetAliases(true)
		if err != nil {
			log.Println("API: Failure getting aliases", err)
			return c.Status(fiber.StatusOK).SendString(response)
		}
		reply, err := json.Marshal(newAliases)
		if err != nil {
			log.Println("API: Failure marshalling aliases", err)
			return c.Status(fiber.StatusOK).SendString(response)
		}
		return c.Status(fiber.StatusOK).SendString(string(reply)) // Notice that JSON can be sent in string format

	case "SAVE_ONLINE_HISTORY":
		err := db.SaveOnline(request.OnlineInfo)
		if err != nil {
			log.Println("API: Failure saving online history", err)
			return c.Status(fiber.StatusOK).SendString(err.Error())
		}
		site, err := db.SaveScanInfo(request.ScanInfo)
		if err != nil {
			log.Println("API: Failure saving mac table", err)
			return c.Status(fiber.StatusOK).SendString(err.Error())
		}
		err = db.SaveAliases(request.Aliases)
		if err != nil {
			log.Println("API: Failure saving aliases", err)
			return c.Status(fiber.StatusOK).SendString(err.Error())
		}
		err = db.GuessMacInfo()
		if err != nil {
			log.Println("API: Failure guessing mac info", err)
		}

		go func() {
			db.BuildMacCorrelationTable(site) // ~360 msec
			db.PopulateVendors()              // Sequential start to avoid DB lock contention
		}()
	}
	return c.Status(fiber.StatusOK).SendString("OKAY")
}

//measureIt(site) /// 350 msec, Total Allocations: 342,254,472, Number of GCs: 105
// func measureIt(site string) {
// 	start := time.Now()
// 	var m1, m2 runtime.MemStats
// 	// Capture memory stats before calling the function
// 	runtime.ReadMemStats(&m1)
// 	// Call the function you want to measure
// 	db.BuildMacCorrelationTable(site)
// 	// Capture memory stats after calling the function
// 	runtime.ReadMemStats(&m2)
// 	// Print the difference in allocated memory
// 	// TotalAlloc is a cumulative count of allocated bytes.
// 	fmt.Printf("Memory Usage for BuildMacCorrelationTable:\n")
// 	fmt.Printf("Total Memory Allocated: %d bytes\n", m2.TotalAlloc-m1.TotalAlloc)
// 	fmt.Printf("Number of GCs: %v\n", m2.NumGC-m1.NumGC)
// 	fmt.Println("Correlation took: " + time.Since(start).String())
// }
