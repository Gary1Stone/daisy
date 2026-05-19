package reports

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/devices"
	"github.com/gbsto/daisy/util"
)

/*
 * Last Seen Report and Last Backed Up Report
 */

func GetLastSeenReport(curUid int, isSeen bool) string {
	items, err := db.GetDevices(curUid, 0, -1)
	if err != nil {
		log.Println(err)
		return ""
	}

	var report strings.Builder
	report.WriteString(buildTableHeader(isSeen))

	for _, item := range items {
		if !isSeen && item.Type != devices.Desktop && item.Type != devices.Laptop {
			continue
		}
		report.WriteString(buildTableRow(item, isSeen))
	}

	report.WriteString("</tbody></table>")
	return report.String()
}

func buildTableHeader(isSeen bool) string {
	title := "Last Backup Report"
	if isSeen {
		title = "Last Seen"
	}
	return `<table class='striped' id='last_table'  
		  
		  
		 
		>
		<thead><tr>
		<th aria-sort="ascending" data-sort="asc">Photo</th>
		<th aria-sort="none">Device</th>
		<th aria-sort="none">` + title + `</th>
		<th aria-sort="none">Location</th>
		<th aria-sort="none">Assigned To</th>
		<th aria-sort="none">Model</th>
		</tr></thead><tbody>`
}

func buildTableRow(item *db.Device, isSeen bool) string {
	var row strings.Builder
	row.WriteString("<tr>")
	row.WriteString("<td>" + util.GetThumbnail(item.Image) + "</td>")
	row.WriteString("<td>" + getButtonCtrl(item.Cid, getLastDays(item, isSeen), item.Type, item.Icon, item.Name) + "</td>")
	row.WriteString("<td>" + buildLastSeenOrBackup(item, isSeen) + "</td>")
	row.WriteString("<td>" + buildLocation(item) + "</td>")
	row.WriteString("<td>" + item.Assigned + "</td>")
	row.WriteString("<td>" + item.Model + "</td>")
	row.WriteString("</tr>")
	return row.String()
}

func getLastDays(item *db.Device, isSeen bool) int {
	if isSeen {
		return item.Last_seen_days
	}
	return item.Last_backup_days
}

func buildLastSeenOrBackup(item *db.Device, isSeen bool) string {
	if isSeen {
		return item.Last_seen_date + " by " + item.Last_seen_by + " (" + strconv.Itoa(item.Last_seen_days) + " days ago)"
	}
	return item.Last_backup_date + " by " + item.Last_backup_by + " (" + strconv.Itoa(item.Last_backup_days) + " days ago)"
}

func buildLocation(item *db.Device) string {
	var location strings.Builder
	location.WriteString(db.GetCodeDescription("SITE", item.Site) + " ")
	location.WriteString(db.GetCodeDescription("OFFICE", item.Office))
	if len(item.Location) > 0 {
		location.WriteString("<p>" + item.Location + "</p>")
	}
	return location.String()
}
