package db

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"
)

// returns the minumum and maximum (now) for the online history
// in this computer's timzone - NOT UTC
func MinMaxHistoryDate(curUid int) (minDate, maxDate string) {
	// Convert current UTC time to the user's local time using the provided offset.
	// Note: The sign of tzoff from the browser's `getTimezoneOffset()` is inverted
	// compared to Go's `time.FixedZone`. JavaScript returns minutes WEST of UTC,
	// while Go expects seconds EAST of UTC. We assume tzoff is already corrected to seconds EAST.
	// browser says add tzoff this to get UTC. Go says subtract this to get current location time
	tzoff := GetTzoff(curUid)
	userLocation := time.FixedZone("UserLocalTime", -tzoff)
	currentTimeInUserTZ := time.Now().In(userLocation)
	maxDate = currentTimeInUserTZ.Format("2006-01-02")

	// Get min timestamp from history table
	query := `SELECT MIN(date) FROM onlinehistory`
	var minDateInt int
	if err := Conn.QueryRow(query).Scan(&minDateInt); err != nil {
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
	err := Conn.QueryRow(query).Scan(&totalCount)
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
	err = Conn.QueryRow(query, date, slot).Scan(&onCount)
	if err != nil {
		log.Println(err)
		return 0, totalCount
	}
	return onCount, totalCount - onCount
}

// Debugging utility
func PrintBitmap(mac string, bmp []int) {
	fmt.Print(mac, " ")
	for i := range bmp {
		fmt.Print(bmp[i])
	}
	fmt.Print("\n")
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

	rows, err := Conn.Query(query, args...)
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
	rows, err := Conn.Query(query)
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
	return mac2Hostname, nil
}

// Build a lookup map of mac to hostname for easier analysis later
// TODO: Add macs where hostanme is null or zero length
func GetMac2Name(randomOnly bool) (map[string]string, error) {
	mac2Name := make(map[string]string)
	query := "SELECT O.mac, COALESCE(COALESCE(M.name, M.hostname), O.mac) FROM onlinehistory O LEFT JOIN macs M on M.mac=O.mac "
	if randomOnly {
		query += "WHERE M.isRandomMac=1  "
	}
	query += "GROUP BY O.mac"
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mac, name string
		if err := rows.Scan(&mac, &name); err != nil {
			log.Println(err)
			return nil, err
		}
		mac2Name[mac] = name
	}

	return mac2Name, nil
}

// get all the online history for the mac from newest first to oldest last
func GetDeviceHistory(tzoff int, mac string) ([]Online, error) {
	onlineInfo, err := getOnlineHistoryForMac(mac) // note: slots will be pre-filled with zeros (offline) here
	if err != nil {
		return nil, err
	}
	if len(onlineInfo) == 0 {
		return nil, nil
	}

	// Calculate date range for system history
	// Add buffer for timezone shifts (at least 1 day)
	maxDateVal := onlineInfo[0].Date
	minDateVal := onlineInfo[len(onlineInfo)-1].Date

	// Convert to time to subtract/add days safely
	minT := yyyymmddUTC2time(minDateVal).AddDate(0, 0, -2)
	maxT := yyyymmddUTC2time(maxDateVal).AddDate(0, 0, 2)

	minDate := int64(dayKey(minT))
	maxDate := int64(dayKey(maxT))

	// get system history in range
	sysHist, err := getSysHistoryInRange(minDate, maxDate)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup of device data
	deviceData := make(map[int64]*Online, len(onlineInfo))
	for i := range onlineInfo {
		deviceData[onlineInfo[i].Date] = &onlineInfo[i]
	}

	// Pre-calculate timezone offset index
	startIdx := genStartindex(tzoff)

	// Helper to fill a section for a given three-days slot buffer
	// defined outside the loop to avoid reallocating the closure each iteration.
	fillSection := func(dayInt int, offset int, threeDaysOfSlots *[slotsPerDay * 3]int) {
		// Check device data
		var am, pm uint64
		if dev, ok := deviceData[int64(dayInt)]; ok {
			am = uint64(dev.Am)
			pm = uint64(dev.Pm)
		}

		// Check system data
		var sysAM, sysPM uint64
		if sys, ok := sysHist[int64(dayInt)]; ok {
			sysAM = uint64(sys.Am)
			sysPM = uint64(sys.Pm)
		}

		// Fill directly into the provided array slice
		fillSlots(threeDaysOfSlots[offset:offset+slotsPerDay], am, pm, sysAM, sysPM)
	}

	// Iterate over the days we have data for and adjust for timezone
	for i := range onlineInfo {
		d := yyyymmddUTC2time(onlineInfo[i].Date)
		if len(onlineInfo[i].Slots) != slotsPerDay {
			onlineInfo[i].Slots = make([]int, slotsPerDay) // 96 elements
		}
		prevDay, currDay, nextDay := gen3Days(d)

		var threeDaysOfSlots [slotsPerDay * 3]int

		fillSection(prevDay, 0, &threeDaysOfSlots)
		fillSection(currDay, slotsPerDay, &threeDaysOfSlots)
		fillSection(nextDay, slotsPerDay*2, &threeDaysOfSlots)

		// Copy the relevant window to the item's slots
		copy(onlineInfo[i].Slots, threeDaysOfSlots[startIdx:startIdx+slotsPerDay])
	}
	return scrubOnline(onlineInfo)
}

// Get all history info for this mac in UTC, key=date int yyyymmdd
func getOnlineHistoryForMac(mac string) ([]Online, error) {
	items := make([]Online, 0)
	query := "SELECT mac, date, am, pm FROM onlinehistory WHERE mac=? ORDER BY date DESC"
	rows, err := Conn.Query(query, mac)
	if err != nil {
		log.Println("Error querying onlinehistory data:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Online
		err := rows.Scan(&item.Mac, &item.Date, &item.Am, &item.Pm)
		if err != nil {
			log.Println("Error scanning onlinehistory data:", err)
			continue
		}
		item.Slots = make([]int, 96) // prefills with zeros
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println("Error iterating onlinehistory data:", err)
		return nil, err
	}

	// Remove duplicate mac/date pairs, merging the AM & PM values
	return scrubOnline(items)
}

// Convert yyyymmdd to time
func yyyymmddUTC2time(yyyymmdd int64) time.Time {
	y := int(yyyymmdd / 10000)
	m := int((yyyymmdd % 10000) / 100)
	d := int(yyyymmdd % 100)
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// System/Probe online/offline information
func getSysHistoryInRange(minDate, maxDate int64) (map[int64]Online, error) {
	items := make(map[int64]Online)
	query := `SELECT mac, date, am, pm FROM onlinehistory WHERE host=1 AND date >= ? AND date <= ?`
	rows, err := Conn.Query(query, minDate, maxDate)
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
