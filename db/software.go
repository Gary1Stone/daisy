package db

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/util"
)

type Software struct {
	Task              string `json:"task"`
	Sid               int    `json:"sid" db:"sid"`                             //Software ID
	Name              string `json:"name" db:"name"`                           //Software name
	Licenses          int    `json:"licenses" db:"licenses"`                   //number of licenses purchased
	Source            string `json:"source" db:"source"`                       //Where the licenses were purchased from (Vendor)
	License_key       string `json:"license_key" db:"license_key"`             //The software license key
	Product           string `json:"product" db:"product"`                     //The product ID number
	Link              string `json:"link" db:"link"`                           //URL to download the software from
	Notes             string `json:"notes" db:"notes"`                         //user notes
	Active            int    `json:"active" db:"active"`                       //If active or deleted
	Last_updated_by   int    `json:"last_updated_by" db:"last_updated_by"`     //Who last changed the record
	Last_updated_time string `json:"last_updated_time" db:"last_updated_time"` //When the record was last changed
	Purchased         string `json:"purchased" db:"purchased"`                 //Date purchased (UTC int in database)
	Reuseable         int    `json:"reuseable" db:"reuseable"`                 //Software license can be reused on diferent computers, not tied to single computer
	Fullname          string `json:"fullname" db:"-"`                          //Full name (display name) of the user who last updated the record
	Color             string `json:"color" db:"color"`                         //Maybe Not used
	Installed         int    `json:"installed" db:"-"`                         //Date the software was installed
	Icon              string `json:"icon" db:"-"`                              //Software icon name
	Inv_name          string `json:"inv_name" db:"inv_name"`                   //Name in inventory
	Pre_installed     int    `json:"pre_installed" db:"pre_installed"`         //Number of licenses that came pre-installed on computers
	Free              int    `json:"free" db:"free"`                           //No license needed to install
	Aid               int    `json:"aid" db:"-"`                               //Used by web page
	Showhistory       int    `json:"showhistory"`                              //Used by web page
}

func (S *Software) trim() {
	S.Name = strings.TrimSpace(S.Name)
	S.Source = strings.TrimSpace(S.Source)
	S.License_key = strings.TrimSpace(S.License_key)
	S.Product = strings.TrimSpace(S.Product)
	S.Link = strings.TrimSpace(S.Link)
	S.Notes = strings.TrimSpace(S.Notes)
}

type SoftwareFilter struct {
	Task   string `json:"task" db:"-"`    //
	Sid    int    `json:"sid" db:"sid"`   // Software ID -1=ignore, 0=all, >0=exact one
	Name   string `json:"name" db:"name"` // Software name (exact match only)
	Search string `json:"search" db:"-"`  // Search term
	Page   int    `json:"page" db:"-"`    //
}

func (f *SoftwareFilter) Init() {
	f.Task = "get_first_page"
	f.Sid = -1
	f.Name = ""
	f.Search = ""
	f.Page = 0
}

// Find the software using the sid (key) as the unique identifier
func GetSoftware(curUid, sid int) (Software, error) {
	var filter SoftwareFilter
	filter.Init()
	filter.Sid = sid
	softwares, err := filter.GetSoftwares(curUid)
	if err != nil {
		log.Println(err)
		return Software{}, err
	}
	if len(softwares) == 0 {
		return Software{}, nil
	}
	addSoftwareLock(curUid, sid)
	return softwares[0], nil
}

// Get slice of Software structs
func (f *SoftwareFilter) GetSoftwares(curUid int) ([]Software, error) {
	var softwares []Software
	var query strings.Builder
	// Nullable Fields: source, license_key, product, link, notes, color, old_name
	query.WriteString(`
		SELECT
			A.sid, A.name, A.licenses, A.source, A.license_key, A.product, A.link, A.notes,
			A.active, COALESCE(A.last_updated_by,'') AS last_updated_by,
			strftime('%Y-%m-%d %H:%M', A.last_updated_time - ? , 'unixepoch') AS updated,
			strftime('%Y-%m-%d', A.purchased - ?, 'unixepoch') AS purchased,
			A.reuseable, B.fullname AS fullname,
			COALESCE(colours.color, 'light') AS color,
			COALESCE(colours.icon, '') AS icon,
			COALESCE(installs.installed, 0) AS installed,
			COALESCE(A.inv_name, "") AS inv_name,
			A.pre_installed, A.free
		FROM software A
		LEFT JOIN profiles B ON B.uid = A.last_updated_by
		LEFT JOIN (
			SELECT C.sid, C.action, D.color, D.icon
			FROM action_log C
			LEFT JOIN icons D ON D.name = C.action
			WHERE (sid_ack IS NULL OR sid_ack = 0 OR sid_ack = '') AND D.is_device = 0
			ORDER BY D.priority LIMIT 1
		) colours ON A.sid = colours.sid
		LEFT JOIN (
			SELECT count(*) as installed, sid from sw_inv
			WHERE sid>0
			GROUP BY sid
		) installs ON A.sid=installs.sid
`)

	tzoff := GetTzoff(curUid) //Time Zone Offset in minutes
	where, params := f.buildWhereClause(tzoff)
	query.WriteString(where)
	query.WriteString("LIMIT 50 OFFSET ")
	query.WriteString(strconv.Itoa(f.Page * 50))
	rows, err := Conn.Query(query.String(), params...)
	if err != nil {
		log.Println(err)
		if err == sql.ErrNoRows {
			return softwares, nil
		}
		return softwares, err
	}
	defer rows.Close()
	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var dto Software
		err := rows.Scan(&dto.Sid, &dto.Name, &dto.Licenses, &dto.Source, &dto.License_key, &dto.Product,
			&dto.Link, &dto.Notes, &dto.Active, &dto.Last_updated_by, &dto.Last_updated_time,
			&dto.Purchased, &dto.Reuseable, &dto.Fullname, &dto.Color, &dto.Icon, &dto.Installed,
			&dto.Inv_name, &dto.Pre_installed, &dto.Free)
		if err != nil {
			log.Println(err)
		} else {
			softwares = append(softwares, dto)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println("ERROR iterating over software rows:", err)
	}
	return softwares, err
}

func (f *SoftwareFilter) buildWhereClause(tzoff int) (string, []any) {
	var where strings.Builder
	var params []any
	where.WriteString("WHERE A.active=1 ")
	params = append(params, tzoff) //two timezone adjustments
	params = append(params, tzoff)
	if f.Sid == 0 {
		where.WriteString("AND A.sid>? ")
		params = append(params, f.Sid)
	} else if f.Sid > 0 {
		where.WriteString("AND A.sid=? ")
		params = append(params, f.Sid)
	}
	if len(f.Name) > 0 {
		where.WriteString("AND A.name=? ")
		params = append(params, f.Name)
	}
	if len(f.Search) > 0 {
		where.WriteString("AND (A.name like ? OR A.source like ?) ")
		params = append(params, wrapSearchTerm(f.Search))
		params = append(params, wrapSearchTerm(f.Search))
	}
	where.WriteString("ORDER BY A.name ")
	return where.String(), params
}

// Check that no other software package has the same name
func (s *Software) IsUnique() bool {
	cnt := 0
	query := "SELECT count(*) as cnt FROM software WHERE name=? and sid<>? COLLATE NOCASE LIMIT 1"
	err := Conn.QueryRow(query, s.Name, s.Sid).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return cnt == 0
}

func (s *Software) isDeletable() bool {
	cnt := 0
	query := "SELECT COUNT(*) AS cnt FROM action_log WHERE sid=?"
	err := Conn.QueryRow(query, s.Sid).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return cnt == 0
}

// Try to delete software if not in action log (installed), else mark as not active
func (s *Software) DeleteRecord(curUid int) bool {
	if s.isDeletable() {
		query := "DELETE FROM software WHERE sid=?"
		_, err := Conn.Exec(query, s.Sid)
		if err != nil {
			log.Println(err)
			return false
		}
		return true
	}
	query := "UPDATE software SET active=0, old_name=name, name=sid, last_updated_time=strftime('%s','now'), last_updated_by=? WHERE sid=?"
	_, err := Conn.Exec(query, curUid, s.Sid)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Save software record
func (s *Software) UpdateRecord(curUid int) bool {
	if curUid < 1 {
		curUid = SYS_PROFILE.Uid
	}
	if isSoftwareLocked(curUid, s.Sid) {
		log.Println("Record was updated by someone else before you tried to save.")
		return false
	}
	s.trim()
	if !s.IsUnique() {
		log.Println("Software name is not unique")
		return false
	}
	// Nullable Fields: source, license_key, product, link, notes, color, old_name
	const query = `
	UPDATE software SET
		name=?, licenses=?, active=?, purchased=?, reuseable=?,
		last_updated_by=?, last_updated_time=strftime('%s','now'),
		source=?, license_key=?, product=?, link=?, notes=?, 
		color=?, inv_name=?, pre_installed=?, free=?
	WHERE sid=?
	`
	_, err := Conn.Exec(query, s.Name, s.Licenses, s.Active,
		util.ToUnixSeconds(s.Purchased), s.Reuseable,
		foreignKey(s.Last_updated_by), s.Source,
		s.License_key, s.Product, s.Link, s.Notes,
		s.Color, s.Inv_name, s.Pre_installed, s.Free, s.Sid)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Add software record
func (s *Software) AddRecord() bool {
	s.trim()
	// Check software name is unique
	if !s.IsUnique() {
		log.Println("Software name was not unique", s.Name)
		return false
	}
	const query = `
		INSERT INTO software (
			name, licenses, active, last_updated_by, last_updated_time, purchased, reuseable,
			source, license_key, product, link, notes, color, inv_name, pre_installed, free, old_name
		) VALUES (?,?,?,?,strftime('%s','now'),?,?,?,?,?,?,?,?,?,?,?,"")
	`
	result, err := Conn.Exec(query, s.Name, s.Licenses, s.Active,
		foreignKey(s.Last_updated_by), util.ToUnixSeconds(s.Purchased),
		s.Reuseable, s.Source, s.License_key, s.Product, s.Link,
		s.Notes, s.Color, s.Inv_name, s.Pre_installed, s.Free)
	if err != nil {
		log.Println(err)
		s.Sid = 0
		return false
	}
	//Get the SID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		s.Sid = 0
		return false
	}
	s.Sid = int(lastInsertID)
	return true
}
