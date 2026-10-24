package online

import (
	"fmt"
	"log"
	"time"

	"github.com/gbsto/daisy/db"
)

// get which devices were online for a given date (YYYYMMDD)
func getOnlineDay(tzoff int, yyyymmdd string) ([]Online, error) {
	input := yyyymmdd + " 00:01" // Convert yyyymmdd hh:nn to time
	layout := "20060102 15:04"
	dayUTC, err := time.Parse(layout, input) // Parse without timezone (naive time)
	if err != nil {
		dayUTC = time.Now().UTC()
	}
	dayUTC = dayUTC.Add(time.Duration(tzoff) * time.Second) // Apply offset: UTC = local + offset
	return getDayOnlineDevices(tzoff, dayUTC)
}

// get the online history for all the devices in the 3 days for timezone shifting
func getDayOnlineDevices(tzoff int, dayUTC time.Time) ([]Online, error) {
	prevDay, currDay, nextDay := gen3Days(dayUTC)
	startIdx := genStartindex(tzoff)                               // Pre-calculate timezone offset index
	onlineInfo, err := getWhatWasOnline(prevDay, currDay, nextDay) // This has been scrubbed
	if err != nil {
		log.Println("Error fetching online data:", err)
		return nil, err
	}

	if len(onlineInfo) == 0 {
		return nil, nil
	}

	// get system history in range in a date indexed map
	sysHist, err := getSysHistoryInRange(int64(prevDay), int64(nextDay)) // Already handled scrubbing
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup of device data
	deviceData := make(map[string]*Online, len(onlineInfo))
	for i := range onlineInfo {
		key := fmt.Sprintf("%s_%d", onlineInfo[i].Mac, onlineInfo[i].Date)
		deviceData[key] = &onlineInfo[i]
	}

	// Eliminate devices that were not active in the user's timezone for the current day
	activeMacs := make([]Online, 0, len(onlineInfo))

	// Iterate over the days we have data for and adjust for timezone
	// We have up to three (days) entries of macs in the list, but we only need one for analysis
	processedMacs := make(map[string]bool) // Track processed macs to avoid duplicates

	for i := range onlineInfo {
		if processedMacs[onlineInfo[i].Mac] {
			continue
		}
		processedMacs[onlineInfo[i].Mac] = true

		days := [3]int{prevDay, currDay, nextDay}
		var deviceDays, systemDays [3]Online
		for dayIndex, day := range days {
			if dev, ok := deviceData[fmt.Sprintf("%s_%d", onlineInfo[i].Mac, day)]; ok {
				deviceDays[dayIndex] = *dev
			}
			systemDays[dayIndex] = sysHist[int64(day)]
		}
		shiftOnlineBits(&onlineInfo[i], deviceDays, systemDays, startIdx)

		if onlineInfo[i].Am|onlineInfo[i].Pm != 0 {
			activeMacs = append(activeMacs, onlineInfo[i])
		}
	}
	return activeMacs, nil
}

// Fetch System (Host) status and all device online data for the 3 days
func getWhatWasOnline(prevDay, currDay, nextDay int) ([]Online, error) {
	items := make([]Online, 0)
	query := `SELECT O.mac, O.date, O.am, O.pm, O.host FROM onlinehistory O 
		LEFT JOIN macs M ON O.mac=M.mac
		WHERE date IN (?, ?, ?)
		ORDER BY  M.intruder DESC, M.name
	`
	rows, err := db.Conn.Query(query, prevDay, currDay, nextDay)
	if err != nil {
		log.Println("Error querying online data:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Online
		err := rows.Scan(&item.Mac, &item.Date, &item.Am, &item.Pm, &item.Host)
		if err != nil {
			log.Println("Error scanning online data:", err)
			continue
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println("Error iterating online data:", err)
		return nil, err
	}
	return scrubOnline(items)
}

type LastOnline struct {
	Name string
	Mac  string
	date int
	Date string
	Cid  int
}

func GetLastOnlineDates(curUid, tzoff int) ([]LastOnline, error) {
	items := make([]LastOnline, 0)

	query := `
		SELECT last_seen, mac, cid, name
		FROM (
			SELECT
				O.date AS last_seen, O.mac, M.cid, COALESCE(M.name, M.hostname, 'Unknown') AS name,
				ROW_NUMBER() OVER (
					PARTITION BY M.cid
					ORDER BY O.date DESC
				) AS rn
			FROM online O
			LEFT JOIN macs M ON M.mac = O.mac
			WHERE (O.AM > 0 OR O.PM > 0) AND M.cid > 0
		)
		WHERE rn = 1
		ORDER BY name DESC
		`

	rows, err := db.Conn.Query(query)
	if err != nil {
		log.Println("Error querying online data:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item LastOnline
		err := rows.Scan(&item.date, &item.Mac, &item.Cid, &item.Name)
		if err != nil {
			log.Println("Error scanning online data:", err)
			continue
		}
		item.Date = formatDate(item.date, tzoff)
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println("Error iterating online data:", err)
		return nil, err
	}
	return items, nil
}
