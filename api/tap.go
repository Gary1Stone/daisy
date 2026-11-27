package api

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/gbsto/daisy/db"
	"github.com/gofiber/fiber/v2"
)

type Msg struct {
	ApiKey  string   `json:"apikey"`  // The API key
	Cmd     string   `json:"cmd"`     // Command: SAVE or HWM (high water mark)
	Host    string   `json:"host"`    // Hostname of this server
	Records []Record `json:"records"` // The database dhcp table data
}

type Record struct {
	Id       int    `json:"id"`       // The dhcp table's primary key (automunber)
	Created  int64  `json:"created"`  // When created in UTC
	Site     string `json:"site"`     // The site name (from the .env file)
	Host     string `json:"host"`     // The hostname of the computer running this tap/probe
	Ssid     string `json:"ssid"`     // The network name (wifi name or ethernet name)
	Bssid    string `json:"bssid"`    // The wifi base station MAC address
	Mac      string `json:"mac"`      // The requesting computer's MAC address
	Hostname string `json:"hostname"` // The requesting computer's hostname
	IP       string `json:"ip"`       // The IP assigned to the computer after the request
	Vendor   string `json:"vendor"`   // Vendor info
	ClientID string `json:"clientid"` // The requesting mobile phone's id
	IsRandom bool   `json:"israndom"` // Is the MAC random
}

// Store where the dhcp requests are kept
func PostTapsApi(c *fiber.Ctx) error {

	msg := new(Msg) // Allocate on heap (address of)
	if err := c.BodyParser(msg); err != nil {
		log.Println("API: parser ", err)
		return c.Status(fiber.StatusBadRequest).SendString("Server error parsing message.")
	}

	// Validate the API key
	if msg.ApiKey != os.Getenv("ISAW_KEY") {
		log.Println("API: Invalid API Key")
		return c.Status(fiber.StatusBadRequest).SendString("Bad API key")
	}

	switch msg.Cmd {

	case "HWM":
		i, err := msg.getTapHwm()
		if err != nil {
			log.Println("API: Failure getting taps high-water marks", err)
			return c.Status(fiber.StatusBadRequest).SendString("Error getting high-water marks")
		}
		return c.Status(fiber.StatusOK).SendString(strconv.Itoa(i))

	case "SAVE":
		err := msg.saveTaps()
		if err != nil {
			log.Println("API: Failure adding taps", err)
			return c.Status(fiber.StatusBadRequest).SendString("Error saving records")
		}

	}

	return c.Status(fiber.StatusOK).SendString("OKAY")

}

func (m *Msg) getTapHwm() (int, error) {
	cnt := 0
	query := "SELECT COALESCE(MAX(src_id), 0) FROM dhcp WHERE host=?"
	err := db.Conn.QueryRow(query, m.Host).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return 0, err
	}
	return cnt, err
}

// in the database, ensure host + src_id are unique in the dhcp table
func (m *Msg) saveTaps() error {
	tx, err := db.Conn.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()
	const Query = `INSERT OR IGNORE INTO dhcp (src_id, created, site, host, ssid, bssid, mac, hostname, ip, vendor, clientid, israndom) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := db.Conn.Prepare(Query)
	if err != nil {
		log.Fatalf("FATAL: Error preparing DHCP insert statement: %v", err)
	}

	for _, r := range m.Records {
		_, err := stmt.Exec(r.Id, r.Created, r.Site, r.Host, r.Ssid, r.Bssid, r.Mac, r.Hostname, r.IP, r.Vendor, r.ClientID, r.IsRandom)
		if err != nil {
			log.Printf("Error saving DHCP record for %v: %v", r.Id, err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}

	return nil
}
