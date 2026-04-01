package db

import (
	"database/sql"
	"log"
	"math"
)

type ActiveUsers struct {
	Task      string `json:"task"`
	Id        int    `json:"id"`
	Uid       int    `json:"uid"`
	Fullname  string `json:"fullname"`
	Country   string `json:"country"`
	State     string `json:"state"`
	City      string `json:"city"`
	Community string `json:"community"`
	Since     string `json:"since"`
}

func GetActiveUsers(curUid int) ([]ActiveUsers, error) {
	items := make([]ActiveUsers, 0)
	tzoff := GetTzoff(curUid)
	query := `
		SELECT max(L.timestamp), L.uid, L.id, P.fullname, L.country, L.state, L.city, L.community, strftime('%Y-%m-%d %H:%M', L.timestamp-?, 'unixepoch') AS since 
		FROM logins L
		LEFT JOIN profiles P ON L.uid=P.uid
		WHERE L.success=1 AND L.timestamp > strftime('%s', datetime('now', '-290 day')) 
		GROUP BY L.uid 
		ORDER BY L.timestamp DESC, P.fullname
	`
	rows, err := Conn.Query(query, tzoff)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var item ActiveUsers
		var timestamp int = 0
		err := rows.Scan(&timestamp, &item.Uid, &item.Id, &item.Fullname, &item.Country, &item.State,
			&item.City, &item.Community, &item.Since)
		if err != nil {
			log.Println(err)
			continue
		} else {
			items = append(items, item)
		}
	}
	if rows.Err() != nil {
		log.Println(rows.Err())
	}
	return items, nil
}

// If the JWT's (JSON Web Token) session ID does not match
// the session ID in the database the JWT cookie is deleted
// and the user needs to log in again
func EndSession(uid int) {
	tx, err := Conn.Begin()
	if err != nil {
		log.Println(err)
		return
	}
	defer tx.Rollback()

	// 1. Clear session in logins table
	if _, err := tx.Exec("UPDATE logins SET session='' WHERE uid=?", uid); err != nil {
		log.Println(err)
		return
	}

	// 2. Remove user's passkeys using a subquery to find the auth_id mapping
	if _, err := tx.Exec("DELETE FROM credentials WHERE auth_id = (SELECT auth_id FROM profiles WHERE uid=?)", uid); err != nil {
		log.Println(err)
		return
	}

	// 3. Clear the auth_id in the profiles table
	if _, err := tx.Exec("UPDATE profiles SET auth_id=null WHERE uid=?", uid); err != nil {
		log.Println(err)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

// Returns the counts of how many unique people logged in within 15 minute periods
// for the last 30 days
func GetLoginCounts(duration, width int) ([]int, error) {
	var points []int

	days := "'-1 days'"
	switch duration {
	case 7:
		days = "'-7 days'"
	case 30:
		days = "'-30 days'"
	}
	query := `
		WITH RECURSIVE
		params AS (
			SELECT
			-- Calculate the Unix timestamp for the start of the 24-hour window,
			-- rounded down to the nearest 15-minute (900 seconds) interval.
			CAST(strftime('%s', 'now', ` + days + `) / 900 AS INTEGER) * 900 AS period_start_ts,
			-- Calculate the Unix timestamp for the start of the 15-minute interval
			-- *following* the current one. This ensures the current interval is included.
			CAST(strftime('%s', 'now') / 900 AS INTEGER) * 900 + 900 AS period_end_ts
		),
		-- Generate a series of 15-minute interval start timestamps
		intervals(interval_start_ts) AS (
			-- Anchor member: start with the beginning of our analysis period
			SELECT period_start_ts FROM params
			UNION ALL
			-- Recursive member: add 15 minutes (900 seconds) to the previous interval's start
			SELECT interval_start_ts + 900
			FROM intervals, params -- Include params to access period_end_ts
			WHERE interval_start_ts + 900 < period_end_ts -- Continue until we reach the end of our period
		)
		-- Select the interval start time (in human-readable and Unix format)
		-- and count the distinct UIDs from the logins table for each interval.
		SELECT
		--datetime(i.interval_start_ts, 'unixepoch') AS interval_start_time_human, -- Human-readable format
		--i.interval_start_ts AS interval_start_time_unix,                     -- Unix timestamp format
		COALESCE(COUNT(DISTINCT l.uid), 0) AS unique_login_count
		FROM intervals i
		LEFT JOIN logins l
		ON l.timestamp >= i.interval_start_ts       -- Login timestamp is on or after interval start
		AND l.timestamp < (i.interval_start_ts + 900) -- Login timestamp is before the next interval start
		AND l.success = 1                             -- Assuming you only want to count successful logins
		GROUP BY i.interval_start_ts
		ORDER BY i.interval_start_ts;
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

// Reduce the point count to match the width of the SVG plot
func bucketDownsample(data []int, pixelWidth int) []int {
	n := len(data)
	if n == 0 || pixelWidth <= 0 {
		return nil
	}
	if pixelWidth >= n {
		return data
	}

	bucketSize := int(math.Ceil(float64(n) / float64(pixelWidth)))
	result := make([]int, 0, pixelWidth)

	for i := 0; i < n; i += bucketSize {
		end := i + bucketSize
		if end > n {
			end = n
		}
		max := data[i]
		for _, v := range data[i:end] {
			if v > max {
				max = v
			}
		}
		result = append(result, max)
	}
	return result
}

func GetActiveUserCount() (int, error) {
	cnt := 0
	query := "SELECT count(*) FROM profiles WHERE active=1"
	err := Conn.QueryRow(query).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return 0, err
	}
	return cnt, nil
}

// Returns the counts of how many unique people logged in within 15 minute periods
// for the last 30 days
func GetHitsCounts(duration, width int) ([]int, error) {
	var points []int
	days := "'-1 days'"
	switch duration {
	case 7:
		days = "'-7 days'"
	case 30:
		days = "'-30 days'"
	}
	query := "SELECT hits FROM hits WHERE timestamp > strftime('%s', 'now', " + days + ") ORDER BY timestamp DESC"
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
