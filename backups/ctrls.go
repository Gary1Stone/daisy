package backups

import (
	"fmt"
	"strings"
)

func (b *BackupInfo) BuildBackupsTable(uid, cid int) string {
	var tbl strings.Builder
	if cid == 0 || uid == 0 {
		return ""
	}

	backups, err := b.GetBackups(uid, cid)
	if err != nil || len(backups) == 0 {
		return ""
	}

	for _, backup := range backups {
		fmt.Fprintf(&tbl, `<p>%s, %s: %s,%s<hr></p>`, backup.Dated, backup.Source, backup.Volume, backup.What)
	}

	return tbl.String()
}
