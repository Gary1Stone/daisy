package dashboards

import (
	"database/sql"
	"log"

	"github.com/gbsto/daisy/db"
)

type DashInfo struct {
	Type          string `json:"type"`  //i.e. DESKTOP, LAPTOP, TABLET, PHONE...
	Label         string `json:"label"` //i.e. Desktop
	Icon          string `json:"icon"`
	Count         int    `json:"count"`         //Number of devices
	Unavailable   int    `json:"unavailable"`   //Out of commisison devices
	Instorage     int    `json:"instorage"`     //Number in storage
	PendingAction int    `json:"pendingAction"` //Pending actions
}

// Generate the dashboard device types and counts
func (d *DashInfo) GetDashboard() []DashInfo {
	dashInfo := make([]DashInfo, 0)
	if err := d.countDashboardDeviceTypes(&dashInfo); err != nil {
		log.Println(err)
	}
	if err := d.countDashboardLostDevices(&dashInfo); err != nil {
		log.Println(err)
	}
	if err := d.coundDashboardStoredDevices(&dashInfo); err != nil {
		log.Println(err)
	}
	if err := d.countDashboardIssues(&dashInfo); err != nil {
		log.Println(err)
	}
	return dashInfo
}

// Count the device types
// Slices are passed by reference, so don't need pointers much
func (d *DashInfo) countDashboardDeviceTypes(dash *[]DashInfo) error {
	query := `
		SELECT count(*) as cnt, A.type, B.description, B.icon 
		FROM devices A 
		LEFT JOIN icons B ON B.name=A.type 
		WHERE active=1 GROUP BY type 
	`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item DashInfo
		err := rows.Scan(&item.Count, &item.Type, &item.Label, &item.Icon)
		if err != nil {
			log.Println(err)
		} else {
			*dash = append(*dash, item)
		}
	}
	err = rows.Err()
	return err
}

// Get count of devices that were lost or died or stolen
func (d *DashInfo) countDashboardLostDevices(dash *[]DashInfo) error {
	query := `
		SELECT COUNT(cid) AS cnt, type 
		FROM devices 
		WHERE active=1  
		AND status IN ('DIED', 'LOST', 'STOLEN')
		GROUP BY type
	`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var ddi DashInfo
		err := rows.Scan(&ddi.Unavailable, &ddi.Type)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
		} else {
			for i := range *dash {
				if (*dash)[i].Type == ddi.Type { //brackets mean dereference
					(*dash)[i].Unavailable = ddi.Unavailable
					break
				}
			}
		}
	}
	err = rows.Err()
	return err
}

// Get count of devices in storage
func (d *DashInfo) coundDashboardStoredDevices(dash *[]DashInfo) error {
	query := `
		SELECT count(*) as cnt, type 
		FROM devices 
		WHERE (office='STORAGE' OR status='STORAGE') AND active=1 
		GROUP BY type
		`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var ddi DashInfo
		err := rows.Scan(&ddi.Instorage, &ddi.Type)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
		} else {
			for i := range *dash {
				if (*dash)[i].Type == ddi.Type {
					(*dash)[i].Instorage = ddi.Instorage
					break
				}
			}
		}
	}
	err = rows.Err()
	return err
}

// get pending actions that are issues
func (d *DashInfo) countDashboardIssues(dash *[]DashInfo) error {
	query := `
		SELECT count(*) AS cnt, B.type 
		FROM action_log A 
		LEFT JOIN devices B ON A.cid=B.cid 
		WHERE A.cid > 0 AND (A.cid_ack IS NULL OR A.cid_ack < 1) 
			AND A.action IN ('BROKEN', 'CARE', 'LOST', 'DIED', 'REQUEST') 
			AND B.active=1 
		GROUP BY B.type
	`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var ddi DashInfo
		err := rows.Scan(&ddi.PendingAction, &ddi.Type)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
		} else {
			for i := range *dash {
				if (*dash)[i].Type == ddi.Type {
					(*dash)[i].PendingAction = ddi.PendingAction
					break
				}
			}
		}
	}
	err = rows.Err()
	return err
}
