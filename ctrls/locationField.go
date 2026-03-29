package ctrls

import (
	"log"

	"github.com/gbsto/daisy/db"
)

func LocationCtrl(curUid, cid int) string {
	var location string
	if cid > 0 {
		if dev, err := db.GetDevice(curUid, cid); err == nil {
			location = dev.Location
		} else {
			log.Println(err)
		}
	}
	return `<input type="text" id="location" name="location" maxlength="100" value="` + location + `" 
        title="Location" data-role="input" data-validate="maxlength=100" >
        <span id="locationError" class="invalid_feedback">Location details are needed.</span>`
}
