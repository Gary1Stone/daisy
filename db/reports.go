package db

import (
	"database/sql"
	"log"
)

type History struct {
	Uid       int    `json:"uid" db:"uid"`
	User      string `json:"user" db:"user"`
	Fullname  string `json:"fullname" db:"fullname"`
	Group     string `json:"group" db:"group"`
	City      string `json:"city" db:"city"`
	Community string `json:"community" db:"community"`
	LastLogin string `json:"last" db:"last"`
	Banned    bool   `json:"banned" db:"banned"`
	Active    int    `json:"active" db:"active"`
	Days      int    `json:"days" db:"days"`
}

// Users last login and where, if banned, not logged in for 90+ days
func GetUsersReport(uid int) ([]History, error) {
	history := make([]History, 0)
	//Get this user's timezone offset
	Tzoff := GetTzoff(uid)
	//Get the non-deleted users
	var query = `
		SELECT A.uid, A.user, A.fullname, A.active, B.description FROM profiles A 
		LEFT JOIN choices B ON B.code = A.gid 
		WHERE B.field = "GROUP" AND A.user <> A.uid ORDER BY A.fullname
	`
	rows, err := Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return history, err
	}
	defer rows.Close()
	for rows.Next() {
		var usr History
		err := rows.Scan(&usr.Uid, &usr.User, &usr.Fullname, &usr.Active, &usr.Group)
		if err != nil {
			log.Println(err)
		}
		history = append(history, usr)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	//Get last logins
	query = `SELECT A.uid, max(A.timestamp) AS max, strftime('%Y-%m-%d %H:%M', A.timestamp-?, 'unixepoch') AS login, 
	B.city_ascii, C.community_ascii, A.ip, cast((strftime('%s', 'now') - A.timestamp) / 86400 AS INTEGER) AS days 
	FROM logins A
	LEFT JOIN cities B ON A.city_id = B.city_id
	LEFT JOIN communities C ON A.community_id = C.community_id
	WHERE A.success=1 GROUP BY A.uid
	`
	rows, err = Conn.Query(query, Tzoff)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return history, err
	}
	defer rows.Close()
	logins := struct {
		Uid       int
		Max       int64
		Login     string
		City      string
		Community string
		Ip        string
		Days      int
	}{}
	for rows.Next() {
		err := rows.Scan(&logins.Uid, &logins.Max, &logins.Login, &logins.City, &logins.Community, &logins.Ip, &logins.Days)
		if err != nil {
			log.Println(err)
		} else {
			for i := range history {
				if history[i].Uid == logins.Uid {
					history[i].LastLogin = logins.Login
					history[i].City = logins.City
					history[i].Community = logins.Community
					history[i].Banned = IsBanned(logins.Ip)
					history[i].Days = logins.Days
					break
				}
			}
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return history, nil
}

// Devices that checked in
type LastSeen struct {
	Cid      int    `json:"cid"`
	Name     string `json:"name"`
	Make     string `json:"make"`
	Model    string `json:"model"`
	Type     string `json:"type"`
	Image    string `json:"image"`
	Days     int    `json:"days"`
	LastSeen string `json:"lastseen"`
}

// Find all devices that never checked in
// or have not checked in with the last 90 days
func GetLastSeenDevices(curUid int) ([]LastSeen, error) {
	tzoff := GetTzoff(curUid)
	items := make([]LastSeen, 0)
	query := `
	SELECT D.cid, D.name, D.type, D.image, coalesce(cast((strftime('%s', 'now') - D.last_audit) / 86400 AS INTEGER), 0) AS days,
	coalesce(strftime('%Y-%m-%d', D.last_audit-?, 'unixepoch'), 'never') AS last_seen, coalesce(C.description, '') AS make, D.model
	FROM devices D
	LEFT JOIN choices C ON  C.code=D.make AND C.field='MAKE'
	WHERE D.active=1
	ORDER BY days desc
	`
	rows, err := Conn.Query(query, tzoff)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item LastSeen
		err := rows.Scan(&item.Cid, &item.Name, &item.Type, &item.Image, &item.Days, &item.LastSeen, &item.Make, &item.Model)
		if err != nil {
			log.Println(err)
		} else {
			items = append(items, item)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return items, err
}

type Gaps struct {
	Cid       int
	Hostname  string
	Timestamp string
	Gap       int
}

func NetworkGaps() ([]Gaps, error) {
	gaps := make([]Gaps, 0)
	query := `
		WITH OrderedPings AS (
			SELECT
				p.cid,
				p.utc,
				LAG(p.utc) OVER (PARTITION BY p.cid ORDER BY p.utc) AS prev_timestamp
			FROM pings p
		),
		Gaps AS (
			SELECT
				o.cid,
				d.name AS device_name,
				DATETIME(o.prev_timestamp, 'unixepoch') AS last_data_time,
				DATETIME(o.utc, 'unixepoch') AS current_data_time,
				(o.utc - o.prev_timestamp) AS gap_in_seconds
			FROM OrderedPings o
			JOIN devices d ON o.cid = d.cid
			WHERE o.prev_timestamp IS NOT NULL AND (o.utc - o.prev_timestamp) > 60
		)
		SELECT
			cid,
			device_name,
			last_data_time,
			gap_in_seconds / 60 AS gap_in_minutes
		FROM Gaps
		ORDER BY cid, last_data_time;
	`
	rows, err := Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return gaps, err
	}
	defer rows.Close()
	for rows.Next() {
		var dto Gaps
		err := rows.Scan(&dto.Cid, &dto.Hostname, &dto.Timestamp, &dto.Gap)
		if err != nil {
			log.Println(err)
		} else {
			gaps = append(gaps, dto)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return gaps, err
}
