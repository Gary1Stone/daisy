package ctrls

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

func BuildAttacksTable(curUid, duration int) string {
	var table strings.Builder

	// Build the table header
	table.WriteString(buildAttacksTableHeader())

	// Fetch attacks items
	items, err := db.GetAttacksDetails(curUid, duration)
	if err != nil {
		log.Println(err)
		table.WriteString("</tbody></table>")
		return table.String()
	}

	// Build table rows
	icon := svg.GetIcon("map")
	for _, item := range items {
		table.WriteString(buildAttacksTableRow(&item, icon))
	}

	table.WriteString("</tbody></table>")
	return table.String()
}

// Helper function to build the table header
func buildAttacksTableHeader() string {
	icon := svg.GetIcon("user")
	return fmt.Sprintf(`<table class='striped' id="attackstable" >
    <thead>
    <tr>
        <th aria-sort="ascending" data-sort="asc">Occurred</th>
        <th aria-sort="none">Attacking IP</th>
        <th aria-sort="none">Attacks</th>
        <th aria-sort="none">Browser</th>
        <th aria-sort="none">Location</th>
		<th aria-sort="none">%s</th>
		<th aria-sort="none">Map</th>
    </tr>
    </thead>
    <tbody>`, icon)
}

// Helper function to build a single table row
func buildAttacksTableRow(item *db.AttackInfo, icon string) string {
	var row strings.Builder

	place := "<p>" + item.Country + "</p><p>" + item.State + " / " + item.City + "</p>"
	if item.City != item.Community {
		place += "<p>" + item.Community + "</p>"
	}

	fmt.Fprintf(&row, `<tr class='row-hover'><td>%s</td><td>%s</td>`, item.Occurred, item.Ip)
	fmt.Fprintf(&row, `<td>%d</td><td>%s</td>`, item.Attack_count, item.First_browser)
	fmt.Fprintf(&row, `<td>%s</td><td>%s</td>`, place, item.Fullname)
	fmt.Fprintf(&row, `<td><a href='https://www.google.com/maps/search/?api=1&query=%s, %s' target='_blank'>%s</a></td></tr>`, strconv.FormatFloat(item.Latitude, 'f', -1, 64), strconv.FormatFloat(item.Longitude, 'f', -1, 64), icon)

	return row.String()
}
