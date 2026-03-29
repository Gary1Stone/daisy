package ctrls

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/db"
)

// Device cards for the devices.html screen
func DeviceCards(curUid int, filter *db.DeviceFilter, isWiz bool) string {
	var card strings.Builder
	items, err := db.SearchDevices(curUid, filter)
	if err != nil {
		log.Println(err)
	}
	for _, item := range items {
		color := colors.Primary
		eyeIcon := "<span class='mif-eye' title='Last seen (days ago)'></span>"
		backupIcon := ""
		assigned := ""
		brokenIcon := ""
		if item.IsMissing {
			eyeIcon = "<p><span class='mif-eye fg-red blinking' title='Last seen (days ago)'></span>"
			color = "alert"
		}
		if item.Last_seen_days >= 0 {
			eyeIcon += " " + strconv.Itoa(item.Last_seen_days) + " Days</p>"
		} else {
			eyeIcon += " never seeen</p>"
		}
		if item.Type == "Laptop" || item.Type == "Desktop" {
			backupIcon = "<p><span class='mif-copy' title='Last backup (days ago)'></span>"
			if item.IsLate {
				backupIcon = "<p><span class='mif-copy fg-red blinking' title='Last backup (days ago)'></span>"
				color = "alert"
			}
			backupIcon += " " + strconv.Itoa(item.Last_backup_days) + " Days</p>"
		}
		// If assigend to a user or group
		if item.Uid > 0 {
			assigned = "<p><span class='mif-user' title='Assigned to person' ></span> " + item.Assigned + "</p>"
		} else if item.Gid > 0 {
			assigned = "<span class='mif-users icon' title='Assigned to group'></span> " + item.Gid_usr + "</p>"
		}
		// Broken Icon
		if item.IsBroken {
			brokenIcon = "<span class='mif-heart-broken fg-red blinking' title='Broken'></span>&nbsp;"
		}

		// Build the card contents - HEADER
		card.WriteString("<div class='miniCard'><header>")
		card.WriteString("<span class='")
		card.WriteString(item.Icon)
		card.WriteString(" icon ")
		if color == "alert" {
			card.WriteString("fg-red")
		}
		card.WriteString("'></span>&nbsp;")
		card.WriteString(mxl25(item.Name))
		card.WriteString("&nbsp;")

		// LHS of title - Alert Icon
		if len(item.Color) > 0 {
			card.WriteString("<span class='mif-bell fg-red blinking float-right' title='Alert'></span>")
		} else {
			card.WriteString("&nbsp;")
		}
		card.WriteString("</header>")

		//PICTURE
		if isWiz {
			card.WriteString("<a href='#' onclick='selectDevice(")
			card.WriteString(strconv.Itoa(item.Cid))
			card.WriteString(");' title='Select Device'>")
			card.WriteString("<img src='images/")
			card.WriteString(item.Small_image)
			card.WriteString("' alt='device photo'>")
			card.WriteString("</a>")
		} else {
			card.WriteString("<a href='device.html?cid=")
			card.WriteString(strconv.Itoa(item.Cid))
			card.WriteString("' title='Show Record'>")
			card.WriteString("<img src='images/")
			card.WriteString(item.Small_image)
			card.WriteString("' alt='photo'>")
			card.WriteString("</a>")
		}

		//BODY
		card.WriteString("<section><p>")
		card.WriteString(item.Make)
		card.WriteString(" (")
		card.WriteString(strconv.Itoa(item.Year))
		card.WriteString(")</p><p>")
		card.WriteString(mxl25(item.Model))
		card.WriteString("</p><p>")
		card.WriteString(brokenIcon)
		card.WriteString(item.Status)
		card.WriteString("</p>")
		card.WriteString("<p>")
		card.WriteString(item.Site)
		card.WriteString(" ")
		card.WriteString(item.Office_usr)
		card.WriteString("</p>")
		card.WriteString(assigned)
		card.WriteString(eyeIcon)
		card.WriteString(backupIcon)
		card.WriteString("</section>")

		//FOOTER
		card.WriteString("<footer>")
		if isWiz {
			deviceJson, err := json.Marshal(item)
			if err != nil {
				log.Println(err)
			}
			card.WriteString("<div style='display: none;' id='cid")
			card.WriteString(strconv.Itoa(item.Cid))
			card.WriteString("'>")
			card.WriteString(string(deviceJson))
			card.WriteString("</div>")
		} else {
			card.WriteString("<button title='Create Report' class='button secondary' onclick=\"popWizards(")
			card.WriteString(strconv.Itoa(item.Cid))
			card.WriteString(", '")
			card.WriteString(mxl25(item.Name))
			card.WriteString("', '")
			card.WriteString(item.Type)
			card.WriteString("');\">")
			card.WriteString("<span class='mif-news icon'></span>&nbsp;Report&hellip; ")
			card.WriteString("</button>")
		}
		card.WriteString("</footer>")
		card.WriteString("</div>")
	}
	return card.String()
}
