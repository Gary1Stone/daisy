package api

import (
	"log"

	"github.com/gbsto/daisy/db"
)

// A single computer can have multiple disk drives
func setDiskInfo(cid int, diskInfo []DiskInfo) error {
	// Delete existing disk data for this device
	query := "DELETE FROM disks WHERE cid=?"
	_, err := db.Conn.Exec(query, cid)
	if err != nil {
		log.Println(err)
	}

	// Add the disk information
	query = "INSERT INTO disks (cid, drive, total, free, used, fill) VALUES (?,?,?,?,?,?)"
	for _, disk := range diskInfo {
		_, err := db.Conn.Exec(query, cid, disk.Drive, disk.Total, disk.Free, disk.Used, disk.Fill)
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
	tzoff := db.GetTzoff(curUid)
	rows, err := db.Conn.Query(query, tzoff, cid)
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

func setDriveHealth(cid int, health []DriveHealth) error {
	// Delete existing disk health data for this device
	query := "DELETE FROM disk_health WHERE cid=?"
	_, err := db.Conn.Exec(query, cid)
	if err != nil {
		log.Println(err)
	}

	// Add the disk information
	query = "INSERT INTO disk_health (cid, name, deviceid, drivetype, healthstatus, operationalstatus) VALUES (?,?,?,?,?,?)"
	for _, disk := range health {
		_, err := db.Conn.Exec(query, cid, disk.Name, disk.DeviceID, disk.DriveType, disk.HealthStatus, disk.OperationalStatus)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func setLetterHealth(cid int, health []DriveLetterHealth) error {
	// Delete existing disk data for this device
	query := "DELETE FROM disk_letter_health WHERE cid=?"
	_, err := db.Conn.Exec(query, cid)
	if err != nil {
		log.Println(err)
	}

	// Add the disk information
	query = "INSERT INTO disk_letter_health (cid, driveletter, healthstatus, operationalstatus, filesystem) VALUES (?,?,?,?,?)"
	for _, disk := range health {
		_, err := db.Conn.Exec(query, cid, disk.DriveLetter, disk.HealthStatus, disk.OperationalStatus, disk.FileSystem)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func setPredictiveFailure(cid int, items []PredictiveFailure) error {
	query := "DELETE FROM disk_predictive_failure WHERE cid=?"
	_, err := db.Conn.Exec(query, cid)
	if err != nil {
		log.Println(err)
	}

	//	NOTE:PredictFailure is true / false / nil (unknown)
	query = "INSERT INTO disk_predictive_failure (cid, Name, DeviceID, PredictFailure, ReasonForFailure) VALUES (?,?,?,?,?)"
	for _, disk := range items {
		onoff := 0 // No failure predicted
		if disk.PredictFailure != nil && *disk.PredictFailure {
			onoff = 1
		}
		_, err := db.Conn.Exec(query, cid, disk.Name, disk.DeviceID, onoff, disk.ReasonForFailure)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
