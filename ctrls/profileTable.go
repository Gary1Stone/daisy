package ctrls

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
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
	return `<table id="profileTable">
    <thead>
    <tr>
        <th>UID</th>
        <th>User ID</th>
        <th>Name</th>
        <th>Queue</th>
        <th>Alerts&sol;Tickets</th>
    </tr>
    </thead>
    <tbody>`
}

// Helper function to build a single table row
func buildProfileTableRow(item *db.Profile) string {
	var row strings.Builder
	row.WriteString("<tr><td>")
	row.WriteString(strconv.Itoa(item.Uid))
	row.WriteString("</td><td><a href='profile.html?uid=")
	row.WriteString(strconv.Itoa(item.Uid))
	row.WriteString("'>")
	row.WriteString(item.User)
	row.WriteString("</a></td><td>")
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
	return "<span class='" + color + "'>" + svg.GetIcon("bell.svg") + "</span> (" + strconv.Itoa(alerts) + ")&nbsp;"
}

// Helper function to build the ticket icon
func buildTicketIcon(color string, tickets int) string {
	return "<span class='" + color + "'>" + svg.GetIcon("ticket.svg") + "</span> (" + strconv.Itoa(tickets) + ")&nbsp;"
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
