package droplists

import (
	"fmt"
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/icons"
)

func BuildDropList(field, selected, parentCode string, withBlank, readOnly bool) string {
	var options []DroplistOption
	switch field {
	case "SOFTWARE":
		options = GetSoftwareList(withBlank)
	case "USER", "USERSEARCH", "USERINFORM": // Profile table search, "GROUPINFORM"
		options = GetUserList(field, selected, parentCode, withBlank)
	case "TYPE", "TYPESEARCH": // parent = wizard, so TYPE needs special processing
		options = GetTypeOptions(field, selected, parentCode, withBlank)
	case "KIND":
		options = GetKindOptions(field, selected, parentCode, withBlank)
	case "MID":
		options = GetMidOptions(field, selected, parentCode, withBlank)
	case "WIZARDS":
		options = GetWizardOptions(field, selected, withBlank)
	case "SITE", "SITESEARCH":
		options = GetSiteOptions(field, selected, withBlank)
	default:
		options = GetOptions(field, selected, parentCode, withBlank)
	}
	return buildCtrl(field, readOnly, options)
}

func buildCtrl(field string, readOnly bool, options []DroplistOption) string {
	droplist := GetDroplistInfo(field)
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
		return BuildDropdown(droplist, options, readOnly, onchange, required)
	}
	return buildSelect(droplist, options, readOnly, onchange, required)
}

// Dropdown with icons with colors
func BuildDropdown(droplist Droplist, options []DroplistOption, readOnly bool, onchange, required string) string {
	var ctrl strings.Builder
	selected := ""
	for option := range options {
		if options[option].Selected {
			selected = options[option].Value
			break
		}
	}
	disabled := ""
	if readOnly {
		disabled = "disabled"
	}
	ariaRequired := "false"
	if required != "" {
		ariaRequired = "true"
	}

	ctrl.WriteString(`<div class="custom-select-container">`)
	// We use type="text" but visually hidden to support native validation and onchange handlers
	fmt.Fprintf(&ctrl, `<input type="text" class="droplist-input" id="%s" name="%s" value="%s" %s %s tabindex="-1" aria-hidden="true" />`, droplist.Id, droplist.Id, selected, onchange, required)
	ctrl.WriteString(`<details role="list" class="dropdown">`)
	fmt.Fprintf(&ctrl, `<summary aria-haspopup="listbox" aria-invalid="false" aria-disabled="%s" aria-required="%s" aria-describedby="%sErr" >%s</summary>`, disabled, ariaRequired, droplist.Id, droplist.Title)
	ctrl.WriteString(`<ul role="listbox">`)
	for _, option := range options {
		icon := ""
		if len(option.Icon) > 0 && len(option.Colour) > 0 {
			icon = SetIconColor(option.Icon, option.Colour, option.Tip)
		} else if len(option.Icon) > 0 {
			icon = icons.GetIcon(option.Icon)
		}
		description := "&nbsp;"
		if len(option.Description) > 0 {
			description = option.Description
		}
		fmt.Fprintf(&ctrl, `<li><a href="#" data-value="%s">%s %s</a></li>`, option.Value, icon, description)
	}
	ctrl.WriteString(`</ul></details></div>`)
	return ctrl.String()
}

// Standard select list, no icons, no colors
func buildSelect(droplist Droplist, options []DroplistOption, readOnly bool, onchange, required string) string {
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

func SetIconColor(iconName string, colour string, tip string) string {
	class := xlateColor(colour)
	return fmt.Sprintf(`<span class="%s" data-tooltip="%s">%s</span>`, class, tip, icons.GetIcon(iconName))
}
