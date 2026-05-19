package ctrls

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
)

func BuildActiveUsersTable(curUid int) string {
	items, err := db.GetActiveUsers(curUid)
	if err != nil || len(items) == 0 {
		log.Println(err)
		return ""
	}
	var table strings.Builder
	table.WriteString(buildActiveUsersTableHeader())

	for _, item := range items {
		table.WriteString("<tr>")
		//Login Time
		table.WriteString("<td>")
		table.WriteString(item.Fullname)
		table.WriteString("</td><td>")
		//days Ago
		table.WriteString(item.Since)
		table.WriteString("</td><td>")
		//location
		table.WriteString(item.Country)
		table.WriteString(", ")
		table.WriteString(item.State)
		table.WriteString(", ")
		table.WriteString(item.City)
		table.WriteString("<p>")
		table.WriteString(item.Community)
		table.WriteString("</p></td><td>")
		// boot off
		table.WriteString(`<button type='button' onclick="endSession('`)
		table.WriteString(strconv.Itoa(item.Id))
		table.WriteString(`');"><span class='mif-settings-power mif-4x fg-red'></span></button>`)
		table.WriteString("</td></tr>")
	}
	table.WriteString("</tbody></table></div>")
	return table.String()

}

// Helper function to build the table header
func buildActiveUsersTableHeader() string {
	return `<div style="max-height: 400px; overflow-y: auto;">
	<table class='striped' data-sortable='true' id="activeuserstable" >
    <thead>
    <tr>
        <th aria-sort="ascending" data-sort="asc">User</th>
        <th aria-sort="none">Since</th>
        <th aria-sort="none">Location</th>
        <th>End Session</th>
    </tr>
    </thead>
    <tbody>`
}
