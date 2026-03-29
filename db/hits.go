package db

import (
	"log"
	"math"
)

// LogHits logs the hit count at the end of a given 15-minute interval.
// The database adds the integer timestamp (UTC) automatically
func LogHits(count uint64) {
	cnt := int64(count)
	if cnt < 0 {
		cnt = math.MaxInt64
	}
	query := "INSERT INTO hits (hits) VALUES (?)"
	_, err := Conn.Exec(query, cnt)
	if err != nil {
		log.Println(err)
	}
}
