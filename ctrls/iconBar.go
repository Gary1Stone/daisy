package ctrls

// Create the command buttons - Save, New and Delete, depending on the user's permissions for this record
// MakeSaveButton creates the HTML for the save button and a hidden input field.
// The button is only rendered if 'update' is true.
// The hidden input 'canSave' reflects the value of 'update' (1 for true, 0 for false).
func MakeSaveButton(update bool) string {
	if update {
		return `<button type='submit' id='btnSave' form='theForm' title='Save Record' style='border: none; outline: none; background: none; cursor: pointer;'><span class='mif-floppy-disk fg-white'></span></button><input type='hidden' id='canSave' value='1'>`
	}
	return `<span id='btnSave'></span><input type='hidden' id='canSave' value='0'>`
}

// Add button
func MakeAddButton(create bool) string {
	if create {
		return `<button id='btnNew' title='Create Record' onclick='addRecord();' style='border: none; outline: none; background: none; cursor: pointer;'><span class='mif-plus fg-white'></span></button><input type='hidden' id='canNew' value='1' >`
	}
	return `<span id='btnNew'></span><input type='hidden' id='canNew' value='0' >`
}

// Delete button- TODO: Only if the device has no outstanding alerts or tracked software
func MakeDeleteButton(delete bool) string {
	if delete {
		return `<button id='btnDelete' title='Delete Record' onclick='deleteRecord();' style='border: none; outline: none; background: none; cursor: pointer;' ><span class='mif-cross fg-white'></span></button><input type='hidden' id='canDelete' value='1' >`
	}
	return `<span id='btnDelete'></span><input type='hidden' id='canDelete' value='0' >`
}

// Seen and missing buttons
func MakeSeeButton() string {
	return `<button id='btnSee' onclick='seeIconClick();' style='border: none; outline: none; background: none; cursor: pointer;'><span id='mif-eye' class='mif-eye fg-white' title='Not seen in 90+ days' ></span><span id='mif-copy' class='mif-copy fg-red' title='Not backed up in 90+ days' style='display:none'></span></button><input type='hidden' id='btnSeeState' value='off' ><input type='hidden' id='islate' value='0' ><input type='hidden' id='ismissing' value='0' >`
}

func MakeSearchBtn() string {
	return `<button id='btnFilter' title='Filter' onclick='popFilters();'; document.body.scrollTop = document.documentElement.scrollTop = 0;' style='border: none; outline: none; background: none; cursor: pointer;'><span class='mif-search fg-white'></span></button>`
}

// ACRUD = Admin (Create, Read, Update, Delete)
func MakeAdminSelectButton(read bool) string {
	if read {
		return `<button id='btnTables' title='Select Admin Table' onclick='showTableSelect();' style='border: none; outline: none; background: none; cursor: pointer;'><span class='mif-map2 fg-white'></span></button>`
	}
	return `&nbsp;`
}

func MakeAdminSaveButton(update bool) string {
	if update {
		return `<button id='btnSave' title='Save Record' onclick='save();' style='border: none; outline: none; background: none; cursor: pointer;'><span class='mif-floppy-disk fg-white'></span></button><input type='hidden' id='canSave' value='1' >`
	}
	return `<span id='btnSave'></span><input type='hidden' id='canSave' value='0' >`
}

func MakeAdminHelpButton() string {
	return `<button id='btnAdminHelp' title='Help' onclick='showHelp();' style='border: none; outline: none; background: none; cursor: pointer;'><span class='mif-help fg-white'></span></button>`
}
