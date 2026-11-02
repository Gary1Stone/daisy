package icons

import (
	"fmt"
	"sync"
)

// Button identifiers as constants to prevent typos and improve maintainability.
const (
	BtnSave      = "btnSave"
	BtnNew       = "btnNew"
	BtnDelete    = "btnDelete"
	BtnSeen      = "btnSeen"
	BtnBackup    = "btnBackup"
	BtnSearch    = "btnSearch"
	BtnHelp      = "btnHelp"
	BtnTables    = "btnTables"
	BtnFilter    = "btnFilter"
	BtnOnline    = "btnOnline"
	BtnDuplicate = "btnDuplicate"
	BtnCorrelate = "btnCorrelate"
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
		BtnSave:      {"btnSave", "Save Record", "", "saveRecord(event);", GetIcon("save")},
		BtnNew:       {"btnNew", "Create Record", "", "addRecord(event);", GetIcon("add")},
		BtnDelete:    {"btnDelete", "Delete Record", "", "deleteRecord(event);", GetIcon("delete")},
		BtnSeen:      {"btnSeen", "Not Seen in 90+ days", "style='color:red;'", "seenClick();", GetIcon("eye")},
		BtnBackup:    {"btnBackup", "Not Backed up in 90+ days", "style='color:red;'", "backupClick();", GetIcon("copy")},
		BtnHelp:      {"btnHelp", "Help", "", "showHelp();", GetIcon("help")},
		BtnTables:    {"btnTables", "Select Admin Table", "", "showTableSelect();", GetIcon("factory")},
		BtnFilter:    {"btnFilter", "Filter...", "", "popFilters();", GetIcon("filter")},
		BtnOnline:    {"btnOnline", "Device Online History", "", `goto("online.html");`, GetIcon("history")},
		BtnDuplicate: {"btnDuplicate", "Find Duplicate devices (MACs)", "", `goto("duplicates.html");`, GetIcon("clone")},
		BtnCorrelate: {"btnCorrelate", "Associate MACs to Devices", "", `goto("correlation.html");`, GetIcon("equals")},
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
	return fmt.Sprintf(`<button type='button' id='%s' class='outline' aria-label='%s' data-tooltip='%s' data-allowed='1' %s onclick='%s'>%s</button>`, btn.id, btn.tooltip, btn.tooltip, btn.style, btn.function, btn.icon)
}
