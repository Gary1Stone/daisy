package ctrls

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

func BuildActiveUsersTable(curUid int) string {
	items, err := db.GetActiveUsers(curUid)
	if err != nil || len(items) == 0 {
		log.Println(err)
		return ""
	}
	var table strings.Builder
	table.WriteString(`<div style="max-height: 400px; overflow-y: auto;">
	<table class='striped' id="activeuserstable" >
    <thead>
    <tr>
        <th aria-sort="ascending" data-sort="asc">User</th>
        <th aria-sort="none">Since</th>
        <th aria-sort="none">Location</th>
        <th>End Session</th>
    </tr>
    </thead>
    <tbody>`)

	logoutIcon := svg.GetIcon("logout")

	for _, item := range items {

		// Login Time
		fmt.Fprintf(&table, `<tr><td>%s</td><td>%s</td>`, item.Fullname, item.Since)
		// location
		fmt.Fprintf(&table, `<td>%s, %s, <p>%s</p></td>`, item.Country, item.State, item.Community)
		// boot off
		fmt.Fprintf(&table, `<td><button type='button' class="outline" onclick="endSession('%d');"><span style='color:red;'>%s</span></button></td></tr>`, item.Id, logoutIcon)
	}
	table.WriteString("</tbody></table></div>")
	return table.String()

}
