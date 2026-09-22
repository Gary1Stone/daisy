package ctrls

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

func ComstatTable(uid int) string {
	var table strings.Builder

	auditLate, err := strconv.Atoi(os.Getenv("AUDIT_LATE"))
	if err != nil {
		auditLate = 28
	}

	seenLate, err := strconv.Atoi(os.Getenv("LAST_SEEN_LATE"))
	if err != nil {
		seenLate = 90
	}

	fileLate, err := strconv.Atoi(os.Getenv("BACKUP_FILE_LATE"))
	if err != nil {
		fileLate = 28
	}
	systemLate, err := strconv.Atoi(os.Getenv("BACKUP_SYSTEM_LATE"))
	if err != nil {
		systemLate = 180
	}
	diskLate, err := strconv.Atoi(os.Getenv("BACKUP_DISK_LATE"))
	if err != nil {
		diskLate = 180
	}

	// Build the table header with search
	table.WriteString(`<table class='striped' id="comstatstable">
    <thead>
    <tr>
        <th aria-sort='ascending' data-sort='asc'>Name</th>
        <th aria-sort='none'>Site</th>
        <th aria-sort='none'>Office</th>
        <th aria-sort='none'>Assigned to</th>
        <th aria-sort='none'>Status</th>
        <th aria-sort='none'>Last Seen</th>
        <th aria-sort='none'>Last Audit</th>
        <th aria-sort='none'>Community</th>
        <th aria-sort='none'>Backups: File</th>
        <th aria-sort='none'>Backups: System</th>
        <th aria-sort='none'>Backups: Disk</th>
        <th aria-sort='none'>Disk Space</th>
    </tr>
    </thead>
    <tbody>`)

	// Fetch all devices
	items, err := db.GetComstat(uid)
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	// Build table rows
	for _, item := range items {
		auditColor := "green"
		if auditLate > 0 && item.AuditDays > auditLate {
			auditColor = "red"
		}
		seenColor := "green"
		if seenLate > 0 && item.SeenDays > seenLate {
			seenColor = "red"
		}
		assigned := item.Assigned
		if assigned == "" {
			assigned = item.Group
		}

		//<tr data-id='%d' item.Cid
		fmt.Fprintf(&table, `<tr><td><a href='device.html?cid=%d'>%s %s</a></td><td>%s</td><td>%s</td><td>%s</td><td>%s</td>`,
			item.Cid, svg.GetIcon(item.Icon), item.Name, item.Site, item.Office, assigned, item.Status)

		fmt.Fprintf(&table, `<td><span style='color:%s;'  title="%d days ago">%s</span></td>`, seenColor, item.SeenDays, item.SeenDate)
		fmt.Fprintf(&table, `<td><span style='color:%s;'  title="%d days ago">%s</span></td>`, auditColor, item.AuditDays, item.AuditDate)
		fmt.Fprintf(&table, `<td><a href='https://www.google.com/maps/search/?api=1&query=%f,%f' target='_blank' title="WARNING: Location accuracy is only 10Km">%s</a></td>`, item.Latitude, item.Longitude, item.Community)

		// Build the backup information for each computer

		colourFile := "green"
		if item.FileDays > fileLate {
			colourFile = "red"
		}
		colourSystem := "green"
		if item.SystemDays > systemLate {
			colourSystem = "red"
		}
		colourDisk := "green"
		if item.DiskDays > diskLate {
			colourDisk = "red"
		}

		fmt.Fprintf(&table, `<td><span style='color:%s;' title="%d days ago">%s</span></td>`, colourFile, item.FileDays, item.FileDate)
		fmt.Fprintf(&table, `<td><span style='color:%s;' title="%d days ago">%s</span></td>`, colourSystem, item.SystemDays, item.SystemDate)
		fmt.Fprintf(&table, `<td><span style='color:%s;' title="%d days ago">%s</span></td>`, colourDisk, item.DiskDays, item.DiskDate)

		// meter bars
		table.WriteString(`<td>`)
		c := 0
		for _, disk := range item.DisksInfo {
			if c > 0 {
				table.WriteString("<br>")
			}
			c++
			fill := strconv.Itoa(int(disk.Fill))
			freeGB := float64(disk.Free) / 1024.0
			totalGB := float64(disk.Total) / 1024.0
			details := fmt.Sprintf("%.0f GB free of %.0f GB", freeGB, totalGB)
			fmt.Fprintf(&table, `<span title="%s">%s %s%% <meter value="%s" min="0" max="100" low="70" high="90" optimum="0"></meter></span>`, details, disk.Drive, fill, fill)
		}
		table.WriteString(`</td></tr>`)
	}

	table.WriteString("</tbody></table>")
	return table.String()

}
