package db

import (
	"database/sql"
	"log"
)

type AssetPrefixes struct {
	Id     int    `json:"id"`
	Type   string `json:"type"`   // Device Type
	Prefix string `json:"prefix"` // Prefix
	Icon   string `json:"icon"`
	Update bool   `json:"update"` // Does the record need to be updated?
}

func GetAssetPrefixes(uid int) ([]AssetPrefixes, error) {
	items := make([]AssetPrefixes, 0)
	query := `
	SELECT A.id, A.description, A.asset_id, B.icon
	FROM choices A
	LEFT JOIN icons B ON A.code = B.name
	WHERE A.field = 'TYPE' AND B.is_device = 1
	ORDER BY A.seq
	`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		if err == sql.ErrNoRows {
			return items, nil
		}
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item AssetPrefixes
		err := rows.Scan(&item.Id, &item.Type, &item.Prefix, &item.Icon)
		if err != nil {
			log.Println(err)
		} else {
			item.Update = false
			items = append(items, item)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return items, err
	}
	return items, nil
}

func SetAssetPrefixes(items *[]AssetPrefixes) bool {
	for _, item := range *items {
		if item.Update {
			query := "UPDATE choices SET asset_id=? WHERE id=?"
			_, err := Conn.Exec(query, item.Prefix, item.Id)
			if err != nil {
				log.Println(err)
				return false
			}
		}
	}
	return true
}
