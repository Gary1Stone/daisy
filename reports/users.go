package reports

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
)

/*
 *
 * Generate the Users report - last login and where, if banned, not logged in for 90+ days
 *
 */
func GetUsersReport(curUid int) string {
	items, err := db.GetUsersReport(curUid)
	if err != nil {
		log.Println(err)
		return ""
	}
	var report strings.Builder
	report.WriteString("<table data-role='table' id='dash_table' data-rows='50' ")
	report.WriteString("data-show-rows-steps='true' data-show-search='true' ")
	report.WriteString("data-show-pagination='true' data-show-table-info='true' ")
	report.WriteString("data-horizontal-scroll='true'")
	report.WriteString("class='table striped table-border row-border row-hover compact' >")
	report.WriteString("<thead>")
	report.WriteString("<tr>")
	report.WriteString("<th data-sortable='true'>User</th>")
	report.WriteString("<th data-sortable='true'>Last Login</th>")
	report.WriteString("<th data-sortable='true'>Location</th>")
	report.WriteString("<th data-sortable='true'>Active</th>")
	report.WriteString("<th data-sortable='true'>Banned</th>")
	report.WriteString("</tr>")
	report.WriteString("</thead>")
	report.WriteString("<tbody>")
	for _, item := range items {
		report.WriteString("<tr>")
		//user
		report.WriteString("<td><a href='profile.html?uid=")
		report.WriteString(strconv.Itoa(item.Uid))
		report.WriteString("'>")
		report.WriteString(item.Fullname)
		report.WriteString("</a><p>")
		report.WriteString(item.Group)
		report.WriteString("</p>")
		report.WriteString("</td>")
		//Login Time
		report.WriteString("<td>")
		report.WriteString(item.LastLogin)
		if item.Days > 0 || len(item.LastLogin) > 0 {
			report.WriteString("<p>(")
			report.WriteString(strconv.Itoa(item.Days))
			report.WriteString(" Days Ago)</p>")
		}
		report.WriteString("</td>")
		//location
		report.WriteString("<td>")
		report.WriteString(item.City)
		report.WriteString("<p>")
		report.WriteString(item.Community)
		report.WriteString("</p>")
		report.WriteString("</td>")
		//Active
		if item.Active == 1 {
			report.WriteString("<td>Yes</td>")
		} else {
			report.WriteString("<td><p class='fg-red'>No</p></td>")
		}
		//Banned
		if item.Banned {
			report.WriteString("<td><p class='fg-red'>Yes</p></td>")
		} else {
			report.WriteString("<td>No</td>")
		}
		report.WriteString("</tr>")
	}
	report.WriteString("</tbody>")
	report.WriteString("</table>")
	return report.String()
}
