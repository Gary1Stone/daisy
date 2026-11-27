package tap

import (
	"log"

	"github.com/gbsto/daisy/db"
)

type Netcounts struct {
	Site    string
	Network string
	Day     string
	Users   int
}

func GetUserCountsPerDayPerNetwork(tzoff int) ([]Netcounts, error) {

	var items []Netcounts
	const query = `SELECT site, ssid, day, COUNT(*) AS count
	FROM (SELECT DISTINCT date(created-?, 'unixepoch') AS day, ssid, mac, IP, site FROM dhcp)
	GROUP BY day, ssid ORDER BY site, day, ssid`

	rows, err := db.Conn.Query(query, tzoff)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var item Netcounts
		err := rows.Scan(&item.Site, &item.Network, &item.Day, &item.Users)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return items, nil
}
