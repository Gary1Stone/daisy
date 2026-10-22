package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gbsto/daisy/aliases"
	"github.com/gbsto/daisy/cities"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/devices"
	"github.com/gbsto/daisy/emails"
	"github.com/gbsto/daisy/inventory"
	"github.com/gbsto/daisy/logins"

	"github.com/gbsto/daisy/wigle"
)

type DiskInfo struct {
	Drive     string  `json:"Drive"`     // Drive name ie: "C:", or "D:",...
	Total     int64   `json:"Total"`     // Total drive space in megabytes. GB := float64(DiskInfo.Total)/(1<<30)
	Free      int64   `json:"Free"`      // Free space in megabytes
	Used      int64   `json:"Used"`      // Used space in megabytes
	Fill      float64 `json:"Fill"`      // Percentage of disk used
	Cid       int     `json:"Cid"`       // Device ID
	Timestamp int     `json:"Timestamp"` // Unix UTC timestamp
	Localtime string  `json:"Localtime"` // Local time
}

type Wifi struct {
	Ssid  string `json:"Ssid"`  // Service Set Identifier is the unique name of a Wi-Fi network that identifies it to users and devices
	Bssid string `json:"Bssid"` // Basic Service Set Identifier is the unique physical address (MAC address) of a Wi-Fi access point
	Rssi  int    `json:"Rssi"`  // Received Signal Strength Indicator
}

type SysInfo struct {
	Hostname                  string     `json:"Hostname"`
	ApiKey                    string     `json:"ApiKey"`
	SoftwareList              []string   `json:"SoftwareList"`
	OS_Name                   string     `json:"OS_Name"`
	OS_Version                string     `json:"OS_Version"`
	OS_Manufacturer           string     `json:"OS_Manufacturer"`
	OS_Configuration          string     `json:"OS_Configuration"`
	OS_Build_Type             string     `json:"OS_Build_Type"`
	Registered_Owner          string     `json:"Registered_Owner"`
	Registered_Organization   string     `json:"Registered_Organization"`
	Product_ID                string     `json:"Product_ID"`
	Original_Install_Date     string     `json:"Original_Install_Date"`
	System_Boot_Time          string     `json:"System_Boot_Time"`
	System_Manufacturer       string     `json:"System_Manufacturer"`
	System_Model              string     `json:"System_Model"`
	System_Type               string     `json:"System_Type"`
	Processors                string     `json:"Processors"`
	BIOS_Version              string     `json:"BIOS_Version"`
	Windows_Directory         string     `json:"Windows_Directory"`
	System_Directory          string     `json:"System_Directory"`
	Boot_Device               string     `json:"Boot_Device"`
	System_Locale             string     `json:"System_Locale"`
	Input_Locale              string     `json:"Input_Locale"`
	Time_Zone                 string     `json:"Time_Zone"`
	Total_Physical_Memory     string     `json:"Total_Physical_Memory"`
	Available_Physical_Memory string     `json:"Available_Physical_Memory"`
	Virtual_Memory_Max_Size   string     `json:"Virtual_Memory_Max_Size"`
	Virtual_Memory_Available  string     `json:"Virtual_Memory_Available"`
	Virtual_Memory_In_Use     string     `json:"Virtual_Memory_In_Use"`
	Page_File_Locations       string     `json:"Page_File_Locations"`
	Domain                    string     `json:"Domain"`
	Logon_Server              string     `json:"Logon_Server"`
	Hotfixs                   string     `json:"Hotfixs"`
	Network_Cards             string     `json:"Network_Cards"`
	HyperV_Requirements       string     `json:"HyperV_Requirements"`
	Battery                   int        `json:"Battery"`
	CpuName                   string     `json:"CpuName"`
	Latitude                  float64    `json:"Latitude"`
	Longitude                 float64    `json:"Longitude"`
	IP_Address                string     `json:"Ip_Address"`
	Mac_Addresses             []string   `json:"Mac_Addresses"`
	WiFi_List                 []Wifi     `json:"Wifi_List"`
	Disk_Info                 []DiskInfo `json:"Disk_Info"`
}

func AddSoftwareList(cid int, sysInfo *SysInfo) error {
	oldList := make(map[string]int)
	var name string
	var id int
	var site, office string
	// get the computer site/office
	query := `SELECT C.description AS site, coalesce(E.description, '') AS office FROM devices D 
		LEFT JOIN choices C ON D.site = C.code AND C.field = 'SITE'
		LEFT JOIN choices E ON D.office = E.code AND E.field = 'OFFICE'
		WHERE D.cid = ?`
	err := db.Conn.QueryRow(query, cid).Scan(&site, &office)
	if err != nil {
		log.Println(err)
	}

	// Prepare query to fetch existing software records
	query = "SELECT name, id FROM sw_inv WHERE cid=?"
	rows, err := db.Conn.Query(query, cid)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()
	// Populate the oldList map with existing software entries
	for rows.Next() {
		if err := rows.Scan(&name, &id); err != nil {
			log.Println(err)
		}
		oldList[name] = id
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}
	// Determine which software to delete and which to insert
	var insertValues []any
	var newSwList []string
	for _, app := range sysInfo.SoftwareList {
		if _, exists := oldList[app]; exists {
			delete(oldList, app)
		} else {
			insertValues = append(insertValues, cid, app)
			newSwList = append(newSwList, app)
		}
	}
	// Any remaining oldlist software packages should be deleted.
	var deleteIDs []int
	for _, id := range oldList {
		deleteIDs = append(deleteIDs, id)
	}
	// Delete outdated software
	if len(deleteIDs) > 0 {
		idStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(deleteIDs)), ","), "[]")
		_, err := db.Conn.Exec(fmt.Sprintf("DELETE FROM sw_inv WHERE id IN (%s)", idStr))
		if err != nil {
			log.Println(err)
		}
	}
	// Insert new software entries
	if len(insertValues) > 0 {
		valuePlaceholders := strings.Repeat("(?, ?),", len(insertValues)/2)
		valuePlaceholders = valuePlaceholders[:len(valuePlaceholders)-1] // Remove trailing comma
		_, err := db.Conn.Exec("INSERT INTO sw_inv (cid, name) VALUES "+valuePlaceholders, insertValues...)
		if err != nil {
			log.Println(err)
		}
	}
	// Update software inventory
	if err := inventory.MatchSoftwareToInventory(0); err != nil {
		log.Println(err)
	}
	//Send new software notification to sysadmin
	if len(newSwList) > 0 {
		var e emails.Emails
		err := e.EmailNewSoftwareList(sysInfo.Hostname, site, office, newSwList)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

// Add the system info from the API to the device record
// These fields are not shared with the device fields
// used in the application
func UpdateSysInfo(cid int, sysInfo *SysInfo) error {
	curUid := db.SYS_PROFILE.Uid
	recordDeviceLocation(cid, sysInfo)
	recordDeviceWiFi(cid, sysInfo)
	db.AddDeviceLock(curUid, cid)
	query := `
		UPDATE devices SET 
		OS_Name=?, OS_Version=?, OS_Manufacturer=?, OS_Configuration=?, OS_Build_Type=?,
		Registered_Owner=?, Registered_Organization=?, Product_ID=?, Original_Install_Date=?,
		System_Boot_Time=?, System_Manufacturer=?, System_Model=?, System_Type=?, Processors=?,
		BIOS_Version=?, Windows_Directory=?, System_Directory=?, Boot_Device=?, System_Locale=?,
		Input_Locale=?, Time_Zone=?, Total_Physical_Memory=?, Available_Physical_Memory=?,
		Virtual_Memory_Max_Size=?, Virtual_Memory_Available=?, Virtual_Memory_In_Use=?,
		Page_File_Locations=?, Domain=?, Logon_Server=?, Hotfixs=?, Network_Cards=?,
		HyperV_Requirements=?, Battery=?, cpu=?, ram=?, last_audit=strftime('%s', 'now')
		WHERE cid=?
	`
	if !db.IsDeviceLocked(curUid, cid) {
		_, err := db.Conn.Exec(query, sysInfo.OS_Name, sysInfo.OS_Version, sysInfo.OS_Manufacturer, sysInfo.OS_Configuration, sysInfo.OS_Build_Type,
			sysInfo.Registered_Owner, db.ForeignKey(sysInfo.Registered_Organization), sysInfo.Product_ID, sysInfo.Original_Install_Date,
			sysInfo.System_Boot_Time, sysInfo.System_Manufacturer, sysInfo.System_Model, sysInfo.System_Type, sysInfo.Processors,
			sysInfo.BIOS_Version, sysInfo.Windows_Directory, sysInfo.System_Directory, sysInfo.Boot_Device, sysInfo.System_Locale,
			sysInfo.Input_Locale, sysInfo.Time_Zone, sysInfo.Total_Physical_Memory, sysInfo.Available_Physical_Memory,
			sysInfo.Virtual_Memory_Max_Size, sysInfo.Virtual_Memory_Available, sysInfo.Virtual_Memory_In_Use,
			sysInfo.Page_File_Locations, sysInfo.Domain, sysInfo.Logon_Server, sysInfo.Hotfixs, sysInfo.Network_Cards,
			sysInfo.HyperV_Requirements, sysInfo.Battery, sysInfo.CpuName,
			convertMemory(sysInfo.Total_Physical_Memory), cid)
		if err != nil {
			log.Println(err)
			return err
		}
		// Save year if missing by using BIOS year
		var d devices.Device
		dev, err := d.GetDevice(curUid, db.GetTzoff(curUid), cid)
		if err != nil {
			log.Println(err)
		}
		if dev.Year < 2000 {
			dev.Year = extractYearFromBios(sysInfo.BIOS_Version)
		}
		query = "UPDATE devices set year=? WHERE cid=?"
		_, err = db.Conn.Exec(query, dev.Year, cid)
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		log.Println("Device record was locked, cannot save record from audit API")
	}
	return nil
}

// Never saw this computer before, or the user changed the computer name
func ApiNewComputer(sysInfo *SysInfo) (int, error) {
	computer := new(devices.Device) // Allocate on heap (address of)
	computer.Name = sysInfo.Hostname
	computer.Year = extractYearFromBios(sysInfo.BIOS_Version)
	if sysInfo.Battery == 1 {
		computer.Type = devices.Laptop
	} else {
		computer.Type = devices.Desktop
	}
	computer.Active = 1
	computer.Make = sysInfo.System_Manufacturer
	computer.Model = sysInfo.System_Model
	computer.Ram = convertMemory(sysInfo.Total_Physical_Memory)
	computer.Status = "WORKING"
	computer.Site = "WKNC"
	// OS version
	computer.Os = apiGuessOS(sysInfo.OS_Name)
	// Manufacturer
	computer.Make = apiGuessMake(sysInfo.System_Manufacturer)
	if len(sysInfo.CpuName) > 0 {
		computer.Cpu = sysInfo.CpuName
	} else {
		computer.Cpu = sysInfo.Processors
	}
	computer.Notes = "Added By API"
	computer.Last_updated_by = db.SYS_PROFILE.Uid
	var d devices.Device
	success := d.AddDevice(computer)
	if !success {
		return computer.Cid, errors.New("error adding record")
	}
	return computer.Cid, nil
}

func apiGuessOS(osName string) string {
	db.AdminCache.RLock()
	defer db.AdminCache.RUnlock()
	osName = strings.ToUpper(osName)
	for _, item := range db.AdminCache.TheSlice {
		if item.Field == "OS" {
			code := strings.ToUpper(item.Code)
			description := strings.ToUpper(item.Description)
			if strings.Contains(osName, code) || strings.Contains(osName, description) {
				return item.Code
			}
		}
	}
	return "WIN11"
}

func apiGuessMake(manufacturer string) string {
	db.AdminCache.RLock()
	defer db.AdminCache.RUnlock()
	manufacturer = strings.ToUpper(manufacturer)
	for _, item := range db.AdminCache.TheSlice {
		if item.Field == "MAKE" {
			code := strings.ToUpper(item.Code)
			description := strings.ToUpper(item.Description)
			if strings.Contains(manufacturer, code) || strings.Contains(manufacturer, description) {
				return item.Code
			}
		}
	}
	return "OTHER"
}

// Convert string with "xxx MB" in it to GB
func convertMemory(input string) int {
	digits := regexp.MustCompile(`[^0-9]`).ReplaceAllString(input, "")
	if digits == "" {
		return 0
	}
	mb, err := strconv.Atoi(digits)
	if err != nil {
		return 0
	}
	gb := float64(mb) / 1024
	return int(math.Ceil(gb))
}

// Define the regex pattern to match a 4-digit year starting with 20
func extractYearFromBios(input string) int {
	re := regexp.MustCompile(`20\d{2}`)
	match := re.FindString(input)
	year, err := strconv.Atoi(match)
	if err != nil {
		year = 1999
	}
	return year
}

func recordDeviceLocation(cid int, sysInfo *SysInfo) {
	var dto logins.Logins
	var found = false

	// Look up to see is we have the lat and lon already
	longitude, latitude, err := getWiFiLocation(sysInfo)
	if err == nil {
		dto.Latitude = latitude
		dto.Longitude = longitude
		found = true
	} else {

		// Look up lat/lon from wigle.net
		for _, wifi := range sysInfo.WiFi_List {
			lat, lon, err := wigle.GetWiFiLocationFromWigle(wifi.Ssid, wifi.Bssid, sysInfo.Latitude, sysInfo.Longitude)
			if err != nil {
				fmt.Println(err)
			} else {
				dto.Latitude = lat
				dto.Longitude = lon
				saveWiFiLocation(wifi.Ssid, wifi.Bssid, wifi.Rssi, lat, lon, "wigle.net")
				found = true
				break
			}
		}

	}
	if !found {
		dto.Latitude = sysInfo.Latitude
		dto.Longitude = sysInfo.Longitude
	}

	var dto1 cities.Cities
	var dto2 cities.Communities
	curSinLat := math.Sin(dto.Latitude * math.Pi / 180)
	curCosLat := math.Cos(dto.Latitude * math.Pi / 180)
	curCosLon := math.Cos(dto.Longitude * math.Pi / 180)
	curSinLon := math.Sin(dto.Longitude * math.Pi / 180)
	var wg sync.WaitGroup
	wg.Add(1)
	go cities.GetCity(curSinLat, curCosLat, curCosLon, curSinLon, &dto1, &wg)
	wg.Add(1)
	go cities.GetCommunity(curSinLat, curCosLat, curCosLon, curSinLon, &dto2, &wg)
	wg.Wait()

	//	dto.City = dto1.City
	dto.City_id = dto1.City_id
	//	dto.State = dto1.State
	//	dto.Country = dto1.Country_code
	//	dto.Country_code = dto1.Country_code
	//	dto.Community = dto2.Community
	dto.Community_id = dto2.Community_id

	query := `
		INSERT INTO tracks (cid, longitude, latitude, ip, city_id, community_id) 
		VALUES (?,?,?,?,?,?)
		`
	_, err = db.Conn.Exec(query, cid, dto.Longitude, dto.Latitude, sysInfo.IP_Address,
		db.ForeignKey(dto.City_id), db.ForeignKey(dto.Community_id))
	if err != nil {
		log.Println(err)
	}
}

func recordDeviceWiFi(cid int, sysInfo *SysInfo) {
	if cid < 1 || len(sysInfo.WiFi_List) == 0 {
		return
	}
	query := "INSERT INTO tracks_wifi (cid, ssid, bssid, rssi) VALUES (?,?,?,?)"
	for _, wifi := range sysInfo.WiFi_List {
		_, err := db.Conn.Exec(query, cid, wifi.Ssid, strings.ToUpper(wifi.Bssid), wifi.Rssi)
		if err != nil {
			log.Println(err)
		}
	}
}

// Returns longitude, latitude for a known wifi access point
func getWiFiLocation(sysInfo *SysInfo) (float64, float64, error) {
	longitude := 0.0
	latitude := 0.0
	query := "SELECT longitude, latitude FROM wifi_locations WHERE ssid=? AND bssid=? LIMIT 1"
	for _, wifi := range sysInfo.WiFi_List {
		err := db.Conn.QueryRow(query, wifi.Ssid, wifi.Bssid).Scan(&longitude, &latitude)
		if err == nil {
			log.Printf("WiFi: %s Location found %v, %v", wifi.Ssid, longitude, latitude)
			return longitude, latitude, nil
		}
	}
	return longitude, latitude, errors.New("no matches")
}

func SetMacAddress(cid int, sysInfo *SysInfo) error {

	// If no macs, nothing to do.
	if len(sysInfo.Mac_Addresses) == 0 {
		return nil
	}

	// Start a new transaction
	tx, err := db.Conn.Begin()
	if err != nil {
		log.Printf("SetMacAddress: failed to begin transaction for cid %d: %v", cid, err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer a rollback in case anything fails.
	// If Commit succeeds, Rollback is a no-op.
	defer tx.Rollback()

	// Mark any macs with a new CID as being changed (set valid_from date = now)
	var newCid strings.Builder
	newCid.WriteString(`UPDATE device_mac SET valid_from=strftime('%s', 'now') WHERE mac IN (`)
	var quoted []string
	for _, mac := range sysInfo.Mac_Addresses {
		quoted = append(quoted, `"`+strings.ToUpper(strings.ReplaceAll(mac, "-", ":"))+`"`)
	}
	newCid.WriteString(strings.Join(quoted, ", "))
	newCid.WriteString(") AND cid != ?")
	log.Println(newCid.String())
	_, err = tx.Exec(newCid.String(), cid)
	if err != nil {
		log.Printf("SetMacAddress: failed to update old MAC addresses for new valid_from for cid changed %d: %v", cid, err)
		//	return fmt.Errorf("failed to update reused MAC addresses: %w", err)
	}

	// Prepare for bulk insert
	numMacs := len(sysInfo.Mac_Addresses)
	valueArgs := make([]any, 0, numMacs*2)
	for _, mac := range sysInfo.Mac_Addresses {
		mac = strings.ToUpper(strings.ReplaceAll(mac, "-", ":"))
		valueArgs = append(valueArgs, db.ForeignKey(cid)) // Use ForeignKey for cid
		valueArgs = append(valueArgs, mac)
	}

	// Construct the placeholder string e.g., "(?, ?),(?, ?)"
	valuePlaceholders := strings.Repeat("(?, ?),", numMacs)
	valuePlaceholders = valuePlaceholders[:len(valuePlaceholders)-1] // Remove trailing comma

	stmt := "INSERT INTO device_mac (cid, mac) VALUES " + valuePlaceholders + " ON CONFLICT(mac) DO UPDATE SET cid = excluded.cid"

	// Execute the bulk insert query within the transaction
	_, err = tx.Exec(stmt, valueArgs...)
	if err != nil {
		log.Printf("SetMacAddress: failed to insert/update new MAC addresses for cid %d: %v\n", cid, err)
		log.Println("QUESTION: Was a NIC card of WIFI adapter reused on a different computer?")
		return fmt.Errorf("failed to insert new MAC addresses: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("SetMacAddress: failed to commit transaction after inserting MACs for cid %d: %v", cid, err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update the mac table with the CID and aliaes table
	err = SetMacAliases(cid, sysInfo)
	if err != nil {
		log.Println(err)
	}

	return nil
}

// SetMacAliases associates a set of MAC addresses with a device, determines a master MAC,
// updates the database, and creates aliases for the other MACs.
func SetMacAliases(cid int, sysInfo *SysInfo) error {
	macs := sysInfo.Mac_Addresses
	if len(macs) == 0 {
		return nil
	}

	// Check if a master MAC already exists for this CID or if any of the incoming MACs are known.
	var masterMac string
	args := make([]any, 0, len(macs)+2)
	args = append(args, cid)
	for _, v := range macs {
		args = append(args, v)
	}
	args = append(args, cid)

	placeholders := strings.Repeat(",?", len(macs))
	if len(placeholders) > 0 {
		placeholders = placeholders[1:]
	}

	query := fmt.Sprintf("SELECT mac FROM macs WHERE cid=? OR mac IN (%s) ORDER BY CASE WHEN cid=? THEN 1 ELSE 0 END DESC, scanned DESC, mac ASC LIMIT 1", placeholders)
	err := db.Conn.QueryRow(query, args...).Scan(&masterMac)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error determining master MAC for cid %d: %v", cid, err)
	}

	// If still no master, pick the lexicographically smallest MAC as the master.
	if masterMac == "" {
		sort.Strings(macs)
		masterMac = macs[0]
	}

	// Use a single transaction to update the CIDs for all provided MAC addresses.
	tx, err := db.Conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for cid %d: %w", cid, err)
	}
	defer tx.Rollback()

	// Prepare arguments for a bulk UPDATE to avoid N+1 queries.
	updateArgs := make([]any, len(macs)+1)
	updateArgs[0] = cid
	places := make([]string, len(macs))
	for i, mac := range macs {
		updateArgs[i+1] = mac
		places[i] = "?"
	}

	updateStmt := fmt.Sprintf("UPDATE macs SET cid=? WHERE mac IN (%s) AND cid IS NULL", strings.Join(places, ","))
	if _, err := tx.Exec(updateStmt, updateArgs...); err != nil {
		return fmt.Errorf("failed to bulk update cids for macs: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for cid %d: %w", cid, err)
	}

	// --- Create aliases for the other MACs ---
	for _, mac := range macs {
		if mac != masterMac {
			var a aliases.Aliases
			if err := a.AddAliasPair(mac, masterMac); err != nil {
				return err
			}
		}
	}

	return nil
}

func saveWiFiLocation(SSID, BSSID string, RSSI int, latitude, longitude float64, source string) {
	query := "INSERT INTO wifi_locations (ssid, bssid, rssi, latitude, longitude, source) VALUES (?,?,?,?,?,?)"
	_, err := db.Conn.Exec(query, SSID, BSSID, RSSI, latitude, longitude, source)
	if err != nil {
		log.Println(err)
	}
}
