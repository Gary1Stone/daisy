package db

import (
	"log"
	"strings"
)

type MacInfo struct {
	Mid         int    `json:"Mid"`         // Mac ID - Primary Key
	Mac         string `json:"Mac"`         // MAC address of the NIC card (Unique)
	Created     int64  `json:"Created"`     // Unix timestamp when record was created
	Hostname    string `json:"Hostname"`    // The name the device has
	Ip          string `json:"Ip"`          // IP (address) Its the computers Internet Protocol address
	Vendor      string `json:"Vendor"`      // The vendor of the NIC card from vendor lookup table
	Online      bool   `json:"Online"`      // From most current arp/pingsweep/Nmap scan
	Scanned     int64  `json:"Scanned"`     // Unix timestamp of last scan that saw this device
	Source      string `json:"Source"`      // Most current record source: arp/pingsweep/Nmap/Host scan. If "Host" its this computer
	Firstseen   string `json:"Firstseen"`   // String timestamp in user's timezone YYYY-MM-DD
	Lastseen    string `json:"Lastseen"`    // String timestamp in user's timezone YYYY-MM-DD
	Intruder    bool   `json:"Intruder"`    // Newly seen device (intrusion flag)
	Active      bool   `json:"Active"`      // Active flag for sudo record deletion
	Updated     int64  `json:"Updated"`     // UTC UNIX timestamp of last update
	Name        string `json:"Name"`        // Manually entered display name
	Kind        string `json:"Kind"`        // Kind of device,: computer, router, switch, watch,...
	Os          string `json:"Os"`          // Operating System
	User        string `json:"User"`        // Assigned user, if any. Manually entered
	Site        string `json:"Site"`        // Assigned site, if any. Manually entered
	Office      string `json:"Office"`      // Assigned office/room, if any. Manually entered
	Location    string `json:"Location"`    // Assigned location, if any. Manually entered
	Note        string `json:"Note"`        // Manually entered note
	Cid         int    `json:"Cid"`         // Computer ID from devices table, if any
	IsSolitary  bool   `json:"IsSolitary"`  // Mac table can have other entries with identical hostnames that are not the same device as this one
	IsRandomMac bool   `json:"IsRandomMac"` // Mac address is random generated
	IsIgnore    bool   `json:"IsIgnore"`    // Ignore this mac for online overlap correlations
}

func getMacInfo(whereclause string, tzoff int, params ...any) ([]MacInfo, error) {
	items := make([]MacInfo, 0)
	prependValues := []any{tzoff, tzoff}      //need two timezone offsets for query
	params = append(prependValues, params...) // add any additional params that were passed in

	query := `
		SELECT Mid, Mac, Created, Name, Hostname, Ip, Kind, Os,
		User, Site, Office, Location, Note, Vendor, 
		Online, Scanned, Source, Intruder, Updated, Active, coalesce(cid, 0) AS cid,
		isSolitary, isRandomMac, isIgnore,		
		strftime('%Y-%m-%d', created-?, 'unixepoch') AS firstseen,
		strftime('%Y-%m-%d', scanned-?, 'unixepoch') AS lastseen
		FROM macs ` + whereclause
	rows, err := Conn.Query(query, params...)
	if err != nil {
		log.Println("Error querying mac info:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item MacInfo
		err := rows.Scan(&item.Mid, &item.Mac, &item.Created, &item.Name, &item.Hostname,
			&item.Ip, &item.Kind, &item.Os, &item.User, &item.Site, &item.Office,
			&item.Location, &item.Note, &item.Vendor, &item.Online, &item.Scanned,
			&item.Source, &item.Intruder, &item.Updated, &item.Active, &item.Cid,
			&item.IsSolitary, &item.IsRandomMac, &item.IsIgnore, &item.Firstseen, &item.Lastseen)
		if err != nil {
			log.Println("Error scanning mac info:", err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// Save the macInfo map to the database, only the user addressable fields, not the scan fields
func SaveMacs(items []MacInfo) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := Conn.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback() // Defer Rollback. It's a no-op if Commit succeeds.

	query := `UPDATE macs SET name=?, ip=?, kind=?, os=?, user=?, site=?, office=?, location=?, note=?, intruder=?, active=?, cid=? WHERE mid=?`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println("Error preparing statement", err)
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err := stmt.Exec(item.Name, item.Ip, item.Kind, item.Os, item.User,
			item.Site, item.Office, item.Location, item.Note, item.Intruder,
			item.Active, foreignKey(item.Cid), item.Mid)
		if err != nil {
			log.Printf("Error saving macs to database for %v. Rolling back transaction. Error: %v", item, err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}
	return nil
}

// Update the macInfo map to the database, only the user addressable fields, not the scan fields
func UpdateMac(item MacInfo) error {
	tx, err := Conn.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback() // Defer Rollback. It's a no-op if Commit succeeds.

	query := `UPDATE macs SET name=?, kind=?, office=?, location=?, note=?, intruder=? WHERE mid=?`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println("Error preparing statement", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(item.Name, item.Kind, item.Office, item.Location, item.Note, item.Intruder, item.Mid)
	if err != nil {
		log.Printf("Error saving mac to database for %v. Rolling back transaction. Error: %v", item, err)
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}
	return nil
}

func GetMacInfoByMid(tzoff, mid int) (MacInfo, error) {
	var item MacInfo
	params := []any{mid}
	items, err := getMacInfo("WHERE mid=?", tzoff, params...)
	if err != nil {
		log.Println(err)
		return item, err
	}
	if len(items) == 0 {
		return item, nil
	}
	return items[0], nil
}

func GetMacInfoByMac(tzoff int, mac string) (MacInfo, error) {
	var item MacInfo
	params := []any{mac}
	//	params = append(params, mac)
	items, err := getMacInfo("WHERE mac=?", tzoff, params...)
	if err != nil {
		log.Println(err)
		return item, err
	}
	if len(items) == 0 {
		return item, nil
	}
	return items[0], nil
}

type ChartMacInfo struct {
	Mid  int    `json:"Mid"`  // Mac ID - Primary Key
	Mac  string `json:"Mac"`  // MAC address of the NIC card (Unique)
	Name string `json:"Name"` // Manually entered display name
}

func GetChartInfo4WhatWasOnline(maclist []string) (map[string]ChartMacInfo, error) {
	items := make(map[string]ChartMacInfo)
	args := make([]any, len(maclist))
	for i, v := range maclist {
		args[i] = v
	}
	query := `SELECT Mid, Mac, Name FROM macs WHERE mac IN (?` + strings.Repeat(",?", len(maclist)-1) + `)`
	rows, err := Conn.Query(query, args...)
	if err != nil {
		log.Println("Error querying mac info:", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item ChartMacInfo
		err := rows.Scan(&item.Mid, &item.Mac, &item.Name)
		if err != nil {
			log.Println("Error scanning mac info:", err)
			continue
		}
		items[item.Mac] = item
	}
	return items, nil
}

// Get the random mac's hostnames for bigram correlation
func GetHostnames() ([]string, error) {
	items := make([]string, 0)
	query := `SELECT Hostname FROM macs WHERE Hostname IS NOT NULL AND Hostname != '' AND isRandomMac = 1 ORDER BY Hostname`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			log.Println(err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// Get the mac's hostnames for correlation
// Only include the hostnames where their macs are not in the alias table
func GetDuplicateHostnames() ([]string, error) {
	items := make([]string, 0)
	query := `SELECT m.mac
		FROM macs AS m
		JOIN (
			SELECT hostname
			FROM macs
			GROUP BY hostname
			HAVING COUNT(*) > 1
		) AS d
			ON m.hostname = d.hostname
		LEFT JOIN aliases a1 ON m.mac = a1.mac
		LEFT JOIN aliases a2 ON m.mac = a2.alias
		WHERE m.active = 1 AND m.isSolitary = 0 AND m.isIgnore = 0
		AND a1.mac IS NULL
		AND a2.alias IS NULL
		ORDER BY m.hostname`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mac string
		err := rows.Scan(&mac)
		if err != nil {
			log.Println(err)
			continue
		}
		items = append(items, mac)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	return items, nil
}

// return list of macs that are active, not in the alias table, not isSolitary ....?
func GetMacList() ([]string, error) {
	items := make([]string, 0)
	query := `SELECT M.mac FROM macs M
		JOIN (SELECT DISTINCT mac FROM onlinehistory) O ON O.mac = M.mac
		WHERE M.active = 1 AND M.isSolitary = 0 AND M.isIgnore = 0
	`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			log.Println(err)
			continue
		}
		items = append(items, item)
	}
	if rows.Err() != nil {
		log.Println(err)
		return nil, err
	}
	return items, nil
}
