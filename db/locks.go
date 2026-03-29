package db

import (
	"database/sql"
	"log"
)

type Locks struct {
	Uid        int `json:"uid" db:"uid"`               //User ID of who locked the record
	Record_id  int `json:"record_id" db:"record_id"`   //ID of the record
	Table_name int `json:"table_name" db:"table_name"` //Table name (Predefined above)
	Display    int `json:"display" db:"display"`       //1=Displayed, 0=Saved
}

// Adds a lock record with timestamp to the locks table
// Records who looks at or saved which record when, in server time
func (recLock Locks) AddLock() {
	if recLock.Uid < 1 {
		recLock.Uid = SYS_PROFILE.Uid
	}
	query := "INSERT INTO locks (uid, record_id, table_name, display) VALUES (?, ?, ?, ?)"
	_, err := Conn.Exec(query, recLock.Uid, recLock.Record_id, recLock.Table_name, recLock.Display)
	if err != nil {
		log.Println(err)
	}
	query = "DELETE FROM locks WHERE timestamp <= strftime('%s', datetime('now', '-31 day'))"
	_, err = Conn.Exec(query)
	if err != nil {
		log.Println(err)
	}
}

// Check if anyone else saved the record, since the current user viewed it.
// return - true if the user's displayed record is stale
func (recLock Locks) isLocked() bool {
	if recLock.Uid < 1 {
		recLock.Uid = SYS_PROFILE.Uid
	}
	var displayedTime int = 0
	var lastSaveTime int = 0
	locked := true //default to the record was locked

	//Get when this user last displayed (fetched) the record
	//Ignore errors because it may never had been displayed
	var query = `
	SELECT timestamp 
		FROM locks 
		WHERE uid=? AND record_id=? AND table_name=? AND display=1 
		ORDER BY timestamp DESC 
	LIMIT 1
	`
	err := Conn.QueryRow(query, recLock.Uid, recLock.Record_id, recLock.Table_name).Scan(&displayedTime)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return false
	}

	//Get when any user saved this record, except this user
	//- they can save multiple times in one viewing
	//Ignore errors because it may never had been written to since displayed
	query = `
	SELECT timestamp 
		FROM locks 
		WHERE record_id=? AND table_name=? AND display=0 AND uid<>? 
		ORDER by timestamp DESC 
	LIMIT 1
	`
	err = Conn.QueryRow(query, recLock.Record_id, recLock.Table_name, recLock.Uid).Scan(&lastSaveTime)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return false
	}
	locked = !(lastSaveTime == 0 || displayedTime == 0 || lastSaveTime < displayedTime)

	if !locked {
		recLock.Display = 0 //0=save 1=Show
		recLock.AddLock()
	}
	return locked
}

func isProfileLocked(curUid, uid int) bool {
	recLock := Locks{
		Uid:        curUid,
		Record_id:  uid,
		Table_name: PROFILE_TABLE,
		Display:    0,
	}
	return recLock.isLocked()
}

func addProfileLock(curUid, uid int) {
	recLock := Locks{
		Uid:        curUid,
		Record_id:  uid,
		Table_name: PROFILE_TABLE,
		Display:    1,
	}
	recLock.AddLock()
}

func isSoftwareLocked(curUid, sid int) bool {
	recLock := Locks{
		Uid:        curUid,
		Record_id:  sid,
		Table_name: SOFTWARE_TABLE,
		Display:    0,
	}
	return recLock.isLocked()
}

func addSoftwareLock(curUid, sid int) {
	recLock := Locks{
		Uid:        curUid,
		Record_id:  sid,
		Table_name: SOFTWARE_TABLE,
		Display:    1,
	}
	recLock.AddLock()
}

func addDeviceLock(curUid, cid int) {
	recLock := Locks{
		Uid:        curUid,
		Record_id:  cid,
		Table_name: DEVICE_TABLE,
		Display:    1,
	}
	recLock.AddLock()
}

func isDeviceLocked(curUid, cid int) bool {
	recLock := Locks{
		Uid:        curUid,
		Record_id:  cid,
		Table_name: DEVICE_TABLE,
		Display:    0,
	}
	return recLock.isLocked()
}

func addActionLock(curUid, aid int) {
	recLock := Locks{
		Uid:        curUid,
		Record_id:  aid,
		Table_name: ACTION_TABLE,
		Display:    1,
	}
	recLock.AddLock()
}

func isActionLocked(curUid, aid int) bool {
	recLock := Locks{
		Uid:        curUid,
		Record_id:  aid,
		Table_name: ACTION_TABLE,
		Display:    0,
	}
	return recLock.isLocked()
}
