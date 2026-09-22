package db

import (
	"log"
	"sort"
	"strings"
)

// Check that the Seen Date and Audit Date are not mixed up.
// UPDATE devices
// SET last_seen = latest.opened
// FROM (
//     SELECT cid, opened
//     FROM action_log
//     WHERE aid IN (
//         SELECT MAX(aid)
//         FROM action_log
//         WHERE action IN ('SIGHTING', 'DIED', 'USING', 'GIVING', 'BROKEN')
//         GROUP BY cid
//     )
// ) AS latest
// WHERE devices.cid = latest.cid

// Computer Status information structure
// You would think this would be easy to cache, but the timezone offset is user specific for all the date fields.
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
	IsLate     bool       // Late backup flag
	IsMissing  bool       // Missing backup flag
	IsBroken   bool       // Broken device flag
	SeenDays   int        // Days since last seen
	SeenDate   string     // YYYY-MM-DD of most recent checkin
	AuditDays  int        // Days since last checkin
	AuditDate  string     // YYYY-MM-DD of most recent checkin
	City       string     // City
	State      string     // State
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

func GetComstat(curUid int) ([]*Comstat, error) {
	computers := make([]*Comstat, 0)
	var query strings.Builder
	tzoff := GetTzoff(curUid) // User Time Zone Offest in minutes
	late := GetLateBackups()
	broken := GetBrokenDevices()
	missing := GetMissingDevices()
	lastAudit, err := getLastAudit(curUid)
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
	    COALESCE(strftime('%Y-%m-%d', A.last_seen-?, 'unixepoch'), 'never') AS last_seen
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

	rows, err := Conn.Query(query.String(), tzoff)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var dto Comstat
		err := rows.Scan(&dto.Cid, &dto.Name, &dto.Type, &dto.Icon, &dto.Site, &dto.Office, &dto.Status, &dto.Group, &dto.Assigned, &dto.SeenDays, &dto.SeenDate)
		if err != nil {
			log.Println(err)
			continue
		} else {
			//Determine if backup is late - over 90 days ago
			j := sort.SearchInts(late, dto.Cid)
			dto.IsLate = j < len(late) && late[j] == dto.Cid
			//Determine if not seen in 90 days
			j = sort.SearchInts(missing, dto.Cid)
			dto.IsMissing = j < len(missing) && missing[j] == dto.Cid
			j = sort.SearchInts(broken, dto.Cid)
			dto.IsBroken = j < len(broken) && broken[j] == dto.Cid
			// Set last audt date and days since last seen
			if audit, ok := lastAudit[dto.Cid]; ok {
				dto.AuditDate = audit.Checkin
				dto.AuditDays = audit.Days
				dto.City = audit.City
				dto.State = audit.State
				dto.Community = audit.Community
				dto.Latitude = audit.Latitude
				dto.Longitude = audit.Longitude
			}
			computers = append(computers, &dto)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	//Get the backup information for all the computers and map it to the corresponding computer
	backupInfo, err := GetLatestBackups(curUid)
	if err != nil {
		log.Println(err)
	} else {
		for _, comp := range computers {
			for _, backup := range backupInfo {
				if comp.Cid == backup.Cid {
					comp.FileDate = backup.FileDate
					comp.FileDays = backup.FileDays
					comp.SystemDate = backup.SystemDate
					comp.SystemDays = backup.SystemDays
					comp.DiskDate = backup.DiskDate
					comp.DiskDays = backup.DiskDays
				}
			}
		}
	}

	// Get the disk information for all the computers and map it to the corresponding computer
	diskInfo, err := GetDiskInfo(curUid, -1)
	if err != nil {
		log.Println(err)
	} else {
		// Map the disk info to the corresponding computer
		for _, comp := range computers {
			for _, disk := range diskInfo {
				if comp.Cid == disk.Cid {
					comp.DisksInfo = append(comp.DisksInfo, disk)
				}
			}
		}
	}
	return computers, nil
}

func getLastAudit(curUid int) (map[int]Tracks, error) {
	var items = make(map[int]Tracks)
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
		`
	rows, err := Conn.Query(query, tzoff)
	if err != nil {
		log.Println(err)
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
			items[item.Cid] = item
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	return items, nil
}
