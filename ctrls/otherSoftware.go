package ctrls

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
)

func BuildOtherSoftwareTable() string {
	var tbl strings.Builder
	tbl.WriteString(`<table class='striped' id='other_table'><thead><tr>
		<th aria-sort="ascending" data-sort="asc">Software</th>
		<th aria-sort="none">Installs</th>
		<th aria-sort="none">Unwanted?</th>
		</tr></thead><tbody>`)

	var otherSoftware db.OtherSoftware
	items, err := otherSoftware.List()
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
