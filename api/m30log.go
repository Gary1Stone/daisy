package api

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/gbsto/daisy/db"
	"github.com/gofiber/fiber/v2"
)

type M30msgs struct {
	Cmd    string      `json:"cmd"`    // HWM (High-water Mark) or SAVE
	ApiKey string      `json:"apikey"` // The API key
	Msgs   []M30status `json:"msgs"`   // The messages
}

type M30status struct {
	Id      int    `json:"id"`      // primary key
	Created int    `json:"created"` // UTC created date
	Msg     string `json:"msg"`     // Summary message of the m30 log file
}

// Store where the backups are kept
func PostM30Api(c *fiber.Ctx) error {

	msg := new(M30msgs) // Allocate on heap (address of)
	if err := c.BodyParser(msg); err != nil {
		log.Println("API: parser ", err)
		return c.Status(fiber.StatusBadRequest).SendString("Server error parsing message.")
	}

	// Validate the API key TODO: move to .env or use the master key
	if msg.ApiKey != "itwasafuntrip2machiupicchu" {
		log.Println("API: Invalid API Key")
		return c.Status(fiber.StatusBadRequest).SendString("Bad API key")
	}

	switch msg.Cmd {

	case "HWM":
		i, err := msg.getM30Hwm()
		if err != nil {
			log.Println("API: Failure getting taps high-water marks", err)
			return c.Status(fiber.StatusBadRequest).SendString("Error getting high-water marks")
		}
		return c.Status(fiber.StatusOK).SendString(strconv.Itoa(i))

	case "SAVE":
		err := msg.saveM30()
		if err != nil {
			log.Println("API: Failure adding taps", err)
			return c.Status(fiber.StatusBadRequest).SendString("Error saving records")
		}

	}

	return c.Status(fiber.StatusOK).SendString("OKAY")

}

func (m *M30msgs) getM30Hwm() (int, error) {
	cnt := 0
	query := "SELECT COALESCE(MAX(src_id), 0) FROM m30status"
	err := db.Conn.QueryRow(query).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return 0, err
	}
	return cnt, err
}

func (m *M30msgs) saveM30() error {
	tx, err := db.Conn.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()
	const Query = `INSERT OR IGNORE INTO m30status (src_id, created, msg) VALUES (?, ?, ?)`

	stmt, err := db.Conn.Prepare(Query)
	if err != nil {
		log.Fatalf("FATAL: Error preparing M30 insert statement: %v", err)
	}

	for _, r := range m.Msgs {
		_, err := stmt.Exec(r.Id, r.Created, r.Msg)
		if err != nil {
			log.Printf("Error saving M30 record for %v: %v", r.Id, err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}

	return nil
}
