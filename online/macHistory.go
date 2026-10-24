package online

import (
	"errors"
	"log"

	"github.com/gbsto/daisy/db"
)

func getMacHistory(tzoff int, mac string) ([]Online, error) {
    if mac == "" {
        return nil, errors.New("zero length mac not permitted")
    }

    onlineInfo, err := getOnlineHistoryForMac(mac)
    if err != nil || len(onlineInfo) == 0 {
        return onlineInfo, err
    }

    onlineInfo, err = handleTimezones(tzoff, onlineInfo)
    if err != nil {
        return nil, err
    }

    return scrubOnline(onlineInfo)
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
