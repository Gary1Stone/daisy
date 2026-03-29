package util

import "strings"

type crud struct {
	Create bool
	Read   bool
	Update bool
	Delete bool
}

type Permissions struct {
	Software crud
	Profile  crud
	Device   crud
	Admin    crud
	Ticket   crud
}

// :SCRUD:DCRU:PCRUD:AR
// Split the permission string by ':'
func (p *Permissions) GetPermissions(permissionString string) {
	//Remove any leading colons
	permissionString = strings.Trim(permissionString, ":")
	//Split the string by ':'
	parts := strings.Split(permissionString, ":")
	for _, part := range parts {
		firstLetter := string(part[0])
		switch firstLetter {
		case "S":
			p.Software = setCRUD(part)
		case "D":
			p.Device = setCRUD(part)
		case "P":
			p.Profile = setCRUD(part)
		case "A":
			p.Admin = setCRUD(part)
		case "T":
			p.Ticket = setCRUD(part)
		}
	}
}

// Helper function to set permissions based on the actions in a string
func setCRUD(str string) crud {
	var perms crud
	if strings.Contains(str, "C") {
		perms.Create = true
	}
	if strings.Contains(str, "R") {
		perms.Read = true
	}
	if strings.Contains(str, "U") {
		perms.Update = true
	}
	if strings.Contains(str, "D") {
		perms.Delete = true
	}
	return perms
}
