package ctrls

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
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
	fmt.Fprintf(&table, "(Last Scan: %s", items[0].ScanDate)
	table.WriteString(`)
		<div style="max-height: 400px; overflow-y: auto;">
		<table class='striped' id="swlist" >
		<thead><tr>
		<th aria-sort="ascending" data-sort="asc">Software</th>
		<th aria-sort="none">OEM License</th>
		</tr></thead>
		<tbody>`)

	for _, item := range items {
		tracked := ""
		if item.IsTracked {
			tracked = svg.GetIcon("steps")
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
	ctrl.WriteString(`<select name='year' id='year' data-tooltip='Year of manufacture'><option value=''></option>`)
	for _, item := range items {
		selected := ""
		if item.Selected {
			selected = "selected"
		}
		fmt.Fprintf(&ctrl, `<option value="%s" %s>%s</option>`, item.Value, selected, item.Description)
	}
	ctrl.WriteString(`</select>`)
	return ctrl.String()
}

func BuildDeviceLog(curUid, cid, hist int) string {
	var table strings.Builder
	table.WriteString(`<table class='striped' id="actionlog">
		<thead><tr>
		<th aria-sort="none">Action</th>
		<th aria-sort="ascending" data-sort="asc">Date</th>
		<th aria-sort="none">Created By</th>
		<th aria-sort="none">Notes</th>
		<th aria-sort="none">Status</th>
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
		// if btnLabel has a space in it, truncate at the space.
		if strings.Contains(btnLabel, " ") {
			btnLabel = strings.Split(btnLabel, " ")[0]
		}
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
		fmt.Fprintf(&table, `<tr><td>%s</td><td>%s</td><td>%s</td><td><div id='notes%d'>%s</div></td><td>%s</td></tr>`, buildButton(button), act.Localtime, act.OriginatorName, act.Aid, act.Notes, calculateStatus(act))
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
	color := ""
	if button.Active == 0 {
		color = `class="secondary"`
	}
	fmt.Fprintf(&btn, `<a href="#" onclick="pop('%d');" role="button" %s >%s %s</a>`, button.Aid, color, svg.GetIcon(button.Icon), button.Label)
	fmt.Fprintf(&btn, `<input type='hidden' id='aid%d' value='%s' >`, button.Aid, string(btnJason))
	return btn.String()
}
