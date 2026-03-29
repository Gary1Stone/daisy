package db

import (
	"database/sql"
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gbsto/daisy/util"
)

type Device struct {
	Task              string `json:"task"`
	Cid               int    `json:"cid"`               //Computer ID
	Name              string `json:"name"`              //Computer Name
	Type              string `json:"type"`              //What type of device, desktop, laptop, tablet, printer, phone,... (pick list)
	Type_usr          string `json:"type_usr"`          //Display name for the type
	Site              string `json:"site"`              //What site the computer is in (pick list)
	Site_usr          string `json:"site_usr"`          //Display name for the Site
	Office            string `json:"office"`            //What office the computer is in (pick list)
	Office_usr        string `json:"office_usr"`        //Display name for the Office
	Location          string `json:"location"`          //Comment on the computer location
	Year              int    `json:"year"`              //Year of manufacture
	Make              string `json:"make"`              //Who made it: Dell, IBM, ASUS,...(pick list)
	Make_usr          string `json:"make_usr"`          //Display name for the make
	Model             string `json:"model"`             //What model it is
	Cpu               string `json:"cpu"`               //What CPU does it have
	Ram               int    `json:"ram"`               //How much RAM it has
	Drivetype         string `json:"drivetype"`         //What type of drive (fixed list)
	Drivetype_usr     string `json:"drivetype_usr"`     //Display name for the drive type
	Drivesize         int    `json:"drivesize"`         //How much storage the drive has
	Cd                int    `json:"cd"`                //Does it have a CD-ROM drive (yes/no)
	Notes             string `json:"notes"`             //Comments
	Cores             int    `json:"cores"`             //How many cores the CPU has
	Cores_usr         string `json:"cores_usr"`         //Display name for the cores
	Gpu               string `json:"gpu"`               //Graphics processing unit description
	Wifi              int    `json:"wifi"`              //Does it have WiFI (yes/no)
	Ethernet          int    `json:"ethernet"`          //Does it have Ethernet (yes/no)
	Usb               int    `json:"usb"`               //How many USB ports
	Uid               int    `json:"uid"`               //User ID of the person assigned/responsible for the computer
	Active            int    `json:"active"`            //It is the computer still in active inventory (yes/no)
	Last_updated_by   int    `json:"last_updated_by"`   //User ID of who last changed the record
	Last_updated_time string `json:"last_updated_time"` //Last time the record was updated
	Image             string `json:"image"`             //Name of the photo in the images directory for this computer
	Small_image       string `json:"small_image"`       //Small version of the image
	Color             string `json:"color"`             //Highlight on screen (NOT USED)
	Speed             int    `json:"speed"`             //How fast is the computer (NOT USED)
	Status            string `json:"status"`            //Working, broken, stolen... (pick list)
	Status_usr        string `json:"status_usr"`        //Display Name for the status
	Os                string `json:"os"`                //What Operating System (pick list)
	Serial_number     string `json:"serial_number"`     //Serial number of the computer
	Gid               int    `json:"gid"`               //ID of the group assigned the computer (pick list)
	Gid_usr           string `json:"gid_usr"`           //Display name for the Group
	Assigned          string `json:"assigned"`          //The full name of the person assigned the computer
	Icon              string `json:"icon"`              //The icon for the computer (tablet, desktop, laptop...)
	IsLate            bool   `json:"islate"`            //Has the computer been seen in the last 90 days
	IsMissing         bool   `json:"ismissing"`         //Has the computer been backed up in the last 90 days
	Lun               string `json:"lun"`               //Last Updated Name (full name of who last updated the reecord)
	Old_name          string `json:"old_name"`          //Original name of the computer, filled in when active set to 0
	Last_seen_date    string `json:"last_seen"`         //Date of when last seen (action_log)
	Last_seen_days    int    `json:"last_seen_days"`    //Days elapsed since last seen (action_log)
	Last_seen_by      string `json:"last_seen_by"`      //Who saw it (action_log)
	Last_backup_date  string `json:"last_backup"`       //Date of when the last backup was performed (action_log)
	Last_backup_days  int    `json:"last_backup_days"`  //Days elapsed since the last backup (action_log)
	Last_backup_by    string `json:"last_backup_by"`    //Who backed it up (action_log)
	IsBroken          bool   `json:"isbroken"`          //Is there a broken report on this device
	Aid               int    `json:"aid"`               //Used by postDevice
	ShowHistory       int    `json:"showhistory"`       //used by postDevice
}

type LastInfo struct {
	Cid       int    `json:"cid" db:"-"`            //Computer ID
	Last_date string `json:"last_seen_date" db:"-"` //Date (string) when last seen YYYY-MM-DD
	Last_days int    `json:"last_seen_days" db:"-"` //Days since last seen
	Last_by   string `json:"last_seen_by" db:"-"`   //Who (fullname) saw it
	Created   int    `json:"created" db:"-"`        //Max of When the record was created (most current)
}

func (D *Device) trim() {
	D.Name = strings.TrimSpace(D.Name)
	D.Location = strings.TrimSpace(D.Location)
	D.Notes = strings.TrimSpace(D.Notes)
	D.Gpu = strings.TrimSpace(D.Gpu)
	D.Serial_number = strings.TrimSpace(D.Serial_number)
}

// get one single device
func GetDevice(curUid, cid int) (Device, error) {
	devices, err := GetDevices(curUid, cid, 0)
	if err != nil {
		log.Println(err)
		return Device{}, err
	}
	if len(devices) == 0 {
		return Device{}, nil
	}
	addDeviceLock(curUid, cid)
	return *devices[0], err
}

// Get the device photo name ensuring it is in the /public/images directory
func GetImage(cid int) string {
	if cid <= 0 {
		return "missing.jpg"
	}
	img := ""
	query := "SELECT image FROM devices WHERE cid=?"
	err := Conn.QueryRow(query, cid).Scan(&img)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	photoMap := util.MapPhotos()
	if len(img) == 0 || !photoMap[img] {
		img = "missing.jpg"
	}
	return img
}

// Get the device's small photo name ensuring it is in the /public/images directory
func GetSmallImage(cid int) string {
	if cid <= 0 {
		return "missing-sm.jpg"
	}
	img := ""
	query := "SELECT image FROM devices WHERE cid=?"
	err := Conn.QueryRow(query, cid).Scan(&img)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	//Create name for the small image
	img = util.AddSuffixBeforeExtension(img, "-sm")
	photoMap := util.MapPhotos()
	if len(img) == 0 || !photoMap[img] {
		img = "missing.jpg"
	}
	return img
}

// Produce slice of devices
func GetDevices(curUid, cid, page int) ([]*Device, error) {
	whereClause := "WHERE A.active=1 AND A.cid>? "
	if cid > 0 {
		whereClause = "WHERE A.active=1 AND A.cid=? "
	}
	return readDeviceTable(curUid, page, whereClause, cid)
}

// Produce slice of devicesfunc
func GetDevicesByType(curUid int, devType string) ([]*Device, error) {
	params := []any{0}
	whereClause := "WHERE A.active=1 AND A.cid>? "
	if len(devType) > 0 {
		whereClause += "AND A.type=? "
		params = append(params, devType)
	}
	return readDeviceTable(curUid, -1, whereClause, params...)
}

// Produce list of devices assigned to a user
func GetAssignedDevices(curUid, uid int) ([]*Device, error) {
	whereClause := "WHERE A.active=1 AND A.uid=? "
	return readDeviceTable(curUid, -1, whereClause, uid)
}

// get the last backup or sighting dates/days and by whom
func getLastInfo(curUid int, seen bool) []*LastInfo {
	lastinfo := make([]*LastInfo, 0)
	tzoff := GetTzoff(curUid)
	var query strings.Builder
	query.WriteString(`
		SELECT 
			max(A.opened) as created, A.cid, coalesce(B.fullname, '') last_by, 
			strftime('%Y-%m-%d', A.opened - ?, 'unixepoch') AS last_date, 
			cast((strftime('%s', 'now') - opened) / 86400 AS INTEGER) AS last_days 
		FROM action_log A 
		LEFT JOIN profiles B ON B.uid=A.originator 
		LEFT JOIN devices C ON C.cid=A.cid 
		WHERE c.active=1 AND 
		`)
	if seen {
		query.WriteString("A.action IN ('BACKUP', 'SIGHTING') ")
	} else {
		query.WriteString("A.action='BACKUP' ")
	}
	query.WriteString("GROUP BY A.cid ")
	rows, err := Conn.Query(query.String(), tzoff)
	if err != nil {
		log.Println(err)
		return lastinfo
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var dto LastInfo
		err := rows.Scan(&dto.Created, &dto.Cid, &dto.Last_by, &dto.Last_date, &dto.Last_days)
		if err != nil {
			log.Println(err)
		} else {
			lastinfo = append(lastinfo, &dto)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
	}
	return lastinfo
}

// Get the device information, returning a pointer to the slice
func readDeviceTable(curUid, page int, whereClause string, params ...any) ([]*Device, error) {
	devices := make([]*Device, 0, 50)
	tzoff := GetTzoff(curUid)    //Time Zone Offest in minutes
	photoMap := util.MapPhotos() //Map (name,true/false) of photos in the /public/images directory
	late := GetLateBackups()
	broken := GetBrokenDevices()
	missing := GetMissingDevices()
	lastseen := getLastInfo(curUid, true)
	lastbackup := getLastInfo(curUid, false)
	var query strings.Builder
	query.WriteString(`
	SELECT 
		A.Cid, A.Name, A.Type, A.Site, A.Office, A.Location, A.Year, A.Make, 
		A.Model, A.Cpu, A.Ram, A.Drivetype, A.Drivesize, A.Cd, A.Notes, A.Cores, 
		A.Gpu, A.Wifi, A.Ethernet, A.Usb, COALESCE(A.Uid, 0) AS uid, A.Active, A.Last_updated_by,  
		strftime('%Y-%m-%d %H:%M', A.last_updated_time-?, 'unixepoch') AS updated, 
		A.Image, A.Speed, A.Status, A.Os, A.Serial_number, COALESCE(A.Gid, 0) AS Gid, 
		COALESCE(B.fullname, '') AS assigned, COALESCE(colours.color, '') AS color, 
		E.icon, COALESCE(E.fullname, '') AS lun, 
		COALESCE(F.description, '') As type_usr, 
		COALESCE(G.description, '') As office_usr,
		COALESCE(H.description, '') As site_usr,
		COALESCE(I.description, '') As status_usr,
		COALESCE(J.description, '') As make_usr,
		COALESCE(K.description, '') As usergroup,
		COALESCE(L.description, '') As drivetype_usr,
		COALESCE(M.description, '') As cores_usr
	FROM devices A  
		LEFT JOIN profiles B ON A.uid=B.uid 
		LEFT JOIN profiles E ON A.last_updated_by=E.uid 
		LEFT JOIN (
			SELECT C.cid, C.action, D.color FROM action_log C 
			LEFT JOIN icons D ON D.name=C.action 
			WHERE (C.cid_ack IS NULL OR C.cid_ack=0 OR C.cid_ack='') AND D.is_device=0 
			ORDER BY D.priority LIMIT 1) colours ON A.Cid=colours.cid 
		LEFT JOIN icons E on E.name=A.type
		LEFT JOIN choices F on F.field='TYPE' AND F.code=A.type
		LEFT JOIN (SELECT code, description FROM choices WHERE field='OFFICE' GROUP BY code) G ON A.office=G.code
		LEFT JOIN choices H on H.field='SITE' AND H.code=A.site
		LEFT JOIN choices I on I.field='STATUS' AND I.code=A.status
		LEFT JOIN choices J on J.field='MAKE' AND J.code=A.make
		LEFT JOIN choices K on K.field='GROUP' AND K.code=A.gid
		LEFT JOIN choices L on L.field='DRIVETYPE' AND L.code=A.drivetype
		LEFT JOIN choices M on M.field='CORES' AND M.code=A.cores
	`)
	query.WriteString(whereClause)
	query.WriteString("ORDER BY A.name, A.Year, A.type ")
	if page >= 0 {
		query.WriteString("LIMIT 50 OFFSET ")
		query.WriteString(strconv.Itoa(page * 50))
	}
	prependValues := []any{tzoff}
	params = append(prependValues, params...)

	//	log.Println(query.String())
	rows, err := Conn.Query(query.String(), params...)
	if err != nil {
		log.Println(err)
		if err == sql.ErrNoRows {
			return devices, nil
		}
		return devices, err
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var dto Device
		err := rows.Scan(&dto.Cid, &dto.Name, &dto.Type, &dto.Site, &dto.Office, &dto.Location,
			&dto.Year, &dto.Make, &dto.Model, &dto.Cpu, &dto.Ram, &dto.Drivetype, &dto.Drivesize,
			&dto.Cd, &dto.Notes, &dto.Cores, &dto.Gpu, &dto.Wifi, &dto.Ethernet, &dto.Usb, &dto.Uid, &dto.Active,
			&dto.Last_updated_by, &dto.Last_updated_time, &dto.Image, &dto.Speed, &dto.Status, &dto.Os,
			&dto.Serial_number, &dto.Gid, &dto.Assigned, &dto.Color, &dto.Icon, &dto.Lun,
			&dto.Type_usr, &dto.Office_usr, &dto.Site_usr, &dto.Status_usr, &dto.Make_usr,
			&dto.Gid_usr, &dto.Drivetype_usr, &dto.Cores_usr)
		if err != nil {
			log.Println(err)
		} else {
			//Determine if backup is late - over 90 days ago
			j := sort.SearchInts(late, dto.Cid)
			dto.IsLate = j < len(late) && late[j] == dto.Cid
			//Determine if not seen in 90 days
			j = sort.SearchInts(missing, dto.Cid)
			dto.IsMissing = j < len(missing) && missing[j] == dto.Cid
			//Determin if image is missing
			if len(dto.Image) == 0 || !photoMap[dto.Image] {
				dto.Image = "missing.jpg"
				dto.Small_image = "missing.jpg"
			}
			//Confirm the small version of the image exists
			dto.Small_image = util.AddSuffixBeforeExtension(dto.Image, "-sm")
			if len(dto.Small_image) == 0 || !photoMap[dto.Small_image] {
				dto.Small_image = "missing.jpg"
			}
			//Fill in user displayed descriptions for the codes
			dto.Last_seen_by, dto.Last_seen_date, dto.Last_seen_days = findLastInfo(dto.Cid, lastseen)
			dto.Last_backup_by, dto.Last_backup_date, dto.Last_backup_days = findLastInfo(dto.Cid, lastbackup)
			j = sort.SearchInts(broken, dto.Cid)
			dto.IsBroken = j < len(broken) && broken[j] == dto.Cid
			devices = append(devices, &dto)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
	}
	return devices, err
}

// Function to find last seen/backup information for a given CID
func findLastInfo(cid int, infoSlice []*LastInfo) (string, string, int) {
	for _, info := range infoSlice {
		if info.Cid == cid {
			return info.Last_by, info.Last_date, info.Last_days
		}
	}
	return "", "", -1
}

// Build list of CIDs (device/Computer IDs) not seen in the last 90 days
func GetMissingDevices() []int {
	var missing []int
	query := `
		SELECT cid FROM devices
		WHERE active = 1 AND (last_audit IS NULL OR last_audit < strftime('%s', datetime('now', '-90 day')))
		ORDER BY cid
		`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return missing
	}
	defer rows.Close()
	var cid int
	for rows.Next() {
		err := rows.Scan(&cid)
		if err != nil {
			log.Println(err)
		} else {
			missing = append(missing, cid)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
	}
	return missing
}

// Build list of CIDs (Computer IDs) not backed up in the last 90 days
func GetLateBackups() []int {
	var late []int
	timestamp := time.Now().AddDate(0, 0, -90).Unix()
	query := `
	SELECT d.cid
	FROM devices d
	LEFT JOIN action_log al
		ON d.cid = al.cid
		AND al.action = 'BACKUP'
		AND al.opened > ?
	WHERE d.active = 1
	AND (d.type = 'DESKTOP' OR d.type = 'LAPTOP')
	AND al.cid IS NULL
	ORDER BY d.cid;
	`
	rows, err := Conn.Query(query, timestamp)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return late
	}
	defer rows.Close()
	var cid int
	for rows.Next() {
		err := rows.Scan(&cid)
		if err != nil {
			log.Println(err)
		} else {
			late = append(late, cid)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
	}
	return late
}

// Check if the device name is already used
func IsUniqueDevice(cid int, name string) bool {
	cnt := 0
	query := "SELECT count(*) as cnt FROM devices WHERE name=? and cid<>? LIMIT 1 COLLATE NOCASE"
	err := Conn.QueryRow(query, name, cid).Scan(&cnt)
	if err != nil {
		log.Println(err)
	}
	return cnt == 0
}

// Save the device record
// TODO: move the note to the action log
func SetDevice(curUid int, dto *Device) bool {
	if curUid < 1 {
		curUid = SYS_PROFILE.Uid
	}
	if isDeviceLocked(curUid, dto.Cid) {
		log.Println("Record was updated by someone else before you tried to save.")
		return false
	}
	dto.trim()
	if !IsUniqueDevice(dto.Cid, dto.Name) {
		log.Println("Device name is not unique")
		return false
	}
	//Insure gid matches uid's group
	if dto.Uid > 0 && dto.Gid == 0 {
		dto.Gid = GetGid(dto.Uid)
	}
	var query = `
		UPDATE devices SET 
			name=?, type=?, site=?, office=?, location=?, year=?, make=?, model=?, cpu=?, 
			cores=?, ram=?, drivetype=?, drivesize=?, notes=?, gpu=?, cd=?, wifi=?, ethernet=?, usb=?,
			active=?, image=?, color=?, speed=?, uid=?, status=?, os=?, serial_number=?, 
			gid=?, last_updated_by=?, last_updated_time=strftime('%s','now') 
		WHERE cid=?
	`
	_, err := Conn.Exec(query, dto.Name, dto.Type, dto.Site, dto.Office, dto.Location, dto.Year,
		dto.Make, dto.Model, dto.Cpu, dto.Cores, dto.Ram, dto.Drivetype, dto.Drivesize, dto.Notes, dto.Gpu,
		dto.Cd, dto.Wifi, dto.Ethernet, dto.Usb, dto.Active, dto.Image, dto.Color, dto.Speed, foreignKey(dto.Uid), dto.Status, dto.Os,
		dto.Serial_number, dto.Gid, foreignKey(dto.Last_updated_by), dto.Cid)
	if err != nil {
		log.Println(err)
		return false
	}
	checkAdminCache()
	return true
}

// Delete device if not used in action log, else mark as not active
func MarkDeviceAsDeleted(curUid, cid int) bool {
	//If any actionlog items, do not delete
	if !IsDeletableDevice(cid) {
		return false
	}
	// Clean-up software inventory
	err := cleanupSwInv(cid)
	if err != nil {
		log.Println(err)
	}
	// Clean-up backups
	err = cleanupBackups(cid)
	if err != nil {
		log.Println(err)
	}

	// TODO: Clean-up others
	// Clean-up Action Log
	// Clean-up Tickets
	// Clean-up Checkins

	if isArchiveableDevice(cid) {
		// Mark device as inactive
		//	query := "UPDATE devices SET active=0, old_name=name, name=cid, last_updated_time=strftime('%s','now'), last_updated_by=? WHERE cid=?"
		query := "UPDATE devices SET active=0, old_name=name, last_updated_time=strftime('%s','now'), last_updated_by=? WHERE cid=?"
		_, err := Conn.Exec(query, curUid, cid)
		if err != nil {
			log.Println(err)
			return false
		}
	} else {
		// Really delete the device record since it hasn't been used anywhere
		query := "DELETE FROM devices WHERE cid=?"
		_, err = Conn.Exec(query, cid)
		if err != nil {
			log.Println(err)
			return false
		}
	}
	return true
}

func cleanupSwInv(cid int) error {
	// Delete all software inventory for this device
	query := "DELETE FROM sw_inv WHERE cid=?"
	_, err := Conn.Exec(query, cid)
	if err != nil {
		log.Println(err)
		return err
	}

	// General cleanup, delete all software inventory where there is no longer an active device
	query = "DELETE FROM sw_inv WHERE NOT EXISTS (SELECT 1 FROM devices d WHERE d.cid=sw_inv.cid AND d.active=1)"
	_, err = Conn.Exec(query)
	if err != nil {
		log.Println(err)
		return err
	}

	// Update action log, setting all installed/tracked software for this device to be marked as uninstalled by this user
	//TODO: Should we really mark the manually tracked software as removed? the license counts should take into account the active flag on the device
	// query = "INSERT INTO ..."
	// _, err = Conn.Exec(query, curUid, cid)
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }

	return nil
}

func cleanupBackups(cid int) error {
	query := "DELETE FROM backups WHERE cid=?"
	_, err := Conn.Exec(query, cid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Check if the device has open action log items
func IsDeletableDevice(cid int) bool {
	cnt := 0
	query := "SELECT count(*) as cnt FROM action_log WHERE closed=0 AND cid=? AND action NOT IN ('GIVING', 'SIGHTING')"
	err := Conn.QueryRow(query, cid).Scan(&cnt)
	if err != nil {
		log.Println(err)
		return false
	}
	return cnt == 0
}

// Check if the device has open or closed action log items
func isArchiveableDevice(cid int) bool {
	cnt := 0
	query := "SELECT count(*) as cnt FROM action_log WHERE cid=?"
	err := Conn.QueryRow(query, cid).Scan(&cnt)
	if err != nil {
		log.Println(err)
		return false
	}
	return cnt > 0
}

// Insert device record
func AddDevice(dto *Device) bool {
	dto.trim()
	if !IsUniqueDevice(dto.Cid, dto.Name) {
		log.Println("Device Name was not unique")
		return false
	}
	var query = `
	INSERT INTO devices 
		(name, type, site, office, location, year, make, model, cpu, cores, ram, 
		drivetype, drivesize, notes, gpu, cd, wifi, ethernet, usb, active, image, color, speed, 
		uid, status, os, serial_number, gid, last_updated_by, old_name, last_updated_time) 
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'',strftime('%s','now'))
		`
	result, err := Conn.Exec(query, dto.Name, dto.Type, dto.Site, dto.Office, dto.Location, dto.Year,
		dto.Make, dto.Model, dto.Cpu, dto.Cores, dto.Ram, dto.Drivetype, dto.Drivesize, dto.Notes, dto.Gpu,
		dto.Cd, dto.Wifi, dto.Ethernet, dto.Usb, dto.Active, dto.Image, dto.Color, dto.Speed, foreignKey(dto.Uid),
		dto.Status, dto.Os, dto.Serial_number, dto.Gid, foreignKey(dto.Last_updated_by))
	if err != nil {
		log.Println(err)
		return false
	}
	//Get the CID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		dto.Cid = 0
		return false
	}
	dto.Cid = int(lastInsertID)
	checkAdminCache()
	return true
}

func GetBrokenDevices() []int {
	var list []int
	query := "SELECT distinct(cid) FROM action_log WHERE active=1 AND cid NOT NULL AND action IN ('BROKEN', 'CARE', 'LOST', 'DIED', 'REQUEST') ORDER BY cid"
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return list
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		err := rows.Scan(&cid)
		if err != nil {
			log.Println(err)
		} else {
			list = append(list, cid)
		}
	}
	return list
}

// Used by the api to verify the computer name
func GetCidByName(name string) (int, error) {
	cid := 0
	query := "SELECT cid FROM devices WHERE name=? LIMIT 1"
	err := Conn.QueryRow(query, name).Scan(&cid)
	return cid, err
}

type User2DeviceReport struct {
	Uid        int
	User       string
	Fullname   string
	Cid        int
	DeviceName string
	Icon       string
}

// For the who has what computer report
func ListUsersDevices() ([]User2DeviceReport, error) {
	var list []User2DeviceReport
	query := `
		SELECT A.cid, A.name, B.icon, C.uid, C.user, C.fullname FROM devices A
		LEFT JOIN icons B ON A.type=B.name
		LEFT JOIN profiles C ON A.uid=C.uid
		WHERE A.active=1 AND B.is_device=1 AND A.uid IS NOT NULL
		ORDER BY C.fullname
	`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return list, err
	}
	defer rows.Close()
	for rows.Next() {
		var item User2DeviceReport
		err := rows.Scan(&item.Cid, &item.DeviceName, &item.Icon, &item.Uid, &item.User, &item.Fullname)
		if err != nil {
			log.Println(err)
		} else {
			list = append(list, item)
		}
	}
	return list, nil
}

func SearchDevices(curUid int, filter *DeviceFilter) ([]*Device, error) {
	var params []any
	var where strings.Builder
	where.WriteString("WHERE A.active=1 ")
	if len(filter.DevType) > 0 {
		params = append(params, filter.DevType)
		where.WriteString("AND A.type=? ")
	}
	if len(filter.Site) > 0 {
		params = append(params, filter.Site)
		where.WriteString("AND A.site=? ")
	}
	if len(filter.Office) > 0 {
		params = append(params, filter.Office)
		where.WriteString("AND A.office=? ")
	}
	if filter.Gid > 0 {
		params = append(params, filter.Gid)
		where.WriteString("AND A.gid=? ")
	}
	if filter.Uid > 0 {
		params = append(params, filter.Uid)
		where.WriteString("AND A.uid=? ")
	}

	if len(filter.SearchTxt) > 0 {
		txt := "%" + filter.SearchTxt + "%"
		where.WriteString(`
			AND ( 
			A.name COLLATE NOCASE LIKE ? 
			OR A.year COLLATE NOCASE LIKE ? 
			OR A.model COLLATE NOCASE LIKE ? 
			OR A.status COLLATE NOCASE LIKE ? 
			OR location COLLATE NOCASE LIKE ? 
			OR usergroup COLLATE NOCASE LIKE ? 
			OR assigned COLLATE NOCASE LIKE ? `)
		params = append(params, txt)
		params = append(params, txt)
		params = append(params, txt)
		params = append(params, txt)
		params = append(params, txt)
		params = append(params, txt)
		params = append(params, txt)

		// If the user is searching through text, also search the select lists (choices)
		AdminCache.RLock()
		defer AdminCache.RUnlock()
		for _, item := range AdminCache.theSlice {
			if item.Active == 1 && isDeviceField(item.Field) &&
				strings.Contains(strings.ToLower(item.Description), strings.ToLower(filter.SearchTxt)) {
				where.WriteString("OR A.")
				if item.Field == "GROUP" {
					where.WriteString("gid")
				} else {
					where.WriteString(item.Field)
				}
				where.WriteString("=? ")
				params = append(params, filter.SearchTxt)
			}
		}
		where.WriteString(") ")
	}
	items, err := readDeviceTable(curUid, filter.Page, where.String(), params...)
	if filter.IsLate || filter.IsMissing {
		items = filterDevices(items, filter.IsLate, filter.IsMissing)
	}
	return items, err
}

// TODO: We could move this into readDeviceTable to keep memory usage low
func filterDevices(items []*Device, isLate, isMissing bool) []*Device {
	result := items[:0] // Keep the same underlying array
	for _, item := range items {
		if (isLate && item.IsLate) || (isMissing && item.IsMissing) {
			result = append(result, item)
		}
	}
	return result
}

// TODO: Can we get this from CHOICES instead of hardcoding it?
func isDeviceField(field string) bool {
	devicefields := []string{"MAKE", "SITE", "TYPE", "OFFICE", "STATUS", "GROUP"}
	field = strings.ToUpper(field)
	for _, devField := range devicefields {
		if devField == field {
			return true
		}
	}
	return false
}

func getPreviousName(cid int, deviceType string) string {
	assetId := ""
	query := "SELECT name FROM devices WHERE cid=? AND type=?"
	err := Conn.QueryRow(query, cid, deviceType).Scan(&assetId)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return assetId
}

// If the device types changes, modfify the device name (assetId-xxx)
func GetNextAssetId(cid int, deviceType string) (string, error) {
	// if the device exists in the database and has the same device type, return the existing asset id
	if cid > 0 {
		assetId := getPreviousName(cid, deviceType)
		if len(assetId) > 0 {
			return assetId, nil
		}
	}
	// Generate a new asset ID
	var maxAssetId sql.NullString
	prefix := getAssetIdByDeviceType(deviceType) + "-"
	query := "SELECT MAX(name) FROM devices WHERE name LIKE ?"

	// Loop until a unique ID is found, or 25 tries exceeded
	i := 0
	for {
		err := Conn.QueryRow(query, prefix+"%").Scan(&maxAssetId)
		if err != nil && err != sql.ErrNoRows {
			return "", err
		}

		// If no max ID found, start with the first ID
		if !maxAssetId.Valid {
			return prefix + "01", nil
		}

		// Extract the integer part from the maxAssetId
		integerPart := strings.TrimPrefix(maxAssetId.String, prefix)
		maxId, err := strconv.Atoi(integerPart)
		if err != nil {
			return "", err
		}
		maxId += 1
		assetId := prefix + strconv.Itoa(maxId)
		if maxId < 10 {
			assetId = prefix + "0" + strconv.Itoa(maxId)
		}

		// Confirm new ID is unique
		cnt := 0
		query = "SELECT count(cid) FROM devices WHERE name=?"
		err = Conn.QueryRow(query, assetId).Scan(&cnt)
		if err != nil {
			return "", err
		}
		if cnt == 0 {
			return assetId, nil
		}

		// Limit the number of tries
		i++
		if i > 25 {
			return "", errors.New("unable to generate asset id")
		}
	}
}

// Find a device's assigned user and group
func GetDeviceAssignedUserGroup(curUid, cid int) (int, int) {
	uid := 0
	gid := 0
	query := "SELECT coalesce(uid, 0) AS uid, coalesce(gid, 0) AS gid FROM devices WHERE cid=?"
	err := Conn.QueryRow(query, cid).Scan(&uid, &gid)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return uid, gid
}

// Find a device's assigned user
func GetDeviceAssignedSiteOffice(curUid, cid int) (string, string) {
	site := ""
	office := ""
	if cid < 1 {
		return site, office
	}
	query := "SELECT coalesce(site, '') AS site, coalesce(office, '') AS office FROM devices WHERE cid=?"
	err := Conn.QueryRow(query, cid).Scan(&site, &office)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return site, office
}

func SetAuditCheckin(cid int) error {
	if cid <= 0 {
		return errors.New("invalid cid")
	}
	query := "UPDATE devices SET last_audit=strftime('%s', 'now') WHERE cid=?"
	_, err := Conn.Exec(query, cid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type DevicesMeta struct {
	Cid   int    `json:"cid"`
	Name  string `json:"name"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Type  string `json:"type"`
	Icon  string `json:"icon"`
}

func GetDeviceMeta() (map[int]DevicesMeta, error) {
	items := make(map[int]DevicesMeta)
	query := `
		SELECT D.cid, D.name, coalesce(C.description, ''), coalesce(D.model, ''),
		D.type, coalesce(I.icon, 'mif-devices')
		FROM devices D
		LEFT JOIN icons I ON D.type=I.name
		LEFT JOIN choices C ON D.make=C.code AND C.field='MAKE'
		WHERE D.active=1 
	`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item DevicesMeta
		err := rows.Scan(&item.Cid, &item.Name, &item.Make, &item.Model, &item.Type, &item.Icon)
		if err != nil {
			log.Println(err)
		} else {
			items[item.Cid] = DevicesMeta{
				Cid:   item.Cid,
				Name:  item.Name,
				Make:  item.Make,
				Model: item.Model,
				Type:  item.Type,
				Icon:  item.Icon,
			}
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
	}
	return items, nil
}

// name is unique in the devices table
func GetDevicesByNames(hostnames []string) (map[string]Device, error) {
	devices := make(map[string]Device)
	placeholders := make([]string, len(hostnames))
	args := make([]any, len(hostnames))

	for i, hostname := range hostnames {
		placeholders[i] = "?"
		args[i] = hostname
	}

	query := `
		SELECT D.cid, D.name, D.type, D.site, D.office, D.location, coalesce(P.fullname, '')
		FROM devices D
		LEFT JOIN profiles P ON D.uid = P.uid
		WHERE D.name IN (` + strings.Join(placeholders, ",") + `)
	`
	rows, err := Conn.Query(query, args...)
	if err != nil {
		log.Println(err)
		return devices, err
	}

	defer rows.Close()
	for rows.Next() {
		var dto Device
		err := rows.Scan(&dto.Cid, &dto.Name, &dto.Type, &dto.Site, &dto.Office, &dto.Location, &dto.Assigned)
		if err != nil {
			log.Println(err)
		} else {
			devices[dto.Name] = dto
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
	}
	return devices, err
}
