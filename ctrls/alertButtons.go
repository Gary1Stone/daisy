package ctrls

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
	"github.com/gbsto/daisy/web/wizards"
)

func GetAlertTable(uid int) string {
	var table strings.Builder
	filter := db.Alert{
		Aid:  -1,  // Get Any ticket
		Gid:  -1,  // Get Any group
		Uid:  uid, // Just this user
		Ack:  0,   // Not Acknowledged
		Wait: 0,   // Not waiting
	}
	// Fetch alerts and handle errors
	items, err := db.GetAlerts(filter)
	if err != nil {
		log.Println(err)
		return ""
	}

	table.WriteString(`<table id="alerttable"><thead><tr><th>Device</th><th>Action</th><th>Dismiss</th></tr></thead><tbody>`)

	for _, item := range items {
		deviceName := item.DeviceName
		if len(deviceName) > 10 {
			deviceName = deviceName[:10]
		}
		fmt.Fprintf(&table, `<tr><td>%s %s</td>`, svg.GetIcon(item.DeviceIcon), deviceName)
		fmt.Fprintf(&table, `<td>%s %s</td>`, svg.GetIcon(item.ActionIcon), xlateAction(item.Action, item.Uid_ack))
		fmt.Fprintf(&table, `<td><button onclick="ackAlert('%d');">Okay</button></td></tr>`, item.Alert.Aid)
	}
	table.WriteString("</tbody></table>")
	return table.String()
}

func GetAlertButtons(uid int) string {
	filter := db.Alert{
		Aid:  -1,  // Get Any ticket
		Gid:  -1,  // Get Any group
		Uid:  uid, // Just this user
		Ack:  0,   // Not Acknowledged
		Wait: 0,   // Not waiting
	}

	// Fetch alerts and handle errors
	items, err := db.GetAlerts(filter)
	if err != nil {
		log.Println(err)
		return ""
	}

	var card strings.Builder
	card.WriteString("<div class='row'>")

	// Build alert buttons
	for i, item := range items {
		if i%12 == 0 && i != 0 {
			card.WriteString("</div><div class='row'>")
		}
		card.WriteString(buildAlertButton(item))
	}

	card.WriteString("</div>")
	return card.String()
}

// Helper function to build a single alert button
func buildAlertButton(item *db.AlertDetails) string {
	var button strings.Builder

	// Truncate device name if necessary
	deviceName := item.DeviceName
	if len(deviceName) > 10 {
		deviceName = deviceName[:10]
	}

	button.WriteString("<div class='cell'><button class='command-button bg-lightRed fg-white rounded m-3' style='width: 200px;' onclick='ackAlert(")
	button.WriteString(strconv.Itoa(item.Alert.Aid))
	button.WriteString(");'><span class='")
	button.WriteString(item.DeviceIcon)
	button.WriteString(" icon'></span>&nbsp;<span class='caption'>")
	button.WriteString(deviceName)
	button.WriteString("<small><span class='")
	button.WriteString(item.ActionIcon)
	button.WriteString("'></span>&nbsp;")
	button.WriteString(xlateAction(item.Action, item.Uid_ack))
	button.WriteString("</small></span></button></div>")

	return button.String()
}

func xlateAction(action string, uid_ack int) string {
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
