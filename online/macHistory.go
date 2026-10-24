package online

import (
	"log"

	"github.com/gbsto/daisy/db"
)

func getMacHistory(tzoff int, mac string) ([]Online, error) {
	onlineInfo, err := getOnlineHistoryForMac(mac) // note: slots will be pre-filled with zeros (offline) here
	if err != nil {
		return nil, err
	}
	if len(onlineInfo) == 0 {
		return nil, nil
	}
	info, err := handleTimezones(tzoff, onlineInfo)
	return scrubOnline(info)
}

// Get all history info for this mac in UTC, key=date int yyyymmdd
func getOnlineHistoryForMac(mac string) ([]Online, error) {
	items := make([]Online, 0)
	query := "SELECT mac, date, am, pm FROM onlinehistory WHERE mac=? ORDER BY date DESC"
	rows, err := db.Conn.Query(query, mac)
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
