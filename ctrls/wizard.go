package ctrls

import (
	"strings"

	"github.com/gbsto/daisy/db"
)

func BuildWizardTitle(wizkey string) string {
	var title strings.Builder
	title.WriteString(`<span class="mif-app icon"></span>&nbsp;Wizard `)
	items, _ := db.GetActionCodes(false)
	for _, item := range items {
		if item.Name == wizkey {
			title.WriteString(`<span class="`)
			title.WriteString(item.Icon)
			title.WriteString(` icon"></span>&nbsp;`)
			title.WriteString(item.Description)
		}
	}
	return title.String()
}
