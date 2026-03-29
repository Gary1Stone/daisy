package ctrls

import (
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/db"
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
	droplist := db.GetDroplistInfo(field)
	droplist.ReadOnly = readOnly
	var ctrl strings.Builder
	ctrl.WriteString("<label for=\"")
	ctrl.WriteString(droplist.Id)
	ctrl.WriteString("\">")
	ctrl.WriteString(droplist.Label)
	ctrl.WriteString("</label>")
	ctrl.WriteString("<select name=\"")
	ctrl.WriteString(droplist.Name)
	ctrl.WriteString("\" id=\"")
	ctrl.WriteString(droplist.Id)
	ctrl.WriteString("\" title=\"")
	ctrl.WriteString(droplist.Title)
	ctrl.WriteString("\" data-role=\"select\" ")
	if droplist.ReadOnly {
		ctrl.WriteString("disabled ")
	}
	if len(options) > 30 {
		ctrl.WriteString("data-filter=\"true\" ")
	} else {
		ctrl.WriteString("data-filter=\"false\" ")
	}
	if len(droplist.Action) > 0 {
		ctrl.WriteString("onchange=\"")
		ctrl.WriteString(droplist.Action)
		ctrl.WriteString("\" ")
	}
	ctrl.WriteString(">")

	for _, option := range options {
		ctrl.WriteString("<option value=\"")
		ctrl.WriteString(option.Value)
		ctrl.WriteString("\" ")
		if option.Selected {
			ctrl.WriteString("selected ")
		}
		if len(option.Icon) > 0 {
			ctrl.WriteString("data-template=\"<span class='")
			ctrl.WriteString(option.Icon)
			ctrl.WriteString(" icon'></span> $1 \" ")
		}
		if len(option.Colour) > 0 {
			ctrl.WriteString("class=\"")
			ctrl.WriteString(xlateColor(option.Colour))
			ctrl.WriteString("\" ")
		}
		ctrl.WriteString(">")
		ctrl.WriteString(option.Description)
		ctrl.WriteString("</option>")
	}
	ctrl.WriteString("</select>")
	ctrl.WriteString("<small id=\"")
	ctrl.WriteString(droplist.Id)
	ctrl.WriteString("Error\" class=\"invalid_feedback\">")
	ctrl.WriteString(droplist.ErrMsg)
	ctrl.WriteString("</small>")
	return ctrl.String()
}

func xlateColor(colour string) string {
	fgColor := "" //foreground color
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
