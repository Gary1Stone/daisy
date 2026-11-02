package droplists

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/icons"
)

type DroplistOption struct {
	Value       string
	Description string
	Icon        string
	Colour      string
	Tip         string
	Selected    bool
}

// Use the choices (admin table) cache
func GetOptions(field, selected, parentCode string, withBlank bool) []DroplistOption {
	db.AdminCache.RLock()
	defer db.AdminCache.RUnlock()
	isSearch := strings.Contains(field, "SEARCH")
	field = strings.TrimSuffix(field, "SEARCH")
	field = strings.TrimSuffix(field, "INFORM")
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	for _, item := range db.AdminCache.TheSlice {
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
		option.Colour = item.Color
		options = append(options, option)
	}
	return options
}

// Use the choices (admin table) cache
func GetSiteOptions(field, selected string, withBlank bool) []DroplistOption {
	db.AdminCache.RLock()
	defer db.AdminCache.RUnlock()
	isSearch := strings.Contains(field, "SEARCH")
	field = strings.TrimSuffix(field, "SEARCH")
	field = strings.TrimSuffix(field, "INFORM")
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	for _, item := range db.AdminCache.TheSlice {
		if item.Field != field ||
			item.Active != 1 ||
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
		option.Colour = item.Color
		options = append(options, option)
	}
	return options
}

// Get the type of device droplist
// list is dependant on wizard AND if in use
// Only shows computers for certian wizards
func GetTypeOptions(field, selected, wizard string, withBlank bool) []DroplistOption {
	db.AdminCache.RLock()
	defer db.AdminCache.RUnlock()
	isSearch := strings.Contains(field, "SEARCH")
	field = strings.TrimSuffix(field, "SEARCH")
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	whatToShow := make(map[string]bool)

	const (
		Sighting = "SIGHTING"
		Using    = "USING"
		Claiming = "CLAIMING"
		Giving   = "GIVING"
		Broken   = "BROKEN"
		Lost     = "LOST"
		Died     = "DIED"
		Care     = "CARE"
		Backup   = "BACKUP"
		Install  = "INSTALL"
		Remove   = "REMOVE"
		Request  = "REQUEST"
		Desktop  = "DESKTOP"
		Laptop   = "LAPTOP"
		Network  = "NETWORK"
		Printer  = "PRINTER"
		Video    = "VIDEO"
		Audio    = "AUDIO"
		Tablet   = "TABLET"
		Phone    = "PHONE"
	)
	if len(wizard) > 0 && (wizard == Backup ||
		wizard == Install ||
		wizard == Remove ||
		wizard == Request) {
		whatToShow[Desktop] = true
		whatToShow[Laptop] = true
	} else {
		for _, item := range db.AdminCache.TheSlice {
			if item.Field == field && item.Active == 1 {
				whatToShow[item.Code] = true
			}
		}
	}
	for _, item := range db.AdminCache.TheSlice {
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
// If the specified UID is no longer active, include it.
// But only show it if the uid's group is correct
// This prevents saving user to wrong group
func GetUserList(field string, selected, parentCode string, withBlank bool) []DroplistOption {
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	field = strings.TrimSuffix(field, "SEARCH")
	field = strings.TrimSuffix(field, "INFORM")
	icon := icons.FindIconNameByName(field)
	uid, gid := toInt(selected, parentCode)
	if uid > 0 {
		var query strings.Builder
		query.WriteString(`SELECT uid, fullname FROM profiles WHERE gid=? AND (uid=? OR (active = 1 AND uid <> ?)) ORDER BY fullname`)
		rows, err := db.Conn.Query(query.String(), gid, uid, uid)
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
				if item.Value == strconv.Itoa(uid) {
					item.Selected = true
				}
				options = append(options, item)
			}
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}
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
	icon := icons.FindIconNameByName("SOFTWARE")
	var query strings.Builder
	query.WriteString("SELECT sid, name FROM software WHERE active=1 ORDER BY name")
	rows, err := db.Conn.Query(query.String())
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
func FindGroupID(curUid, usr_gid, cid, dev_uid, dev_gid int, defaultToCurUser bool) int {
	db.AdminCache.RLock()
	defer db.AdminCache.RUnlock()
	if defaultToCurUser {
		return usr_gid
	}

	// If no computer, return the first available active group
	if cid == 0 {
		for _, item := range db.AdminCache.TheSlice {
			if item.Field == "GROUP" && item.Active == 1 {
				gid, err := strconv.Atoi(item.Code)
				if err != nil {
					return 0
				}
				return gid
			}
		}
	}

	// If the device has an assigned group, use that
	if dev_gid > 0 {
		return dev_gid
	}

	// if the user assigned to the device, use their group ID
	if dev_uid > 0 {
		return usr_gid
	}
	return 0 // No group default
}

// Find the first group in the group select list to use as the default group for the user select list
func GetDefaultGroup() string {
	for _, item := range db.AdminCache.TheSlice {
		if item.Field == "GROUP" && item.Active == 1 && item.Parent == "" {
			return item.Code
		}
	}
	return ""
}

// Use the choices (admin table) cache
func GetKindOptions(field, selected, parentCode string, withBlank bool) []DroplistOption {
	db.AdminCache.RLock()
	defer db.AdminCache.RUnlock()
	var options []DroplistOption
	if withBlank {
		var option DroplistOption
		options = append(options, option)
	}
	for _, item := range db.AdminCache.TheSlice {
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
	rows, err := db.Conn.Query(query)
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

func GetWizardOptions(field, selected string, withBlank bool) []DroplistOption {
	var options []DroplistOption
	if withBlank {
		options = append(options, DroplistOption{})
	}
	list := []DroplistOption{
		{Value: "SIGHTING", Description: "I saw...", Colour: "green", Icon: "eye", Selected: false},
		{Value: "USING", Description: "I used...", Colour: "green", Icon: "tag", Selected: false},
		{Value: "CLAIMING", Description: "I claim...", Colour: "green", Icon: "tag", Selected: false},
		{Value: "GIVING", Description: "I gave...", Colour: "green", Icon: "eye", Selected: false},
		{Value: "CARE", Description: "It needs care..", Colour: "green", Icon: "wrench", Selected: false},
		{Value: "BROKEN", Description: "It broke...", Colour: "green", Icon: "broken", Selected: false},
		{Value: "DIED", Description: "It died...", Colour: "green", Icon: "stethoscope", Selected: false},
		{Value: "LOST", Description: "It is lost...", Colour: "green", Icon: "steps", Selected: false},
		{Value: "BACKUP", Description: "I backed up...", Colour: "green", Icon: "copy", Selected: false},
		{Value: "INSTALL", Description: "I installed software...", Colour: "green", Icon: "software", Selected: false},
		{Value: "REQUEST", Description: "I need software...", Colour: "green", Icon: "software", Selected: false},
	}

	for i := range list {
		if list[i].Value == selected || (!withBlank && i == 0) {
			list[i].Selected = true
		}
	}
	return append(options, list...)
}
