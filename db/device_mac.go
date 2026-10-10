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

type M2D struct {
	Mid         int `json:"mid"` // MAC ID
	Mac         string
	MacName     string `json:"name"` // Computer Name (not hostname)
	Cid         int
	Parent      int
	Created     string
	Hostname    string
	IP          string
	Kind        string
	Os          string
	User        string
	Site        string
	Office      string
	Location    string
	Note        string
	Scanned     int
	Vendor      string
	Online      bool
	Source      string
	Intruder    bool
	Updated     int
	Active      bool
	IsSolitary  bool
	IsRandomMac bool
	IsIgnore    bool
}

type D2M struct {
	Node
	macs        []M2D
	Cid         int    `json:"cid"`         // Computer ID
	Name        string `json:"name"`        // Computer Name (not hostname)
	Icon        string `json:"icon"`        // The icon for the kind (role) of the device
	Parent      int    `json:"parent"`      // The parent device's cid
	Office      string `json:"office"`      // The office the device is in
	OfficeTitle string `json:"officetitle"` // User visible description of the office
	Kind        string `json:"kind"`        // The role (switch, router, desktop...) of the device
	KindTitle   string `json:"kindtitle"`   // User visible description of the kind
	Model       string `json:"model"`       // Device model
	IsOnline    bool   `json:"isonline"`    // Is the device currently online
	Site        string `json:"site"`        // The node's site
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

const d2mQuery = `SELECT D.cid, D.name, D.model,
COALESCE(D.parent, 0) AS parent,
COALESCE(D.office, '') AS office, 
COALESCE(D.kind, '') AS kind,
COALESCE(I.icon,'') AS icon, 
COALESCE(O.description, '') AS officetitle, 
COALESCE(K.description, D.kind, '') AS kindtitle,
COALESCE(online, 0) AS online
FROM devices D
LEFT JOIN icons I ON D.kind = I.name
LEFT JOIN choices O ON D.office = O.code AND O.field='OFFICE' AND O.parent=?
LEFT JOIN choices K ON D.kind = K.code AND K.field='KIND' 
LEFT JOIN (SELECT cid, MAX(online) AS online FROM macs GROUP BY cid) M2 ON D.cid = M2.cid
WHERE D.active=1 AND D.status != 'STORAGE' `

func BuildDev2M(cid int, site string) (D2M, error) {

	query := d2mQuery + ` AND D.cid=?`
	var node D2M
	err := Conn.QueryRow(query, site, cid).Scan(&node.Cid, &node.Name, &node.Model, &node.Parent, &node.Office, &node.Kind, &node.Icon, &node.OfficeTitle, &node.KindTitle, &node.IsOnline)
	if err != nil {
		log.Println(err)
		return node, err
	}
	return node, nil
}
