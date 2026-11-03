package software

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/icons"
)

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
		fmt.Fprintf(&table, `<tr><td><a href='device.html?cid=%d'>%s %s</a></td><td>%s</td><td>%s</td><td>%s</td></tr>`, item.Cid, icons.GetIcon(item.Icon), item.Computer, item.SiteName, item.OfficeName, item.Software)
	}

	table.WriteString("</tbody></table>")
	return table.String()
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
