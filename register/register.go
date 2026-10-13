package register

import (
	"database/sql"
	"log"

	"github.com/gbsto/daisy/db"
)

/*	Used to register a user for JWT
	The API Codes are the codes emailed to a user to log into this application
*/

type RegisterCode struct {
	Apicode string
	Ip      string
}

// Set the apicode and ip before calling this function
func (r *RegisterCode) SetApiCode() error {
	query := "INSERT INTO api_codes (api_code, ip) VALUES (?,?)"
	_, err := db.Conn.Exec(query, r.Apicode, r.Ip)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Set the apicode and ip before calling this function
func (r *RegisterCode) IsApiCode() bool {
	r.purgeApiCodes()
	cnt := 0
	query := "SELECT count(*) FROM api_codes WHERE api_code=? and ip=?"
	err := db.Conn.QueryRow(query, r.Apicode, r.Ip).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return cnt > 0
}

// Purge codes older than 1 hour
func (r *RegisterCode) purgeApiCodes() error {
	query := "DELETE FROM api_codes WHERE timestamp < strftime('%s', 'now') - 3600"
	_, err := db.Conn.Exec(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return nil
}
