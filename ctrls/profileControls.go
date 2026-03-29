package ctrls

import (
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
)

/*
 *
 * Generate the login report
 *
 */
func GetProfileLogins(curUid, uid int) string {
	items, err := db.GetLogins(curUid, uid)
	if err != nil {
		log.Println(err)
		return ""
	}
	var report strings.Builder
	report.WriteString(`<div style="max-height: 400px; overflow-y: auto;">
		<table data-role='table' id='dash_table' data-rows='50' 
		data-show-rows-steps='true' data-show-search='true' 
		data-show-pagination='true' data-show-table-info='true' 
		data-horizontal-scroll='true'
		class='table striped table-border row-border row-hover compact' >
		<thead>
		<tr>
		<th data-sortable='true'>Date/Time</th>
		<th data-sortable='true'>Days Ago</th>
		<th data-sortable='true'>Location</th>
		<th data-sortable='true'>Distance (Km)</th>
		</tr>
		</thead>
		<tbody>`)
	for _, item := range items {
		report.WriteString("<tr>")
		//Login Time
		report.WriteString("<td>")
		report.WriteString(item.Last_login)
		report.WriteString("</td><td>")
		//days Ago
		report.WriteString(strconv.Itoa(item.Days))
		report.WriteString("</td><td>")
		//location
		report.WriteString(item.Country)
		report.WriteString(", ")
		report.WriteString(item.State)
		report.WriteString(", ")
		report.WriteString(item.City)
		report.WriteString("<p>")
		report.WriteString(item.Community)
		report.WriteString("</p></td><td>")
		// distance
		report.WriteString(strconv.Itoa(item.Distance))
		report.WriteString("</td></tr>")
	}
	report.WriteString("</tbody></table></div>")
	return report.String()
}

// Create list of generic radius from center point
func BuildRadiusOptions(radius int) string {
	var ctrl strings.Builder
	ctrl.WriteString("<option value='' ></option>")
	rads := []int{1, 5, 10, 25, 50, 100}
	for _, rad := range rads {
		str := strconv.Itoa(rad)
		ctrl.WriteString("<option value='")
		ctrl.WriteString(str)
		ctrl.WriteString("' ")
		if rad == radius {
			ctrl.WriteString("selected")
		}
		ctrl.WriteString(">")
		ctrl.WriteString(str)
		ctrl.WriteString(" Km</option>")
	}
	return ctrl.String()
}

// On the profile page, list all the assigned computers
func BuildAssignedDevices(curUid, uid int) string {
	devices, err := db.GetAssignedDevices(curUid, uid)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(devices) == 0 {
		return ""
	}
	var table strings.Builder
	table.WriteString("<ul data-role='listview' id='devlist'>")
	for _, item := range devices {
		table.WriteString("<a href='device.html?cid=")
		table.WriteString(strconv.Itoa(item.Cid))
		table.WriteString("'>")
		table.WriteString("<li data-icon=\"<span class='")
		table.WriteString(item.Icon)
		table.WriteString("'>\" data-caption='")
		table.WriteString(item.Name + " " + item.Model)
		table.WriteString("'></li>")
		table.WriteString("</a>")
	}
	table.WriteString("</ul>")
	return table.String()
}

// PCRUD: - Profile (Read), (Create), (U)pdate, (D)elete
func BuildActiveCheckbox(isActive int, isDisabled bool) string {
	var ctrl strings.Builder
	ctrl.WriteString("<input type='checkbox' id='active' name='active' data-role='checkbox' data-caption='Active' data-caption-position='left' ")
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
	ctrl.WriteString("<input type='checkbox' id='notify' name='notify' data-role='checkbox' data-caption='Notify' data-caption-position='left' ")
	if isDisabled {
		ctrl.WriteString("disabled ")
	}
	if isActive == 1 {
		ctrl.WriteString("checked ")
	}
	ctrl.WriteString(">")
	return ctrl.String()
}
