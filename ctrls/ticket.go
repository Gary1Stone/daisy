package ctrls

import (
	"html"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
)

func BuildDeviceCtrl(dev db.Device) string {
	year := ""
	if dev.Cid != 0 {
		year = " (" + strconv.Itoa(dev.Year) + ")"
	} else {
		dev.Icon = "mif-display"
	}

	var builder strings.Builder
	builder.WriteString(`<input type="text" id="name" value="`)
	builder.WriteString(html.EscapeString(dev.Name))
	builder.WriteString(`" readonly title="Device reported" data-role="input" data-prepend="<span class='`)
	builder.WriteString(html.EscapeString(dev.Icon))
	builder.WriteString(`'></span>" ><p>`)
	builder.WriteString(html.EscapeString(dev.Make))
	builder.WriteString(year)
	builder.WriteString(`</p><p>`)
	builder.WriteString(html.EscapeString(dev.Model))
	builder.WriteString(`</p>`)
	return builder.String()
}

// permissions = TCRUD = Ticket: Create, Read, Update, Delete
func BuildRouteButton(canUpdate bool) string {
	if canUpdate {
		return `<button onclick="showRouteDialog();" id="btnSave" title="Save Record" style="border: none; outline: none; background: none; cursor: pointer;"><span class="mif-floppy-disk fg-white"></span></button><input type="hidden" id="canSave" value="1" >`
	}
	return `<span id="btnSave"></span><input type="hidden" id="canSave" value="0" >`
}

// filepath: c:\Users\gbsto\go\src\devices\cmd\getTicket.go
func BuildAckCheckbox(isChecked int, isReadonly bool, label, id string) string {
	// Ensure label and id are not empty and properly escaped
	if id == "" {
		id = "defaultCheckbox"
	}
	if label == "" {
		label = "Acknowledgment"
	}

	var ctrl strings.Builder
	ctrl.WriteString(`<input type="checkbox" id="`)
	ctrl.WriteString(html.EscapeString(id))
	ctrl.WriteString(`" name="`)
	ctrl.WriteString(html.EscapeString(id))
	ctrl.WriteString(`" value="ack" data-role="switch" data-caption="`)
	ctrl.WriteString(html.EscapeString(label))
	ctrl.WriteString(`"`)

	// Add checked attribute if isChecked is 1
	if isChecked == 1 {
		ctrl.WriteString(" checked")
	}

	// Add disabled attribute if isReadonly is true
	if isReadonly {
		ctrl.WriteString(" disabled")
	}

	ctrl.WriteString(" >")
	return ctrl.String()
}

func BuildWorklog(curUid, aid int) string {
	// Fetch worklog items from the database
	items, err := db.GetLogs(curUid, aid)
	if err != nil {
		log.Println(err)
		return "<p>Error loading worklog data.</p>"
	}

	// Start building the worklog table
	var table strings.Builder
	table.WriteString(`<div>
		<table data-role="table" id="worklog" 
		data-rows="-1" 
		data-show-rows-steps="false" 
		data-show-search="true" 
		data-table-search-title="<span class='mif-search'></span>" 
		data-show-pagination="false" 
		data-show-table-info="false" 
		data-horizontal-scroll="true" 
		class="table striped table-border row-border row-hover">
		<thead><tr>
		<th data-sortable="false">Steps Taken</th>
		</tr></thead>
		<tbody>`)

	// Populate table rows with worklog items
	for _, item := range items {
		table.WriteString("<tr><td><p>")
		table.WriteString(html.EscapeString(item.Cmd))
		table.WriteString("&nbsp;")
		table.WriteString(html.EscapeString(item.Timestamp))
		table.WriteString("</p><p>")
		table.WriteString(html.EscapeString(item.Fullname))
		table.WriteString("</p><p>")
		table.WriteString(html.EscapeString(item.Note))
		table.WriteString("</p></td></tr>")
	}

	// Close table tags
	table.WriteString("</tbody>")
	table.WriteString("</table>")
	table.WriteString("</div>")

	return table.String()
}

func BuildReportCtrl(report string, readonly bool) string {
	var ctrl strings.Builder

	// Start building the input field
	ctrl.WriteString(`<input type="text" id="report" name="report" value="`)
	ctrl.WriteString(html.EscapeString(report)) // Escape the report value to prevent XSS
	ctrl.WriteString(`" placeholder="Enter a descriptive comment" title="Trouble reported" required minlength="6" maxlength="100" `)

	// Add readonly attribute if the field is readonly or already has a value
	if readonly || len(report) > 0 {
		ctrl.WriteString(`readonly `)
	}

	// Add additional attributes for validation and styling
	ctrl.WriteString(`data-role="input" data-prepend="<span class='mif-news'></span>" 
					data-validate="required, minlength=6, maxlength=100" >
					<small class="invalid_feedback">Required, 6-100 characters.</small>`)

	return ctrl.String()
}

func BuildInformList(aid int) string {
	// Define the filter for fetching alerts
	filter := db.Alert{
		Aid:  aid, // Get alerts for this ticket
		Gid:  -1,  // Get any group
		Uid:  -1,  // Get any user
		Ack:  0,   // Not acknowledged
		Wait: 1,   // Waiting for closure (0 = not waiting, 1 = waiting, -1 = all)
	}

	// Fetch alerts from the database
	items, err := db.GetAlerts(filter)
	if err != nil {
		log.Println(err)
		return "<p class='fg-red'>Error loading inform list</p>"
	}

	// Start building the select element
	var ctrl strings.Builder
	ctrl.WriteString(`<select id="informs" size="5" style="width:100%; border:1px solid Gainsboro">`)

	// Populate the select options with alert items
	for _, item := range items {
		ctrl.WriteString(`<option value="`)
		ctrl.WriteString(strconv.Itoa(item.Alert.Uid)) // Escape the UID
		ctrl.WriteString(`">`)
		ctrl.WriteString(html.EscapeString(item.Fullname)) // Escape the Fullname to prevent XSS
		ctrl.WriteString(`</option>`)
	}

	// Close the select element
	ctrl.WriteString("</select>")
	return ctrl.String()
}
