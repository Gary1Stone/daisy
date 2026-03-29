package ctrls

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
)

// Generate profile table for wide screens
func ProfilesTable(curUid int, filter db.ProfileFilter) string {
	var table strings.Builder

	// Build the table header
	table.WriteString(buildProfileTableHeader())

	// Fetch profile items
	items, err := filter.GetProfiles(curUid)
	if err != nil {
		log.Println(err)
		table.WriteString("</tbody></table>")
		return table.String()
	}

	// Build table rows
	for _, item := range items {
		table.WriteString(buildProfileTableRow(&item))
	}

	table.WriteString("</tbody></table>")
	return table.String()
}

// Helper function to build the table header
func buildProfileTableHeader() string {
	return `<table data-role="table" id="profiletable" 
    data-rows="50" data-show-rows-steps="false" 
    data-show-search="true" 
    data-table-search-title="<span class='mif-search'></span>" 
    data-show-pagination="true" 
    data-show-table-info="true" 
    data-horizontal-scroll="true" 
    class="table striped table-border row-border row-hover">
    <thead>
    <tr>
        <th data-sortable="true" data-show="false">UID</th>
        <th data-sortable="true">User ID</th>
        <th data-sortable="true">Name</th>
        <th data-sortable="true">Queue</th>
        <th data-sortable="true">Alerts&sol;Tickets</th>
    </tr>
    </thead>
    <tbody>`
}

// Helper function to build a single table row
func buildProfileTableRow(item *db.Profile) string {
	var row strings.Builder
	row.WriteString("<tr class='row-hover'><td data-label='UID'>")
	row.WriteString(strconv.Itoa(item.Uid))
	row.WriteString("</td><td><span class='mif-user icon'></span> ")
	row.WriteString(item.User)
	row.WriteString("</td><td>")
	row.WriteString(mxl25(item.Fullname))
	row.WriteString("</td><td>")
	row.WriteString(mxl25(item.Group))
	row.WriteString("</td><td>&nbsp;")

	// Add alerts if present
	if item.Alerts > 0 {
		row.WriteString(buildAlertIcon(item.Color, item.Alerts))
	}

	// Add tickets if present
	if item.Tickets > 0 {
		row.WriteString(buildTicketIcon(item.Color, item.Tickets))
	}

	row.WriteString("</td></tr>")
	return row.String()
}

// Helper function to build the alert icon
func buildAlertIcon(color string, alerts int) string {
	return `<span class='mif-bell icon mif-1x ` + color + `'></span> (` + strconv.Itoa(alerts) + `)&nbsp;`
}

// Helper function to build the ticket icon
func buildTicketIcon(color string, tickets int) string {
	return `<span class='mif-news icon mif-1x ` + color + `'></span> (` + strconv.Itoa(tickets) + `)`
}

// set string to max length of 25 characters
func mxl25(str string) string {
	if len(str) > 25 {
		return str[:25] + "&hellip;"
	}
	return str
}

// set string to max length of 50 characters
func mxl50(str string) string {
	if len(str) > 50 {
		str = str[:50] + "&hellip;"
	}
	return str
}
