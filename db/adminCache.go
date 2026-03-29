package db

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// Cache of admin table codes and icons
type adminStruct struct {
	sync.RWMutex
	theSlice []adminInfo
}

var AdminCache adminStruct

// A debounce mechanism
var reloadTimer *time.Timer
var reloadMutex sync.Mutex
var isReloading bool

type adminInfo struct {
	Id          int    `json:"id" db:"id"`                   // Row ID
	Field       string `json:"field" db:"field"`             // Field name
	Code        string `json:"code" db:"code"`               // Field code (put into various tables)
	Description string `json:"description" db:"description"` // What the user sees
	Seq         int    `json:"seq" db:"seq"`                 // Order of display
	Active      int    `json:"active" db:"active"`           // Still can be used or not
	Parent      string `json:"parent" db:"parent"`           // parent-child droplist inter-relationship
	Icon        string `json:"icon" db:"icon"`               // Icon for the choice
	Count       int    `json:"count" db:"cnt"`               // Count of how may times this field/code is used
	AssetId     string `json:"assetid" db:"asset_id"`        // Starting characters dependant on device type
	Permissions string `json:"permissions" db:"permissions"` // CRUD permssions per group
}

/*
	Okay, lets talk about this.
	The device search screen uses "SITE", "OFFICE", "TYPE", "UID", "GID"
	For searches, we should only show the values for these if they are actually in use.

	The devices table has foreign keys columns for "SITE", "OFFICE", "TYPE", "UID", "GID"
	User (UID) is from the profiles table
	Site, Office, Type and Group (GID) are from the choices (admin) table.

	Because of database referential integrity, the devices table cannot hold values for
	Site, Office, Type, and Group that are not in the admin table, so we do not
	need to worry about device fields that do not have a corresponding admin table entry.

	For device searching, we need to build droplists for inuse: Site, Office, Type and Group.
	When updating records, we need to include the not inuse as well (cnt=0, the complete list),
	but not inactive. That is why we have an count (cnt) field.
	If the cnt is > 0, then that field/code is in use in the devices table.

	The AdminCache need to be rebuilt when the admin table changes.

	Database triggers on the devices table keep the counts mostly accurate.
	Reload of the AdminCache also causes the counts to happen again.

	When a device record is created, modified, or deleted, then we may need to update the AdminCache
	to keep it accurate if any of the counts drop to zero or increase above zero.
	This is done with database triggers.

	This initialization/population function has a mutex to keep it thread safe.

*/

// Load the database table into a temporary slice,
// Then lock the real slice (AdminCache), assign the temporary slice to it, then release the lock
// Otherwise some reads would happen before the slice is fully populated
func (a *adminStruct) loadAdmin(count int) {
	tempSlice := make([]adminInfo, count)
	query := `
		SELECT A.id, A.field, A.code, A.description, A.seq, A.active, A.parent, 
		coalesce(B.icon,"") AS icon, coalesce(C.icon,"") AS alticon, A.cnt, A.asset_id, A.permissions
		FROM choices A 
		LEFT JOIN icons B ON A.code=B.name
		LEFT JOIN icons C ON A.field=C.name
		WHERE A.active=1 
		ORDER BY A.field, A.seq, A.description
	`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item adminInfo
		altIcon := ""
		err := rows.Scan(&item.Id, &item.Field, &item.Code, &item.Description,
			&item.Seq, &item.Active, &item.Parent, &item.Icon, &altIcon,
			&item.Count, &item.AssetId, &item.Permissions)
		if err != nil {
			log.Println(err)
			continue
		} else {
			if len(item.Icon) == 0 {
				item.Icon = altIcon
			}
			tempSlice = append(tempSlice, item)
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}
	// Lock the slice while we replace it
	a.Lock()
	defer a.Unlock()
	a.theSlice = tempSlice
}

func (a *adminStruct) resetCacheFlag() {
	_, err := Conn.Exec("UPDATE cache_dirty SET is_dirty = 0 WHERE id = 1")
	if err != nil {
		log.Println("Error resetting cache dirty flag:", err)
	}
}

func (a *adminStruct) getAdminCount() int {
	count := 155
	err := Conn.QueryRow("SELECT count(*) FROM choices").Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return count
}

// Reset the usage counters in the choices table and the profiles table
func (a *adminStruct) resetUsageCounts() {
	query := `
	UPDATE choices
	SET cnt = COALESCE((
		SELECT COUNT(*)
		FROM devices
		WHERE 
			(choices.field = 'TYPE' AND devices.type IS NOT NULL AND devices.type = choices.code) OR
			(choices.field = 'SITE' AND devices.site IS NOT NULL AND devices.site = choices.code) OR
			(choices.field = 'OFFICE' AND devices.office IS NOT NULL AND devices.office = choices.code) OR
			(choices.field = 'GROUP' AND devices.gid IS NOT NULL AND devices.gid = choices.code)
	), 0);
	`
	_, err := Conn.Exec(query)
	if err != nil {
		log.Println(err)
		return
	}

	query = `
		UPDATE profiles
		SET cnt = COALESCE((
			SELECT COUNT(*) 
			FROM devices 
			WHERE devices.uid = profiles.uid
		), 0);
	`
	_, err = Conn.Exec(query)
	if err != nil {
		log.Println(err)
		return
	}
}

// Available externally
func buildAdminCache() {
	// Prevent calling BuildAdmin() again from the reload timer
	reloadMutex.Lock()
	isReloading = true
	reloadMutex.Unlock()
	defer func() {
		reloadMutex.Lock()
		isReloading = false
		reloadMutex.Unlock()
	}()
	AdminCache.resetUsageCounts()
	count := AdminCache.getAdminCount()
	AdminCache.loadAdmin(count) // This function has the mutex lock, Read Blocking
	AdminCache.resetCacheFlag()
}

// Use a timer function to reload the cache, allowing for
// Many updates to happen before starting the cache reload
func scheduleCacheReload() {
	reloadMutex.Lock()
	defer reloadMutex.Unlock()
	if isReloading {
		return
	}
	if reloadTimer != nil {
		reloadTimer.Stop()
	}
	reloadTimer = time.AfterFunc(1000*time.Millisecond, buildAdminCache)
}

// Generic search for description given field and code
func GetCodeDescription(field string, code any) string {
	AdminCache.RLock()
	defer AdminCache.RUnlock()
	codeStr := fmt.Sprint(code)
	for _, item := range AdminCache.theSlice {
		if item.Field == field && item.Code == codeStr {
			return item.Description
		}
	}
	return ""
}

// Return the asset for a device type
func getAssetIdByDeviceType(deviceType string) string {
	for _, item := range AdminCache.theSlice {
		if item.Field == "TYPE" && item.Code == deviceType {
			return item.AssetId
		}
	}
	return deviceType
}

// Created a database trigger to look at the choices table counts
// and set the is_dirty flag.
// If count transitions from 0 to positive, or if count transitions
// from positive to zero, set is_dirty flag
func checkAdminCache() {
	var isDirty bool
	err := Conn.QueryRow("SELECT is_dirty FROM cache_dirty WHERE id = 1").Scan(&isDirty)
	if err != nil {
		log.Println("Error checking cache dirty flag:", err)
	}
	if isDirty {
		scheduleCacheReload() // Reload data from DB
	}
}

/* This is used to cache the droplists in the browser
 * so ajax calls are not necessary
 * The struct names are kept short to resuce the size of the JSON
 */

type FieldInfo struct {
	Code        string `json:"c"` // Code
	Description string `json:"d"` // Description
	Parent      string `json:"p"` // Parent
	Seq         int    `json:"s"` // Sequence
}

func BuildFieldList(field string) string {
	offices := readFromCache(field) //The adminCache is pre-sorted
	json, err := json.Marshal(offices)
	if err != nil {
		log.Panicln(err)
		return ""
	}
	return string(json)
}

// Use the admin table cache
func readFromCache(field string) []FieldInfo {
	items := make([]FieldInfo, 0)
	AdminCache.RLock()
	defer AdminCache.RUnlock()

	for _, item := range AdminCache.theSlice {
		if item.Field != field || item.Active != 1 {
			continue
		}
		var field FieldInfo
		field.Code = item.Code
		field.Description = item.Description
		field.Parent = item.Parent
		field.Seq = item.Seq
		items = append(items, field)
	}
	return items
}
