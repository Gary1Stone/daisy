package ctrls

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

// Generate profile table for wide screens
func SoftwaresTable(curUid int, filter db.SoftwareFilter) string {
	var table strings.Builder

	// Build the table header with search
	table.WriteString(`<table class='striped' id="softwaretable">
    <thead>
    <tr>
        <th aria-sort='ascending' data-sort='asc'>Name</th>
        <th aria-sort='none'>Vendor</th>
        <th aria-sort='none'>Licenses&sol;Installed</th>
        <th aria-sort='none'>OEM Licenses</th>
    </tr>
    </thead>
    <tbody>`)

	// Fetch software items
	items, err := filter.GetSoftwares(curUid)
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	icon := svg.GetIcon("software")

	// Build table rows
	for _, item := range items {
		licenses := ""
		if item.Free > 0 {
			licenses = "Unlimited"
		} else {
			licenses = strconv.Itoa(item.Licenses)
		}
		fmt.Fprintf(&table, `<tr data-id='%d'><td>%s %s</td><td>%s</td><td>%s &sol; %d</td><td>%d</td></tr>`, item.Sid, icon, item.Name, mxl25(item.Source), licenses, item.Installed, item.Installed)
	}

	table.WriteString("</tbody></table>")
	return table.String()
}

// On the Device page, list all the installed software
func BuildInstalledList(curUid, sid int) string {
	items, err := db.GetInstalledComputers(curUid, sid)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(items) == 0 {
		return ""
	}
	var table strings.Builder
	table.WriteString(`<label>Installed On</label>
		<div style="max-height: 400px; overflow-y: auto;">
		<table class='striped' id='swlist' >
		<thead><tr><th aria-sort='ascending' data-sort='asc'>Device</th></tr></thead>
		<tbody>`)
	for _, item := range items {
		fmt.Fprintf(&table, "<tr><td><a href='device.html?cid=%d'>%s %s %s</a></td></tr>", item.Cid, svg.GetIcon(item.Icon), item.Name, item.Model)
	}
	table.WriteString("</tbody></table></div>")
	return table.String()
}

func BuildSoftwareLog(curUid, sid, hist int) string {
	var table strings.Builder
	table.WriteString(`<table class='striped' id='softwarelog'><thead><tr>
	<th aria-sort='ascending' data-sort='asc'>Action</th>
	<th aria-sort='none'>Date</th>
	<th aria-sort='none'>Computer</th>
	<th aria-sort='none'>Installer</th>
	<th aria-sort='none'>Notes</th>
	<th aria-sort='none'>Status</th>
	</tr></thead><tbody>`)
	//table body
	filter := new(db.ActionFilter)
	filter.Active = -1    // -1 = Dont care if opened or closed
	filter.Pending = hist // Show or not not acknowledged actions
	filter.Cid = 0
	filter.Uid = 0
	filter.Sid = sid
	actions, _ := filter.GetActions(curUid)
	for _, act := range actions {
		btnColor := colors.Light
		if act.Active == 1 {
			btnColor = act.Color
		}
		btnLabel := act.ActionDescription
		if len(btnLabel) > 15 {
			btnLabel = string(btnLabel[0:12]) + "..."
		}
		var button db.Popbutton
		button.Color = btnColor
		button.Action = act.Action
		button.Label = btnLabel
		button.Aid = act.Aid
		button.Active = act.Active
		button.Cid_ack = act.Cid_ack
		button.Iid_ack = act.Inform_ack
		button.Sid_ack = act.Sid_ack
		button.Uid_ack = act.Uid_ack
		button.Wlog = act.Wlog
		button.Icon = act.Icon
		fmt.Fprintf(&table, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td>", buildButton(button), act.Localtime, act.Devicename, act.OriginatorName)
		fmt.Fprintf(&table, "<td><div id='notes%d'>%s</div></td><td>%s</td></tr>", act.Aid, act.Notes, calculateStatus(act))
	}
	table.WriteString("</tbody></table>")
	return table.String()
}

func calculateStatus(act *db.Action) string {
	retVal := ""
	if act.Uid > 0 && act.Uid_ack == 0 {
		retVal += ": User"
	}
	if act.Cid > 0 && act.Cid_ack == 0 {
		retVal += ": Device"
	}
	if act.Sid > 0 && act.Sid_ack == 0 {
		retVal += ": Software"
	}
	if act.Inform > 0 && act.Inform_ack == 0 {
		retVal += ": Inform"
	}
	if len(retVal) > 0 {
		retVal = "Pending" + retVal
	} else if act.Active == 1 {
		retVal = "Open but completed"
	} else {
		retVal = "Closed"
	}
	return retVal
}
