package ctrls

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/util"
	"github.com/gbsto/daisy/web/wizards"
)

// We want only BROKEN, DIED, LOST, CARE, REQUEST actions
// They are only the ones that require follow-up
// Generate tickets table for wide screens
func TicketsTable(curUid int) string {
	var table strings.Builder
	var filter db.ActionFilter
	filter.Page = 0

	// Build the table header
	table.WriteString(buildTicketsTableHeader())

	// Fetch profile and handle errors
	profile, err := db.GetProfile(curUid, curUid)
	if err != nil {
		log.Println(err)
		profile.Gid = 0
	}

	// Fetch actionable actions and handle errors
	items, err := db.GetAllActionableActions(curUid)
	if err != nil {
		log.Println(err)
		table.WriteString("</tbody></table>")
		return table.String()
	}

	// Build table rows
	for _, item := range items {
		table.WriteString(buildTicketsTableRow(curUid, item, &profile, &filter))
	}

	table.WriteString("</tbody></table>")
	return table.String()
}

// Helper function to build the table header
func buildTicketsTableHeader() string {
	return `<table data-role="table" id="tickettable" data-rows="50" 
    data-show-rows-steps="false" data-show-search="true" 
    data-table-search-title="<span class='mif-search'></span>" 
    data-show-pagination="true" data-show-table-info="true" 
    data-horizontal-scroll="true" 
    class="table striped table-border row-border row-hover">
    <thead>
    <tr>
        <th data-sortable="true" data-show="false">AID</th>
        <th data-sortable="true">Device</th>
        <th data-sortable="true">Queue</th>
        <th data-sortable="true">Assigned</th>
        <th data-sortable="true">Trouble</th>
        <th data-sortable="true">Report</th>
        <th data-sortable="true">Origin</th>
        <th data-sortable="true">Duration</th>
        <th data-sortable="true">ACK'd</th>
    </tr>
    </thead>
    <tbody>`
}

// Helper function to build a single table row
func buildTicketsTableRow(curUid int, item *db.Action, profile *db.Profile, filter *db.ActionFilter) string {
	var row strings.Builder

	row.WriteString("<tr class='row-hover'>")
	row.WriteString(buildTableCell("AID", strconv.Itoa(item.Aid)))
	row.WriteString(buildDeviceCell(item))
	row.WriteString(buildTableCell("QUEUE", item.AssignedGroupName))
	row.WriteString(buildTableCell("ASSIGNED", item.AssignedUserName))
	row.WriteString(buildTableCell("TROUBLE", mxl25(item.TroubleDescription)))
	row.WriteString(buildTableCell("REPORT", mxl50(item.Report)))
	row.WriteString(buildTableCell("ORIGIN", mxl25(item.OriginatorName)))
	row.WriteString(buildTableCell("DURATION", util.CalcDuration(item.OpenedInt, item.ClosedInt)))
	row.WriteString(buildAckCell(curUid, item, profile, filter))
	row.WriteString("</tr>")

	return row.String()
}

// Helper function to build a generic table cell
func buildTableCell(label, content string) string {
	return `<td data-label='` + label + `'>` + content + `</td>`
}

// Helper function to build the device cell
func buildDeviceCell(item *db.Action) string {
	var cell strings.Builder
	cell.WriteString("<td data-label='DEVICE'>")
	if len(item.Action) > 0 {
		cell.WriteString("<span class='")
		cell.WriteString(item.Color)
		cell.WriteString(" ")
		cell.WriteString(item.Icon)
		cell.WriteString(" icon'></span> ")
	}
	cell.WriteString("<span class='")
	cell.WriteString(item.DeviceIcon)
	cell.WriteString("'></span> ")
	cell.WriteString(mxl25(item.Devicename))
	cell.WriteString("</td>")
	return cell.String()
}

// Helper function to build the ACK cell
func buildAckCell(curUid int, item *db.Action, profile *db.Profile, filter *db.ActionFilter) string {
	var cell strings.Builder
	cell.WriteString("<td data-label='ACK'>")
	if isInformAck(curUid, item, profile, filter) || isAssignedAck(curUid, item, profile, filter) {
		cell.WriteString("<button id='cmd' class='button primary' onclick='acceptAction(\"")
		cell.WriteString(strconv.Itoa(item.Aid))
		cell.WriteString("\");'>Accept</button>")
	}
	cell.WriteString("</td>")
	return cell.String()
}

// Show the ack button to the inform person
// Check the inform flag to see if this user can acknowledge this alert
func isInformAck(curUid int, act *db.Action, profile *db.Profile, filter *db.ActionFilter) bool {
	// Already acknowledged, don't show the button
	if act.Inform_ack > 0 {
		return false
	}
	// if this user is the informed, show the button
	if act.Inform == curUid {
		return true
	}
	// if the inform is not assigned to a user, but to a group, and this user belongs to that same group
	if act.Inform == 0 && act.Inform_gid == profile.Gid {
		return true
	}
	// No filter set, therefore no button
	if filter.Uid == 0 {
		return false
	}
	// If the filtered user is the informed person, show the button
	if act.Inform == filter.Uid {
		return true
	}
	// if the inform is not assigned to a user, but to a group, and the filter is set to this group
	if act.Inform == 0 && act.Inform_gid == filter.Gid {
		return true
	}
	return false
}

// Show the ack button to the assigned person (UID) after they close the worklog (do something) if BROKEN,LOST,DIED,CARE,REQUEST
// Show the ack button to the assigned person (UID) if NOT:  BROKEN,LOST,DIED,CARE,REQUEST
// Show the ack button to the assigned person's (UID) GROUP after they close the worklog (do something) if  BROKEN,LOST,DIED,CARE,REQUEST
// Show the ack button to the assigned person's (UID) GROUP if NOT:  BROKEN,LOST,DIED,CARE,REQUEST
// Determine if the UID field for notification needs to be acknowledged by this person
func isAssignedAck(curUid int, act *db.Action, profile *db.Profile, filter *db.ActionFilter) bool {

	// Already acknowledged, don't show the button
	if act.Uid_ack > 0 {
		return false
	}

	// If the actions require a worklog entry
	switch act.Action {
	case wizards.Broken, wizards.Lost, wizards.Died, wizards.Care, wizards.Request:
		// Is the worklog closed?
		if act.Wlog == 0 {
			return false
		}
		// if this user is assigned a task (UID) and not ack'd and worklog is closed and it requires a worklog entry
		if act.Uid == curUid {
			return true
		}
		// Anyone in the same group can also close the action item, IF alert NOT ASSIGNED TO A USER
		if act.Gid == profile.Gid && act.Uid == 0 {
			return true
		}
		// Now handle the filtered user
		if filter.Uid == 0 {
			return false
		}
		// If this user is assigned a task (UID) and not ack'd and worklog is closed and it requires a worklog entry
		if act.Uid == filter.Uid {
			return true
		}
		// Anyone in the same group can also close the action item
		if act.Gid == filter.Gid && act.Uid == 0 {
			return true
		}
	default:
		// Show the ack button to the assigned person (UID) if action is not requiring a work log
		if act.Uid == curUid {
			return true
		}
		// Show the ack button to the assigned person (UID) if action is not requiring a work log
		if act.Gid == profile.Gid {
			return true
		}
		if filter.Uid == 0 {
			return false
		}
		// Show the ack button to the assigned person (UID) if action is not requiring a work log
		if act.Uid == filter.Uid {
			return true
		}
		// Show the ack button to the assigned group
		if act.Gid == filter.Gid && act.Uid == 0 {
			return true
		}
	}
	return false
}
