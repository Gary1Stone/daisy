package software

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
)

type Other struct {
	Id                    int    `json:"id"`
	Name                  string `json:"name"`
	Installs              int    `json:"installs"`
	Active                bool   `json:"active"`
	IsBad                 bool   `json:"isbad"`
	ActiveInstalls        int    `json:"activeinstalls"`
	DecomissionedInstalls int    `json:"decomissioninstalls"`
}

// Get the software list not being tracked
func (o *Other) List() ([]Other, error) {
	items := make([]Other, 0)
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
	rows, err := db.Conn.Query(query)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Other
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

func (o *Other) BuildOtherTable() string {
	var tbl strings.Builder
	tbl.WriteString(`<table class='striped' id='other_table'><thead><tr>
		<th aria-sort="ascending" data-sort="asc">Software</th>
		<th aria-sort="none">Installs</th>
		<th aria-sort="none">Unwanted?</th>
		</tr></thead><tbody>`)
	items, err := o.List()
	//items, err := o.List()
	if err != nil {
		log.Println(err)
		return ""
	}

	row := 0
	for _, item := range items {
		row++
		cnt := item.ActiveInstalls + item.DecomissionedInstalls
		if item.IsBad {
			fmt.Fprintf(&tbl, `<tr><td>%s</td><td>%d</td><td>👎</td></tr>`, item.Name, cnt)
		} else {
			fmt.Fprintf(&tbl, `<tr><td>%s</td><td>%d</td><td id="%d"><a href="#" style="text-decoration: none;"  title="add to unwanted software list" onclick='addUnwanted(%d,"%s");'>➕</a></td></tr>`, item.Name, cnt, row, row, item.Name)
		}
	}
	fmt.Fprintf(&tbl, "</tbody></table>")
	return tbl.String()
}
