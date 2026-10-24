package online

import (
	"log"

	"github.com/gbsto/daisy/db"
)

func getComputerHistory(cid, tzoff int) ([]Online, error) {
	onlineInfo, err := getComputerOnlineHistory(cid) // note: slots will be pre-filled with zeros (offline) here
	if err != nil {
		return nil, err
	}
	if len(onlineInfo) == 0 {
		return nil, nil
	}
	return shiftTimezones(tzoff, onlineInfo)
}

func getComputerOnlineHistory(cid int) (items []Online, err error) {
	seen := make(map[int64]int)
	query := `SELECT COALESCE(A.alias, O.mac) AS macs, O.date, O.am, O.pm, o.host 
		FROM online O 
		LEFT JOIN aliases A ON O.mac=A.mac
		WHERE macs IN
		(SELECT mac FROM device_mac WHERE cid=? UNION SELECT mac FROM macs WHERE cid=? AND active=1)
		ORDER BY date DESC `
	rows, err := db.Conn.Query(query, cid, cid)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item Online
		if err = rows.Scan(&item.Mac, &item.Date, &item.Am, &item.Pm, &item.Host); err != nil {
			log.Println(err)
			return
		}

		if idx, ok := seen[item.Date]; ok {
			// Duplicate. Merge with the item already in the result slice. - Only the LAST MAC of the day is retained
			items[idx].Am |= item.Am
			items[idx].Pm |= item.Pm
		} else {
			// New item. Add to items slice and record its index.
			items = append(items, item)
			seen[item.Date] = len(items) - 1
		}
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
	}

	return
}
