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
	case "WIZARDS":
		options = db.GetWizardOptions(field, selected, withBlank)
	default:
		options = db.GetOptions(field, selected, parentCode, withBlank)
	}
	return buildCtrl(field, readOnly, options)
}

func buildCtrl(field string, readOnly bool, options []db.DroplistOption) string {
	droplist := db.GetDroplistInfo(field)
	isDropDown := false
	for option := range options {
		if len(options[option].Icon) > 0 || len(options[option].Colour) > 0 {
			isDropDown = true
			break
		}
	}
	onchange := ""
	if len(droplist.Action) > 0 {
		onchange = `onchange="` + droplist.Action + `"`
	}
	required := ""
	if field == "TYPE" {
		required = "required"
	}
	if isDropDown {
		return buildDropdown(droplist, options, readOnly, onchange, required)
	}
	return buildSelect(droplist, options, readOnly, onchange, required)
}

// Dropdown with icons and colors
func buildDropdown(droplist db.Droplist, options []db.DroplistOption, readOnly bool, onchange, required string) string {
	var ctrl strings.Builder
	selected := ""
	for option := range options {
		if options[option].Selected {
			selected = options[option].Value
			break
		}
	}
	disabled := "false"
	if readOnly {
		disabled = "true"
	}
	ariaRequired := ""
	if required != "" {
		ariaRequired = `aria-required="true"`
	}
	ctrl.WriteString(`<div class="custom-select-container">`)
	// Use type="text" but visually hidden to support native validation and onchange handlers
	fmt.Fprintf(&ctrl, `<input type="text" class="droplist-input" id="%s" name="%s" value="%s" %s %s tabindex="-1" aria-hidden="true" />`, droplist.Id, droplist.Id, selected, onchange, required)
	ctrl.WriteString(`<details role="list" class="dropdown">`)
	fmt.Fprintf(&ctrl, `<summary aria-haspopup="listbox" aria-invalid="false" aria-disabled="%s" %s aria-describedby="%sErr" >%s</summary>`, disabled, ariaRequired, droplist.Id, droplist.Title)
	ctrl.WriteString(`<ul role="listbox">`)
	for _, option := range options {
		icon := ""
		if len(option.Icon) > 0 {
			icon = svg.GetIcon(option.Icon)
		}
		description := "&nbsp;"
		if len(option.Description) > 0 {
			description = option.Description
		}
		fmt.Fprintf(&ctrl, `<li><a href="#" class="%s" data-value="%s">%s %s</a></li>`, xlateColor(option.Colour), option.Value, icon, description)
	}
	ctrl.WriteString(`</ul></details></div>`)
	return ctrl.String()
}

// Standard select list, no icons, no colors
func buildSelect(droplist db.Droplist, options []db.DroplistOption, readOnly bool, onchange, required string) string {
	var ctrl strings.Builder
	disabled := ""
	if readOnly {
		disabled = "disabled"
	}
	fmt.Fprintf(&ctrl, `<select id="%s" name="%s" data-tooltip="%s" %s %s %s aria-invalid="false" aria-describedby="%sErr" >`, droplist.Id, droplist.Id, droplist.Title, disabled, onchange, required, droplist.Id)
	for _, option := range options {
		selected := ""
		if option.Selected {
			selected = "selected"
		}
		fmt.Fprintf(&ctrl, `<option value="%s" %s>%s</option>`, option.Value, selected, option.Description)
	}
	ctrl.WriteString("</select>")
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
