package ctrls

import (
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
)

func BuildInventoryList() string {
	list, err := db.GetInventoryList()
	if err != nil {
		log.Println(err)
		return ""
	}
	var table strings.Builder
	table.WriteString(`<table id='inv_table' 
		class='table striped table-border row-border row-hover compact'>
		<thead><tr>
		<th>Software Titles</th>
		</tr></thead><tbody>`)
	for _, item := range list {
		table.WriteString(`<tr><td><a href='#' class='no-decor fg-black' onclick='fillSearch("`)
		table.WriteString(item)
		table.WriteString(`");'>`)
		table.WriteString(item)
		table.WriteString("</a></td></tr>")
	}
	table.WriteString("</tbody></table>")
	return table.String()
}
