package db

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type AttackInfo struct {
	Id               int     `json:"id"`
	Timestamp        int64   `json:"timestamp"`
	Ip               string  `json:"ip"`
	Uid              int     `json:"uid"`
	Method           string  `json:"method"`
	Path             string  `json:"path"`
	Browser          string  `json:"browser"`
	Longitude        float64 `json:"longitude"`
	Latitude         float64 `json:"latitude"`
	City_id          int     `json:"city_id"`
	Community_id     int     `json:"community_id"`
	Business_name    string  `json:"business_name"`
	Business_website string  `json:"business_website"`
	Ip_name          string  `json:"ip_name"`
	Ip_type          string  `json:"ip_type"`
	Isp              string  `json:"isp"`
	Org              string  `json:"org"`
	Continent        string  `json:"continent"`
	Country          string  `json:"country"`
	State            string  `json:"state"`
	City             string  `json:"city"`
	Community        string  `json:"community"`
	Occurred         string  `json:"occurred"`
	Fullname         string  `json:"fullname"`
	First_browser    string  `json:"first_browser"`
	Attack_count     int     `json:"attack_counts"`
}

// Record when external user asks for a file that does not exist.
func RecordAttack(ipAddress, method, path, browser string) {
	query := "INSERT INTO attacks (ip, method, path, browser) VALUES (?,?,?,?)"
	result, err := Conn.Exec(query, ipAddress, method, path, browser)
	if err != nil {
		log.Println(err)
		return // Don't proceed if insert fails
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return // Don't proceed if no record ID
	}
	recId := int(lastInsertID)
	if recId < 1 {
		log.Println("Bad attack record ID")
		return // Don't proceed
	}

	var loc = AttackInfo{
		Id:      recId,
		Ip:      ipAddress,
		Method:  method,
		Path:    path,
		Browser: browser,
	}

	go findAndUpdateLocation(&loc)
}

// findAndUpdateLocation orchestrates finding the location and updating the attacks table.
func findAndUpdateLocation(loc *AttackInfo) {
	if loc.Id < 1 || loc.Ip == "" {
		return
	}

	if !findLocationInAttacks(loc) { // Check if we already have it in attacks
		if !findLocationInLogins(loc) { // Check if we already have it in logins
			if !geolocation(loc) { // geolocation is fully free
				return // never found
			}
		}
	}

	if loc.Longitude == 0.0 && loc.Latitude == 0.0 {
		return // No results found, can't do anything
	}
	loc.City_id, loc.Community_id = SearchNearestPlace(loc.Longitude, loc.Latitude)
	updateAttacks(loc)
}

// Search the attacks table for previous matches
func findLocationInAttacks(loc *AttackInfo) bool {
	query := `SELECT uid, longitude, latitude, city_id, community_id,
		business_name, business_website, ip_name, ip_type, isp, org
		FROM attacks 
		WHERE timestamp=(
			SELECT MAX(timestamp) 
			FROM attacks 
			WHERE ip=? AND id<>? AND longitude<>0.0 AND latitude<>0.0 
			AND city_id>0 AND timestamp > strftime('%s', 'now', '-7 days')
		) LIMIT 1`
	err := Conn.QueryRow(query, loc.Ip, loc.Id).Scan(&loc.Uid, &loc.Longitude, &loc.Latitude,
		&loc.City_id, &loc.Community_id, &loc.Business_name, &loc.Business_website, &loc.Ip_name,
		&loc.Ip_type, &loc.Isp, &loc.Org)
	if loc.Uid > 0 {
		log.Println("User session ended because of hacking.", loc.Fullname)
		EndSession(loc.Uid) // Kick out user for hacking
	}
	return err == nil
}

// Search the logins table for previous matches
func findLocationInLogins(loc *AttackInfo) bool {
	query := `SELECT uid, longitude, latitude, city_id, community_id FROM logins 
		WHERE timestamp=(SELECT MAX(timestamp) FROM logins 
		WHERE ip=? AND timestamp > strftime('%s', 'now', '-1 days')) LIMIT 1`
	err := Conn.QueryRow(query, loc.Ip, loc.Id).Scan(&loc.Uid, &loc.Longitude, &loc.Latitude, &loc.City_id, &loc.Community_id)
	if loc.Uid > 0 {
		log.Println("User session ended because of hacking.", loc.Fullname)
		EndSession(loc.Uid) // Kick out user for hacking
	}
	return err == nil
}

// Consider defining this globally or passing it around
var httpClient = &http.Client{
	Timeout: 60 * time.Second, // Example: 60-second timeout
}

func updateAttacks(loc *AttackInfo) {
	query := `UPDATE attacks SET uid=?, longitude=?, latitude=?, city_id=?, community_id=? WHERE id=?`
	_, err := Conn.Exec(query, foreignKey(loc.Uid), loc.Longitude, loc.Latitude, foreignKey(loc.City_id), foreignKey(loc.Community_id), loc.Id)
	if err != nil {
		log.Println(err)
	}
}

// Geolocation-db.com
// Our IP geolocation API is completely free and unlimited requests per day are allowed.
func geolocation(loc *AttackInfo) bool {
	url := os.Getenv("GEOLOCATION_URL") + os.Getenv("GEOLOCATION_KEY") + "/" + loc.Ip

	// Notes: This API JSON can contain null for some items,
	// so we have to use pointers when defining the struct
	var geo = struct {
		Country_code *string `json:"country_code"` //"country_code":"CA",
		Country_name *string `json:"country_name"` //"country_name":"Canada",
		City         *string `json:"city"`         //"city":"Etobicoke",
		Postal       *string `json:"postal"`       //"postal":"M9V",
		Latitude     float64 `json:"latitude"`     //"latitude":43.7432,
		Longitude    float64 `json:"longitude"`    //"longitude":-79.5876,
		Ipv4         string  `json:"ipv4"`         //"IPv4":"142.112.223.79",
		State        *string `json:"state"`        //"state":"Ontario"
	}{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Error creating GET request: %v", err)
		return false
	}

	resp, err := httpClient.Do(req) // Use the client
	if err != nil {
		log.Printf("Error performing lookup for %s: %v", loc.Ip, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("API Response Body: %s", string(bodyBytes))
		return false
	}

	body, err := io.ReadAll(resp.Body) // response body is []byte
	if err != nil {
		return false
	}

	if err := json.Unmarshal(body, &geo); err != nil { // Parse []byte to the go struct pointer
		return false
	}

	loc.Latitude = geo.Latitude
	loc.Longitude = geo.Longitude
	return true
}

func GetAttacksDetails(curUid, duration int) ([]AttackInfo, error) {
	items := make([]AttackInfo, 0)
	tzoff := GetTzoff(curUid)
	days := "'-1 days'"
	switch duration {
	case 7:
		days = "'-7 days'"
	case 30:
		days = "'-30 days'"
	}
	query := `SELECT strftime('%Y-%m-%d %H:%M', A.timestamp-?, 'unixepoch') AS occurred, 
		A.ip, coalesce(A.uid, 0) as uid, A.method, A.path, A.browser, A.latitude, A.longitude,
		coalesce(B.continent,'') as continent, coalesce(B.country,'') as country, 
		coalesce(B.state,'') as state, coalesce(city_ascii,'') as city, 
		coalesce(C.community_ascii,'') as community, coalesce(D.fullname, '') as fullname,
		CASE WHEN INSTR(A.browser, ' ') > 0 THEN SUBSTR(A.browser, 1, INSTR(A.browser, ' ') - 1)
		ELSE A.browser
		END AS first_browser
		FROM attacks A
		LEFT JOIN cities B ON A.city_id=B.city_id
		LEFT JOIN communities C ON A.community_id=C.community_id
		LEFT JOIN profiles D ON A.uid=D.uid
		WHERE A.timestamp >= strftime('%s', 'now', ` + days + `) 
		GROUP BY ip, first_browser ORDER BY timestamp DESC`
	rows, err := Conn.Query(query, tzoff)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var item AttackInfo
		err := rows.Scan(&item.Occurred, &item.Ip, &item.Uid, &item.Method, &item.Path,
			&item.Browser, &item.Latitude, &item.Longitude, &item.Continent, &item.Country,
			&item.State, &item.City, &item.Community, &item.Fullname, &item.First_browser)
		if err != nil {
			log.Println(err)
			continue
		} else {
			items = append(items, item)
		}
	}
	if rows.Err() != nil {
		log.Println(rows.Err())
	}

	type Attacks struct {
		Count   int
		Ip      string
		Browser string
	}
	var attacks []Attacks

	query = `WITH Attacking AS (
		SELECT ip, timestamp, 
			CASE
				WHEN INSTR(browser, ' ') > 0
				THEN SUBSTR(browser, 1, INSTR(browser, ' ') - 1)
				ELSE browser
			END AS first_browser
		FROM attacks
		WHERE timestamp >= strftime('%s', 'now', ` + days + `)
	)
	SELECT
		COUNT(*) AS attack_count, ip, first_browser
	FROM Attacking
	GROUP BY ip, first_browser`

	rows, err = Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var attack Attacks
		err := rows.Scan(&attack.Count, &attack.Ip, &attack.Browser)
		if err != nil {
			log.Println(err)
			continue
		} else {
			attacks = append(attacks, attack)
		}
	}
	if rows.Err() != nil {
		log.Println(rows.Err())
	}

	for i := range items {
		for _, attack := range attacks {
			if items[i].Ip == attack.Ip && items[i].First_browser == attack.Browser {
				items[i].Attack_count = attack.Count
				break
			}
		}
	}

	return items, nil
}
