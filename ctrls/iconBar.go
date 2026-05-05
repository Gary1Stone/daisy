package ctrls

import "github.com/gbsto/daisy/svg"

// Create the command buttons - Save, New and Delete, depending on the user's permissions for this record
// MakeSaveButton creates the HTML for the save button and a hidden input field.
// The button is only rendered if 'update' is true.
// The hidden input 'canSave' reflects the value of 'update' (1 for true, 0 for false).
func MakeSaveButton(update bool) string {
	if update {
		return `<button id='btnSave' type='submit' class='outline secondary' form='theForm' title='Save Record'>` + svg.GetIcon("save") + `</button><input type='hidden' id='canSave' value='1'>`
	}
	return `<span id='btnSave'></span><input type='hidden' id='canSave' value='0'>`
}

// Add button
func MakeAddButton(create bool) string {
	if create {
		return `<button id='btnNew' class='outline secondary' title='Create Record' onclick='addRecord();'>` + svg.GetIcon("add") + `</button><input type='hidden' id='canNew' value='1' >`
	}
	return `<span id='btnNew'></span><input type='hidden' id='canNew' value='0' >`
}

// Delete button
func MakeDeleteButton(delete bool) string {
	if delete {
		return `<button id='btnDelete' title='Delete Record' class='outline secondary' onclick='deleteRecord();'>` + svg.GetIcon("delete") + `</button><input type='hidden' id='canDelete' value='1' >`
	}
	return `<span id='btnDelete'></span><input type='hidden' id='canDelete' value='0' >`
}

// Seen and missing buttons are combined
func MakeSeeButton() string {
	ico := `<button id='btnSee' class='outline secondary' onclick='seeIconClick();' title='Not seen in 90+ days' style='color:red; display:block;'>` + svg.GetIcon("eye") + `</button>`
	ico += `<button id='mif-copy' class='outline secondary' onclick='seeIconClick();' title='Not backed up in 90+ days' style='color:red; display:none;'>` + svg.GetIcon("copy") + `</button>`
	ico += `<input type='hidden' id='btnSeeState' value='off' ><input type='hidden' id='islate' value='0' ><input type='hidden' id='ismissing' value='0' >`
	return ico
}

func MakeSearchBtn() string {
	return `<button class='outline secondary' title='Search' onclick='onclick='popFilters();'>` + svg.GetIcon("search") + `</a>`
}

// ACRUD = Admin (Create, Read, Update, Delete)
func MakeAdminSelectButton(read bool) string {
	if read {
		return `<button id='btnTables' class='outline secondary' title='Select Admin Table' onclick='showTableSelect();'><span class='mif-map2 fg-white'></span></button>`
	}
	return `&nbsp;`
}

func MakeAdminSaveButton(update bool) string {
	if update {
		return `<button id='btnSave' title='Save Record' onclick='save();'>` + svg.GetIcon("save") + `</button><input type='hidden' id='canSave' value='1' >`
	}
	return `<span id='btnSave'></span><input type='hidden' id='canSave' value='0' >`
}

func MakeAdminHelpButton() string {
	return `<button id='btnAdminHelp' title='Help' onclick='showHelp();'>` + svg.GetIcon("help") + `</button>`
}
