package db

import (
	"database/sql"
	"log"
	"math"
	"sync"
	"time"
)

type Logins struct {
	Id            int     `json:"id"`
	Timestamp     string  `json:"timestamp"`
	Uid           int     `json:"uid"`
	User          string  `json:"user"`
	Fullname      string  `json:"fullname"`
	Tzoff         int     `json:"tzoff"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	Country_code  string  `json:"country_code"`
	Country       string  `json:"country"`
	State         string  `json:"state"`
	Community_id  int     `json:"community_id"`
	Community     string  `json:"community"`
	City_id       int     `json:"city_id"`
	City          string  `json:"city"`
	Ip            string  `json:"ip"`
	Success       int     `json:"success"`
	Session       string  `json:"session"`
	Weekday       string  `json:"weekday"`       // Day of the week
	Days          int     `json:"days"`          // Days since last login
	Last_login    string  `json:"last_login"`    // Last login date/time
	Credential_id string  `json:"credential_id"` // Credentials ID
	Permissions   string  `json:"permissions"`   // Permissions
	Distance      int     `json:"distance"`      // Distance from home office
	Home          string  `json:"home"`          // Home office name
	Timezone      string  `json:"timezone"`      // Timezone name of the user
}

// Suggest the last user that logged in from this IP address
func GetLastUserByIP(ip string) string {
	var user string = ""
	const query = `
		SELECT B.user 
		FROM logins A
		LEFT JOIN profiles B on A.uid=B.uid
		WHERE A.success=1 AND A.ip=? 
		ORDER BY timestamp DESC LIMIT 1
	`
	err := Conn.QueryRow(query, ip).Scan(&user)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return user
}

// We need a session id to prevent two user from sharing the id/password
func GetLastSessionByUid(uid int) (string, string, error) {
	var session string = ""
	var ip string = ""
	query := "SELECT ip, session FROM logins WHERE uid=? AND success=1 ORDER BY timestamp DESC LIMIT 1"
	err := Conn.QueryRow(query, uid).Scan(&ip, &session)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return session, ip, nil
}

// Keep a log of the login successes and failures
// Used to ban IPs and find the user's current timezone offset
// Note the non-normalization of data (Country, State, City, Community)
// We needed to do this to keep the queries fast.
// Joins with the very big cities and communities tables were
// taking over 30ms, with only 400 entries in the logins table
func SaveLogin(dto *Logins) error {
	curSinLat := math.Sin(dto.Latitude * math.Pi / 180)
	curCosLat := math.Cos(dto.Latitude * math.Pi / 180)
	curCosLon := math.Cos(dto.Longitude * math.Pi / 180)
	curSinLon := math.Sin(dto.Longitude * math.Pi / 180)
	var wg sync.WaitGroup
	wg.Add(1)
	go getCity(curSinLat, curCosLat, curCosLon, curSinLon, dto, &wg)
	wg.Add(1)
	go getCommunity(curSinLat, curCosLat, curCosLon, curSinLon, dto, &wg)
	wg.Wait()

	// Calculate the distance from the user's home office or the WKNC main office
	usr, err := GetProfile(dto.Uid, dto.Uid)
	if err != nil {
		log.Println(err)
	}

	// Set the user's base office location
	dto.Home = "WKNC (Weston King)"
	if len(usr.Geo_fence) > 0 {
		dto.Home = GetCodeDescription("GEOFENCE", usr.Geo_fence)
	}

	// Calculate the distance from the home office in KM
	if _, dto.Distance = IsInsideGeoFence(usr.Geo_fence, usr.Geo_radius, dto.Latitude, dto.Longitude); dto.Distance == -1 {
		dto.Distance = int(HowFar(43.7001873335, -79.5165298401, dto.Latitude, dto.Longitude)) // WKNC Location
	}

	const query = `
		INSERT INTO logins 
		(uid, tzoff, longitude, latitude, community_id, city_id, ip, 
		success, session, distance, home, country, state, city, 
		community, timezone) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`
	_, err = Conn.Exec(query, foreignKey(dto.Uid), dto.Tzoff, dto.Longitude,
		dto.Latitude, foreignKey(dto.Community_id), foreignKey(dto.City_id),
		dto.Ip, dto.Success, dto.Session, dto.Distance, dto.Home, dto.Country,
		dto.State, dto.City, dto.Community, dto.Timezone)
	if err != nil {
		log.Println(err)
		return err
	}
	//Get rid of records older than 2 years
	_, err = Conn.Exec("DELETE FROM logins WHERE timestamp <= strftime('%s', datetime('now', '-730 day'))")
	if err != nil {
		log.Println(err)
		return err
	}
	//Should user be deactivated?
	deActivateUser(dto.Uid)
	//Rebuild the banned IP list
	if dto.Success == 0 {
		go buildBanned(dto.Ip)
	}
	return err
}

// If the user has too many failed login attempts in the last 60 mins, deactivate the profile
func deActivateUser(uid int) error {
	var cnt = 0
	const query = "SELECT count(*) AS cnt FROM logins WHERE uid=? AND success=0 AND timestamp > strftime('%s', datetime('now', '-60 minutes'))"
	err := Conn.QueryRow(query, uid).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	if cnt < 10 {
		return nil
	}
	_, err = Conn.Exec("UPDATE profile SET active=0 WHERE uid=?", uid)
	if err != nil {
		log.Println(err)
	}
	return err
}

var bannedIPs = make(map[string]bool)
var bannedStartTime = time.Now()
var bannedMutex sync.RWMutex

func buildBanned(ip string) {

	bannedMutex.Lock()
	defer bannedMutex.Unlock()

	if time.Since(bannedStartTime).Hours() > 24 || len(bannedIPs) == 0 {
		for k := range bannedIPs {
			delete(bannedIPs, k)
		}
		bannedStartTime = time.Now()   //Reset the clock
		bannedIPs["127.0.0.1"] = false //Stop redoing the search if there are none
		query := "SELECT ip FROM logins WHERE success=0 AND timestamp > strftime('%s', datetime('now', '-7 day')) GROUP BY ip HAVING count(ip) > 10 ORDER By ip"
		rows, err := Conn.Query(query)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			var ips string = ""
			err2 := rows.Scan(&ips)
			if err2 != nil {
				err2 = nil
			}
			bannedIPs[ips] = true
		}
	} else {
		cnt := 0
		query := "SELECT count(*) AS cnt FROM logins WHERE ip=? AND success=0 AND timestamp > strftime('%s', datetime('now', '-7 day'))"
		err := Conn.QueryRow(query, ip).Scan(&cnt)
		if err != nil && err != sql.ErrNoRows {
			log.Println(err)

		}
		if cnt > 5 {
			bannedIPs[ip] = true
		}
	}
}

func IsBanned(ip string) bool {
	bannedMutex.RLock()
	defer bannedMutex.RUnlock()
	_, found := bannedIPs[ip]
	return found
}

// Used by home page to find last login details to show to the user
func GetLastLogin(uid int) (Logins, error) {
	dayOfWeek := [...]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	var dto Logins
	var weekday int = 0
	const query = `
		SELECT A.id, B.user, A.ip, A.tzoff, A.longitude, A.latitude, A.country, 
			A.community, A.city, A.city_id, A.state, 
			strftime('%Y-%m-%d %H:%M', (A.timestamp - A.tzoff), 'unixepoch') AS timestamp, 
			strftime('%w',(A.timestamp - A.tzoff), 'unixepoch') AS weekday, 
			A.success, A.session, A.distance, A.home, A.timezone 
		FROM logins A
		LEFT JOIN profiles B on A.uid=B.uid
		WHERE A.uid=? AND A.success=1
		ORDER by TIMESTAMP DESC 
		LIMIT 1 OFFSET 1
`
	err := Conn.QueryRow(query, uid).Scan(&dto.Id, &dto.User, &dto.Ip, &dto.Tzoff, &dto.Longitude, &dto.Latitude,
		&dto.Country, &dto.Community, &dto.City, &dto.City_id, &dto.State,
		&dto.Timestamp, &weekday, &dto.Success, &dto.Session, &dto.Distance, &dto.Home, &dto.Timezone)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return dto, err
	}
	dto.Weekday = dayOfWeek[weekday]
	return dto, nil
}

// Return the User's Time Zone offset in minutes
func GetTzoff(curUid int) int {
	// we want only the most recent entry
	if curUid < 1 {
		curUid = SYS_PROFILE.Uid
	}
	var tzoff int = 0
	query := "SELECT tzoff FROM logins WHERE uid=? AND success=1 ORDER BY timestamp DESC LIMIT 1"
	err := Conn.QueryRow(query, curUid).Scan(&tzoff) // If user never signed in, no logs to look at, so ignore err (no rows returned)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return tzoff
}

//Return the User's timzone name
func GetTimezoneName(curUid int) string {
	if curUid < 1 {
		curUid = SYS_PROFILE.Uid
	}
	var timezone string = ""
	query := "SELECT timezone FROM logins WHERE uid=? AND success=1 ORDER BY timestamp DESC LIMIT 1"
	err := Conn.QueryRow(query, curUid).Scan(&timezone)	
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return timezone
}

func CheckUsersLastIpBanned(uid int) bool {
	const query = "SELECT ip FROM logins WHERE uid=? ORDER BY timestamp DESC LIMIT 1"
	ip := ""
	err := Conn.QueryRow(query, uid).Scan(&ip)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return IsBanned(ip)
}

func UnBanUser(uid int) error {
	//Get the user's last IP address and reset it in the banned list
	var query = "SELECT ip FROM logins WHERE uid=? ORDER BY timestamp DESC LIMIT 1"
	ip := ""
	err := Conn.QueryRow(query, uid).Scan(&ip)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if len(ip) > 0 {
		bannedIPs[ip] = false
	}
	//Clear entries in logins table less than 7 days old
	query = "DELETE FROM logins WHERE uid=? AND success=0 AND timestamp > strftime('%s', datetime('now', '-7 day'))"
	_, err = Conn.Exec(query, uid)
	return err
}
