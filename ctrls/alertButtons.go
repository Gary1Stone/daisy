package ctrls

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/web/wizards"
)

func GetAlertTable(uid int) string {
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

	var table strings.Builder
	table.WriteString(buildAlertTableHeader())
	// table.WriteString("<table class='table striped hover'>")
	// table.WriteString("<thead><tr>")
	// table.WriteString("<th>Device</th>")
	// table.WriteString("<th>Action</th>")
	// table.WriteString("<th>Dismiss</th>")
	// table.WriteString("</tr>")
	// table.WriteString("</thead>")
	// table.WriteString("<tbody>")

	for _, item := range items {
		deviceName := item.DeviceName
		if len(deviceName) > 10 {
			deviceName = deviceName[:10]
		}
		table.WriteString("<tr>")
		table.WriteString("<td>")
		table.WriteString("<span class='")
		table.WriteString(item.DeviceIcon)
		table.WriteString(" icon'></span>&nbsp;<span class='caption'>")
		table.WriteString(deviceName)
		table.WriteString("</td>")
		table.WriteString("<td>")
		table.WriteString("<span class='")
		table.WriteString(item.ActionIcon)
		table.WriteString(" icon'></span>&nbsp;")
		table.WriteString(xlateAction(item.Action, item.Uid_ack))
		table.WriteString("</td>")
		table.WriteString("<td>")
		table.WriteString(`<button class="button primary" onclick="ackAlert('`)
		table.WriteString(strconv.Itoa(item.Alert.Aid))
		table.WriteString(`');">`)
		table.WriteString("Okay</button>")
		table.WriteString("</td>")
		table.WriteString("</tr>")
	}
	table.WriteString("</tbody>")
	table.WriteString("</table>")
	return table.String()
}

// Helper function to build the table header
func buildAlertTableHeader() string {
	return `<table data-role="table" id="alerttable" 
    data-rows="-1" data-show-rows-steps="false" 
    data-show-search="false" 
    data-table-search-title="<span class='mif-search'></span>" 
    data-show-pagination="false" 
    data-show-table-info="false" 
    data-horizontal-scroll="true" 
    class="table striped table-border row-border row-hover">
    <thead>
    <tr>
        <th>Device</th>
        <th>Action</th>
        <th>Dismiss</th>
    </tr>
    </thead>
    <tbody>`
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
