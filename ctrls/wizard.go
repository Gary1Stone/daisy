package ctrls

import (
	"fmt"
	"strings"

	"github.com/gbsto/daisy/actions"
	"github.com/gbsto/daisy/icons"
)

func BuildWizardTitle(wizkey string) string {
	var title strings.Builder
	title.WriteString(icons.GetIcon("wizard"))
	title.WriteString(`&nbsp;Wizard `)
	items, _ := actions.GetActionCodes(false)
	for _, item := range items {
		if item.Name == wizkey {
			fmt.Fprintf(&title, `%s&nbsp;%s`, icons.GetIcon(item.Icon), item.Description)
		}
	}
	return title.String()
}
