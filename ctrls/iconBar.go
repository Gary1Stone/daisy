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
	BtnFilter = "btnFilter"
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
		BtnFilter: {"btnFilter", "Filter...", "", "popFilters();", svg.GetIcon("filter")},
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
