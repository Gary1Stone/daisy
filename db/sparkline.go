package db

import (
	"database/sql"
	"log"
	"time"
)

// Returns a slice of ints for every 15-minute interval with the specified range
func GetAttacks(duration, width int) ([]int, error) {
	var points []int
	intervalSeconds := int64(15 * 60)              // 15 minutes in seconds
	totalSeconds := int64(duration * 24 * 60 * 60) // 24 hours in seconds * num of days
	now := time.Now().Unix()
	startTime := now - totalSeconds

	// Prepare the query once
	query := "SELECT COUNT(id) FROM attacks WHERE timestamp >= ? AND timestamp < ?"
	stmt, err := Conn.Prepare(query)
	if err != nil {
		log.Println(err)
		return points, err
	}
	defer stmt.Close()

	// Iterate through each 15-minute interval in the last 24 hours
	for intervalStart := startTime; intervalStart < now; intervalStart += intervalSeconds {
		intervalEnd := intervalStart + intervalSeconds

		// Ensure the last interval doesn't go beyond 'now'
		if intervalEnd > now {
			intervalEnd = now
		}

		var point int
		// Execute the prepared statement for the current interval
		err := stmt.QueryRow(intervalStart, intervalEnd).Scan(&point)
		if err != nil {
			// sql.ErrNoRows is expected if count is 0, handle it gracefully
			if err == sql.ErrNoRows {
				point = 0 // Explicitly set to 0 if no rows found (COUNT returns 0 anyway)
			} else {
				// Log other potential errors
				log.Printf("Error querying interval %d-%d: %v\n", intervalStart, intervalEnd, err)
				// Depending on requirements, you might want to append a marker like -1 or skip
				points = append(points, 0) // Append 0 on error for consistency
				continue
			}
		}
		points = append(points, point)
	}

	return bucketDownsample(points, width), nil
}
