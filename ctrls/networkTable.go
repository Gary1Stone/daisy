package ctrls

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
)

func BuildNetworkTable(tzoff int, site string) string {
	var tbl strings.Builder

	items, err := db.GetMacs(tzoff, site)
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	tbl.WriteString("<table class='striped' id='networktable'>")
	tbl.WriteString("<thead>")
	tbl.WriteString("<tr>")
	tbl.WriteString("<th>Mac</th><th>Name</th><th>Kind</th><th>IP</th><th>Office</th><th>Online</th>")
	tbl.WriteString("</tr>")
	tbl.WriteString("</thead>")
	tbl.WriteString("<tbody>")

	for _, item := range items {
		fmt.Fprintf(&tbl, `<tr data-id="%d"><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`, item.Mid, item.Mac, item.Name, item.Kind, item.Ip, item.Office, isOnline(item.Online))
	}

	tbl.WriteString("</tbody>")
	tbl.WriteString("<table>")

	return tbl.String()
}

func isOnline(online bool) string {
	if online {
		return "⚪"
	}
	return "🔴"
}
