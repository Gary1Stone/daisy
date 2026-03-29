package db

import (
	"database/sql"
	"log"
)

//The API Codes are the codes emailed to a user to log into this application

func SetApiCode(apicode, ip string) error {
	query := "INSERT INTO api_codes (api_code, ip) VALUES (?,?)"
	_, err := Conn.Exec(query, apicode, ip)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func IsApiCode(apicode, ip string) bool {
	purgeApiCodes()
	cnt := 0
	query := "SELECT count(*) FROM api_codes WHERE api_code=? and ip=?"
	err := Conn.QueryRow(query, apicode, ip).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return cnt > 0
}

// Purge codes older than 1 hour
func purgeApiCodes() error {
	query := "DELETE FROM api_codes WHERE timestamp < strftime('%s', 'now') - 3600"
	_, err := Conn.Exec(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return nil
}
