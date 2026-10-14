package ctrls

import (
	"github.com/gbsto/daisy/devices"
)

func ImageLink(cid int) string {
	var d devices.Device
	return "images/" + d.GetImage(cid)
}

func SmallImageLink(cid int) (string, error) {
	var d devices.Device
	return "images/" + d.GetSmallImage(cid), nil
}
