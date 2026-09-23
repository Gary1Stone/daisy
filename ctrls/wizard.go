package ctrls

import (
	"fmt"
	"strings"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/svg"
)

func BuildWizardTitle(wizkey string) string {
	var title strings.Builder
	title.WriteString(svg.GetIcon("wizard"))
	title.WriteString(`&nbsp;Wizard `)
	items, _ := db.GetActionCodes(false)
	for _, item := range items {
		if item.Name == wizkey {
			fmt.Fprintf(&title, `%s&nbsp;%s`, svg.GetIcon(item.Icon), item.Description)
		}
	}
	return title.String()
}
