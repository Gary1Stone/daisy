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
		table.WriteString(`<button onclick="endSession('`)
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
	<table data-role="table" id="activeuserstable" 
    data-rows="-1" data-show-rows-steps="false" 
    data-show-search="true" 
    data-table-search-title="<span class='mif-search'></span>" 
    data-show-pagination="false" 
    data-show-table-info="false" 
    data-horizontal-scroll="true" 
    class="table striped table-border row-border row-hover">
    <thead>
    <tr>
        <th data-sortable="true">User</th>
        <th data-sortable="true">Since</th>
        <th data-sortable="true">Location</th>
        <th data-sortable="false">End Session</th>
    </tr>
    </thead>
    <tbody>`
}
