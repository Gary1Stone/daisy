package ctrls

import (
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

func BuildMacTable(tzoff int, site string) string {
	var tbl strings.Builder

	items, err := db.GetMacs(tzoff, site)
	if err != nil {
		log.Println(err)
		return err.Error()
	}

	tbl.WriteString("<table class='striped' id='mactable'><thead><tr>")
	tbl.WriteString("<th>Name</th><th>Mac</th><th>IP</th><th>Office</th><th>Online</th>")
	tbl.WriteString("</tr></thead><tbody>")

	for _, item := range items {
		icon := svg.GetIcon(item.Kind)
		fmt.Fprintf(&tbl, `<tr data-id="%d"><td><span title="%s">%s</span> %s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`, item.Mid, item.Kind, icon, item.Name, item.Mac, item.Ip, item.Office, isOnline(item.Online))
	}

	tbl.WriteString("</tbody><table>")

	return tbl.String()
}
