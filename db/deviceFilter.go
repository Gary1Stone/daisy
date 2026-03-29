package db

import (
	"database/sql"
	"log"
)

type DeviceFilter struct {
	Id        int    `json:"id"`        // Database row id
	Owner     int    `json:"owner"`     // Who owns the filter
	Task      string `json:"task"`      // Action the user wants to take, get page, get next page,...
	Page      int    `json:"page"`      // Which page of the results to return
	Cid       int    `json:"cid"`       // Device (computer ID)
	DevType   string `json:"devtype"`   // What type of device, desktop, laptop, tablet, printer, phone,... (pick list)
	Site      string `json:"site"`      // What site the computer is in (pick list)
	Office    string `json:"office"`    // What office the computer is in (pick list)
	Gid       int    `json:"gid"`       // User Group ID
	Uid       int    `json:"uid"`       // User ID
	SearchTxt string `json:"searchtxt"` // User's seach phrase
	IsLate    bool   `json:"islate"`    // True if it has not been backed up in 90 days
	IsMissing bool   `json:"ismissing"` // True if it has not been seen in 90 days
}

func GetDeviceFilter(curUid int) (DeviceFilter, error) {
	var filter DeviceFilter
	late := 0
	missing := 0
	query := `
		SELECT 
			id, coalesce(owner, 0) owner, task, page, coalesce(cid, 0) cid, devtype, site, office, gid, coalesce(uid, 0) uid, searchtxt, islate, ismissing 
		FROM 
			device_filter 
		WHERE owner=?
		`
	err := Conn.QueryRow(query, curUid).Scan(&filter.Id, &filter.Owner, &filter.Task, &filter.Page,
		&filter.Cid, &filter.DevType, &filter.Site, &filter.Office, &filter.Gid, &filter.Uid,
		&filter.SearchTxt, &late, &missing)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	} else {
		if late > 0 {
			filter.IsLate = true
		}
		if missing > 0 {
			filter.IsMissing = true
		}
	}
	return filter, err
}

func (filter *DeviceFilter) SetDeviceFilter(curUid int) error {
	late := 0
	missing := 0
	if filter.IsLate {
		late = 1
	}
	if filter.IsMissing {
		missing = 1
	}
	query := `
	INSERT INTO device_filter (owner, task, page, cid, devtype, site, office, gid, uid, searchtxt, islate, ismissing) 
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
	ON CONFLICT(owner) 
	DO UPDATE SET 
		task=excluded.task,
		page=excluded.page,
		cid=excluded.cid,
		devtype=excluded.devtype,
		site=excluded.site, 
		office=excluded.office, 
		gid=excluded.gid, 
		uid=excluded.uid, 
		searchtxt=excluded.searchtxt, 
		islate=excluded.islate, 
		ismissing=excluded.ismissing
	`
	_, err := Conn.Exec(query, foreignKey(curUid), filter.Task, filter.Page,
		foreignKey(filter.Cid), filter.DevType, filter.Site, filter.Office, filter.Gid, foreignKey(filter.Uid),
		filter.SearchTxt, late, missing)
	if err != nil {
		log.Println(err)
	}
	return err
}
