package db

import (
	"database/sql"
	"log"
)

type OtherSoftware struct {
	Id                    int    `json:"id"`
	Name                  string `json:"name"`
	Installs              int    `json:"installs"`
	Active                bool   `json:"active"`
	IsBad                 bool   `json:"isbad"`
	ActiveInstalls        int    `json:"activeinstalls"`
	DecomissionedInstalls int    `json:"decomissioninstalls"`
}

// Get the software list not being tracked
func (o *OtherSoftware) List() ([]OtherSoftware, error) {
	items := make([]OtherSoftware, 0)
	query := `SELECT COUNT(*) AS cnt, A.name, D.active,
			CASE
				WHEN EXISTS (SELECT 1 FROM sw_bad B WHERE B.name = A.name) THEN 1 ELSE 0
			END AS bad
		FROM sw_inv A
		LEFT JOIN devices D ON D.cid = A.cid
		WHERE sid IS NULL
		GROUP BY A.name
		ORDER BY A.name
	`
	rows, err := Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item OtherSoftware
		var active, cnt int
		err := rows.Scan(&cnt, &item.Name, &active, &item.IsBad)
		if err != nil {
			log.Println(err)
		} else {
			if active == 1 {
				item.ActiveInstalls = cnt
			} else {
				item.DecomissionedInstalls = cnt
			}
			items = append(items, item)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return items, err
}
