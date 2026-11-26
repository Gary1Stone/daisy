package api

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/gbsto/daisy/db"
	"github.com/gofiber/fiber/v2"
)

type SpeedPayload struct {
	Cmd     string      `json:"cmd"`     // HWM (High-water Mark) or SAVE
	ApiKey  string      `json:"apikey"`  // The API key
	Records []Speedtest `json:"records"` // The messages
}

type Speedtest struct {
	Id               int     `json:"Id"`
	Created          string  `json:"Created"`
	IsWiFi           bool    `json:"isWiFi"`
	Ssid             string  `json:"Ssid"`
	Bssid            string  `json:"Bssid"`
	Rssi             int     `json:"Rssi"`
	UserIp           string  `json:"UserIp"`
	UserLat          float64 `json:"UserLat"`
	UserLon          float64 `json:"UserLon"`
	UserIsp          string  `json:"UserIsp"`
	Server           string  `json:"Server"`
	ServerLat        float64 `json:"ServerLat"`
	ServerLon        float64 `json:"ServerLon"`
	Name             string  `json:"Name"`
	Country          string  `json:"Country"`
	Sponsor          string  `json:"Sponsor"`
	ServerId         int     `json:"ServerId"`
	Host             string  `json:"Host"`
	Distance         float64 `json:"Distance"`
	Latency          int     `json:"Latency"`       // MilliSeconds
	MaxLatency       int     `json:"MaxLatency"`    // MilliSeconds
	MinLatency       int     `json:"MinLatency"`    // MilliSeconds
	Jitter           int     `json:"Jitter"`        // MilliSeconds
	DownloadSpeed    float64 `json:"DownloadSpeed"` // Mbps (Mega bytes per second)
	UploadSpeed      float64 `json:"UploadSpeed"`   // Mbytes per second
	TestDuration     int     `json:"TestDuration"`
	PingDuration     int     `json:"PingDuration"`
	DownloadDuration int     `json:"DownloadDuration"`
	UploadDuration   int     `json:"UploadDuration"`
	PacketLossSent   int     `json:"PacketLossSent"`
	PacketLossDup    int     `json:"PacketLossDup"`
	PacketLossMax    int     `json:"PacketLossMax"`
	Sync             int     `json:"Sync"`
}

// Store where the backups are kept
func PostSpeedApi(c *fiber.Ctx) error {

	msg := new(SpeedPayload) // Allocate on heap (address of)
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
		i, err := msg.getSpeedHwm()
		if err != nil {
			log.Println("API: Failure getting taps high-water marks", err)
			return c.Status(fiber.StatusBadRequest).SendString("Error getting high-water marks")
		}
		return c.Status(fiber.StatusOK).SendString(strconv.Itoa(i))

	case "SAVE":
		err := msg.saveSpeed()
		if err != nil {
			log.Println("API: Failure adding taps", err)
			return c.Status(fiber.StatusBadRequest).SendString("Error saving records")
		}

	}

	return c.Status(fiber.StatusOK).SendString("OKAY")

}

func (m *SpeedPayload) getSpeedHwm() (int, error) {
	cnt := 0
	query := "SELECT COALESCE(MAX(src_id), 0) FROM speedtest"
	err := db.Conn.QueryRow(query).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return 0, err
	}
	return cnt, err
}

func (m *SpeedPayload) saveSpeed() error {
	tx, err := db.Conn.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()

	const query = `INSERT OR IGNORE INTO speedtest (src_id, created,
	IsWiFi, Ssid, Bssid, Rssi, UserIp, UserLat, UserLon, UserIsp, Server, 
	ServerLat, ServerLon, Name, Country, Sponsor, ServerId, Host, Distance, 
	Latency, MaxLatency, MinLatency, Jitter, DownloadSpeed, UploadSpeed, 
	TestDuration, PingDuration, DownloadDuration, UploadDuration, 
	PacketLossSent, PacketLossDup, PacketLossMax
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := db.Conn.Prepare(query)
	if err != nil {
		log.Fatalf("FATAL: Error preparing speedtest insert statement: %v", err)
	}

	for _, r := range m.Records {
		_, err := stmt.Exec(r.Id, r.Created, r.IsWiFi, r.Ssid, r.Bssid, r.Rssi,
			r.UserIp, r.UserLat, r.UserLon, r.UserIsp, r.Server,
			r.ServerLat, r.ServerLon, r.Name, r.Country, r.Sponsor, r.ServerId, r.Host, r.Distance,
			r.Latency, r.MaxLatency, r.MinLatency, r.Jitter, r.DownloadSpeed, r.UploadSpeed,
			r.TestDuration, r.PingDuration, r.DownloadDuration, r.UploadDuration,
			r.PacketLossSent, r.PacketLossDup, r.PacketLossMax)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}

	return nil
}
