package ctrls

import "github.com/gbsto/daisy/svg"

// Create the command buttons - Save, New and Delete, depending on the user's permissions for this record
// MakeSaveButton creates the HTML for the save button and a hidden input field.
// The button is only rendered if 'update' is true.
// The hidden input 'canSave' reflects the value of 'update' (1 for true, 0 for false).
func MakeSaveButton(update bool) string {
	if update {
		return `<button id='btnSave' type='submit' aria-busy='false' form='theForm' data-tooltip='Save Record'>` + svg.GetIcon("save") + ` Save</button><input type='hidden' id='canSave' value='1'>`
	}
	return `<span id='btnSave'></span><input type='hidden' id='canSave' value='0'>`
}

// Add button
func MakeAddButton(create bool) string {
	if create {
		return `<button type='button' class='secondary' id='btnNew' data-tooltip='Create Record' onclick='addRecord();'>` + svg.GetIcon("add") + ` New</button><input type='hidden' id='canNew' value='1' >`
	}
	return `<span id='btnNew'></span><input type='hidden' id='canNew' value='0' >`
}

// Delete button
func MakeDeleteButton(delete bool) string {
	if delete {
		return `<button type='button' id='btnDelete' data-tooltip='Delete Record' class='outline secondary' data-target="delete-dialog" onclick='deleteRecord(event);'>` + svg.GetIcon("delete") + ` Delete</button><input type='hidden' id='canDelete' value='1' >`
	}
	return `<span id='btnDelete'></span><input type='hidden' id='canDelete' value='0' >`
}

// Delete button
func MakeCancelButton(read bool) string {
	if read {
		return `<buttontype='button' id='btnCancel' data-tooltip='cancel changes' class='secondary' onclick='window.location.href=window.location.href;'>` + svg.GetIcon("cancel") + ` Cancel</button><input type='hidden' id='canCancel' value='1' >`
	}
	return `<span id='btnCancel'></span><input type='hidden' id='canCancel' value='0' >`
}

// Seen and missing buttons are combined
func MakeSeeButton() string {
	ico := `<button type='button' id='btnSee' onclick='seeIconClick();' data-tooltip='Not seen in 90+ days' style='color:red;'>` + svg.GetIcon("eye") + `</button> `
	ico += `<button type='button' id='mif-copy' onclick='seeIconClick();' data-tooltip='Not backed up in 90+ days' style='color:red;'>` + svg.GetIcon("copy") + `</button>`
	ico += `<input type='hidden' id='btnSeeState' value='off' ><input type='hidden' id='islate' value='0' ><input type='hidden' id='ismissing' value='0' >`
	return ico
}

func MakeSearchBtn() string {
	return `<button type='button' class='contrast' data-tooltip='Search' onclick='popFilters();'>` + svg.GetIcon("search") + `</button>`
}

// ACRUD = Admin (Create, Read, Update, Delete)
func MakeAdminSelectButton(read bool) string {
	if read {
		return `<button type='button' id='btnTables' class='secondary' data-tooltip='Select Admin Table' onclick='showTableSelect();'>` + svg.GetIcon("site") + `</button>`
	}
	return `&nbsp;`
}

func MakeAdminSaveButton(update bool) string {
	if update {
		return `<button type='button' id='btnSave' data-tooltip='Save Record' onclick='save();'>` + svg.GetIcon("save") + `</button><input type='hidden' id='canSave' value='1' >`
	}
	return `<span id='btnSave'></span><input type='hidden' id='canSave' value='0' >`
}

func MakeAdminHelpButton() string {
	return `<button type='button' id='btnAdminHelp' class='secondary' data-tooltip='Help' onclick='showHelp();'>` + svg.GetIcon("help") + `</button>`
}
