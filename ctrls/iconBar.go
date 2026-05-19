package ctrls

import (
	"fmt"
	"sync"

	"github.com/gbsto/daisy/svg"
)

// Button identifiers as constants to prevent typos and improve maintainability.
const (
	BtnSave   = "btnSave"
	BtnNew    = "btnNew"
	BtnDelete = "btnDelete"
	BtnSeen   = "btnSeen"
	BtnBackup = "btnBackup"
	BtnSearch = "btnSearch"
	BtnHelp   = "btnHelp"
	BtnTables = "btnTables"
)

type buttonInfo struct {
	id       string
	tooltip  string
	style    string
	function string
	icon     string
}

var (
	btnInfoMap map[string]buttonInfo
	once       sync.Once
)

func loadBtnInfo() {
	btnInfoMap = map[string]buttonInfo{
		BtnSave:   {"btnSave", "Save Record", "", "saveRecord(event);", svg.GetIcon("save")},
		BtnNew:    {"btnNew", "Create Record", "", "addRecord(event);", svg.GetIcon("add")},
		BtnDelete: {"btnDelete", "Delete Record", "", "deleteRecord(event);", svg.GetIcon("delete")},
		BtnSeen:   {"btnSeen", "Not Seen in 90+ days", "style='color:red;'", "seenClick();", svg.GetIcon("eye")},
		BtnBackup: {"btnBackup", "Not Backed up in 90+ days", "style='color:red;'", "backupClick();", svg.GetIcon("copy")},
		BtnHelp:   {"btnHelp", "Help", "", "showHelp();", svg.GetIcon("help")},
		BtnTables: {"btnTables", "Select Admin Table", "", "showTableSelect();", svg.GetIcon("factory")},
	}
}

// MakeButton creates a command button depending on the user's permissions.
// The button is only rendered if 'permission' is true.
func MakeButton(name string, permission bool) string {
	once.Do(loadBtnInfo)

	btn, ok := btnInfoMap[name]
	if !ok {
		return ""
	}

	if !permission {
		return fmt.Sprintf("<span id='%s' data-allowed='0'></span>", btn.id)
	}

	return fmt.Sprintf(`<button type='button' id='%s' class='outline secondary' aria-label='%s' data-tooltip='%s' data-allowed='1' %s onclick='%s'>%s</button>`, btn.id, btn.tooltip, btn.tooltip, btn.style, btn.function, btn.icon)
}

// Specialized helper functions for common button types that include hidden inputs
// used by the application's legacy JavaScript for state and permission checks.

func MakeSaveButton(update bool) string {
	val := "0"
	if update {
		val = "1"
	}
	return MakeButton(BtnSave, update) + fmt.Sprintf(`<input type='hidden' id='canSave' value='%s'>`, val)
}

func MakeAddButton(create bool) string {
	val := "0"
	if create {
		val = "1"
	}
	return MakeButton(BtnNew, create) + fmt.Sprintf(`<input type='hidden' id='canNew' value='%s'>`, val)
}

func MakeDeleteButton(canDelete bool) string {
	if canDelete {
		once.Do(loadBtnInfo)
		btn := btnInfoMap[BtnDelete]
		return fmt.Sprintf(`<button type='button' id='%s' class='outline secondary' aria-label='%s' data-tooltip='%s' data-target="delete-dialog" data-allowed='1' onclick='%s'>%s</button><input type='hidden' id='canDelete' value='1'>`,
			btn.id, btn.tooltip, btn.tooltip, btn.function, btn.icon)
	}
	return MakeButton(BtnDelete, false) + `<input type='hidden' id='canDelete' value='0'>`
}

func MakeSeeButton() string {
	once.Do(loadBtnInfo)
	ico := MakeButton(BtnSeen, true) + " " + MakeButton(BtnBackup, true)
	ico += `<input type='hidden' id='btnSeeState' value='off'><input type='hidden' id='islate' value='0'><input type='hidden' id='ismissing' value='0'>`
	return ico
}

func MakeSearchBtn() string {
	return MakeButton(BtnSearch, true)
}

func MakeAdminSelectButton(read bool) string {
	if read {
		return MakeButton(BtnTables, true)
	}
	return `&nbsp;`
}

func MakeAdminSaveButton(update bool) string {
	return MakeSaveButton(update)
}

func MakeAdminHelpButton() string {
	return MakeButton(BtnHelp, true)
}
