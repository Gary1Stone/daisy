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
	return `<table data-role="table" id="attackstable" 
    data-rows="-1" data-show-rows-steps="false" 
    data-show-search="true" 
    data-table-search-title="<span class='mif-search'></span>" 
    data-show-pagination="false" 
    data-show-table-info="true" 
    data-horizontal-scroll="true" 
    class="table striped table-border row-border row-hover">
    <thead>
    <tr>
        <th data-sortable="true">Occurred</th>
        <th data-sortable="true">Attacking IP</th>
        <th data-sortable="true">Attacks</th>
        <th data-sortable="true">Browser</th>
        <th data-sortable="true">Location</th>
		<th data-sortable="true"><span class='mif-user icon'></span></th>
		<th data-sortable="true">Map</th>
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
