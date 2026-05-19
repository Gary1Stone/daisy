package reports

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
	"github.com/gbsto/daisy/util"
)

func GetDashboardReport() string {
	items := db.GetDashboard()
	var report strings.Builder
	report.WriteString(`<table class='striped' data-sortable='true' id="dashboardTable">
    <thead>
    <tr>
        <th aria-sort="ascending" data-sort="asc">Device</th>
        <th aria-sort="none">Quantity</th>
        <th aria-sort="none">Lost&sol;Died</th>
        <th aria-sort="none">Issues</th>
        <th aria-sort="none">Storage</th>
    </tr>
    </thead>
    <tbody>`)

	for _, item := range items {
		pendingAction := "0"
		if item.PendingAction > 0 {
			pendingAction = fmt.Sprintf(`<a href="#" onclick="getReport('ISSUES_REPORT', '%s');">%d</a>", item.Type, item.PendingAction'`, item.Type, item.PendingAction)
		}
		fmt.Fprintf(&report, `<tr><td><a href="#" onclick="getReport('DEVICES_REPORT', '%s');">%s %s</a></td>`, item.Type, svg.GetIcon(item.Icon), item.Label)
		fmt.Fprintf(&report, `<td>%d</td><td>%d</td><td>%s</td><td>%d</td></tr>`, item.Count, item.Unavailable, pendingAction, item.Instorage)
	}
	report.WriteString("</tbody></table>")
	return report.String()
}

func GetDeviceCounts() string {
	items := db.GetDashboard()
	var report strings.Builder
	report.WriteString(`<table class='striped' data-sortable='true' id="dash_table">
    <thead>
    <tr>
        <th aria-sort="ascending" data-sort="asc">Device</th>
        <th aria-sort="none">Quantity</th>
        <th aria-sort="none">Issues</th>
        <th aria-sort="none">Storage</th>
    </tr>
    </thead>
    <tbody>`)
	for _, item := range items {
		fmt.Fprintf(&report, `<tr><td>%s %s</td><td>%d</td><td>%d</td><td>%d</td></tr>`, svg.GetIcon(item.Icon), item.Label, item.Count, item.PendingAction, item.Instorage)
	}
	report.WriteString("</tbody></table>")
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
	report.WriteString(`<table class='striped' data-sortable='true' id="dash_table">
    <thead>
    <tr>
        <th aria-sort="ascending" data-sort="asc">Device</th>
        <th aria-sort="none">Opened</th>
        <th aria-sort="none">Issue</th>
    </tr>
    </thead>
    <tbody>`)
	for _, item := range items {
		bttn := getButtonCtrl(item.Cid, 0, item.Devicetype, item.Icon, item.Devicename)
		impact := db.GetCodeDescription("IMPACT", item.Impact)
		fmt.Fprintf(&report, `<tr><td>%s</td><td>%s<p> with impact %s</p></td><td>%s</td></tr>`, bttn, item.Localtime, impact, item.Notes)
	}
	report.WriteString("</tbody></table>")
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
	report.WriteString(`<table class='striped' data-sortable='true' id="dash_table">
    <thead>
    <tr>
        <th aria-sort="ascending" data-sort="asc">Photo</th>
        <th aria-sort="none">Device</th>
        <th aria-sort="none">Detail</th>
        <th aria-sort="none">Location</th>
        <th aria-sort="none">Manufactured</th>
    </tr>
    </thead>
    <tbody>`)

	for _, item := range items {

		img := util.GetThumbnail(item.Image)
		bttn := getButtonCtrl(item.Cid, 0, item.Type, item.Icon, item.Name)
		make := db.GetCodeDescription("MAKE", item.Make) + " " + item.Model
		status := db.GetCodeDescription("STATUS", item.Status)
		ram := ""
		if item.Ram > 0 {
			ram = fmt.Sprintf("RAM: %d GB ", item.Ram)
		}
		cpu := ""
		if len(item.Cpu) > 0 {
			cpu = fmt.Sprintf("CPU: %s ", item.Cpu)
		}
		drivesize := ""
		if item.Drivesize > 0 {
			drivesize = fmt.Sprintf("Drive: %d GB %s ", item.Drivesize, item.Drivetype)
		}
		details := ""
		if item.Ram > 0 || len(item.Cpu) > 0 || item.Drivesize > 0 {
			details = fmt.Sprintf("<li>%s %s %s</li>", ram, cpu, drivesize)
		}
		assigned := ""
		if len(item.Assigned) > 0 {
			assigned = fmt.Sprintf("<li>Used by: %s</li>", item.Assigned)
		}
		notes := ""
		if len(item.Notes) > 0 {
			notes = fmt.Sprintf("<li>%s</li>", item.Notes)
		}
		os := ""
		if item.Type == "DESKTOP" || item.Type == "LAPTOP" {
			os = fmt.Sprintf("<li>%s</li>", db.GetCodeDescription("OS", item.Os))
		}
		site := db.GetCodeDescription("SITE", item.Site)
		office := db.GetCodeDescription("OFFICE", item.Office)
		location := ""
		if len(item.Location) > 0 {
			location = fmt.Sprintf("<p>%s</p>", item.Location)
		}
		year := ""
		if item.Year > 0 {
			year = fmt.Sprintf("<p>%d</p>", item.Year)
		}
		fmt.Fprintf(&report, `<tr><td>%s</td><td>%s</td><td>%s <p><i>(%s)</i></p>`, img, bttn, make, status)
		fmt.Fprintf(&report, `<ul>%s %s %s %s</ul></td>`, details, assigned, notes, os)
		fmt.Fprintf(&report, `<td>%s %s %s</td><td>%s</td></tr>`, site, office, location, year)
	}
	report.WriteString("</tbody></table>")
	return report.String()
}

/*
 *
 * Generate the link button to the device record with color
 *
 */
// <a class='button alert' href='device.html?cid=21' role='button' ><span class='mif-laptop icon'></span>&nbsp;WKNC-20</a>
func getButtonCtrl(cid, days int, devType, icon, name string) string {
	if len(icon) == 0 {
		icon = svg.GetIcon(db.FindIconNameByName(devType))
	} else {
		icon = svg.GetIcon(icon)
	}
	if days > 90 {
		icon = "<span class='fg-red`>" + icon + "</span>"
	}
	return fmt.Sprintf(`<a href='device.html?cid=%d'>%s %s</a>`, cid, icon, name)
}
