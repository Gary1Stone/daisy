package ctrls

import (
	"fmt"
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
		alert := "" // Add alerts if present
		if item.Alerts > 0 {
			alert = buildAlertIcon(item.Alerts)
		}
		ticket := "" // Add tickets if present
		if item.Tickets > 0 {
			ticket = buildTicketIcon(item.Tickets)
		}
		fmt.Fprintf(&table, "<tr><td><a href='profile.html?uid=%d'>%s</a></td><td>%s</td><td>%s</td><td> %s %s</td></tr>", item.Uid, item.User, mxl25(item.Fullname), mxl25(item.Group), alert, ticket)
	}

	table.WriteString("</tbody></table>")
	return table.String()
}

/*
   ⏶ &#9206; &#x23f6; Alphabetic sort
   ⏷ &#9207; &#x23f7; Reverse Alphabetic Sort
	⏴
	⏵
*/
// Helper function to build the table header
func buildProfileTableHeader() string {
	return `<table id="profileTable">
    <thead>
    <tr>
        <th aria-sort="ascending" data-sort="asc">User ID</th>
        <th aria-sort="none">Name</th>
        <th aria-sort="none">Group</th>
        <th aria-sort="none">Alerts&sol;Tickets</th>
    </tr>
    </thead>
    <tbody>`
}

// Helper function to build the alert icon
func buildAlertIcon(alerts int) string {
	return svg.GetIcon("bell") + " (" + strconv.Itoa(alerts) + ") "
}

// Helper function to build the ticket icon
func buildTicketIcon(tickets int) string {
	return svg.GetIcon("ticket") + " (" + strconv.Itoa(tickets) + ") "
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
