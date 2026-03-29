package db

import "log"

type Tracks struct {
	LastSeen  int64   `json:"last_seen"` // UTC time
	Checkin   string  `json:"checkin"`   // YYYY-MM-DD of most recent checkin
	Cid       int     `json:"cid"`       // Computer ID
	Name      string  `json:"name"`      // Computer name
	City      string  `json:"city"`      // City
	State     string  `json:"state"`     // State
	Community string  `json:"community"` // Community
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Days      int     `json:"days"` // Days since last checkin
}

func GetLastTracks(curUid int) ([]Tracks, error) {
	var items []Tracks
	tzoff := GetTzoff(curUid)
	query := `
		SELECT 
			MAX(t.timestamp) AS max_timestamp,
			strftime('%Y-%m-%d', MAX(t.timestamp)-?, 'unixepoch') AS checkin,
			t.cid, d.name, c.city_ascii, c.state, a.community_ascii, t.latitude, t.longitude,
			coalesce(cast((strftime('%s', 'now') - t.timestamp) / 86400 AS INTEGER), 0) AS days
		FROM tracks t
		JOIN devices d ON t.cid = d.cid
		JOIN cities c ON t.city_id = c.city_id
		JOIN communities a ON t.community_id = a.community_id
		WHERE d.active = 1
		GROUP BY  t.cid
		ORDER BY d.name
		`
	rows, err := Conn.Query(query, tzoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Tracks
		err := rows.Scan(&item.LastSeen, &item.Checkin, &item.Cid, &item.Name, &item.City, &item.State, &item.Community, &item.Latitude, &item.Longitude, &item.Days)
		if err != nil {
			log.Println(err)
			continue
		} else {
			items = append(items, item)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	return items, nil
}
