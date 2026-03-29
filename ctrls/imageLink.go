package ctrls

import (
	"github.com/gbsto/daisy/db"
)

func ImageLink(cid int) string {
	return "images/" + db.GetImage(cid)
}

func SmallImageLink(cid int) (string, error) {
	return "images/" + db.GetSmallImage(cid), nil
}
