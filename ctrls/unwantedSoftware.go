package ctrls

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/icons"
)

// Build table for Software page
func BuildUnwantedTable() string {
	var tbl strings.Builder
	var unwanted db.Unwanted
	items, err := unwanted.List()
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
func BuildUnwantedSoftwareTable() string {
	var table strings.Builder

	// Build the table header with search
	table.WriteString(`<table class='striped' id="unwantedtable"><thead><tr>
        <th aria-sort='ascending' data-sort='asc'>Computer</th>
        <th aria-sort='none'>Site</th>
        <th aria-sort='none'>Office</th>
        <th aria-sort='none'>Unwanted Software</th>
    </tr></thead><tbody>`)

	// Fetch software items
	var unwanted db.Unwanted
	items, err := unwanted.ListDevicesWithUnwantedSoftware()
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
