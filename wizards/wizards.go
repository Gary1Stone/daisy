package wizards

import (
	"fmt"
	"strings"

	"github.com/gbsto/daisy/actions"
	"github.com/gbsto/daisy/icons"
)

const (
	Sighting = "SIGHTING"
	Using    = "USING"
	Claiming = "CLAIMING"
	Giving   = "GIVING"
	Broken   = "BROKEN"
	Lost     = "LOST"
	Died     = "DIED"
	Care     = "CARE"
	Backup   = "BACKUP"
	Install  = "INSTALL"
	Remove   = "REMOVE"
	Request  = "REQUEST"
)

func BuildWizardTitle(wizkey string) string {
	var title strings.Builder
	fmt.Fprintf(&title, `%s&nbsp;Wizard `, icons.GetIcon("wizard"))
	items, _ := actions.GetActionCodes(false)
	for _, item := range items {
		if item.Name == wizkey {
			fmt.Fprintf(&title, `%s&nbsp;%s`, icons.GetIcon(item.Icon), item.Description)
			break
		}
	}
	return title.String()
}
