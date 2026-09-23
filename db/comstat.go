package db

import (
	"log"
	"sort"
	"strings"
)

// Computer Status (Comstat) information structure
type Comstat struct {
	Cid        int        // Computer ID
	Name       string     // Computer Name
	Type       string     // Computer Type
	Icon       string     // Icon for the type
	Site       string     // Site description
	Office     string     // Office description
	Status     string     // Status description
	Group      string     // Assigned Group name
	Assigned   string     // Assigned User
	IsBroken   bool       // Broken device flag
	SeenDays   int        // Days since last seen
	SeenDate   string     // YYYY-MM-DD of most recent seen report
	AuditDays  int        // Days since last checkin
	AuditDate  string     // YYYY-MM-DD of most recent automated checkin
	Community  string     // Community
	Latitude   float64    // Where the device was last audited
	Longitude  float64    // Where the device was last audited
	FileDays   int        //
	FileDate   string     // YYYY-MM-DD of most recent file backup
	SystemDays int        //
	SystemDate string     // YYYY-MM-DD of most recent system backup
	DiskDays   int        //
	DiskDate   string     // YYYY-MM-DD of most recent disk backup
	DisksInfo  []DiskInfo // Drive information for this computer
}

type lastBackups struct {
	cid        int
	fileDate   string
	fileDays   int
	systemDate string
	systemDays int
	diskDate   string
	diskDays   int
}

type audits struct {
	cid       int     // Computer ID
	auditDate string  // YYYY-MM-DD of most recent checkin audit
	auditDays int     // Days since last checkin audit
	community string  // Community
	latitude  float64 // Where the device was last audited
	longitude float64 // Where the device was last audited
}

func GetComstat(curUid int) ([]*Comstat, error) {
	//	start := time.Now()
	computers := make([]*Comstat, 0)
	var query strings.Builder
	tzoff := GetTzoff(curUid)
	broken := GetBrokenDevices()

	// Get the last audit information for all activecomputers (map)
	lastAudit, err := getLastAudit(curUid)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Get the last backup information for all activecomputers (map)
	backupInfo, err := GetAllLatestBackups(curUid)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Get the disk information for all active computers (slice)
	driveInfo, err := GetAllDisksInfo(curUid)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	query.WriteString(`
	SELECT 
		A.cid, A.name, A.type, E.icon2 as icon,
		COALESCE(H.description, '') As site,
		COALESCE(G.description, '') As office,
		COALESCE(I.description, '') As status,
		COALESCE(K.description, '') As groupname,
		COALESCE(B.fullname, '') AS assigned, 
        COALESCE(cast((strftime('%s', 'now') - A.last_seen) / 86400 AS INTEGER), 0) AS seen_days,
	    COALESCE(strftime('%Y-%m-%d', A.last_seen-?, 'unixepoch'), 'never') AS seen_date,
        COALESCE(cast((strftime('%s', 'now') - A.last_audit) / 86400 AS INTEGER), 0) AS audit_days,
	    COALESCE(strftime('%Y-%m-%d', A.last_audit-?, 'unixepoch'), 'never') AS audit_date
	FROM devices A  
		LEFT JOIN profiles B ON A.uid=B.uid 
		LEFT JOIN icons E on E.name=A.type
		LEFT JOIN (SELECT code, description FROM choices WHERE field='OFFICE' GROUP BY code) G ON A.office=G.code
		LEFT JOIN choices H on H.field='SITE' AND H.code=A.site
		LEFT JOIN choices I on I.field='STATUS' AND I.code=A.status
		LEFT JOIN choices J on J.field='MAKE' AND J.code=A.make
		LEFT JOIN choices K on K.field='GROUP' AND K.code=A.gid
	WHERE A.active=1 AND (A.type='LAPTOP' OR A.type='DESKTOP')
	ORDER BY A.name 
	`)

	rows, err := Conn.Query(query.String(), tzoff, tzoff)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var dto Comstat
		err := rows.Scan(&dto.Cid, &dto.Name, &dto.Type, &dto.Icon, &dto.Site, &dto.Office, &dto.Status, &dto.Group, &dto.Assigned, &dto.SeenDays, &dto.SeenDate, &dto.AuditDays, &dto.AuditDate)
		if err != nil {
			log.Println(err)
			continue
		} else {
			j := sort.SearchInts(broken, dto.Cid)
			dto.IsBroken = j < len(broken) && broken[j] == dto.Cid
			// Set last audt date and days since last seen
			if audit, ok := lastAudit[dto.Cid]; ok {
				dto.AuditDate = audit.auditDate
				dto.AuditDays = audit.auditDays
				dto.Community = audit.community
				dto.Latitude = audit.latitude
				dto.Longitude = audit.longitude
			}

			// Set backup information for this computer if it exists in the backupInfo map
			if backup, ok := backupInfo[dto.Cid]; ok {
				dto.FileDate = backup.fileDate
				dto.FileDays = backup.fileDays
				dto.SystemDate = backup.systemDate
				dto.SystemDays = backup.systemDays
				dto.DiskDate = backup.diskDate
				dto.DiskDays = backup.diskDays
			}

			// Set disk information for this computer if it exists in the driveInfo map.
			dto.DisksInfo = driveInfo[dto.Cid]

			computers = append(computers, &dto)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	//	log.Println("comstat time: " + time.Since(start).String())
	return computers, nil
}

func getLastAudit(curUid int) (map[int]audits, error) {
	var items = make(map[int]audits)
	tzoff := GetTzoff(curUid)
	query := `
		SELECT 
			t.cid, a.community_ascii, t.latitude, t.longitude,
			strftime('%Y-%m-%d', MAX(t.timestamp)-?, 'unixepoch') AS audit_date,
			coalesce(cast((strftime('%s', 'now') - t.timestamp) / 86400 AS INTEGER), 0) AS audit_days
		FROM tracks t
		JOIN devices d ON t.cid = d.cid
		JOIN communities a ON t.community_id = a.community_id
		WHERE d.active = 1
		GROUP BY  t.cid
		`
	rows, err := Conn.Query(query, tzoff)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item audits
		err := rows.Scan(&item.cid, &item.community, &item.latitude, &item.longitude, &item.auditDate, &item.auditDays)
		if err != nil {
			log.Println(err)
			continue
		} else {
			items[item.cid] = item
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	return items, nil
}

// Reurn a map that contains a slice of drives for each computer
func GetAllDisksInfo(curUid int) (map[int][]DiskInfo, error) {
	disks := make(map[int][]DiskInfo)
	query := "SELECT cid, drive, total, free, used, fill, timestamp, strftime('%Y-%m-%d', timestamp-?, 'unixepoch') as Localtime FROM disks "
	rows, err := Conn.Query(query, GetTzoff(curUid))
	if err != nil {
		log.Println(err)
		return disks, err
	}
	defer rows.Close()
	for rows.Next() {
		var disk DiskInfo
		err := rows.Scan(&disk.Cid, &disk.Drive, &disk.Total, &disk.Free, &disk.Used, &disk.Fill, &disk.Timestamp, &disk.Localtime)
		if err != nil {
			log.Println(err)
		} else {
			disks[disk.Cid] = append(disks[disk.Cid], disk)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return disks, nil
}

// List the most recent backups for all the computers, for FILE and SYSTEM and DISK
func GetAllLatestBackups(curUid int) (map[int]lastBackups, error) {
	items := make(map[int]lastBackups)
	tzoff := GetTzoff(curUid)
	query := `
		WITH RankedBackups AS (
			SELECT cid, what, date,
				ROW_NUMBER() OVER(PARTITION BY cid, what ORDER BY date DESC) as rn
			FROM backups
			WHERE cid IS NOT NULL AND what IN ('Files', 'System', 'Disk')
		)
		SELECT
			rb.cid,
			MAX(CASE WHEN rb.what = 'Files' THEN strftime('%Y-%m-%d', rb.date - ?, 'unixepoch') ELSE '' END) AS filedate,
			MAX(CASE WHEN rb.what = 'Files' THEN CAST((strftime('%s', 'now') - rb.date) / 86400 AS INTEGER) ELSE 0 END) AS filedays,
			MAX(CASE WHEN rb.what = 'System' THEN strftime('%Y-%m-%d', rb.date - ?, 'unixepoch') ELSE '' END) AS systemdate,
			MAX(CASE WHEN rb.what = 'System' THEN CAST((strftime('%s', 'now') - rb.date) / 86400 AS INTEGER) ELSE 0 END) AS systemdays,
			MAX(CASE WHEN rb.what = 'Disk' THEN strftime('%Y-%m-%d', rb.date - ?, 'unixepoch') ELSE '' END) AS diskdate,
			MAX(CASE WHEN rb.what = 'Disk' THEN CAST((strftime('%s', 'now') - rb.date) / 86400 AS INTEGER) ELSE 0 END) AS diskdays
		FROM RankedBackups rb
		WHERE rb.rn = 1 AND rb.cid IS NOT NULL
		GROUP BY rb.cid
	`
	rows, err := Conn.Query(query, tzoff, tzoff, tzoff)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item lastBackups
		err := rows.Scan(&item.cid, &item.fileDate, &item.fileDays, &item.systemDate, &item.systemDays, &item.diskDate, &item.diskDays)
		if err != nil {
			log.Println(err)
			continue
		} else {
			items[item.cid] = item
		}
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return items, nil
}
