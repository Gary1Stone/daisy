package db

import (
	"log"
)

type OnlineAPI struct {
	ApiKey     string    `json:"ApiKey"`
	Command    string    `json:"Command"`
	OnlineInfo []Online  `json:"OnlineInfo"`
	ScanInfo   []MacInfo `json:"ScanInfo"`
	Aliases    []Aliases `json:"Aliases"`
}

type HighWaterMarks struct {
	Macs   int64 `json:"macs"`
	Online int64 `json:"online"`
	Alias  int64 `json:"alias"`
}

type OnlineDetails struct {
	Online  Online  `json:"Online"`
	MacInfo MacInfo `json:"MacInfo"`
	UtcDate int     `json:"UtcDate"`
	Slots   []int   `json:"Slots"` // 96 15-minute timeslots for the day, 1=online, 0=offline, -1=outage
}

type Online struct {
	Mac     string `json:"Mac"`     // Mac ID unique
	Date    int64  `json:"Date"`    // Unix timestamp UTC (Mac + Date = Primary Key in the form of YYYYMMDD)
	Am      int64  `json:"Am"`      // Morning (midnight to noon) Binary
	Pm      int64  `json:"Pm"`      // Afternoon (noon to midnight) Binary
	Host    bool   `json:"Host"`    // Is the entry from the host (network scanning) server
	Updated int64  `json:"Updated"` // Monotonic counter of last update for table replication, filled in by db trigger only
	Slots   []int  `json:"Slots"`   // 96 15-minute timeslots for the day, 1=online, 0=offline, -1=outage
}

func SaveOnline(items []Online) error {
	if len(items) == 0 {
		return nil
	}
	// Start a transaction
	tx, err := Conn.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback()

	// Prepare the statement
	query := `INSERT INTO online (mac, date, am, pm, host) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(mac, date) DO UPDATE SET
    	am = excluded.am,
    	pm = excluded.pm,
		host = excluded.host`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println("Error preparing statement:", err)
		return err
	}
	defer stmt.Close()

	// Execute the statement for each item
	for _, item := range items {
		_, err = stmt.Exec(item.Mac, item.Date, item.Am, item.Pm, item.Host)
		if err != nil {
			log.Printf("Error updating online status for %v. Rolling back. Error: %v", item, err)
			return err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}
	return nil
}

// Save the Scan Info map to the database
// Limited to the scan fields only
// Returns the site ID
func SaveScanInfo(items []MacInfo) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

	tx, err := Conn.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return "", err
	}
	defer tx.Rollback() // Defer Rollback. It's a no-op if Commit succeeds.

	query := `INSERT INTO macs
		(mid, mac, created, hostname, ip, vendor, online, scanned, source, intruder, active, os, site, updated, isRandomMac)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(mid) DO UPDATE SET
		mac=excluded.mac, created=excluded.created, hostname=excluded.hostname, 
		ip=excluded.ip, vendor=excluded.vendor, online=excluded.online, scanned=excluded.scanned,
		source=excluded.source, intruder=excluded.intruder, active=excluded.active, 
		os=excluded.os, site=excluded.site, updated=excluded.updated, isRandomMac=excluded.isRandomMac`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println("Error preparing statement:", err)
		return "", err
	}
	defer stmt.Close()

	for _, item := range items {
		item.Active = true                       // Mark as active
		item.IsRandomMac = isRandomMac(item.Mac) // Mark as random or not
		// Let isSolitary default to false (0) on insert or remain unchanged on update
		_, err := stmt.Exec(item.Mid, item.Mac, item.Created, item.Hostname, item.Ip, item.Vendor, item.Online,
			item.Scanned, item.Source, item.Intruder, item.Active, item.Os, item.Site, item.Updated, item.IsRandomMac)
		if err != nil {
			log.Printf("Error saving macs to database for %v. Rolling back transaction. Error: %v", item, err)
			return "", err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return "", err
	}
	return items[0].Site, nil
}

// Sends hightest updated count in these tables to the API for it to determine if it needs to request new data
func GetLastUpdated() HighWaterMarks {
	var hwm HighWaterMarks
	hwm.Macs = 0
	hwm.Online = 0
	hwm.Alias = 0
	query := `SELECT COALESCE((SELECT MAX(updated) FROM macs), 0) AS maxMacs,
		COALESCE((SELECT MAX(updated) FROM online), 0) AS maxOnline,
		COALESCE((SELECT MAX(updated) FROM aliases), 0) AS maxAliases`
	err := Conn.QueryRow(query).Scan(&hwm.Macs, &hwm.Online, &hwm.Alias)
	if err != nil {
		log.Println(err)
	}
	return hwm
}
