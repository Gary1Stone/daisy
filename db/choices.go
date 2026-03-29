package db

import (
	"database/sql"
	"log"
)

type ChoiceAdmin struct {
	Id          int    `json:"id" db:"id"`                   // Row ID
	Field       string `json:"field" db:"field"`             // Field name
	Code        string `json:"code" db:"code"`               // Field code (put into various tables)
	Description string `json:"description" db:"description"` // What the user sees
	Parent      string `json:"parent" db:"parent"`           // parent-child definition
	Active      int    `json:"active" db:"active"`           // Still can be used or not
	Sequence    int    `json:"sequence" db:"sequence"`       // Order of display
	Update      bool   `json:"update" db:"update"`           // Update pending
	Add         bool   `json:"add" db:"add"`                 // Add record pending
	Delete      bool   `json:"delete" db:"delete"`           // Delete record pending
	Inuse       bool   `json:"inuse" db:"inuse"`             // Code is used in various tables
	Task        string `json:"task" db:"task"`               // What task the code is for
	AssetId     string `json:"assetid" db:"asset_id"`        // Starting characters dependant on device type
	Permissions string `json:"permissions" db:"permissions"` // CRUD permsiions per group
}

// API struct
type Choices struct {
	Id          int    `json:"id"`
	Field       string `json:"field"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Seq         int    `json:"seq"`
	Active      int    `json:"active"`
	Parent      string `json:"parent"`
	Cnt         int    `json:"cnt"`
	Asset_id    string `json:"asset_id"`
	Permissions string `json:"permissions"`
}

// Admin table (choices) administration screens handling
func GetChoicesAdmin(field string) []ChoiceAdmin {
	//First get list of all the unused codes
	var items []ChoiceAdmin
	unused := make(map[string]bool)
	query := "SELECT DISTINCT(code) FROM choices WHERE field=? "

	switch field {
	case "IMPACT":
		query += "AND code NOT IN (SELECT DISTINCT impact FROM action_log WHERE impact IS NOT NULL)"
	case "GROUP":
		query += "AND code NOT IN (SELECT DISTINCT gid FROM profiles WHERE gid IS NOT NULL)"
	case "TROUBLE":
		query += "AND code NOT IN (SELECT DISTINCT trouble FROM action_log WHERE trouble IS NOT NULL)"
	case "GEOFENCE":
		query += ""
	case "KIND":
		query += ""
	default:
		query += "AND code NOT IN (SELECT DISTINCT " + field + " FROM devices WHERE " + field + " IS NOT NULL)"
	}

	rows, err := Conn.Query(query, field)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return items
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		err := rows.Scan(&code)
		if err != nil {
			log.Println(err)
		} else {
			unused[code] = true
		}
	}

	query = "SELECT id, field, code, description, seq, active, parent, asset_id, permissions FROM choices WHERE field=? ORDER BY seq"
	rows, err = Conn.Query(query, field)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var item ChoiceAdmin
		err := rows.Scan(&item.Id, &item.Field, &item.Code, &item.Description, &item.Sequence, &item.Active, &item.Parent, &item.AssetId, &item.Permissions)
		if err != nil {
			log.Println(err)
		} else {
			item.Add = false
			item.Delete = false
			item.Update = false
			item.Inuse = true
			if _, exists := unused[item.Code]; exists {
				item.Inuse = false // The code exists in the unused map (it is not used anywhere)
			}
			items = append(items, item)
		}
	}
	return items
}

// Add, Update, Delete functions for admin (choices) table administration
func (c *ChoiceAdmin) AddRecord() error {
	query := "INSERT INTO Choices (field, code, description, seq, active, parent, asset_id, permissions) VALUES (?,?,?,?,?,?,?,?)"
	_, err := Conn.Exec(query, c.Field, c.Code, c.Description, c.Sequence, 1, c.Parent, c.AssetId, c.Permissions)
	if err != nil {
		log.Println(err)
		return err
	}
	scheduleCacheReload()
	return nil
}

func (c *ChoiceAdmin) DeleteRecord() error {
	query := "DELETE FROM Choices WHERE id=?"
	_, err := Conn.Exec(query, c.Id)
	if err != nil {
		log.Println(err)
		return err
	}
	scheduleCacheReload()
	return nil
}

func (c *ChoiceAdmin) UpdateRecord() error {
	query := "UPDATE Choices SET field=?, code=?, description=?, seq=?, active=?, parent=?, asset_id=?, permissions=? WHERE id=?"
	_, err := Conn.Exec(query, c.Field, c.Code, c.Description, c.Sequence, c.Active, c.Parent, c.AssetId, c.Permissions, c.Id)
	if err != nil {
		log.Println(err)
		return err
	}
	scheduleCacheReload()
	return nil
}

// API
func GetApiChoices() ([]Choices, error) {
	var choices []Choices
	query := `
	SELECT id, field, code, description, seq, active, parent, cnt, asset_id, permissions 
	FROM choices
	WHERE field IN ('SITE', 'OFFICE', 'GROUP', 'MAKE', 'TYPE', 'STATUS', 'KIND')
	ORDER BY field, seq ASC
`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var choice Choices
		err := rows.Scan(&choice.Id, &choice.Field, &choice.Code, &choice.Description, &choice.Seq, &choice.Active, &choice.Parent, &choice.Cnt, &choice.Asset_id, &choice.Permissions)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		choices = append(choices, choice)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	return choices, nil
}
