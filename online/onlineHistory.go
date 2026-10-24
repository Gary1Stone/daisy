package online

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/gbsto/daisy/db"
)

// returns the minumum and maximum (now) for the online history
// in this computer's timzone - NOT UTC
func MinMaxHistoryDate(curUid, tzoff int) (minDate, maxDate string) {
	// Convert current UTC time to the user's local time using the provided offset.
	// Note: The sign of tzoff from the browser's `getTimezoneOffset()` is inverted
	// compared to Go's `time.FixedZone`. JavaScript returns minutes WEST of UTC,
	// while Go expects seconds EAST of UTC. We assume tzoff is already corrected to seconds EAST.
	// browser says add tzoff this to get UTC. Go says subtract this to get current location time
	// tzoff := GetTzoff(curUid)
	userLocation := time.FixedZone("UserLocalTime", -tzoff)
	currentTimeInUserTZ := time.Now().In(userLocation)
	maxDate = currentTimeInUserTZ.Format("2006-01-02")

	// Get min timestamp from history table
	query := `SELECT MIN(date) FROM onlinehistory`
	var minDateInt int
	if err := db.Conn.QueryRow(query).Scan(&minDateInt); err != nil {
		log.Println("cannot query min date from online table:", err)
		return "2023-01-01", maxDate // Fallback to a reasonable default if the query fails
	}

	// convert UTC to user time
	minDateUTC, err := time.Parse("20060102", fmt.Sprintf("%d", minDateInt))
	if err != nil {
		log.Printf("Error parsing min date from database '%d': %v", minDateInt, err)
		return "2023-01-01", maxDate
	}

	// Convert the UTC time to the user's local timezone before formatting.
	// This correctly handles cases where the user's timezone would shift the date.
	minDateInUserTZ := minDateUTC.In(userLocation)
	return minDateInUserTZ.Format("2006-01-02"), maxDate
}

// Used by get network page to show online.offline counts
func GetCurrentOnOffCounts() (int, int) {
	var onCount, totalCount int
	query := "SELECT COUNT(DISTINCT COALESCE(A.alias, M.mac)) FROM macs M LEFT JOIN aliases A ON A.mac=M.mac"
	err := db.Conn.QueryRow(query).Scan(&totalCount)
	if err != nil {
		log.Println(err)
		return 0, 0
	}
	ts := time.Now().UTC()
	date := ts.Year()*10000 + int(ts.Month())*100 + ts.Day()
	seconds := ts.Hour()*3600 + ts.Minute()*60 + ts.Second()
	slot := seconds / (15 * 60) // 0–95
	column := "am"
	if slot > 48 {
		column = "pm"
		slot -= 48
	}
	query = fmt.Sprintf("SELECT count(*) FROM onlinehistory O LEFT JOIN aliases A ON A.mac=O.mac WHERE date=? AND (%s & (1 << ?)) != 0", column)
	err = db.Conn.QueryRow(query, date, slot).Scan(&onCount)
	if err != nil {
		log.Println(err)
		return 0, totalCount
	}
	// If the current timeslot has not been recorded yet, get the previous timeslot
	if onCount < 1 {
		// subtract 15 minutes from the current UTC time
		ts = time.Now().UTC().Add(-15 * time.Minute)
		date := ts.Year()*10000 + int(ts.Month())*100 + ts.Day()
		seconds := ts.Hour()*3600 + ts.Minute()*60 + ts.Second()
		slot := seconds / (15 * 60) // 0–95
		column := "am"
		if slot > 48 {
			column = "pm"
			slot -= 48
		}
		query = fmt.Sprintf("SELECT count(*) FROM onlinehistory O LEFT JOIN aliases A ON A.mac=O.mac WHERE date=? AND (%s & (1 << ?)) != 0", column)
		err = db.Conn.QueryRow(query, date, slot).Scan(&onCount)
		if err != nil {
			log.Println(err)
			return 0, totalCount
		}
	}
	return onCount, totalCount - onCount
}

// Get the Random MAC addresses online information
func GetOnlineHistoryInfo(macList []string) ([]Online, []int64, error) {
	var items []Online
	var dateList []int64
	if len(macList) == 0 {
		return nil, nil, nil
	}

	args := make([]any, len(macList))
	placeholders := make([]string, len(macList))
	for i, mac := range macList {
		args[i] = mac
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(`
		SELECT COALESCE(A.mac, O.mac) as mac, O.date, O.am, O.pm
		FROM onlinehistory O
		LEFT JOIN aliases A ON O.mac = A.alias
		WHERE O.mac IN (%s)
		ORDER BY O.date`, strings.Join(placeholders, ","))

	rows, err := db.Conn.Query(query, args...)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	defer rows.Close()

	dateSet := make(map[int64]struct{})
	for rows.Next() {
		var item Online
		if err := rows.Scan(&item.Mac, &item.Date, &item.Am, &item.Pm); err != nil {
			log.Println(err)
			return nil, nil, err
		}
		items = append(items, item)
		dateSet[item.Date] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, nil, err
	}
	scrubbed, err := scrubOnline(items)
	if err != nil {
		return nil, nil, err
	}

	for date := range dateSet {
		dateList = append(dateList, date)
	}
	slices.Sort(dateList)

	return scrubbed, dateList, nil
}

// Build a lookup map of mac to hostname for easier analysis later
// TODO: Add macs where hostanme is null or zero length
func GetMac2Hostname(randomOnly bool) (map[string]string, error) {
	mac2Hostname := make(map[string]string)
	query := "SELECT O.mac, COALESCE(M.Hostname, o.mac) FROM onlinehistory O LEFT JOIN macs M on M.mac=O.mac "
	if randomOnly {
		query += "WHERE M.isRandomMac=1 "
	}
	query += "GROUP BY O.mac"
	rows, err := db.Conn.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mac, hostname string
		if err := rows.Scan(&mac, &hostname); err != nil {
			log.Println(err)
			return nil, err
		}
		mac2Hostname[mac] = hostname
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return mac2Hostname, err
	}
	return mac2Hostname, nil
}

// System/Probe online/offline information, for the entire date range (0's for missing days)
func getSysHistoryInRange(minDate, maxDate int64) (map[int64]Online, error) {
	items := make(map[int64]Online)
	lastDate := yyyymmddUTC2time(maxDate)
	for date := yyyymmddUTC2time(minDate); !date.After(lastDate); date = date.AddDate(0, 0, 1) {
		dateKey := int64(dayKey(date))
		items[dateKey] = Online{Date: dateKey}
	}

	query := `SELECT mac, date, am, pm FROM onlinehistory WHERE host=1 AND date >= ? AND date <= ?`
	rows, err := db.Conn.Query(query, minDate, maxDate)
	if err != nil {
		log.Println("Error querying online data:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Online
		err := rows.Scan(&item.Mac, &item.Date, &item.Am, &item.Pm)
		if err != nil {
			log.Println("Error scanning online data:", err)
			continue
		}
		// Handle having multiple hosts within the same day
		if oldItem, ok := items[item.Date]; ok {
			item.Am |= oldItem.Am
			item.Pm |= oldItem.Pm
		}
		items[item.Date] = item
	}
	if err = rows.Err(); err != nil {
		log.Println("Error iterating online data:", err)
		return nil, err
	}
	return items, nil
}

// The onlinehistory table is a synthetic view of the online table, with macs replaced by their aliases
// Duplicate Macs on the same day are combined here
func scrubOnline(items []Online) ([]Online, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// A map to track seen keys and their index in the result slice.
	seen := make(map[string]int)
	result := make([]Online, 0, len(items))

	for _, item := range items {
		key := fmt.Sprintf("%s_%d", item.Mac, item.Date)
		if index, ok := seen[key]; ok {
			// Duplicate. Merge with the item already in the result slice.
			result[index].Am |= item.Am
			result[index].Pm |= item.Pm
		} else {
			// New item. Add to result slice and record its index.
			result = append(result, item)
			seen[key] = len(result) - 1
		}
	}

	return result, nil
}
