package m30

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/util"
)

type M30msgs struct {
	Cmd     string      `json:"cmd"`     // HWM (High-water Mark) or SAVE
	ApiKey  string      `json:"apikey"`  // The API key
	Records []M30status `json:"records"` // The messages
}

type M30status struct {
	Id         int    `json:"id"`         // primary key
	Created    int    `json:"created"`    // UTC created date
	Report     int    `json:"Report"`     // YYYYMMDDHH of the report (groups rows into a report)
	Title      string `json:"title"`      // Name of the issue found
	Severity   int    `json:"severity"`   // How severe is the issue (1 to 4)
	Occurances int    `json:"occurances"` // How many times the issue was in the log file
	Occurs     int    `json:"occurs"`     // The first ocurance of the issue in the log file
	Sample1    string `json:"sample1"`    // An example of the issue copied from the log file
	Sample2    string `json:"sample2"`    // A second example
	Sample3    string `json:"sample3"`    // A third example
	SrcId      int    `json:"srcid"`      // Source ID
	ReportDate string `json:"reportdate"` // YYYY-MM-DD
}

func (m *M30msgs) GetM30Hwm() (int, error) {
	cnt := 0
	query := "SELECT COALESCE(MAX(src_id), 0) FROM m30status" //If table is empty, would get a null error
	err := db.Conn.QueryRow(query).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return 0, err
	}
	return cnt, err
}

func (m *M30msgs) SaveM30() error {
	tx, err := db.Conn.Begin()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()

	//Prepare the insert statement
	const Query = "INSERT OR IGNORE INTO m30status (src_id, created, report, title, severity, occurances, occurs, sample1, sample2, sample3) VALUES (?,?,?,?,?,?,?,?,?,?)"
	stmt, err := db.Conn.Prepare(Query)
	if err != nil {
		log.Printf("Error preparing M30 insert statement: %v", err)
		return err
	}

	for _, r := range m.Records {
		_, err = stmt.Exec(r.Id, r.Created, r.Report, r.Title, r.Severity, r.Occurances, r.Occurs, r.Sample1, r.Sample2, r.Sample3)

		if err != nil {
			log.Printf("Error saving M30 record for %v: %v", r.Id, err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
	}

	return nil
}

func getM30status(tzoff int) ([]M30status, error) {
	var items []M30status
	query := `SELECT id, created, strftime('%Y-%m-%d %H', created-?, 'unixepoch') AS reportdate, src_id, 
		report, title, severity, occurances, occurs, sample1, sample2, sample3
		FROM m30status
		WHERE report = (SELECT report FROM m30status WHERE id = (SELECT MAX(id) FROM m30status))
		ORDER BY src_id` //created > strftime('%s', 'now', '-7 days')

	rows, err := db.Conn.Query(query, tzoff)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var item M30status
		err := rows.Scan(&item.Id, &item.Created, &item.ReportDate, &item.SrcId, &item.Report, &item.Title,
			&item.Severity, &item.Occurances, &item.Occurs, &item.Sample1, &item.Sample2, &item.Sample3)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		item.Sample1 = util.Sanitize(item.Sample1)
		item.Sample2 = util.Sanitize(item.Sample2)
		item.Sample3 = util.Sanitize(item.Sample3)
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return items, nil
}

func BuildM30Report(tzoff int) string {
	var ctrl strings.Builder

	items, err := getM30status(tzoff)
	if err != nil {
		return ""
	}

	for _, item := range items {
		fmt.Fprintf(&ctrl, `<details><summary>%s<b>%s</b> [%s]</summary>`, item.ReportDate, item.Title, sev2txt(item.Severity))
		fmt.Fprintf(&ctrl, `<p>%d Occurrences, First line: %d </p>`, item.Occurances, item.Occurs)
		cnt := 0
		if len(item.Sample1) > 0 {
			fmt.Fprintf(&ctrl, `<p>%s</p>`, item.Sample1)
			cnt++
		}
		if len(item.Sample2) > 0 {
			fmt.Fprintf(&ctrl, `<p>%s</p>`, item.Sample2)
			cnt++
		}
		if len(item.Sample3) > 0 {
			fmt.Fprintf(&ctrl, `<p>%s</p>`, item.Sample3)
			cnt++
		}
		if item.Occurances-cnt > 0 {
			fmt.Fprintf(&ctrl, `<p> and %d more occurances</p>`, item.Occurances-cnt)
		}
		fmt.Fprintf(&ctrl, `</details>`)
	}

	return ctrl.String()
}

func sev2txt(sev int) string {
	switch sev {
	case 4:
		return "Low"
	case 3:
		return "Medium"
	case 2:
		return "High"
	case 1:
		return "Critical"
	}
	return ""
}

type summary struct {
	severity int
	title    string
	total    int
}

func getM30Summary() (items []summary, err error) {
	query := `SELECT severity, title, COUNT(*) AS total
		FROM m30status
		WHERE severity > 0 AND created > strftime('%s', 'now', '-7 days')
		GROUP BY severity, title
		ORDER BY severity, title`

	rows, err := db.Conn.Query(query)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	// Loop through rows, using Scan to assign column data to struct fields.
	for rows.Next() {
		var item summary
		err = rows.Scan(&item.severity, &item.title, &item.total)
		if err != nil {
			log.Println(err)
			return
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}

	return
}

func BuildM30Summary() string {
	var ctrl strings.Builder
	items, err := getM30Summary()
	if err != nil {
		log.Println(err)
		return ""
	}

	fmt.Fprintf(&ctrl, `<table><thead><tr><th>Severity</th><th>Title</th><th>Count</th></thead><tbody>`)

	for _, item := range items {
		fmt.Fprintf(&ctrl, `<tr><td>%s</td><td>%s</td><td>%d</td></tr>`, sev2txt(item.severity), item.title, item.total)
	}

	fmt.Fprintf(&ctrl, `</tbody></table>`)

	return ctrl.String()
}
