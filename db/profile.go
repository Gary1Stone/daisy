package db

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var SYS_PROFILE Profile

type Profile struct {
	Task              string `json:"task"`
	Uid               int    `json:"id" db:"uid"`                              // User ID
	User              string `json:"user" db:"user"`                           // User unique identifier (email address)
	Pwd               string `json:"-" db:"pwd"`                               // Password, Using - to prevent convertion to JSON. prevents sending out to users
	Gid               int    `json:"gid" db:"gid"`                             // Group ID
	Fullname          string `json:"fullname" db:"fullname"`                   // Display name for the user
	First             string `json:"first" db:"first"`                         // First name for the user
	Last              string `json:"last" db:"last"`                           // Last name for the user
	Active            int    `json:"active" db:"active"`                       // 1=active 0=deleted
	Last_updated_by   int    `json:"last_updated_by" db:"last_updated_by"`     // Who last updated the record
	Last_updated_time string `json:"last_updated_time" db:"last_updated_time"` // When the record was last updated
	Color             string `json:"color" db:"color"`                         // Highlight color
	Picture           string `json:"picture" db:"picture"`                     // not used
	Geo_fence         string `json:"geo_fence" db:"geo_fence"`                 // around what geographic obect the user is tied to
	Geo_radius        int    `json:"geo_radius" db:"geo_radius"`               // how far from the geo_fence can be before being prevented from loggin in
	Pwd_reset         int    `json:"pwd_reset" db:"pwd_reset"`                 // flag for the passowrd being reset
	Otp               string `json:"otp" db:"otp"`                             // One Time Password
	Mins              int    `json:"mins" db:"-"`                              // Elapsed minutes since last updated
	Group             string `json:"group" db:"-"`                             // Permission Group name
	Lun               string `json:"lun" db:"-"`                               // Last updated by display name
	Deleteable        bool   `json:"deletable" db:"-"`                         // Is the profile deletable 1=yes, 0=no
	Alerts            int    `json:"alerts" db:"-"`                            // Count of user's outstanding alerts
	Notify            int    `json:"notify" db:"notify"`                       // Can the user be sent notifications via email
	Tickets           int    `json:"tickets" db:"tickets"`                     // Count of user's open tickets
}

func (p *Profile) trim() {
	p.User = strings.TrimSpace(p.User)
	p.Fullname = strings.TrimSpace(p.Fullname)
	p.First = strings.TrimSpace(p.First)
	p.Last = strings.TrimSpace(p.Last)
	p.Picture = strings.TrimSpace(p.Picture)
}

type ProfileFilter struct {
	Task   string `json:"task" db:"-"`    // Task/command from web page
	Uid    int    `json:"id" db:"uid"`    // User ID -1=ignore, 0=any id, >0=exact UID
	Gid    int    `json:"gid" db:"gid"`   // Group ID -1=ignore, 0=any id, >0=exact UID
	User   string `json:"user" db:"user"` // User email address
	Search string `json:"search" db:"-"`  // Search Term for: User or Group or fullname
	Page   int    `json:"page" db:"-"`    // Display page of results
}

func (f *ProfileFilter) Init() {
	f.Task = "get_first_page"
	f.Uid = -1
	f.Gid = -1
	f.User = ""
	f.Search = ""
	f.Page = 0
}

// SetSystemProfile
// Initialize one user to be the Global system profile
// then test database access and configuration
func setupSystemProfile() {
	sysUser := os.Getenv("SYS_USER")
	if sysUser == "" {
		log.Panic("FATAL: Environment variable SYS_USER not set in .env file")
	}
	var err error
	SYS_PROFILE, err = GetUserByEmail(sysUser)
	if err != nil {
		log.Panic("FATAL: Database not accessible.", err)
	}
	if SYS_PROFILE.User == "" {
		log.Panicf("FATAL: System User %v is not configured in the database profiles table", sysUser)
	}
}

// Get one or multiple user profiles by uid (user ID), uid=0 means get all
func (filter ProfileFilter) GetProfiles(curUid int) ([]Profile, error) {
	var profiles []Profile
	var query strings.Builder
	query.WriteString(`
    SELECT A.uid, A.user, A.pwd, A.gid, A.fullname, A.first, A.last, A.active, A.last_updated_by,
        A.picture, A.geo_fence, A.geo_radius, A.pwd_reset, A.otp, 
		grp, coalesce(colours.color, 'light') color,
        strftime('%Y-%m-%d %H:%M', last_updated_time - ? , 'unixepoch') AS updated,
        strftime('%s','now') as cur, last_updated_time as lut, coalesce(usr.fullname, '') lun, 
		COALESCE(E.alerts, 0) alerts, A.notify, COALESCE(G.tickets, 0) tickets
    FROM profiles A
    LEFT JOIN (
        SELECT description as grp, code FROM choices WHERE field='GROUP' and active=1
    ) groups ON A.gid=groups.code
    LEFT JOIN (
        SELECT C.action, D.color, C.uid, C.inform FROM action_log C
        LEFT JOIN icons D ON D.name=C.action
        WHERE C.active=1 AND ((C.uid IS NOT NULL AND (C.uid_ack IS NULL OR C.uid_ack=0)) OR
              (C.inform IS NOT NULL AND (C.inform_ack IS NULL or C.inform_ack=0)))
        ORDER BY D.priority
        LIMIT 1
    	) colours ON A.uid=colours.uid OR A.uid=colours.inform
    LEFT JOIN (
        SELECT uid, fullname from profiles
    	) usr ON A.last_updated_by=usr.uid
	LEFT JOIN (
		SELECT count(*) as alerts, E.uid FROM alerts E WHERE ack IS NULL
		GROUP BY E.uid
	) E ON E.uid=A.uid 
	LEFT JOIN (
	SELECT count(*) as tickets, G.uid FROM action_log G WHERE G.active=1 AND G.uid_ack=0  AND G.action IN ('BROKEN', 'DIED', 'LOST', 'CARE', 'REQUEST') GROUP BY G.uid
	) G ON G.uid=A.uid 
	`)

	tzoff := GetTzoff(curUid) //Time Zone Offest in minutes
	where, params := filter.buildWhereClause(tzoff)
	query.WriteString(where)
	rows, err := Conn.Query(query.String(), params...)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return profiles, err
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var usr Profile
		var now, lut int //now=current time, lut=last update time
		err := rows.Scan(&usr.Uid, &usr.User, &usr.Pwd, &usr.Gid, &usr.Fullname, &usr.First, &usr.Last,
			&usr.Active, &usr.Last_updated_by, &usr.Picture, &usr.Geo_fence, &usr.Geo_radius, &usr.Pwd_reset, &usr.Otp,
			&usr.Group, &usr.Color, &usr.Last_updated_time, &now, &lut, &usr.Lun, &usr.Alerts, &usr.Notify, &usr.Tickets)
		if err != nil {
			log.Println(err)
			return profiles, err
		}
		usr.Mins = ((now - lut) / 60) //Minutes since last updated. Is now server timezone, yes, but so is lut, and we only want the difference
		usr.Deleteable = usr.isDeletable()
		profiles = append(profiles, usr)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
	}
	return profiles, err
}

func (f ProfileFilter) buildWhereClause(tzoff int) (string, []any) {
	var where strings.Builder
	var params []any
	where.WriteString("WHERE A.active=1 ")
	params = append(params, tzoff)
	if f.Uid == 0 {
		where.WriteString("AND A.uid>? ")
		params = append(params, f.Uid)
	} else if f.Uid > 0 {
		where.WriteString("AND A.uid=? ")
		params = append(params, f.Uid)
	}
	if f.Gid == 0 {
		where.WriteString("AND A.gid>? ")
		params = append(params, f.Gid)
	} else if f.Gid > 0 {
		where.WriteString("AND A.gid=? ")
		params = append(params, f.Gid)
	}
	if len(f.User) > 0 {
		where.WriteString("AND A.user=? ")
		params = append(params, f.User)
	}
	if len(f.Search) > 0 {
		where.WriteString("AND (A.user like ? OR A.fullname like ? OR grp like ?) ")
		params = append(params, wrapSearchTerm(f.Search))
		params = append(params, wrapSearchTerm(f.Search))
		params = append(params, wrapSearchTerm(f.Search))
	}
	where.WriteString("ORDER BY A.user ")
	where.WriteString("LIMIT 50 OFFSET ")
	where.WriteString(strconv.Itoa(f.Page * 50))
	return where.String(), params
}

func wrapSearchTerm(str string) string {
	return "%" + str + "%"
}

// Find the user profile using the uid (key) as the unique identifier
func GetProfile(curUid, uid int) (Profile, error) {
	var profile Profile
	if curUid < 1 {
		curUid = SYS_PROFILE.Uid
	}
	if uid < 1 {
		log.Println("ERROR: uid is 0 or negative")
		return Profile{}, nil
	}
	var filter ProfileFilter
	filter.Init()
	filter.Uid = uid
	profiles, err := filter.GetProfiles(curUid)
	if err != nil {
		log.Println(err)
		return profile, err
	}
	if len(profiles) < 1 {
		return profile, nil
	}
	addProfileLock(curUid, uid)
	return profiles[0], nil
}

// Find the user profile using their email address as the unique identifier
func GetUserByEmail(email string) (Profile, error) {
	var profile Profile
	var filter ProfileFilter
	uid := SYS_PROFILE.Uid
	filter.Init()
	filter.User = email
	profiles, err := filter.GetProfiles(uid)
	if err != nil {
		log.Println(err)
		return profile, err
	}
	if len(profiles) < 1 {
		return profile, nil
	}
	return profiles[0], nil
}

// Save the user's profile. item is the data transfer object (struct)
func (p *Profile) UpdateRecord(curUid int) error {
	if curUid < 1 {
		curUid = SYS_PROFILE.Uid
	}
	if isProfileLocked(curUid, p.Uid) {
		log.Println("Record was updated by someone else before you tried to save.")
		return errors.New("record updated by someone else before you tried to save")
	}
	if !p.IsUnique() {
		log.Println("ERROR: Failure on checking Unique User ID ")
		return errors.New("failure on checking unique user id")
	}
	p.trim()
	var params []any
	var query strings.Builder
	query.WriteString("UPDATE profiles SET ")
	if len(p.Pwd) > 0 {
		params = append(params, p.Pwd)
		query.WriteString("pwd=?, ")
	}
	query.WriteString(`pwd_reset=?, user=?, gid=?, first=?, last=?, active=?, color=?, last_updated_by=?, 
	last_updated_time=strftime('%s','now'), fullname=?, picture=?, geo_fence=?, geo_radius=?, otp=?, notify=? 
	WHERE uid=?`)
	params = append(params, p.Pwd_reset, p.User, p.Gid, p.First,
		p.Last, p.Active, p.Color, p.Last_updated_by, p.Fullname,
		p.Picture, p.Geo_fence, p.Geo_radius, p.Otp, p.Notify, p.Uid)
	_, err := Conn.Exec(query.String(), params...)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Set the user's profile notification color
func SetProfileColor(uid int, color string) (string, int) {
	_, err := Conn.Exec("UPDATE profiles SET color=? WHERE uid=?", color, uid)
	if err != nil {
		log.Println(err)
	}
	query := "SELECT fullname, gid FROM profiles WHERE uid=?"
	fullname := ""
	gid := 0
	err = Conn.QueryRow(query, uid).Scan(&fullname, &gid)
	if err != nil {
		log.Println(err)
	}
	return fullname, gid
}

// Get permissions CAPS - C-Device/Computer table, A-Admin Tables, P-Profile table, S-Software table T- Ticket (Action_log)
func GetPermissions(gid int) (string, error) {
	var permissions = ""
	var query = "SELECT parent FROM choices WHERE field='GROUP' AND code=?"
	err := Conn.QueryRow(query, gid).Scan(&permissions)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return permissions, err
}

func GetUidPermissions(uid int) (string, error) {
	var permissions = ""
	query := `SELECT A.permissions FROM choices A
		LEFT JOIN profiles B ON B.gid=A.code
		WHERE A.field='GROUP' AND B.uid=?
	`
	err := Conn.QueryRow(query, uid).Scan(&permissions)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return permissions, err
}

// Check if it is a deleteable user
// If profile has assigned devices, return false
// Cannot delete user marked as the system profile
func (p *Profile) isDeletable() bool {
	if p.Uid == SYS_PROFILE.Uid {
		return false
	}
	assignedDevices := 0
	query := "SELECT count(*) as assignedDevices FROM devices WHERE uid=?"
	err := Conn.QueryRow(query, p.Uid).Scan(&assignedDevices)
	if err != nil {
		log.Println(err)
	}
	return assignedDevices == 0
}

// Check if the email address the user want to use is already in use.
func (p *Profile) IsUnique() bool {
	cnt := 1
	query := "SELECT count(*) as cnt FROM profiles WHERE user=? and uid<>? LIMIT 1 COLLATE NOCASE"
	err := Conn.QueryRow(query, p.User, p.Uid).Scan(&cnt)
	if err != nil {
		log.Println(err)
	}
	return cnt == 0
}

// Mark profile as deleted by setting active=0, old_user=user, user=uid
func (p *Profile) DeleteRecord(curUid int) bool {
	if !p.isDeletable() {
		return false
	}
	//if profile has assigned actions, mark them acknowledged
	query := "UPDATE action_log SET uid_ack=1 WHERE uid=?"
	_, err := Conn.Exec(query, p.Uid)
	if err != nil {
		log.Println(err)
	}
	query = "UPDATE action_log SET inform_ack=1 WHERE inform=?"
	_, err = Conn.Exec(query, p.Uid)
	if err != nil {
		log.Println(err)
	}
	query = "UPDATE profiles SET old_user=user, active=0, user=uid, last_updated_time=strftime('%s','now'), last_updated_by=? WHERE uid=?"
	_, err = Conn.Exec(query, curUid, p.Uid)
	if err != nil {
		log.Println(err)
		return false
	}
	query = "UPDATE devices SET uid=null WHERE uid=?"
	_, err = Conn.Exec(query, p.Uid)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Create a new profile
func (p *Profile) AddRecord(curUid int) bool {
	if curUid < 1 {
		curUid = SYS_PROFILE.Uid
	}
	if !p.IsUnique() {
		log.Println("ERROR: Failure on checking Unique User ID ")
		return false
	}
	p.trim()
	var query = `
	INSERT INTO profiles 
		(user, pwd, gid, fullname, first, last, last_updated_by, color, picture, 
		geo_fence, geo_radius, pwd_reset, otp, active, old_user, notify,  last_updated_time) 
	VALUES 
		(?,?,?,?,?,?,?,?,?,?,?,?,?,1,'',?, strftime('%s','now'))
	`
	result, err := Conn.Exec(query, p.User, p.Pwd, p.Gid, p.Fullname, p.First, p.Last,
		curUid, p.Color, p.Picture, p.Geo_fence, p.Geo_radius, p.Pwd_reset, p.Otp, p.Notify)
	if err != nil {
		log.Println(err)
		return false
	}
	LastInsertId, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		p.Uid = 0
		return false
	}
	p.Uid = int(LastInsertId)
	return true
}

// If user's Group (Gid) changed, cascade new groupID (gid) in devices and action tables
func (p *Profile) CascadeGidChange() {
	tx, err := Conn.Begin() // Start a transaction
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return
	}

	queries := []struct {
		query string
		args  []any
	}{
		{"UPDATE devices SET gid=? WHERE uid=?", []any{p.Gid, p.Uid}},
		{"UPDATE action_log SET gid=? WHERE uid=?", []any{p.Gid, p.Uid}},
		{"UPDATE action_log SET inform_gid=? WHERE inform=?", []any{p.Gid, p.Uid}},
	}

	for _, q := range queries {
		_, err := tx.Exec(q.query, q.args...)
		if err != nil {
			log.Println("Failed to execute query:", q.query, "Error:", err)
			tx.Rollback() // Rollback the transaction on error
			return
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println("Failed to commit transaction:", err)
	}
}

// Get the user's group ID
func GetGid(uid int) int {
	var gid = 0
	if uid > 0 {
		err := Conn.QueryRow("SELECT gid FROM profiles WHERE uid=?", uid).Scan(&gid)
		if err != nil {
			log.Println(err)
			return gid
		}
	}
	return gid
}

// Login Report
func GetLogins(curUid, uid int) ([]*Logins, error) {
	Login_log := make([]*Logins, 0)
	tzoff := GetTzoff(curUid)
	// This groups by community - confusing, but interesting
	// const query = `
	// 	SELECT strftime('%Y-%m-%d %H:%M', A.timestamp-?, 'unixepoch') AS login,
	// 		max(A.timestamp) as timestamp, A.uid, B.user, C.country, C.state, C.city, D.community,
	// 		cast((strftime('%s', 'now') - A.timestamp) / 86400 AS INTEGER) AS days, A.distance
	// 	FROM logins A
	// 	LEFT JOIN profiles B on A.uid=B.uid
	// 	LEFT JOIN cities C on A.city_id=C.city_id
	// 	LEFT JOIN communities D on A.community_id=D.community_id
	// 	WHERE A.success=1 AND A.uid=?
	// 	GROUP BY A.community_id
	// 	ORDER BY A.timestamp DESC
	// `

	// Get the UNIX UTC timestamp for 90 days ago
	now := time.Now()
	daysAgo := now.AddDate(0, 0, -90)
	nintyDaysAgo := daysAgo.Unix()
	const query = `
		SELECT strftime('%Y-%m-%d %H:%M', A.timestamp-?, 'unixepoch') AS login, 
		A.timestamp, A.country, A.state, A.city, A.community, 
		cast((strftime('%s', 'now') - A.timestamp) / 86400 AS INTEGER) AS days, A.distance
		FROM logins A
		WHERE A.success=1 AND A.uid=? AND A.timestamp>?
		ORDER BY A.timestamp DESC
		LIMIT 500`
	rows, err := Conn.Query(query, tzoff, uid, nintyDaysAgo)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return Login_log, err
	}
	defer rows.Close()
	for rows.Next() {
		var dto Logins
		err := rows.Scan(&dto.Last_login, &dto.Timestamp, &dto.Country, &dto.State,
			&dto.City, &dto.Community, &dto.Days, &dto.Distance)
		if err != nil {
			log.Println(err)
		} else {
			dto.Uid = uid
			Login_log = append(Login_log, &dto)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return Login_log, err
}
