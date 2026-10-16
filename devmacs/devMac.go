package devmacs

import (
	"database/sql"
	"errors"
	"log"
	"sort"

	"github.com/gbsto/daisy/db"
)

type DeviceInfo struct {
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

const deviceQuery = `SELECT D.cid, D.name, D.model, COALESCE(D.kind, '') AS kind, COALESCE(K.description, '') AS kindname, 
		site, COALESCE(S.description, "") AS sitename, Office, COALESCE(O.description, '') AS officeName,  
		COALESCE(I.icon,'') AS icon
		FROM devices D
		LEFT JOIN choices AS S ON S.field='SITE' AND S.code=D.SITE
		LEFT JOIN choices AS O ON O.field='OFFICE' AND O.code=D.OFFICE AND O.parent=D.site
		LEFT JOIN choices AS K ON K.field='KIND' AND K.code=D.kind
		LEFT JOIN icons AS I ON I.name=D.kind
		WHERE D.active=1 `

type MacInfo struct {
	Mid         int    `json:"mid"`         // MAC ID
	Mac         string `json:"mac"`         //
	MacName     string `json:"name"`        // Computer Name (not hostname)
	Cid         int    `json:"cid"`         //
	Parent      int    `json:"parent"`      //
	Created     string `json:"created"`     //
	Hostname    string `json:"hostname"`    //
	IP          string `json:"ip"`          //
	Kind        string `json:"kind"`        //
	Os          string `json:"os"`          //
	User        string `json:"user"`        //
	Site        string `json:"site"`        //
	Office      string `json:"office"`      //
	Location    string `json:"location"`    //
	Note        string `json:"note"`        //
	Scanned     int    `json:"scanned"`     //
	Vendor      string `json:"vendor"`      //
	IsOnline    bool   `json:"isonline"`    //
	Source      string `json:"source"`      //
	Intruder    bool   `json:"intruder"`    //
	Updated     int    `json:"updated"`     //
	Active      bool   `json:"active"`      //
	IsSolitary  bool   `json:"issolitary"`  //
	IsRandomMac bool   `json:"israndommac"` //
	IsIgnore    bool   `json:"isignore"`    //
	Icon        string `json:"icon"`
}

func GetDevice(cid int) (item DeviceInfo, err error) {
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

func GetDevices(site string) (items []DeviceInfo, err error) {
	rows, err := db.Conn.Query(deviceQuery+` AND D.site=? ORDER BY D.Name`, site)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item DeviceInfo
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

type DeviceMacs struct {
	Mac         string `json:"mac"`            // MAC  address
	Mid         int    `json:"mid"`            //    // If = 0, then it is not not in the MACS table, only the self-reporting table
	Hostname    string `json:"hostname"`       //
	Ip          string `json:"ip"`             //
	Kind        string `json:"kind"`           //
	KindName    string `json:"kindname"`       //
	Icon        string `json:"icon"`           //
	IsOnline    bool   `json:"isonline"`       // // Seen online in the last 15 minutes
	IsDeletable bool   `json:"isnotdeletable"` //
	FirstSeen   string `json:"firstseen"`      //
	LastSeen    string `json:"lastseen"`       //
}

const DeviceMacsQuery = `
		SELECT M.mid, M.mac, COALESCE(M.hostname, '') AS hostname, 
		COALESCE(M.ip, '') AS ip, 
		COALESCE(M.kind, '') AS kind,  COALESCE(K.description, '') AS kindname,
		COALESCE(I.icon, '') AS icon, 
		COALESCE(M.online, 0) AS online,
		strftime('%Y-%m-%d', M.created-?, 'unixepoch') AS firstseen,
		strftime('%Y-%m-%d', M.scanned-?, 'unixepoch') AS lastseen
		FROM macs M
		LEFT JOIN icons AS I ON I.name = M.kind
		LEFT JOIN choices AS K ON K.field='KIND' AND K.code=M.kind `

func GetMacInfoByCid(cid, tzoff int) (items []DeviceMacs, err error) {
	// First read all the macs for the computer that are in the self-reporting (audit.exe) table
	query := `
		SELECT COALESCE(M.mid, 0) AS mid, D.mac, COALESCE(M.hostname, '') AS hostname, 
		COALESCE(M.ip, '') AS ip, 
		COALESCE(M.kind, '') AS kind, COALESCE(K.description, '') AS kindname, 
		COALESCE(I.icon, '') AS icon, COALESCE(M.online, 0) as online
		FROM device_mac D
		LEFT JOIN macs M ON M.Mac=D.mac
		LEFT JOIN icons I ON I.name = M.kind
		LEFT JOIN choices AS K ON K.field='KIND' AND K.code=M.kind
		WHERE D.cid=?
		`
	rows, err := db.Conn.Query(query, cid)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item DeviceMacs
		err = rows.Scan(&item.Mid, &item.Mac, &item.Hostname, &item.Ip, &item.Kind, &item.KindName, &item.Icon, &item.IsOnline)
		if err != nil {
			log.Println(err)
			return
		}
		item.IsDeletable = false // from the self reporting table, so cannot delete these
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}

	//Second, find all cid macs in the Macs table that are not in the mac_device table
	query = DeviceMacsQuery + "WHERE M.cid=? AND M.mac NOT IN (SELECT mac FROM device_mac WHERE cid=?)"
	rows, err = db.Conn.Query(query, tzoff, tzoff, cid, cid)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item DeviceMacs
		err = rows.Scan(&item.Mid, &item.Mac, &item.Hostname, &item.Ip, &item.Kind, &item.KindName, &item.Icon, &item.IsOnline, &item.FirstSeen, &item.LastSeen)
		if err != nil {
			log.Println(err)
			return
		}
		item.IsDeletable = true
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}

	// Sort by mac alphabetically
	sort.Slice(items, func(i, j int) bool {
		return items[i].Mac < items[j].Mac
	})

	return
}

func GetUnmatchedMacs(cid, tzoff int, site string) (items []DeviceMacs, err error) {
	query := DeviceMacsQuery + "WHERE M.cid IS NULL AND site=? AND M.mac NOT IN (SELECT mac FROM device_mac WHERE cid=?) AND M.kind NOT IN ('CELLPHONE', 'TABLET', 'WATCH', 'PTR')"
	rows, err := db.Conn.Query(query, tzoff, tzoff, site, cid)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item DeviceMacs
		err = rows.Scan(&item.Mid, &item.Mac, &item.Hostname, &item.Ip, &item.Kind, &item.KindName, &item.Icon, &item.IsOnline, &item.FirstSeen, &item.LastSeen)
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
