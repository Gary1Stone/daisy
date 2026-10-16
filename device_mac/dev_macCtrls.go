package device_mac

import (
	"fmt"
	"strings"

	"github.com/gbsto/daisy/icons"
)

func BuildDeviceList(site string) string {
	items, err := GetDevices(site)
	if err != nil {
		return ""
	}
	var tbl strings.Builder
	tbl.WriteString(`<table id="listingTable" class="striped"><thead><tr>
	<th aria-sort="ascending" data-sort="asc">Name</th>
	<th aria-sort="none">Model</th>
	<th aria-sort="none">Kind</th>
	<th aria-sort="none">Office</th>
	</tr></thead><tbody>`)
	for _, item := range items {
		fmt.Fprintf(&tbl, `<tr data-id="%d"><td><a href="device_mac.html?cid=%d">%s %s</a></td><td>%s</td><td>%s</td><td>%s</td></tr>`, item.Cid, item.Cid, icons.GetIcon(item.Icon), item.Name, item.Model, item.KindName, item.OfficeName)
	}
	tbl.WriteString(`</tbody></table>`)
	return tbl.String()
}

// // Device 2 MAC
// func BuildD2MDeviceList(site string) string {

// 	items, err := device_mac.GetDevices(site)
// 	if err != nil {
// 		log.Println(err)
// 		return err.Error()
// 	}

// 	var ctrl strings.Builder
// 	ctrl.WriteString(`<details class="dropdown" id="devicelist"><summary id="selected">Select a device</summary><ul>`)
// 	for _, item := range items {
// 		tip := item.Kind
// 		color := colors.Alert
// 		if item.IsLinked {
// 			color = colors.Success
// 		}

// 		fmt.Fprintf(&ctrl, `<li><a href="#" onclick="cidSelected(%d)"><span id="cid%d">%s %s %s<span></a></li>`, item.Cid, item.Cid, setIconColor(item.Kind, color, tip), item.Name, item.Model)
// 	}
// 	ctrl.WriteString(`</ul></details>`)

// 	return ctrl.String()
// }
