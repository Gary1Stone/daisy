package db

import (
	"database/sql"
	"errors"
	"log"
)

type Unwanted struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type UnwantedDevices struct {
	Cid        int
	Computer   string
	Icon       string
	Site       string
	SiteName   string
	Office     string
	OfficeName string
	Software   string
}

func (u *Unwanted) ListDevicesWithUnwantedSoftware() ([]UnwantedDevices, error) {
	var items []UnwantedDevices
	const query = `SELECT D.cid, D.name AS computer, D.Site, D.Office, S.name AS software, 
		COALESCE(I.icon,'') AS icon,
		COALESCE(O.description, '') AS sitename,
		COALESCE(P.description, '') AS officename
		FROM SW_INV S
		LEFT JOIN devices D ON D.cid=S.cid
		LEFT JOIN icons AS I ON D.kind = I.name
		LEFT JOIN choices O ON D.site = O.code AND O.field='SITE'
		LEFT JOIN choices P ON D.office = P.code AND P.field='OFFICE' AND P.parent=D.site
		WHERE S.name IN (SELECT name FROM sw_bad)
		ORDER BY D.name`
	rows, err := Conn.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return items, nil
		}
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item UnwantedDevices
		err := rows.Scan(&item.Cid, &item.Computer, &item.Site, &item.Office, &item.Software, &item.Icon, &item.SiteName, &item.OfficeName)
		if err != nil {
			log.Println(err)
		} else {
			items = append(items, item)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println("ERROR iterating over software rows:", err)
	}

	return items, nil
}

// Get the bad software list
func (u *Unwanted) List() ([]Unwanted, error) {
	items := make([]Unwanted, 0)
	query := `SELECT id, name FROM sw_bad ORDER BY Name`
	rows, err := Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Unwanted
		err := rows.Scan(&item.Id, &item.Name)
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

func (u *Unwanted) Delete() error {
	if u.Id < 1 {
		return errors.New("invalid id")
	}
	_, err := Conn.Exec("DELETE FROM sw_bad WHERE id=?", u.Id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (u *Unwanted) Add() error {
	if len(u.Name) == 0 {
		return errors.New("invalid name")
	}
	_, err := Conn.Exec("INSERT INTO sw_bad (name) VALUES (?)", u.Name)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
