package ctrls

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/devices"
	"github.com/gbsto/daisy/icons"
	"github.com/gbsto/daisy/logins"
	"github.com/gbsto/daisy/profiles"
	"github.com/gbsto/daisy/util"
)

// Generate profile table for wide screens
func ProfilesTable(curUid int, filter profiles.Profile) string {
	var table strings.Builder

	// Build the table header
	table.WriteString(buildProfileTableHeader())

	// Fetch profile items
	items, err := filter.GetProfiles(curUid)
	if err != nil {
		log.Println(err)
		table.WriteString("</tbody></table>")
		return table.String()
	}

	// Build table rows
	for _, item := range items {
		alert := "" // Add alerts if present
		if item.Alerts > 0 {
			alert = buildAlertIcon(item.Alerts)
		}
		ticket := "" // Add tickets if present
		if item.Tickets > 0 {
			ticket = buildTicketIcon(item.Tickets)
		}
		fmt.Fprintf(&table, "<tr data-id='%d'><td><a href='profile.html?uid=%d'>%s</a></td><td>%s</td><td>%s</td><td> %s %s</td></tr>", item.Uid, item.Uid, item.User, util.Mx25(item.Fullname), util.Mx25(item.Group), alert, ticket)
	}

	table.WriteString("</tbody></table>")
	return table.String()
}

/*
   ⏶ &#9206; &#x23f6; Alphabetic sort
   ⏷ &#9207; &#x23f7; Reverse Alphabetic Sort
	⏴	⏵
*/
// Helper function to build the table header
func buildProfileTableHeader() string {
	return `<table class='striped' id="profileTable">
    <thead>
    <tr>
        <th aria-sort="ascending" data-sort="asc">User ID</th>
        <th aria-sort="none">Name</th>
        <th aria-sort="none">Group</th>
        <th aria-sort="none">Alerts&sol;Tickets</th>
    </tr>
    </thead>
    <tbody>`
}

// Helper function to build the alert icon
func buildAlertIcon(alerts int) string {
	return icons.GetIcon("bell") + " (" + strconv.Itoa(alerts) + ") "
}

// Helper function to build the ticket icon
func buildTicketIcon(tickets int) string {
	return icons.GetIcon("ticket") + " (" + strconv.Itoa(tickets) + ") "
}

/*
 *
 * Generate the login report
 *
 */
func GetProfileLogins(curUid, uid int) string {
	items, err := logins.GetLogins(curUid, uid)
	if err != nil {
		log.Println(err)
		return ""
	}
	var tbl strings.Builder
	tbl.WriteString(`<div style="max-height: 400px; overflow-y: auto;">
		<table class='striped' id='dash_table'>
		<thead>
		<tr>
		<th aria-sort="ascending" data-sort="asc">Date/Time</th>
		<th aria-sort="none">Days Ago</th>
		<th aria-sort="none">Location</th>
		<th aria-sort="none">Distance (Km)</th>
		</tr>
		</thead>
		<tbody>`)
	for _, item := range items {
		fmt.Fprintf(&tbl, `<tr><td>%s</td><td>%d</td><td>%s, %s, %s<p>%s</p></td><td>%d</td></tr>`, item.Last_login, item.Days, item.Country, item.State, item.City, item.Community, item.Distance)
	}
	tbl.WriteString("</tbody></table></div>")
	return tbl.String()
}

// Create list of generic radius from center point
func BuildRadiusOptions(radius int) string {
	var ctrl strings.Builder
	ctrl.WriteString("<option value='' ></option>")
	rads := []int{1, 5, 10, 25, 50, 100}
	for _, rad := range rads {
		selected := ""
		if rad == radius {
			selected = "selected"
		}
		fmt.Fprintf(&ctrl, `<option value='%d' %s>%d Km</option>`, rad, selected, rad)
	}
	return ctrl.String()
}

// On the profile page, list all the assigned computers
func BuildAssignedDevices(curUid, tzoff, uid int) string {
	var d devices.Device
	devices, err := d.GetAssignedDevices(curUid, tzoff, uid)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(devices) == 0 {
		return ""
	}
	var ctrl strings.Builder
	ctrl.WriteString(`<div class="scrollarea">`)
	for _, item := range devices {
		fmt.Fprintf(&ctrl, `<p style="text-decoration: none;"><a href='device.html?cid=%d'>%s %s %s</a></p>`, item.Cid, icons.GetIcon(item.Icon), item.Name, item.Model)
	}
	ctrl.WriteString(`</div>`)

	return ctrl.String()
}

// PCRUD: - Profile (Read), (Create), (U)pdate, (D)elete
func BuildActiveCheckbox(isActive int, isDisabled bool) string {
	var ctrl strings.Builder
	ctrl.WriteString("<input type='checkbox' id='active' name='active' ")
	if isDisabled {
		ctrl.WriteString("disabled ")
	}
	if isActive == 1 {
		ctrl.WriteString("checked ")
	}
	ctrl.WriteString(">")
	return ctrl.String()
}

func BuildNotifyCheckbox(isActive int, isDisabled bool) string {
	var ctrl strings.Builder
	ctrl.WriteString("<input type='checkbox' id='notify' name='notify' ")
	if isDisabled {
		ctrl.WriteString("disabled ")
	}
	if isActive == 1 {
		ctrl.WriteString("checked ")
	}
	ctrl.WriteString(">")
	return ctrl.String()
}
