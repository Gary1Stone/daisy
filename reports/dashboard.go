package reports

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
	"github.com/gbsto/daisy/util"
)

func GetDashboardReport() string {
	items := db.GetDashboard()
	var report strings.Builder
	report.WriteString("<table data-role='table' id='dash_table' data-rows='-1' ")
	report.WriteString("data-show-rows-steps='false' data-show-search='false' ")
	report.WriteString("data-show-pagination='false' data-show-table-info='false' ")
	report.WriteString("class='table striped table-border row-border row-hover compact' >")
	report.WriteString("<thead>")
	report.WriteString("<tr>")
	report.WriteString("<th data-sortable='true'>Device</th>")
	report.WriteString("<th data-sortable='true'>Quantity</th>")
	report.WriteString("<th data-sortable='true'>Lost/Died</th>")
	report.WriteString("<th data-sortable='true'>Issues</th>")
	report.WriteString("<th data-sortable='true'>In Storage</th>")
	report.WriteString("</tr>")
	report.WriteString("</thead>")
	report.WriteString("<tbody>")
	for _, item := range items {
		report.WriteString("<tr>")
		report.WriteString("<td><a href=\"#\" onclick=\"getReport('DEVICES_REPORT', '")
		report.WriteString(item.Type)
		report.WriteString("');\">")
		report.WriteString("<span class='")
		report.WriteString(item.Icon)
		report.WriteString("'></span>&nbsp;")
		report.WriteString(item.Label)
		report.WriteString("</a></td>")
		report.WriteString("<td>")
		report.WriteString(strconv.Itoa(item.Count))
		report.WriteString("</td>")
		report.WriteString("<td>")
		report.WriteString(strconv.Itoa(item.Unavailable))
		report.WriteString("</td>")
		if item.PendingAction == 0 {
			report.WriteString("<td>0</td>")
		} else {
			report.WriteString("<td><a href='#' onclick=\"getReport('ISSUES_REPORT', '")
			report.WriteString(item.Type)
			report.WriteString("');\">")
			report.WriteString(strconv.Itoa(item.PendingAction))
			report.WriteString("</a></td>")
		}
		report.WriteString("<td>")
		report.WriteString(strconv.Itoa(item.Instorage))
		report.WriteString("</td>")
		report.WriteString("</tr>\n")
	}
	report.WriteString("</tbody>")
	report.WriteString("</table>")
	return report.String()
}

func GetDeviceCounts() string {
	items := db.GetDashboard()
	var report strings.Builder
	report.WriteString("<table data-role='table' id='dash_table' data-rows='-1' ")
	report.WriteString("data-show-rows-steps='false' data-show-search='false' ")
	report.WriteString("data-show-pagination='false' data-show-table-info='false' ")
	report.WriteString("class='table striped table-border row-border row-hover compact' >")
	report.WriteString("<thead>")
	report.WriteString("<tr>")
	report.WriteString("<th data-sortable='true'>Device</th>")
	report.WriteString("<th data-sortable='true'>Quantity</th>")
	report.WriteString("<th data-sortable='true'>Issues</th>")
	report.WriteString("<th data-sortable='true'>Storage</th>")
	report.WriteString("</tr>")
	report.WriteString("</thead>")
	report.WriteString("<tbody>")
	for _, item := range items {
		report.WriteString("<tr>")
		report.WriteString("<td>")
		report.WriteString("<span class='")
		report.WriteString(item.Icon)
		report.WriteString("'></span>&nbsp;")
		report.WriteString(item.Label)
		report.WriteString("</td>")
		report.WriteString("<td>")
		report.WriteString(strconv.Itoa(item.Count))
		report.WriteString("</td>")
		if item.PendingAction == 0 {
			report.WriteString("<td>0</td>")
		} else {
			report.WriteString("<td>")
			report.WriteString(strconv.Itoa(item.PendingAction))
			report.WriteString("</td>")
		}
		report.WriteString("<td>")
		report.WriteString(strconv.Itoa(item.Instorage))
		report.WriteString("</td>")
		report.WriteString("</tr>\n")
	}
	report.WriteString("</tbody>")
	report.WriteString("</table>")
	return report.String()
}

/*
 * Issues Report
 *
 */
func GetIssuesReport(curUid int, devType string) string {
	var report strings.Builder
	items, err := db.GetDeviceIssues(curUid, devType)
	if err != nil {
		log.Println(err)
		return "ERROR: Cannot Generate Issues Report"
	}
	report.WriteString(GetDashboardReport())
	report.WriteString("<table data-role='table' id='dash_table' data-rows='-1' ")
	report.WriteString("data-show-rows-steps='false' data-show-search='false' ")
	report.WriteString("data-show-pagination='true' data-show-table-info='true' ")
	report.WriteString("data-horizontal-scroll='true'")
	report.WriteString("class='table striped table-border row-border row-hover compact' >")
	report.WriteString("<thead>")
	report.WriteString("<tr>")
	report.WriteString("<th data-sortable='true'>Device ID</th>")
	report.WriteString("<th data-sortable='true'>Opened</th>")
	report.WriteString("<th data-sortable='true'>Issue</th>")
	report.WriteString("</tr>")
	report.WriteString("</thead>")
	report.WriteString("<tbody>")
	for _, item := range items {
		//Go To Button
		report.WriteString("<td>")
		report.WriteString(getButtonCtrl(item.Cid, 0, item.Devicetype, item.Devicename))
		report.WriteString("</td>")
		//Details
		report.WriteString("<td>")
		report.WriteString(item.Localtime)
		report.WriteString("<p> with impact ")
		report.WriteString(db.GetCodeDescription("IMPACT", item.Impact))
		report.WriteString("</p></td>")
		//Details
		report.WriteString("<td>")
		report.WriteString(item.Notes)
		report.WriteString("</td>")
		report.WriteString("</tr>")
	}
	report.WriteString("</tbody>")
	report.WriteString("</table>")
	return report.String()
}

/*
 * Devices Report
 *
 */
func GetDeviceReport(curUid int, devType string) string {
	var report strings.Builder
	items, err := db.GetDevicesByType(curUid, devType)
	if err != nil {
		log.Println(err)
	}
	report.WriteString(GetDashboardReport())

	report.WriteString("<table data-role='table' id='dash_table' data-rows='-1' ")
	report.WriteString("data-show-rows-steps='false' data-show-search='false' ")
	report.WriteString("data-show-pagination='true' data-show-table-info='true' ")
	report.WriteString("data-horizontal-scroll='true'")
	report.WriteString("class='table striped table-border row-border row-hover compact' >")
	report.WriteString("<thead>")
	report.WriteString("<tr>")
	report.WriteString("<th data-sortable='true'>Photo</th>")
	report.WriteString("<th data-sortable='true'>Device</th>")
	report.WriteString("<th data-sortable='true'>Details</th>")
	report.WriteString("<th data-sortable='true'>Location</th>")
	report.WriteString("<th data-sortable='true'>Manufactured</th>")
	report.WriteString("</tr>")
	report.WriteString("</thead>")
	report.WriteString("<tbody>")

	for _, item := range items {
		report.WriteString("<tr>")
		//Photo
		report.WriteString("<td>")
		report.WriteString(util.GetThumbnail(item.Image))
		report.WriteString("</td>")
		//Go To Button
		report.WriteString("<td>")
		report.WriteString(getButtonCtrl(item.Cid, 0, item.Type, item.Name))
		report.WriteString("</td>")
		//Details
		report.WriteString("<td>")
		report.WriteString("<p>")
		report.WriteString(db.GetCodeDescription("MAKE", item.Make))
		report.WriteString(" ")
		report.WriteString(item.Model)
		report.WriteString(" <i>(")
		report.WriteString(db.GetCodeDescription("STATUS", item.Status))
		report.WriteString(")</i></p>")
		report.WriteString("<ul>")
		if item.Ram > 0 || len(item.Cpu) > 0 || item.Drivesize > 0 {
			report.WriteString("<li>")

			if item.Ram > 0 {
				report.WriteString("RAM: ")
				report.WriteString(strconv.Itoa(item.Ram))
				report.WriteString(" GB ")
			}

			if len(item.Cpu) > 0 {
				report.WriteString("CPU: ")
				report.WriteString(item.Cpu)
				report.WriteString(" ")
			}

			if item.Drivesize > 0 {
				report.WriteString(strconv.Itoa(item.Drivesize))
				report.WriteString(" GB ")
				report.WriteString(item.Drivetype)
			}
			report.WriteString("</li>\n")
		}
		if len(item.Assigned) > 0 {
			report.WriteString("<li>Used by: ")
			report.WriteString(item.Assigned)
			report.WriteString("</li>\n")
		}

		if len(item.Notes) > 0 {
			report.WriteString("<li>")
			report.WriteString(item.Notes)
			report.WriteString("</li>\n")
		}
		if item.Type == "DESKTOP" || item.Type == "LAPTOP" {
			report.WriteString("<li>")
			report.WriteString(db.GetCodeDescription("OS", item.Os))
			report.WriteString("</li>\n")
		}
		report.WriteString("</ul>")
		report.WriteString("</td>")
		//Location
		report.WriteString("<td>")
		report.WriteString(db.GetCodeDescription("SITE", item.Site))
		report.WriteString(" ")
		report.WriteString(db.GetCodeDescription("OFFICE", item.Office))
		if len(item.Location) > 0 {
			report.WriteString("<p>")
			report.WriteString(item.Location)
			report.WriteString("</p>")
		}
		report.WriteString("</td>")
		report.WriteString("<td>")
		if item.Year > 0 {
			report.WriteString(strconv.Itoa(item.Year))
		} else {
			report.WriteString("&nbsp;")
		}
		report.WriteString("</td>")
		report.WriteString("</tr>")
	}
	report.WriteString("</tbody>")
	report.WriteString("</table>")
	return report.String()
}

/*
 *
 * Generate the link button to the device record with color
 *
 */
// <a class='button alert' href='device.html?cid=21' role='button' ><span class='mif-laptop icon'></span>&nbsp;WKNC-20</a>
func getButtonCtrl(cid, days int, devType, name string) string {
	var ctrl strings.Builder
	ctrl.WriteString("<a class='button")
	if days > 90 {
		ctrl.WriteString(" alert")
	}
	ctrl.WriteString("' href='device.html?cid=")
	ctrl.WriteString(strconv.Itoa(cid))
	ctrl.WriteString("' role='button' ><span class='")
	ctrl.WriteString(svg.GetIcon(db.FindIconNameByName(devType)))
	ctrl.WriteString(" icon'></span>&nbsp;")
	ctrl.WriteString(name)
	ctrl.WriteString("</a>")
	return ctrl.String()
}
