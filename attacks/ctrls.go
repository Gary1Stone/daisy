package attacks

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/icons"
)

func (a *AttackInfo) BuildAttacksTable(curUid, duration int) string {
	var table strings.Builder

	// Build the table header
	//table.WriteString(buildAttacksTableHeader())
	table.WriteString(fmt.Sprintf(`<table class='striped' id="attackstable" ><thead><tr>
        <th aria-sort="ascending" data-sort="asc">Occurred</th>
        <th aria-sort="none">Attacking IP</th>
        <th aria-sort="none">Attacks</th>
        <th aria-sort="none">Browser</th>
        <th aria-sort="none">Location</th>
		<th aria-sort="none">%s</th>
		<th aria-sort="none">Map</th>
    </tr></thead><tbody>`, icons.GetIcon("user")))

	// Fetch attacks items
	var attacks AttackInfo
	items, err := attacks.GetAttacksDetails(curUid, duration)
	if err != nil {
		log.Println(err)
		table.WriteString("</tbody></table>")
		return table.String()
	}

	// Build table rows
	icon := icons.GetIcon("map")
	for _, item := range items {
		table.WriteString(a.buildAttacksTableRow(&item, icon))
	}

	table.WriteString("</tbody></table>")
	return table.String()
}

// Helper function to build a single table row
func (a *AttackInfo) buildAttacksTableRow(item *AttackInfo, icon string) string {
	var row strings.Builder

	place := "<p>" + item.Country + "</p><p>" + item.State + " / " + item.City + "</p>"
	if item.City != item.Community {
		place += "<p>" + item.Community + "</p>"
	}

	fmt.Fprintf(&row, `<tr class='row-hover'><td>%s</td><td>%s</td>`, item.Occurred, item.Ip)
	fmt.Fprintf(&row, `<td>%d</td><td>%s</td>`, item.Attack_count, item.First_browser)
	fmt.Fprintf(&row, `<td>%s</td><td>%s</td>`, place, item.Fullname)
	fmt.Fprintf(&row, `<td><a href='https://www.google.com/maps/search/?api=1&query=%s, %s' target='_blank'>%s</a></td></tr>`, strconv.FormatFloat(item.Latitude, 'f', -1, 64), strconv.FormatFloat(item.Longitude, 'f', -1, 64), icon)

	return row.String()
}
