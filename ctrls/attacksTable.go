package ctrls

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
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
	for _, item := range items {
		table.WriteString(buildAttacksTableRow(&item))
	}

	table.WriteString("</tbody></table>")
	return table.String()
}

// Helper function to build the table header
func buildAttacksTableHeader() string {
	return `<table class="striped" id="attackstable" >
    <thead>
    <tr>
        <th aria-sort="none">Occurred</th>
        <th aria-sort="none">Attacking IP</th>
        <th aria-sort="none">Attacks</th>
        <th aria-sort="none">Browser</th>
        <th aria-sort="none">Location</th>
		<th aria-sort="none"><span class='mif-user icon'></span></th>
		<th aria-sort="none">Map</th>
    </tr>
    </thead>
    <tbody>`
}

// Helper function to build a single table row
func buildAttacksTableRow(item *db.AttackInfo) string {
	var row strings.Builder
	place := "<p>" + item.Country + "</p><p>" + item.State + " / " + item.City + "</p>"
	if item.City != item.Community {
		place += "<p>" + item.Community + "</p>"
	}

	row.WriteString("<tr class='row-hover'><td>")
	row.WriteString(item.Occurred)
	row.WriteString("</td><td>")
	row.WriteString(item.Ip)
	row.WriteString("</td><td>")
	row.WriteString(strconv.Itoa(item.Attack_count))
	row.WriteString("</td><td><div class='gwrap'>")
	row.WriteString(item.First_browser)
	row.WriteString("</div></td><td>")
	row.WriteString(place)
	row.WriteString("</td><td><div class='gwrap'>")
	row.WriteString(item.Fullname)
	row.WriteString("</div></td><td>")
	row.WriteString("<a href='https://www.google.com/maps/search/?api=1&query=")
	row.WriteString(strconv.FormatFloat(item.Latitude, 'f', -1, 64))
	row.WriteString(",")
	row.WriteString(strconv.FormatFloat(item.Longitude, 'f', -1, 64))
	row.WriteString("' target='_blank'><span class='mif-map icon'></span>...</a>")
	row.WriteString("</td></tr>")
	return row.String()
}
