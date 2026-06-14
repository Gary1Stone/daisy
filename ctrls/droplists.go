package ctrls

import (
	"fmt"
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

func BuildDropList(field, selected, parentCode string, withBlank, readOnly bool) string {
	var options []db.DroplistOption
	switch field {
	case "SOFTWARE":
		options = db.GetSoftwareList(withBlank)
	case "USER", "USERSEARCH", "USERINFORM": // Profile table search, "GROUPINFORM"
		options = db.GetUserList(field, selected, parentCode, withBlank)
	case "TYPE", "TYPESEARCH": // parent = wizard, so TYPE needs special processing
		options = db.GetTypeOptions(field, selected, parentCode, withBlank)
	case "KIND":
		options = db.GetKindOptions(field, selected, parentCode, withBlank)
	case "MID":
		options = db.GetMidOptions(field, selected, parentCode, withBlank)
	default:
		options = db.GetOptions(field, selected, parentCode, withBlank)
	}
	return buildSelectCtrl(field, readOnly, options)
}

func buildSelectCtrl(field string, readOnly bool, options []db.DroplistOption) string {
	var ctrl strings.Builder
	droplist := db.GetDroplistInfo(field)
	addIcons := false
	for option := range options {
		if len(options[option].Icon) > 0 {
			addIcons = true
			break
		}
	}
	disabled := ""
	if readOnly {
		disabled = "disabled"
	}
	onchange := ""
	if len(droplist.Action) > 0 {
		onchange = `onchange="` + droplist.Action + `"`
	}
	if addIcons {
		fmt.Fprintf(&ctrl, `<div class="custom-select">`)
	}
	required := ""
	if field == "TYPE" {
		required = "required"
	}
	fmt.Fprintf(&ctrl, `<select name="%s" id="%s" data-tooltip="%s" %s %s %s aria-invalid="false" aria-describedby="%sErr" >`, droplist.Name, droplist.Id, droplist.Title, disabled, onchange, required, droplist.Id)

	for _, option := range options {
		selected := ""
		if option.Selected {
			selected = "selected"
		}
		icon := ""
		if len(option.Icon) > 0 {
			icon = svg.GetIcon(option.Icon) + " "
		}
		color := ""
		if len(option.Colour) > 0 {
			color = fmt.Sprintf(`data-color="%s" `, xlateColor(option.Colour))
		}
		fmt.Fprintf(&ctrl, `<option value="%s" %s %s>%s%s</option>`, option.Value, selected, color, icon, option.Description)
	}
	ctrl.WriteString("</select>")
	if addIcons {
		fmt.Fprintf(&ctrl, `</div>`)
	}
	return ctrl.String()
}

func xlateColor(colour string) string {
	fgColor := colour //foreground color
	switch colour {
	case (colors.Success):
		fgColor = "fg-green"
	case (colors.Info):
		fgColor = "fg-cyan"
	case (colors.Alert):
		fgColor = "fg-red"
	case (colors.Primary):
		fgColor = "fg-blue"
	case (colors.Secondary):
		fgColor = "fg-gray"
	case (colors.Dark):
		fgColor = "fg-dark"
	case (colors.Light):
		fgColor = "fg-light"
	case (colors.Warning):
		fgColor = "fg-orange"
	case (colors.Yellow):
		fgColor = "fg-yellow"
	}
	return fgColor
}
