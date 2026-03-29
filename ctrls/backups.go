package ctrls

import (
	"strings"

	"github.com/gbsto/daisy/db"
)

func BuildBackups(uid, cid int) string {
	var tbl strings.Builder
	if cid == 0 || uid == 0 {
		return ""
	}

	backups, err := db.GetBackups(uid, cid)
	if err != nil || len(backups) == 0 {
		return ""
	}

	for _, backup := range backups {
		tbl.WriteString("<p>")
		tbl.WriteString(backup.Dated + ", ")
		tbl.WriteString(backup.Source + ": ")
		tbl.WriteString(backup.Volume + ", ")
		tbl.WriteString(backup.What)
		tbl.WriteString("<hr></p>")
	}

	return tbl.String()
}
