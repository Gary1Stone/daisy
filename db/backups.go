package db

import (
	"log"
)

type BackupInfo struct {
	Id        int    `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Source    string `json:"source"`
	Cid       int    `json:"cid"`
	Volume    string `json:"volume"`
	Computer  string `json:"computer"`
	Date      int    `json:"date"`
	Size      int    `json:"size"`
	Method    string `json:"method"`
	What      string `json:"what"`
	Dated     string `json:"dated"`
	Created   string `json:"created"`
	Days      int    `json:"days"`
}

// API sends listings of all the existing backup files to SaveBackups
func SaveBackups(backups []BackupInfo, source string) error {

	if len(backups) == 0 {
		return nil
	}

	query := "DELETE FROM backups WHERE source=?"
	_, err := Conn.Exec(query, source)
	if err != nil {
		log.Println(err)
	}

	query = "INSERT INTO backups (source, volume, computer, date, size, method, what) VALUES (?, ?, ?, ?, ?, ?, ?)"
	for _, backup := range backups {
		_, err := Conn.Exec(query, source, backup.Volume, backup.Computer, backup.Date, backup.Size, backup.Method, backup.What)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	// Try to determine CID if possible from computer name
	query = `
		UPDATE Backups
		SET cid = (SELECT Devices.cid FROM Devices WHERE Devices.name=Backups.computer)
		WHERE Backups.computer IN (SELECT name FROM Devices)
	`
	_, err = Conn.Exec(query)
	if err != nil {
		log.Println(err)
	}

	// Update Laptops server to have its backup name of WKNC-63
	query = `
		UPDATE Backups SET cid = (SELECT cid FROM devices WHERE name='LAPTOPS' LIMIT 1)
		WHERE cid IS NULL
		AND computer='WKNC-63'
	`
	_, err = Conn.Exec(query)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func GetBackups(curUid, cid int) ([]BackupInfo, error) {
	items := make([]BackupInfo, 0)
	tzoff := GetTzoff(curUid) // Adjust for user's timezone
	query := `
		SELECT cid, source, timestamp, strftime('%Y-%m-%d %H:%M', timestamp-?, 'unixepoch') AS created, 
		volume, computer, date, strftime('%Y-%m-%d', date-?, 'unixepoch') AS dated, 
		size, method, what FROM backups WHERE cid=?
		ORDER by date DESC`
	rows, err := Conn.Query(query, tzoff, tzoff, cid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item BackupInfo
		err := rows.Scan(&item.Cid, &item.Source, &item.Timestamp, &item.Created, &item.Volume, &item.Computer, &item.Date,
			&item.Dated, &item.Size, &item.Method, &item.What)
		if err != nil {
			log.Println(err)
			continue
		} else {
			items = append(items, item)
		}
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return items, nil
}

type LatestBackups struct {
	Cid        int    `json:"cid"`
	Computer   string `json:"computer"`
	FileDate   string `json:"filedate"`
	FileDays   int    `json:"filedays"`
	FileSize   int    `json:"filesize"`
	SystemDate string `json:"systemdate"`
	SystemDays int    `json:"systemdays"`
	SystemSize int    `json:"systemsize"`
	DiskDate   string `json:"diskdate"`
	DiskDays   int    `json:"diskdays"`
	DiskSize   int    `json:"disksize"`
	Make       string `json:"make"`
	Model      string `json:"model"`
	Type       string `json:"type"`
}

// List all the most recent backups for all the computers, both FILE and SYSTEM
// Flatten the Files row and System row into a single row
func GetLatestBackups(curUid int) ([]LatestBackups, error) {
	items := make([]LatestBackups, 0)
	tzoff := GetTzoff(curUid) // Adjust for user's timezone

	query := `
		WITH RankedBackups AS (
			SELECT cid, computer, what, date, size,
				ROW_NUMBER() OVER(PARTITION BY cid, what ORDER BY date DESC) as rn
			FROM backups
			WHERE cid IS NOT NULL AND what IN ('Files', 'System', 'Disk')
		)
		SELECT
			rb.cid,
			rb.computer,
			MAX(CASE WHEN rb.what = 'Files' THEN strftime('%Y-%m-%d', rb.date - ?, 'unixepoch') ELSE '' END),
			MAX(CASE WHEN rb.what = 'Files' THEN CAST((strftime('%s', 'now') - rb.date) / 86400 AS INTEGER) ELSE 0 END),
			MAX(CASE WHEN rb.what = 'Files' THEN rb.size ELSE 0 END),
			MAX(CASE WHEN rb.what = 'System' THEN strftime('%Y-%m-%d', rb.date - ?, 'unixepoch') ELSE '' END),
			MAX(CASE WHEN rb.what = 'System' THEN CAST((strftime('%s', 'now') - rb.date) / 86400 AS INTEGER) ELSE 0 END),
			MAX(CASE WHEN rb.what = 'System' THEN rb.size ELSE 0 END),
			MAX(CASE WHEN rb.what = 'Disk' THEN strftime('%Y-%m-%d', rb.date - ?, 'unixepoch') ELSE '' END),
			MAX(CASE WHEN rb.what = 'Disk' THEN CAST((strftime('%s', 'now') - rb.date) / 86400 AS INTEGER) ELSE 0 END),
			MAX(CASE WHEN rb.what = 'Disk' THEN rb.size ELSE 0 END)
		FROM RankedBackups rb
		WHERE rb.rn = 1 AND rb.cid IS NOT NULL
		GROUP BY rb.cid, rb.computer
		ORDER BY rb.computer
	`
	rows, err := Conn.Query(query, tzoff, tzoff, tzoff)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item LatestBackups
		err := rows.Scan(
			&item.Cid,
			&item.Computer,
			&item.FileDate,
			&item.FileDays,
			&item.FileSize,
			&item.SystemDate,
			&item.SystemDays,
			&item.SystemSize,
			&item.DiskDate,
			&item.DiskDays,
			&item.DiskSize,
		)
		if err != nil {
			log.Println(err)
			continue
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return items, nil
}
