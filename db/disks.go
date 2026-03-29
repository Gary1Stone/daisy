package db

import "log"

// A single computer can have multiple disk drives
func SetDiskInfo(cid int, diskInfo []DiskInfo) error {
	// Delete existing disk data for this device
	query := "DELETE FROM disks WHERE cid=?"
	_, err := Conn.Exec(query, cid)
	if err != nil {
		log.Println(err)
	}

	// Add the disk information
	query = "INSERT INTO disks (cid, drive, total, free, used, fill) VALUES (?,?,?,?,?,?)"
	for _, disk := range diskInfo {
		_, err := Conn.Exec(query, cid, disk.Drive, disk.Total, disk.Free, disk.Used, disk.Fill)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

// cid < 0 means get all
func GetDiskInfo(curUid, cid int) ([]DiskInfo, error) {
	disks := make([]DiskInfo, 0)
	query := "SELECT cid, drive, total, free, used, fill, timestamp, strftime('%Y-%m-%d', timestamp-?, 'unixepoch') as Localtime FROM disks "
	if cid < 0 {
		query += "WHERE cid>? "
	} else {
		query += "WHERE cid=? "
	}
	query += "ORDER BY cid, drive"
	tzoff := GetTzoff(curUid)
	rows, err := Conn.Query(query, tzoff, cid)
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
			disks = append(disks, disk)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return disks, nil
}
