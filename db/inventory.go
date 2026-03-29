package db

import (
	"database/sql"
	"log"
)

type Sw_list struct {
	Sid      int
	Name     string
	Inv_name string
}

// Return a unique list of inventory software names
func GetInventoryList() ([]string, error) {
	list := make([]string, 0)
	rows, err := Conn.Query("SELECT DISTINCT(name) FROM sw_inv")
	if err != nil {
		return list, err
	}
	defer rows.Close()
	var item string
	for rows.Next() {
		err := rows.Scan(&item)
		if err != nil {
			log.Println(err)
			continue
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func GetUsedInventoryNames() ([]string, error) {
	list := make([]string, 0)
	query := `SELECT inv_name FROM software WHERE inv_name IS NOT NULL ORDER BY inv_name`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return list, err
	}
	defer rows.Close()
	var item string
	for rows.Next() {
		err := rows.Scan(&item)
		if err != nil {
			log.Println(err)
			continue
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

type Sw_inv struct {
	Id                         int    `json:"id" db:"id"`
	Cid                        int    `json:"cid" db:"cid"`
	Name                       string `json:"name" db:"name"`
	Scandate                   int    `json:"scandate" db:"scandate"`
	Sid                        int    `json:"sid" db:"sid"`
	CountActiveInstalls        int    `json:"countactiveinstalls" db:"-"`
	CountDecomissionedInstalls int    `json:"countdecomissionedinstalls" db:"-"`
	Licenses                   int    `json:"licenses" db:"-"`
	Active                     int    `json:"active" db:"-"`
	Manual                     int    `json:"manual" db:"-"`
	Manual_inactive            int    `json:"Manual_inactive" db:"-"`
}

// Update software inventory table (sw_inv) setting the SID
// where the software sw_inv.name begins with the software.inv_name
func MatchSoftwareToInventory(sid int) error {
	_, err := Conn.Exec("UPDATE sw_inv SET sid=NULL WHERE sid=?", sid)
	if err != nil { // Improvement, do it only for the sid provided?
		return err
	}
	query := `
		WITH matched_sw_inv AS (
			SELECT A.rowid AS sw_inv_rowid, B.sid AS new_sid
			FROM sw_inv A
			LEFT JOIN software B 
			ON TRIM(LOWER(A.name)) LIKE TRIM(LOWER(B.inv_name)) || '%'
			WHERE length(B.inv_name)>0
		)
		UPDATE sw_inv
		SET sid = (
			SELECT new_sid
			FROM matched_sw_inv
			WHERE sw_inv.rowid = matched_sw_inv.sw_inv_rowid
		)
		WHERE EXISTS (
			SELECT 1
			FROM matched_sw_inv
			WHERE sw_inv.rowid = matched_sw_inv.sw_inv_rowid
		)
	`
	_, err = Conn.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

type SoftwareOnComputers struct {
	Cid        int    // Computer ID
	Sid        int    // Software ID
	Dev_active int    // Is the computer active 1=yes 0=decomissioned
	Dev_name   string // Computer Name
	Sw_active  int    // Is the software active 1=yes 0=no
	Sw_name    string // Software name
}

// Get list of manually tracked software per computer (or cid=0 for all)
func GetManuallyTrackedSoftwareOnComputers(cid int) ([]SoftwareOnComputers, error) {
	var items []SoftwareOnComputers
	query := `
		WITH LatestActions AS (
			SELECT cid, sid, MAX(opened) AS latestTime
			FROM action_log
			GROUP BY cid, sid
		),
		CurrentInstalls AS (
			SELECT a.cid, a.sid, a.opened, a.Action
			FROM action_log a
			INNER JOIN LatestActions la
			ON a.cid=la.cid AND a.sid=la.sid AND a.opened=la.latestTime
			WHERE a.Action = 'INSTALL'
		)
		SELECT A.cid, A.sid, B.name as device, B.old_name as dev_old_name, B.active as dev_active, 
		C.name as software, C.old_name as sw_old_name, C.active as sw_active 
		FROM CurrentInstalls A
		LEFT JOIN devices B ON A.cid=B.cid
		LEFT JOIN software C ON A.sid=C.sid
		WHERE b.cid`
	if cid < 1 {
		query += ">?"
	} else {
		query += "=?"
	}
	rows, err := Conn.Query(query, cid)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item SoftwareOnComputers
		var oldComputer string
		var oldSoftware string
		err := rows.Scan(&item.Cid, &item.Sid, &item.Dev_name, &oldComputer, &item.Dev_active, &item.Sw_name, &oldSoftware, &item.Sw_active)
		if err != nil {
			log.Println(err)
		} else {
			if item.Dev_active == 0 {
				item.Dev_name = oldComputer
			}
			if item.Sw_active == 0 {
				item.Sw_name = oldSoftware
			}
			items = append(items, item)
		}
	}
	err = rows.Err()
	return items, err
}

// Get the software list not being tracked
func GetOtherSoftware() ([]Sw_inv, error) {
	items := make([]Sw_inv, 0)
	query := `
		SELECT count(*) as cnt, A.name, B.active FROM sw_inv A
		LEFT JOIN devices B ON B.cid=A.cid
		WHERE sid IS NULL
		GROUP BY A.Name
		ORDER BY A.Name
	`
	rows, err := Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Sw_inv
		var active, cnt int
		err := rows.Scan(&cnt, &item.Name, &active)
		if err != nil {
			log.Println(err)
		} else {
			if active == 1 {
				item.CountActiveInstalls = cnt
			} else {
				item.CountDecomissionedInstalls = cnt
			}
			items = append(items, item)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return items, err
}

// Get a MAP of the tracked software (in the software table)
// And the counts of installs (from inventory and from action_log)
func GetTrackedSoftware() (map[int]Sw_inv, error) {
	items := make(map[int]Sw_inv)
	// Step 1: Get list of tracked software
	query := `
		SELECT sid, name, active, licenses, coalesce(old_name, "") old_name 
		FROM software
		ORDER BY Name
	`
	if err := fetchTrackedSoftware(items, query); err != nil {
		return items, err
	}
	// Step 2: Get count of actual installed software
	query = `
		SELECT B.sid, count(*) as cnt, C.active FROM sw_inv A 
		LEFT JOIN software B ON B.sid=A.sid
		LEFT join devices C ON C.cid=A.cid
		WHERE A.sid>0
		GROUP BY A.sid
	`
	if err := updateSoftwareCount(items, query); err != nil {
		return items, err
	}
	// Step 3: Get manually tracked software installs
	installed, err := GetManuallyTrackedSoftwareOnComputers(0)
	if err != nil {
		log.Println(err)
		return items, err
	}
	// Step 5: Count and update manually tracked software
	swCount := make(map[int]int)
	inactive := make(map[int]int)
	for _, item := range installed {
		if item.Dev_active == 1 {
			swCount[item.Sid] += 1
		} else {
			inactive[item.Sid] += 1
		}
	}
	for sid, cnt := range swCount {
		if item, exists := items[sid]; exists {
			item.Manual = cnt
			items[sid] = item
		}
	}
	for sid, cnt := range inactive {
		if item, exists := items[sid]; exists {
			item.Manual_inactive = cnt
			items[sid] = item
		}
	}
	return items, nil
}

// Fetch tracked software and populate the items map
func fetchTrackedSoftware(items map[int]Sw_inv, query string) error {
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var item Sw_inv
		var oldName string
		if err := rows.Scan(&item.Sid, &item.Name, &item.Active, &item.Licenses, &oldName); err != nil {
			log.Println(err)
			continue
		}
		if item.Active == 0 {
			item.Name = oldName
		}
		item.CountActiveInstalls = 0
		item.CountDecomissionedInstalls = 0
		items[item.Sid] = item
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Update the software count based on a query
func updateSoftwareCount(items map[int]Sw_inv, query string) error {
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var sid, cnt, active int
		if err := rows.Scan(&sid, &cnt, &active); err != nil {
			log.Println(err)
			continue
		}
		if item, exists := items[sid]; exists {
			if active == 1 {
				item.CountActiveInstalls = cnt
			} else {
				item.CountDecomissionedInstalls = cnt
			}
			items[sid] = item
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Set the pre-installed flag in the software inventory table: sw_inv
type PreInstalled struct {
	Id  int `json:"id"`
	Chk int `json:"chk"`
}

func SetPreInstalled(items []PreInstalled) string {
	if len(items) == 0 {
		return "okay"
	}
	query := "UPDATE sw_inv SET preinstalled=? WHERE id=?"
	for _, item := range items {
		_, err := Conn.Exec(query, item.Chk, item.Id)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	// Reset the pre_installed count in the software table: software
	query = `UPDATE software
	SET pre_installed = COALESCE(counts.cnt, 0)
	FROM (
		SELECT sid, COUNT(*) AS cnt
		FROM sw_inv
		WHERE sid > 0 AND preinstalled = 1
		GROUP BY sid
	) AS counts
	LEFT JOIN software s2 ON s2.sid = counts.sid
	WHERE software.sid = s2.sid OR counts.sid IS NULL
	`
	_, err := Conn.Exec(query)
	if err != nil {
		log.Println(err)
		return err.Error()
	}
	return "okay"
}
