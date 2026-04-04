package reports

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gbsto/daisy/db"
)

func TrackedSoftware() string {
	var report strings.Builder
	report.WriteString(`<table data-role='table' id='software_table' data-rows='50' 
		data-show-rows-steps='true' data-show-search='true' 
		data-show-pagination='true' data-show-table-info='true' 
		data-horizontal-scroll='true'
		class='table striped table-border row-border row-hover compact' >
		<thead>
		<tr>
		<th data-sortable='true'>Software</th>
		<th data-sortable='true'>Licenses</th>
		<th data-sortable='true'>Active Computer Installs</th>
		<th data-sortable='true'>Decomissioned Computer Installs</th>
		<th data-sortable='true'>Manually Tracked</th>
		<th data-sortable='true'>Decomissioned Tracked</th>
		</tr>
		</thead>
		<tbody>`)

	items, err := db.GetTrackedSoftware()
	if err != nil {
		log.Println(err)
		report.WriteString(`</tbody></table>`)
		return report.String()
	}

	for _, item := range items {
		btn := fmt.Sprintf(`<a class='button' href='software.html?sid=%d' role='button' ><span class='mif-app icon'></span>&nbsp;%s</a>`, item.Sid, item.Name)
		fmt.Fprintf(&report, `<tr><td>%s</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr>`, btn, item.Licenses, item.CountActiveInstalls, item.CountDecomissionedInstalls, item.Manual, item.Manual_inactive)
	}
	fmt.Fprintf(&report, `</tbody></table>`)
	return report.String()
}

func OtherSoftware() string {
	var report strings.Builder
	report.WriteString(`<table data-role='table' id='other_table' data-rows='50'
		data-show-rows-steps='true' data-show-search='true'
		data-show-pagination='true' data-show-table-info='true'
		data-horizontal-scroll='true'
		class='table striped table-border row-border row-hover compact' >
		<thead>
		<tr>
		<th data-sortable='true'>Software</th>
		<th data-sortable='true'>Installs</th>
		</tr>
		</thead>
		<tbody>`)

	items, err := db.GetOtherSoftware()
	if err != nil {
		log.Println(err)
		report.WriteString(`</tbody></table>`)
		return report.String()
	}

	for _, item := range items {
		cnt := item.CountActiveInstalls + item.CountDecomissionedInstalls
		fmt.Fprintf(&report, `<tr><td>%s</td><td>%d</td></tr>`, item.Name, cnt)
	}
	fmt.Fprintf(&report, "</tbody></table>")
	return report.String()
}

func UsersAssignedDevices() string {
	var report strings.Builder
	report.WriteString(`<table data-role='table' id='assigned_table' data-rows='50'
		data-show-rows-steps='true' data-show-search='true'
		data-show-pagination='true' data-show-table-info='true'
		data-horizontal-scroll='true'
		class='table striped table-border row-border row-hover compact' >
		<thead>
		<tr>
		<th data-sortable='true'>User</th>
		<th data-sortable='true'>Name</th>
		<th data-sortable='true'>Device</th>
		</tr>
		</thead>
		<tbody>`)
	items, err := db.ListUsersDevices()
	if err != nil {
		log.Println(err)

		report.WriteString(`</tbody></table>`)
		return report.String()
	}
	for _, item := range items {
		fmt.Fprintf(&report, `<tr><td>%s</td><td>%s</td><td>%s</tr>`, item.User, item.Fullname, item.DeviceName)
	}
	fmt.Fprintf(&report, "</tbody></table>")
	return report.String()
}

func NetworkGaps() string {
	var report strings.Builder
	report.WriteString(`<table data-role='table' id='gaps_table' data-rows='50'
		data-show-rows-steps='true' data-show-search='true'
		data-show-pagination='true' data-show-table-info='true'
		data-horizontal-scroll='true'
		class='table striped table-border row-border row-hover compact' >
		<thead>
		<tr>
		<th data-sortable='true'>Computer</th>
		<th data-sortable='true'>Occurance</th>
		<th data-sortable='true'>Gap (Minutes)</th>
		</tr>
		</thead>
		<tbody>`)
	items, err := db.NetworkGaps()
	if err != nil {
		log.Println(err)
		report.WriteString(`</tbody></table>`)
		return report.String()
	}
	for _, item := range items {
		fmt.Fprintf(&report, `<tr><td>%s</td><td>%s</td><td>%d</td></tr>`, item.Hostname, item.Timestamp, item.Gap)
	}
	fmt.Fprintf(&report, "</tbody></table>")
	return report.String()
}

func LastSeenDevices(curUid int, devInfo map[int]db.DevicesMeta) string {
	var report strings.Builder
	report.WriteString(`<table data-role='table' id='last_table' data-rows='-1'
		data-show-rows-steps='false' data-show-search='false'
		data-show-pagination='false' data-show-table-info='false'
		data-horizontal-scroll='true'
		class='table striped table-border row-border row-hover compact' >
		<thead>
		<tr>
		<th data-sortable='true'>Device</th>
		<th data-sortable='true'>Last Seen</th>
		</tr>
		</thead>
		<tbody>`)

	seenLate, err := strconv.Atoi(os.Getenv("LAST_SEEN_LATE"))
	if err != nil {
		seenLate = 90
	}

	items, err := db.GetLastSeenDevices(curUid)
	if err != nil {
		log.Println(err)
		report.WriteString(`</tbody></table>`)
		return report.String()
	}
	for _, item := range items {
		colour := "green"
		if item.Days > seenLate {
			colour = "red"
		}
		// Column 1 - Device
		fmt.Fprintf(&report, `<tr><td><a href='device.html?cid=%d' title='%s %s' ><span class='%s icon'></span>&nbsp;%s</a></td>`, item.Cid, item.Make, item.Model, devInfo[item.Cid].Icon, item.Name)
		// Column 2 - Last seen date
		fmt.Fprintf(&report, `<td><span style='color:%s;' title='%d days ago'>%s</span></td></tr>`, colour, item.Days, item.LastSeen)
	}
	fmt.Fprintf(&report, "</tbody></table>")
	return report.String()
}

func Checkins(curUid int, devInfo map[int]db.DevicesMeta) string {
	var report strings.Builder
	report.WriteString(`<table data-role='table' id='last_table' data-rows='-1'
		data-show-rows-steps='false' data-show-search='false'
		data-show-pagination='false' data-show-table-info='false'
		data-horizontal-scroll='true'
		class='table striped table-border row-border row-hover compact' >
		<thead>
		<tr>
		<th data-sortable='true'>Device</th>
		<th data-sortable='true'>Date</th>
		<th data-sortable='true'>Community</th>
		</tr>
		</thead>
		<tbody>`)

	auditLate, err := strconv.Atoi(os.Getenv("AUDIT_LATE"))
	if err != nil {
		auditLate = 28
	}

	items, err := db.GetLastTracks(curUid)
	if err != nil {
		log.Println(err)
		report.WriteString(`</tbody></table>`)
		return report.String()
	}

	for _, item := range items {
		colour := "green"
		if item.Days > auditLate {
			colour = "red"
		}
		title := fmt.Sprintf("%s %s", devInfo[item.Cid].Make, devInfo[item.Cid].Model)
		// Column 1 - Device
		fmt.Fprintf(&report, `<tr><td><a href='device.html?cid=%d' title="%s"><span class="%s icon"></span>&nbsp;%s</a></td>`, item.Cid, title, devInfo[item.Cid].Icon, devInfo[item.Cid].Name)
		// Column 2 - Audit Date
		fmt.Fprintf(&report, `<td><span style='color:%s;'  title="%d days ago">%s</span></td>`, colour, item.Days, item.Checkin)
		// Column 3 - Community
		fmt.Fprintf(&report, `<td><a href='https://www.google.com/maps/search/?api=1&query=%f,%f' target='_blank' title="WARNING: Location accuracy is only 10Km">%s&nbsp;<span class='mif-map icon'></span>...</a></td>`, item.Latitude, item.Longitude, item.Community)
	}
	fmt.Fprintf(&report, `</tbody></table>`)
	return report.String()
}

func Backups(curUid int, devInfo map[int]db.DevicesMeta) string {
	var report strings.Builder
	report.WriteString(`<table data-role='table' id='last_table' data-rows='-1' 
		data-show-rows-steps='false' data-show-search='false' 
		data-show-pagination='false' data-show-table-info='false' 
		data-horizontal-scroll='true'
		class='table striped table-border row-border row-hover compact' >
		<thead>
		<tr>
		<th data-sortable='true'>Device</th>
		<th data-sortable='true'>File</th>
		<th data-sortable='true'>System</th>
		<th data-sortable='true'>Disk</th>
		</tr>
		</thead>
		<tbody>`)

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

	// Get the latest backup info
	items, err := db.GetLatestBackups(curUid)
	if err != nil {
		log.Println(err)
		report.WriteString("</tbody></table>")
		return report.String()
	}
	for _, item := range items {
		dev, ok := devInfo[item.Cid]
		if !ok {
			continue
		}
		colour := "green"
		if item.FileDays > fileLate {
			colour = "red"
		}
		colour2 := "green"
		if item.SystemDays > systemLate {
			colour2 = "red"
		}
		colour3 := "green"
		if item.DiskDays > diskLate {
			colour3 = "red"
		}
		// Column 1 - link to device record
		fmt.Fprintf(&report, `<tr><td><a href='device.html?cid=%d' title="%s %s" ><span class="%s icon"></span>&nbsp;%s</a></td>`, item.Cid, dev.Make, dev.Model, dev.Icon, item.Computer)
		// Column 2 Files backup date
		fmt.Fprintf(&report, `<td><span style='color:%s;' title="%d days ago">%s</span></td>`, colour, item.FileDays, item.FileDate)
		// Column 3 System backup date
		fmt.Fprintf(&report, `<td><span style='color:%s;' title="%d days ago">%s</span></td>`, colour2, item.SystemDays, item.SystemDate)
		// Column 3 Disk backup date
		fmt.Fprintf(&report, `<td><span style='color:%s;' title="%d days ago">%s</span></td></tr>`, colour3, item.DiskDays, item.DiskDate)
	}
	fmt.Fprintf(&report, `</tbody></table>`)
	return report.String()
}

func Drivespace(curUid int, devInfo map[int]db.DevicesMeta) string {
	var report strings.Builder
	report.WriteString(`<table data-role='table' id='drive_table' data-rows='-1'
		data-show-rows-steps='false' data-show-search='false' 
		data-show-pagination='false' data-show-table-info='false' 
		data-horizontal-scroll='true'
		class='table striped table-border row-border row-hover compact' >
		<thead>
		<tr>
		<th data-sortable='true'>Device</th>
		<th data-sortable='true'>Drive Space</th>
		<th data-sortable='true'>Date</th>
		</tr>
		</thead>
		<tbody>`)

	// Get the latest drive space info
	items, err := db.GetDiskInfo(curUid, -1)
	if err != nil {
		log.Println(err)
		report.WriteString("</tbody></table>")
		return report.String()
	}

	for _, item := range items {
		// Skip if devInfo for the cid does not exists in the MAP
		dev, ok := devInfo[item.Cid]
		if !ok {
			continue
		}
		// Calculate drive space
		fill := strconv.Itoa(int(item.Fill))
		freeGB := float64(item.Free) / 1024.0
		totalGB := float64(item.Total) / 1024.0
		details := fmt.Sprintf("%.0f GB free of %.0f GB", freeGB, totalGB)
		// Get days difference between timestamp and now
		now := time.Now().UTC()
		days := int(now.Sub(time.Unix(int64(item.Timestamp), 0)).Hours() / 24)
		colour := "green"
		if days > 7 {
			colour = "red"
		}
		// Column 1 - link to device record
		fmt.Fprintf(&report, `<tr><td><a href="device.html?cid=%d" title="%s %s" ><span class="%s icon"></span>&nbsp;%s </a></td>`, item.Cid, dev.Make, dev.Model, dev.Icon, dev.Name)
		// Column 2 - meter bar
		fmt.Fprintf(&report, `<td><span title="%s">%s <meter value="%s" min="0" max="100" low="70" high="90" optimum="0"></meter> %s%%</span></td>`, details, item.Drive, fill, fill)
		// Column 3 - date of last reading
		fmt.Fprintf(&report, `<td><span style="color:%s;" title="%d days ago">%s</span></td></tr>`, colour, days, item.Localtime)
	}
	fmt.Fprintf(&report, `</tbody></table>`)
	return report.String()
}
