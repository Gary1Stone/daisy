package ctrls

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/db"
)

// Generate profile table for wide screens
func SoftwaresTable(curUid int, filter db.SoftwareFilter) string {
	var table strings.Builder

	// Build the table header
	table.WriteString(buildSoftwareTableHeader())

	// Fetch software items
	items, err := filter.GetSoftwares(curUid)
	if err != nil {
		log.Println(err)
		table.WriteString("</tbody></table>")
		return table.String()
	}

	// Build table rows
	for _, item := range items {
		table.WriteString(buildSoftwareTableRow(&item))
	}

	table.WriteString("</tbody></table>")
	return table.String()
}

// Helper function to build the table header
func buildSoftwareTableHeader() string {
	return `<table data-role="table" id="softwaretable" 
    data-rows="50" data-show-rows-steps="false" 
    data-show-search="true" data-horizontal-scroll="true" 
    data-table-search-title="<span class='mif-search'></span>" 
    data-show-pagination="true" data-show-table-info="true" 
    class="table striped table-border row-border row-hover">
    <thead>
    <tr>
        <th data-sortable="true" data-show="false">SID</th>
        <th data-sortable="true">Name</th>
        <th data-sortable="true">Vendor</th>
        <th data-sortable="true">Licenses&sol;Installed</th>
        <th data-sortable="true">OEM Licenses</th>
    </tr>
    </thead>
    <tbody>`
}

// Helper function to build a single table row
func buildSoftwareTableRow(item *db.Software) string {
	var row strings.Builder
	row.WriteString("<tr class='row-hover'><td data-label='SID'>")
	row.WriteString(strconv.Itoa(item.Sid))
	row.WriteString("</td><td><span class='mif-apps icon'></span> ")
	row.WriteString(item.Name)
	row.WriteString("</td><td>")
	row.WriteString(mxl25(item.Source))
	row.WriteString("</td><td>")
	if item.Free > 0 {
		row.WriteString("Unlimited")
	} else {
		row.WriteString(strconv.Itoa(item.Licenses))
	}
	row.WriteString("&nbsp;&sol;&nbsp;")
	row.WriteString(strconv.Itoa(item.Installed))
	row.WriteString("</td><td>")
	row.WriteString(strconv.Itoa(item.Pre_installed))
	row.WriteString("</td></tr>")
	return row.String()
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
		<table data-role="table" id="swlist" 
		data-rows="-1" 
		data-show-rows-steps="false" 
		data-show-table-info="false" 
		data-show-pagination="false" 
		data-horizontal-scroll="true" 
		class="table striped table-border row-border row-hover compact">
		<thead><tr>
		<th data-sortable="true">Device</th>
		</tr></thead>
		<tbody>`)

	for _, item := range items {
		table.WriteString("<tr><td><a href='device.html?cid=")
		table.WriteString(strconv.Itoa(item.Cid))
		table.WriteString("'><span class='")
		table.WriteString(item.Icon)
		table.WriteString(" icon'></span>&nbsp;")
		table.WriteString(item.Name)
		table.WriteString("</a>&nbsp;")
		table.WriteString(item.Model)
		table.WriteString("</td></tr>")
	}
	table.WriteString("</tbody></table></div>")
	return table.String()
}

func BuildSoftwareLog(curUid, sid, hist int) string {
	var table strings.Builder
	table.WriteString("<table data-role='table' id='install_table' ")
	table.WriteString("data-rows='50' ")
	table.WriteString("data-show-rows-steps='true' ")
	table.WriteString("data-horizontal-scroll='true' ")
	table.WriteString("class='table striped table-border row-border row-hover compact'>\n")
	table.WriteString("<thead>\n<tr>\n")
	table.WriteString("<th data-sortable='true'>Action</th>\n")
	table.WriteString("<th data-sortable='true'>Date</th>\n")
	table.WriteString("<th data-sortable='true'>Computer</th>\n")
	table.WriteString("<th data-sortable='true'>Installer</th>\n")
	table.WriteString("<th data-sortable='true'>Notes</th>\n")
	table.WriteString("<th data-sortable='true'>Status</th>\n")
	table.WriteString("</tr>\n</thead>\n")
	table.WriteString("<tbody>\n")
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
		table.WriteString("<tr>\n")
		table.WriteString("<td>")
		table.WriteString(buildButton(button))
		table.WriteString("</td>")
		table.WriteString("<td><div class='gwrap'>")
		table.WriteString(act.Localtime)
		table.WriteString("</div></td>")
		table.WriteString("<td><div class='gwrap'>")
		table.WriteString(act.Devicename)
		table.WriteString("</div></td>")
		table.WriteString("<td><div class='gwrap'>")
		table.WriteString(act.OriginatorName)
		table.WriteString("</div></td>")
		table.WriteString("<td><div id='notes")
		table.WriteString(strconv.Itoa(act.Aid))
		table.WriteString("' class='gwrap'>")
		table.WriteString(act.Notes)
		table.WriteString("</div></td>")
		table.WriteString("<td><div class='gwrap'>")
		table.WriteString(calculateStatus(act))
		table.WriteString("</div></td>")
		table.WriteString("</tr>\n")
	}
	table.WriteString("</tbody>\n")
	table.WriteString("</table>\n")
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
