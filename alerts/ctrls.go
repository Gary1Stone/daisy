package alerts

import (
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/wizards"
)

// filter := Alert{
// 	Aid:     -1,  // Get Any ticket
// 	Gid:     -1,  // Get Any group
// 	Uid:     uid, // Just this user
// 	Ack:     0,   // Not Acknowledged
// 	Wait:    0,   // Not waiting
// 	Cid_ack: 0,   // don't set acks
// 	Sid_ack: 0,   // don't set acks
// }

func (a *Alert) GetAlertTable() string {
	// Fetch alerts and handle errors
	items, err := a.GetAlerts()
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	var table strings.Builder
	table.WriteString(`<table class='striped' id="alerttable"><thead><tr><th>Device</th><th>Action</th><th> </th></tr></thead><tbody>`)
	for _, item := range items {
		deviceName := item.DeviceName
		if len(deviceName) > 10 {
			deviceName = deviceName[:10]
		}
		fmt.Fprintf(&table, `<tr><td>%s %s</td>`, icons.GetIcon(item.DeviceIcon), deviceName)
		fmt.Fprintf(&table, `<td>%s %s</td>`, icons.GetIcon(item.ActionIcon), a.xlateAction(item.Action, item.Uid_ack))
		fmt.Fprintf(&table, `<td><button type="button" onclick="ackAlert('%d');">Dismiss</button></td></tr>`, item.Aid)
	}
	table.WriteString("</tbody></table>")
	return table.String()
}

func (a *Alert) xlateAction(action string, uid_ack int) string {
	msg := ""
	switch action {
	case wizards.Sighting: // "SIGHTING":
		msg = "Device was seen"
	case wizards.Broken: // "BROKEN":
		msg = "Device was repaired"
		if uid_ack == 0 {
			msg = "Device is broken"
		}
	case wizards.Backup: //UP":
		msg = "Backup was done"
	case wizards.Care: //"CARE":
		msg = "Device was set right"
		if uid_ack == 0 {
			msg = "Attention is needed"
		}
	case wizards.Giving: //"GIVING":
		msg = "Device given to someone"
	case wizards.Claiming: //"CLAIMING":
		msg = "Device claimed"
	case wizards.Lost: //"LOST":
		msg = "Device lost"
	case wizards.Died: //"DIED":
		msg = "Device's death was investigated"
		if uid_ack == 0 {
			msg = "Device reported as dead"
		}
	case wizards.Using: //"USING":
		msg = "Device being used"
	case wizards.Install: //"INSTALL":
		msg = "Software installation request completed"
		if uid_ack == 0 {
			msg = "Computer needs software"
		}
	case wizards.Remove: //"REMOVE":
		msg = "Software was removed"
		if uid_ack == 0 {
			msg = "Computer needs software removed"
		}
	case wizards.Request: //"REQUEST":
		msg = "Software was requested"
		if uid_ack == 0 {
			msg = "Computer needs software added"
		}
	}
	return msg
}

func (a *Alert) BuildInformList() string {
	if a.Aid < 1 {
		return "error: AID missing"
	}

	items, err := a.GetAlerts()
	if err != nil {
		log.Println(err)
		return "<p class='fg-red'>Error loading alert inform list</p>"
	}

	// Start building the select element
	var ctrl strings.Builder
	ctrl.WriteString(`<select id="informs" size="5" style="width:100%; border:1px solid Gainsboro">`)

	// Populate the select options with alert items
	for _, item := range items {
		ctrl.WriteString(`<option value="`)
		ctrl.WriteString(strconv.Itoa(item.Uid)) // Escape the UID
		ctrl.WriteString(`">`)
		ctrl.WriteString(html.EscapeString(item.Fullname)) // Escape the Fullname to prevent XSS
		ctrl.WriteString(`</option>`)
	}

	// Close the select element
	ctrl.WriteString("</select>")
	return ctrl.String()
}
