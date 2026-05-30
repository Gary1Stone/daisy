package ctrls

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

// Device cards for the devices.html screen
func DeviceCards(curUid int, filter *db.DeviceFilter, isWiz bool) string {
	var card strings.Builder
	items, err := db.SearchDevices(curUid, filter)
	if err != nil {
		log.Println(err)
	}
	for _, item := range items {
		backupIcon := ""
		assigned := ""
		brokenIcon := ""
		color := ""
		if item.IsMissing {
			color = "fg-red"
		}
		detail := "Never Seen"
		if item.Last_seen_days >= 0 {
			detail = "Last Seen: " + strconv.Itoa(item.Last_seen_days) + " days ago"
		}
		eyeIcon := fmt.Sprintf("<span class='%s' data-tooltip='Last Seen'>%s</span> %s", color, svg.GetIcon("eye"), detail)

		if item.Type == "Laptop" || item.Type == "Desktop" {
			color = ""
			if item.IsLate {
				color = "fg-red"
			}
			detail = "Not backed up"
			if item.Last_backup_days >= 0 {
				detail = "Backed up " + strconv.Itoa(item.Last_seen_days) + " days ago"
			}
			backupIcon = fmt.Sprintf("<span class='%s' data-tooltip='Backup'`>%s</span>%s", color, svg.GetIcon("copy"), detail)
		}

		// If assigend to a user or group
		if item.Uid > 0 {
			assigned = "<span data-tooltip='Assigned to person' >" + svg.GetIcon("user") + "</span> " + item.Assigned
		} else if item.Gid > 0 {
			assigned = "<span data-tooltip='Assigned to group'>" + svg.GetIcon("profiles") + "</span> " + item.Gid_usr
		}
		// Broken Icon
		if item.IsBroken {
			brokenIcon = "<span class='fg-red' data-tooltip='Broken'>" + svg.GetIcon("broken") + "</span>"
		}

		// Build the card contents - HEADER
		color = ""
		if item.IsMissing || item.IsLate || item.IsBroken {
			color = "fg-red"
		}
		fmt.Fprintf(&card, "<article><header><span class='%s' >%s</span> %s ", color, svg.GetIcon(item.Icon), mxl25(item.Name))

		// LHS of title - Alert Icon
		if len(item.Color) > 0 {
			fmt.Fprintf(&card, "<span style='color: red; float: right !important;' data-tooltip='Alert'>%s</span></header>", svg.GetIcon("bell"))
		} else {
			card.WriteString("&nbsp;</header>")
		}

		//PICTURE
		if isWiz {
			fmt.Fprintf(&card, "<a href='#' onclick='selectDevice('%d');' data-tooltip='Show Record'>", item.Cid)
		} else {
			fmt.Fprintf(&card, "<a href='device.html?cid=%d' data-tooltip='Show Record'>", item.Cid)
		}
		fmt.Fprintf(&card, "<img src='images/%s' alt='device photo' width='100%%'></a>", item.Small_image)

		//BODY
		fmt.Fprintf(&card, "<section><p>%s (%d)</p><p>%s</p><p>%s %s</p>", item.Make, item.Year, mxl25(item.Model), brokenIcon, item.Status)
		fmt.Fprintf(&card, "<p>%s %s</p><p>%s %s</p>", svg.GetIcon("site"), item.Site_usr, svg.GetIcon("office"), item.Office_usr)
		fmt.Fprintf(&card, "<p>%s</p><p>%s</p><p>%s</p></section>", assigned, eyeIcon, backupIcon)

		//FOOTER
		card.WriteString("<footer>")
		if isWiz {
			deviceJson, err := json.Marshal(item)
			if err != nil {
				log.Println(err)
			}
			fmt.Fprintf(&card, `<div style='display: none;' id='cid%d'>%s</div>`, item.Cid, string(deviceJson))
		} else {
			fmt.Fprintf(&card, `<button title='Create Report' class='secondary' onclick="popWizards('%d', '%s', '%s');">`, item.Cid, mxl25(item.Name), item.Type)
			fmt.Fprintf(&card, `%s&nbsp;Report&hellip; </button>`, svg.GetIcon("news"))
		}
		card.WriteString("</footer></article>")
	}
	return card.String()
}
