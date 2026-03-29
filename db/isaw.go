package db

import "log"

type Profiles struct {
	Uid   int
	Gid   int
	Email string
	Name  string
	Pin   string
}

func GetProfiles() ([]Profiles, error) {
	var profiles []Profiles
	query := `SELECT uid, gid, user, first FROM profiles WHERE active=1 ORDER BY fullname`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return profiles, err
	}
	defer rows.Close()
	for rows.Next() {
		var profile Profiles
		err := rows.Scan(&profile.Uid, &profile.Gid, &profile.Email, &profile.Name)
		if err != nil {
			log.Println(err)
		} else {
			profiles = append(profiles, profile)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return profiles, nil
}

type Computers struct {
	Cid      int    `json:"cid"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Site     string `json:"site"`
	Office   string `json:"office"`
	Location string `json:"location"`
	Status   string `json:"status"`
	Make     string `json:"make"`
	Model    string `json:"model"`
	Year     string `json:"year"`
	Gid      int    `json:"gid"`
	Uid      int    `json:"uid"`
	Image    string `json:"image"`
}

func GetComputers() ([]Computers, error) {
	var computers []Computers
	query := `
		SELECT cid, name, type, coalesce(site, '') AS site, coalesce(office, '') AS office,
		location, status, make, model, year, coalesce(gid, 0) AS gid, coalesce(uid, 0) AS uid, image
		FROM devices 
		WHERE type='LAPTOP' AND active=1 
		ORDER BY name
	`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return computers, err
	}
	defer rows.Close()
	for rows.Next() {
		var computer Computers
		err := rows.Scan(&computer.Cid, &computer.Name, &computer.Type, &computer.Site, &computer.Office,
			&computer.Location, &computer.Status, &computer.Make, &computer.Model, &computer.Year,
			&computer.Gid, &computer.Uid, &computer.Image)
		if err != nil {
			log.Println(err)
		} else {
			computers = append(computers, computer)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return computers, nil
}
