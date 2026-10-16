package device_mac

import (
	"database/sql"
	"errors"
	"log"

	"github.com/gbsto/daisy/db"
)

type Device struct {
	Cid        int    `json:"cid"`        // Computer ID
	Name       string `json:"name"`       // Computer name
	Model      string `json:"model"`      // Computer model
	Kind       string `json:"kind"`       // Computer role (switch, desktop, router...) of the device
	KindName   string `json:"kindname"`   // User visible description of the kind
	Icon       string `json:"icon"`       // Name of svg icon for the role/kind
	Site       string `json:"site"`       //
	SiteName   string `json:"sitename"`   // The node's site
	Office     string `json:"office"`     //
	OfficeName string `json:"officename"` // User visible description of the office
	IsLinked   bool   `json:"islinked"`   // Does the computer have any MACs associated with it
	IsOnline   bool   `json:"isonline"`   // Is the device currently online
}

type Mac struct {
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

const deviceQuery = `SELECT D.cid, D.name, D.model, COALESCE(D.kind, '') AS kind, COALESCE(K.description, '') AS kindname, 
		site, COALESCE(S.description, "") AS sitename, Office, COALESCE(O.description, '') AS officeName,  
		COALESCE(I.icon,'') AS icon
		FROM devices D
		LEFT JOIN choices AS S ON S.field='SITE' AND S.code=D.SITE
		LEFT JOIN choices AS O ON O.field='OFFICE' AND O.code=D.OFFICE AND O.parent=D.site
		LEFT JOIN choices AS K ON K.field='KIND' AND K.code=D.kind
		LEFT JOIN icons AS I ON I.name=D.kind
		WHERE D.active=1 `

func GetDevice(cid int) (item Device, err error) {
	if cid < 1 {
		err = errors.New("invalid id")
		return
	}
	err = db.Conn.QueryRow(deviceQuery+` AND D.cid=?`, cid).Scan(&item.Cid, &item.Name, &item.Model, &item.Kind, &item.KindName, &item.Site, &item.SiteName, &item.Office, &item.OfficeName, &item.Icon)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Println(err)
		return
	}
	return
}

func GetDevices(site string) (items []Device, err error) {
	rows, err := db.Conn.Query(deviceQuery+` AND D.site=? ORDER BY D.Name`, site)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item Device
		err = rows.Scan(&item.Cid, &item.Name, &item.Model, &item.Kind, &item.KindName, &item.Site, &item.SiteName, &item.Office, &item.OfficeName, &item.Icon)
		if err != nil {
			log.Println(err)
			return
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}
	return
}

func GetDevicesMacs(site string) (items []Device, err error) {
	query := `SELECT D.cid, D.name, D.model, COALESCE(D.kind, '') AS kind, COALESCE(I.icon,'') AS icon, COALESCE(M.cid, 0) AS linked
		FROM devices D
		LEFT JOIN icons AS I ON D.kind = I.name
		LEFT JOIN macs AS M ON D.cid = M.cid
		WHERE D.active=1 AND D.site='?' GROUP BY M.cid ORDER BY D.name`
	rows, err := db.Conn.Query(query, site)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item Device
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

// const d2mQuery = `SELECT D.cid, D.name, D.model,
// COALESCE(D.parent, 0) AS parent,
// COALESCE(D.office, '') AS office,
// COALESCE(D.kind, '') AS kind,
// COALESCE(I.icon,'') AS icon,
// COALESCE(O.description, '') AS officetitle,
// COALESCE(K.description, D.kind, '') AS kindtitle,
// COALESCE(online, 0) AS online
// FROM devices D
// LEFT JOIN icons I ON D.kind = I.name
// LEFT JOIN choices O ON D.office = O.code AND O.field='OFFICE' AND O.parent=?
// LEFT JOIN choices K ON D.kind = K.code AND K.field='KIND'
// LEFT JOIN (SELECT cid, MAX(online) AS online FROM macs GROUP BY cid) M2 ON D.cid = M2.cid
// WHERE D.active=1 AND D.status != 'STORAGE' `

// func BuildDev2M(cid int, site string) (Device, error) {

// 	query := d2mQuery + ` AND D.cid=?`
// 	var node Device
// 	err := db.Conn.QueryRow(query, site, cid).Scan(&node.Cid, &node.Name, &node.Model, &node.Parent, &node.Office, &node.Kind, &node.Icon, &node.OfficeName, &node.KindName, &node.IsOnline)
// 	if err != nil {
// 		log.Println(err)
// 		return node, err
// 	}
// 	return node, nil
// }
