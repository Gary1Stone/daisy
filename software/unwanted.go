package software

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/icons"
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

func (u *Unwanted) listDevicesWithUnwantedSoftware() ([]UnwantedDevices, error) {
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
	rows, err := db.Conn.Query(query)
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
func (u *Unwanted) list() ([]Unwanted, error) {
	items := make([]Unwanted, 0)
	query := `SELECT id, name FROM sw_bad ORDER BY Name`
	rows, err := db.Conn.Query(query)
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
	_, err := db.Conn.Exec("DELETE FROM sw_bad WHERE id=?", u.Id)
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
	_, err := db.Conn.Exec("INSERT INTO sw_bad (name) VALUES (?)", u.Name)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Build table for Software page
func (u *Unwanted) BuildUnwantedEditTable() string {
	var tbl strings.Builder
	items, err := u.list()
	if err != nil {
		log.Println(err)
		return ""
	}

	tbl.WriteString(`<table class='striped' id='unwanted_table'><thead><tr>
		<th aria-sort="ascending" data-sort="asc">Unwanted</th><th aria-sort="none">Remove</th>
		</tr></thead><tbody>`)

	for _, item := range items {
		fmt.Fprintf(&tbl, `<tr><td>%s</td><td><a href="#" style="text-decoration: none;" title="remove from unwanted software list" onclick="deleteUnwanted(%d);">❌</a></td></tr>`, item.Name, item.Id)
	}
	fmt.Fprintf(&tbl, "</tbody></table>")
	return tbl.String()
}

// Generate profile table for wide screens
func (u *Unwanted) BuildUnwantedTable() string {
	var table strings.Builder

	// Build the table header with search
	table.WriteString(`<table class='striped' id="unwantedtable"><thead><tr>
        <th aria-sort='ascending' data-sort='asc'>Computer</th>
        <th aria-sort='none'>Site</th>
        <th aria-sort='none'>Office</th>
        <th aria-sort='none'>Unwanted Software</th>
    </tr></thead><tbody>`)

	// Fetch software items
	items, err := u.listDevicesWithUnwantedSoftware()
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	// Build table rows
	for _, item := range items {
		fmt.Fprintf(&table, `<tr><td><a href='computer.html?cid=%d'>%s %s</a></td><td>%s</td><td>%s</td><td>%s</td></tr>`, item.Cid, icons.GetIcon(item.Icon), item.Computer, item.SiteName, item.OfficeName, item.Software)
	}

	table.WriteString("</tbody></table>")
	return table.String()
}
