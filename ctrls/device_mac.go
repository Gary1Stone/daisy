package ctrls

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/db"
)

func BuildD2MDeviceList(site string) string {

	items, err := db.GetD2MDevices(site)
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	var ctrl strings.Builder
	ctrl.WriteString(`<details class="dropdown" id="devicelist"><summary id="selected">Select a device</summary><ul>`)
	for _, item := range items {
		tip := item.Kind
		color := colors.Alert
		if item.IsLinked {
			color = colors.Success
		}

		fmt.Fprintf(&ctrl, `<li><a href="#" onclick="cidSelected(%d)"><span id="cid%d">%s %s %s<span></a></li>`, item.Cid, item.Cid, setIconColor(item.Kind, color, tip), item.Name, item.Model)
	}
	ctrl.WriteString(`</ul></details>`)

	return ctrl.String()
}
