package ctrls

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/db"
)

func MakeImageCtrl(dev *db.Device, isUpdatePerm bool) string {
	var ctrl strings.Builder
	if dev.Cid == 0 {
		ctrl.WriteString(`<img src="images/missing.jpg" alt="photo" width="338px" height="253px" id="displayedImage" >`)
	} else {
		if isUpdatePerm {
			ctrl.WriteString(`<a href="#" onclick="changePic();" >`)
		}
		ctrl.WriteString(`<img src="images/`)
		if len(dev.Image) > 0 {
			ctrl.WriteString(dev.Image)
		} else {
			ctrl.WriteString("missing.jpg")
		}
		ctrl.WriteString(`" +  alt="photo" width="338px" height="253px" id="displayedImage" >`)
		if isUpdatePerm {
			ctrl.WriteString("</a>")
		}
	}
	ctrl.WriteString(`<input type="hidden" name="image" id="image" class="" value="`)
	ctrl.WriteString(dev.Image)
	ctrl.WriteString(`" >`)
	return ctrl.String()
}

// On the Device page, list all the installed software
func BuildSoftwareList(curUid, cid int) string {
	items, err := db.GetInstalledSoftware(curUid, cid)
	if err != nil {
		log.Println(err)
		return ""
	}
	if len(items) == 0 {
		return ""
	}

	var table strings.Builder
	table.WriteString("<label>(Last Scan: ")
	table.WriteString(items[0].ScanDate)
	table.WriteString(`)</label>
		<div style="max-height: 400px; overflow-y: auto;">
		<table data-role="table" id="swlist" 
		data-rows="-1" 
		data-show-rows-steps="false" 
		data-show-table-info="false" 
		data-show-pagination="false" 
		data-horizontal-scroll="true" 
		class="table striped table-border row-border row-hover compact">
		<thead><tr>
		<th data-sortable="true">Software</th>
		<th data-sortable="true">OEM License</th>
		</tr></thead>
		<tbody>`)

	for _, item := range items {
		tracked := ""
		if item.IsTracked {
			tracked = `&nbsp;<span class="mif-steps icon"></span>`
		}
		table.WriteString("<tr><td>")
		table.WriteString("")
		if item.Sid > 0 {
			table.WriteString(`<a href="software.html?sid=`)
			table.WriteString(strconv.Itoa(item.Sid))
			table.WriteString(`"><span class="mif-apps icon"></span>&nbsp;`)
			table.WriteString(item.Name)
			table.WriteString(tracked)
			table.WriteString("</a>")
		} else {
			table.WriteString(`<span class="mif-apps icon"></span>&nbsp;`)
			table.WriteString(item.Name)
			table.WriteString(tracked)
		}
		table.WriteString("</td>")
		table.WriteString("<td>")
		if item.Sid > 0 { // Only for tracked software
			checkboxId := "chk" + strconv.Itoa(item.Id)
			table.WriteString(`<input type="checkbox" id="` + checkboxId + `" `)
			table.WriteString(`onclick="setPreInstalled(` + strconv.Itoa(item.Id) + `, this.checked)" `)
			if item.PreInstalled {
				table.WriteString("checked")
			}
			table.WriteString(` >`)
		}
		table.WriteString("</td></tr>")
	}
	table.WriteString("</tbody></table></div>")
	return table.String()
}

func BuildYearsSelect(selected int) string {
	year := time.Now().Year()
	var items []db.DroplistOption
	for i := year; i > year-25; i-- {
		var item db.DroplistOption
		item.Value = strconv.Itoa(i)
		item.Description = strconv.Itoa(i)
		if i == selected {
			item.Selected = true
		}
		items = append(items, item)
	}
	var ctrl strings.Builder
	ctrl.WriteString(`<label for='year'>Year</label><select name='year' id='year' title='Year of manufacture' 
		data-role='select' data-filter='false' ><option value=''>&nbsp;</option>`)
	for _, item := range items {
		ctrl.WriteString(`<option value="`)
		ctrl.WriteString(item.Value)
		ctrl.WriteString(`" `)
		if item.Selected {
			ctrl.WriteString(`selected `)
		}
		// item.Icon = svg.GetIcon("calendar_today")
		// //	item.Icon = svg.GetIcon(db.FindIconNameByName("YEARS"))
		// // item.Icon = svg.LookupIcon("YEARS")
		// ctrl.WriteString(`data-template="<span class='`)
		// ctrl.WriteString(item.Icon)
		// ctrl.WriteString(` icon'></span> $1 " >`)
		ctrl.WriteString(`>`)
		ctrl.WriteString(item.Description)
		ctrl.WriteString(`</option>`)
	}
	ctrl.WriteString(`</select><small id="yearError" class="invalid_feedback">Select year the device was manufactured</small>`)
	return ctrl.String()
}

func BuildDeviceLog(curUid, cid, hist int) string {
	var table strings.Builder
	table.WriteString(`<table data-role='table' id='install_table' 
		data-rows='50' 
		data-show-rows-steps='true' 
		data-horizontal-scroll='true' 
		class='table striped table-border row-border row-hover compact'>
		<thead><tr>
		<th data-sortable='true'>Action</th>
		<th data-sortable='true'>Date</th>
		<th data-sortable='true'>Created By</th>
		<th data-sortable='true'>Notes</th>
		<th data-sortable='true'>Status</th>
		</tr></thead>
		<tbody>`)
	filter := new(db.ActionFilter)
	filter.Active = -1    // Disable Active filter: -1=Don't care if opened or closed
	filter.Pending = hist // Show/don't show not acknowledged actions
	filter.Cid = cid
	filter.Uid = 0
	filter.Sid = 0
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
		button.Uid_ack = act.Uid_ack
		button.Iid_ack = act.Inform_ack
		button.Cid_ack = act.Cid_ack
		button.Sid_ack = act.Sid_ack
		button.Wlog = act.Wlog
		button.Icon = act.Icon
		table.WriteString("<tr><td>")
		table.WriteString(buildButton(button))
		table.WriteString("</td><td><div class='gwrap'>")
		table.WriteString(act.Localtime)
		table.WriteString("</div></td><td><div class='gwrap'>")
		table.WriteString(act.OriginatorName)
		table.WriteString("</div></td><td><div id='notes")
		table.WriteString(strconv.Itoa(act.Aid))
		table.WriteString("' class='gwrap'>")
		table.WriteString(act.Notes)
		table.WriteString("</div></td><td><div class='gwrap'>")
		table.WriteString(calculateStatus(act))
		table.WriteString("</div></td></tr>")
	}
	table.WriteString("</tbody></table>")
	return table.String()
}

// Button to pop up the modal dialog with JSON object of the paramenters to use when the button is pressed
func buildButton(button db.Popbutton) string {
	var btn strings.Builder
	btnJason, err := json.Marshal(button)
	if err != nil {
		log.Println(err)
	}
	btn.WriteString("<input type='hidden' id='aid")
	btn.WriteString(strconv.Itoa(button.Aid))
	btn.WriteString("' value='")
	btn.WriteString(string(btnJason))
	btn.WriteString("' >")
	btn.WriteString("<a class=\"button ")
	btn.WriteString(button.Color)
	btn.WriteString("\" href=\"#\" ")
	btn.WriteString("onclick=\"pop('")
	btn.WriteString(strconv.Itoa(button.Aid))
	btn.WriteString("');\" role=\"button\" >")
	btn.WriteString("<span class=\"")
	btn.WriteString(button.Icon)
	btn.WriteString(" icon\"></span>&nbsp;")
	btn.WriteString(button.Label)
	btn.WriteString("</a>")
	return btn.String()
}
