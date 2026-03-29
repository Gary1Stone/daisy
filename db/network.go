package db

import (
	"database/sql"
	"log"
)

// Returns the counts of how many unique people logged in within 15 minute periods
// for the last 30 days
func GetNetworkDeviceCountsPerDay(duration, width int) ([]int, error) {
	var points []int
	query := `WITH RECURSIVE dates AS (
		SELECT DATE('now', '-30 days') AS d
		UNION ALL
		SELECT DATE(d, '+1 day')
		FROM dates
		WHERE d < DATE('now')
		)
		SELECT COALESCE(COUNT(online.date), 0) AS count
		FROM dates
		LEFT JOIN online ON online.date = REPLACE(dates.d, '-', '')
		GROUP BY dates.d
		ORDER BY dates.d DESC
		`
	rows, err := Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return points, err
	}
	defer rows.Close()
	for rows.Next() {
		var cnt int
		err := rows.Scan(&cnt)
		if err != nil {
			log.Println(err)
			continue
		} else {
			points = append(points, cnt)
		}
	}
	if rows.Err() != nil {
		log.Println(rows.Err())
		return points, rows.Err()
	}

	return bucketDownsample(points, width), nil
}
