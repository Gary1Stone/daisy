package db

import "log"

type D2MDevice struct {
	Cid      int    `json:"cid"`      // Computer ID
	Name     string `json:"name"`     // Computer name
	Model    string `json:"model"`    // Computer model
	Kind     string `json:"kind"`     // Computer role (switch, desktop, router...)
	Icon     string `json:"icon"`     // Name of svg icon for the role/kind
	IsLinked bool   `json:"islinked"` // Does the computer have any MACs associated with it
}

func GetD2MDevices(site string) (items []D2MDevice, err error) {
	query := `SELECT D.cid, D.name, D.model, COALESCE(D.kind, '') AS kind, COALESCE(I.icon,'') AS icon, COALESCE(M.cid, 0) AS linked
		FROM devices D
		LEFT JOIN icons AS I ON D.kind = I.name
		LEFT JOIN macs AS M ON D.cid = M.cid
		WHERE D.active=1 AND D.site='WKNC' GROUP BY M.cid ORDER BY D.name`
	rows, err := Conn.Query(query, site)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item D2MDevice
		link := 0
		err = rows.Scan(&item.Cid, &item.Name, &item.Model, &item.Kind, &item.Icon, &link)
		if err != nil {
			log.Println(err)
			return
		}
		item.IsLinked = false
		if link > 0 {
			item.IsLinked = true
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}
	return
}
