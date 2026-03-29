package db

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/devices"
	"github.com/gbsto/daisy/web/wizards"
)

type DroplistOption struct {
	Value       string
	Description string
	Icon        string
	Colour      string
	Selected    bool
}

// Use the choices (admin table) cache
func GetOptions(field, selected, parentCode string, withBlank bool) []DroplistOption {
	AdminCache.RLock()
	defer AdminCache.RUnlock()
	isSearch := strings.Contains(field, "SEARCH")
	field = strings.TrimSuffix(field, "SEARCH")
	field = strings.TrimSuffix(field, "INFORM")
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	for _, item := range AdminCache.theSlice {
		if item.Field != field ||
			item.Active != 1 ||
			item.Parent != parentCode ||
			(isSearch && item.Count <= 0) {
			continue
		}
		var option DroplistOption
		option.Value = item.Code
		option.Description = item.Description
		option.Selected = false
		if item.Code == selected {
			option.Selected = true
		}
		option.Icon = item.Icon
		option.Colour = ""
		options = append(options, option)
	}
	return options
}

// Get the type of device droplist
// list is dependant on wizard AND if in use
// Only shows computers for certian wizards
func GetTypeOptions(field, selected, wizard string, withBlank bool) []DroplistOption {
	AdminCache.RLock()
	defer AdminCache.RUnlock()
	isSearch := strings.Contains(field, "SEARCH")
	field = strings.TrimSuffix(field, "SEARCH")
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	whatToShow := make(map[string]bool)
	if wizard == wizards.Backup ||
		wizard == wizards.Install ||
		wizard == wizards.Remove ||
		wizard == wizards.Request {
		whatToShow[devices.Desktop] = true
		whatToShow[devices.Laptop] = true
	} else {
		for _, item := range AdminCache.theSlice {
			if item.Field == field && item.Active == 1 {
				whatToShow[item.Code] = true
			}
		}
	}
	for _, item := range AdminCache.theSlice {
		if item.Field != field ||
			item.Active != 1 ||
			!whatToShow[item.Code] ||
			(isSearch && item.Count <= 0) {
			continue
		}
		var option DroplistOption
		option.Value = item.Code
		option.Description = item.Description
		option.Selected = false
		if item.Code == selected {
			option.Selected = true
		}
		option.Icon = item.Icon
		option.Colour = ""
		options = append(options, option)
	}
	return options
}

// Okay, this selects the groups from the profiles table, given the user id in the devices table
// So if the group is blank, but the user is filled in, then include that group,
// But also include all the groups from the devices table.
// The above non-query includes all the groups from the devices table only.
// This should be okay, because we force the user to pick a group before selecting a user.

// Droplist population
// Fetch list of users (uid, name) for the given group from the profiles table.
// If the specified UID is no longer active, include it. If it
// But only show it if the uid's group is correct
// This prevents saving user to wrong group
//
// TODO: Limit the user's group selection to only ones used in the devices table.
func GetUserList(field string, selected, parentCode string, withBlank bool) []DroplistOption {
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	isSearch := strings.Contains(field, "SEARCH")
	field = strings.TrimSuffix(field, "SEARCH")
	field = strings.TrimSuffix(field, "INFORM")
	icon := GetIcon(field)
	uid, gid := toInt(selected, parentCode)
	if uid > 0 {
		var item DroplistOption
		var myquery strings.Builder
		myquery.WriteString("SELECT uid, fullname FROM profiles WHERE uid=? AND gid=? ")
		if isSearch {
			myquery.WriteString("AND cnt>0 ")
		}
		err := Conn.QueryRow(myquery.String(), uid, gid).Scan(&item.Value, &item.Description)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
		} else {
			item.Selected = true
			item.Icon = icon
			if len(item.Value) > 0 { // Prevent empty entries
				options = append(options, item)
			}
		}
	}
	var query strings.Builder
	query.WriteString("SELECT uid, fullname FROM profiles WHERE active=1 AND gid=? AND uid<>? ")
	if isSearch {
		query.WriteString("AND cnt>0 ")
	}
	query.WriteString("ORDER BY fullname")
	rows, err := Conn.Query(query.String(), gid, uid)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var item DroplistOption
		err := rows.Scan(&item.Value, &item.Description)
		if err != nil {
			log.Println(err)
		} else {
			item.Icon = icon
			options = append(options, item)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return options
}

func toInt(userId, groupId string) (int, int) {
	uid, err := strconv.Atoi(userId)
	if err != nil {
		uid = 0
	}
	gid, err := strconv.Atoi(groupId)
	if err != nil {
		gid = 0
	}
	return uid, gid
}

func GetSoftwareList(withBlank bool) []DroplistOption {
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	icon := GetIcon("SOFTWARE")
	var query strings.Builder
	query.WriteString("SELECT sid, name FROM software WHERE active=1 ORDER BY name")
	rows, err := Conn.Query(query.String())
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var item DroplistOption
		err := rows.Scan(&item.Value, &item.Description)
		if err != nil {
			log.Println(err)
		} else {
			item.Icon = icon
			options = append(options, item)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return options
}

// FindGroupID attempts to resolve the appropriate group ID
// based on the current user ID, device ID, and the defaultToCurUser flag.
func FindGroupID(curUid, cid int, defaultToCurUser bool) int {
	AdminCache.RLock()
	defer AdminCache.RUnlock()
	if defaultToCurUser {
		return GetGid(curUid)
	}
	if cid == 0 {
		for _, item := range AdminCache.theSlice {
			if item.Field == "GROUP" && item.Active == 1 {
				gid, _ := strconv.Atoi(item.Code)
				return gid
			}
		}
	}
	uid, gid := GetDeviceAssignedUserGroup(curUid, cid)
	if gid > 0 {
		return gid
	}
	if uid > 0 {
		return GetGid(uid)
	}
	return gid
}

// Find the first group in the group select list to use as the default group for the user select list
func GetDefaultGroup() string {
	for _, item := range AdminCache.theSlice {
		if item.Field == "GROUP" && item.Active == 1 && item.Parent == "" {
			return item.Code
		}
	}
	return ""
}

// Use the choices (admin table) cache
func GetKindOptions(field, selected, parentCode string, withBlank bool) []DroplistOption {
	AdminCache.RLock()
	defer AdminCache.RUnlock()
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	for _, item := range AdminCache.theSlice {
		if item.Field != field || item.Active != 1 {
			continue
		}
		var option DroplistOption
		option.Value = item.Code
		option.Description = item.Description
		option.Selected = false
		if item.Code == selected {
			option.Selected = true
		}
		option.Icon = item.Icon
		option.Colour = ""
		options = append(options, option)
	}
	return options
}

func GetMidOptions(field, selected, parentCode string, withBlank bool) []DroplistOption {
	query := `SELECT M.mid, M.name, M.source FROM macs M
		JOIN (
			SELECT DISTINCT mac FROM onlinehistory
		) O ON O.mac = M.mac
		WHERE M.active = 1
		ORDER BY M.name`
	rows, err := Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	for rows.Next() {
		var item DroplistOption
		var source string
		err := rows.Scan(&item.Value, &item.Description, &source)
		if err != nil {
			log.Println(err)
		} else {
			item.Selected = false
			if item.Value == selected {
				item.Selected = true
			}
			if source == "M30_Guest" {
				item.Description += " (Last detected on the guest network)"
			}
			options = append(options, item)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return options
}
