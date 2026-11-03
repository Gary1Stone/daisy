package actions

type Popbutton struct {
	Color   string `json:"color"`
	Action  string `json:"action"`
	Label   string `json:"label"`
	Icon    string `json:"icon"`
	Active  int    `json:"active"`
	Aid     int    `json:"aid"`
	Cid_ack int    `json:"cid_ack"`
	Iid_ack int    `json:"iid_ack"`
	Sid_ack int    `json:"sid_ack"`
	Uid_ack int    `json:"uid_ack"`
	Wlog    int    `json:"wlog"` // Can the action be closed: 1=true, 0=false >0=closed
}

func CalculateStatus(act *Action) string {
	retVal := ""
	if act.Uid > 0 && act.Uid_ack == 0 {
		retVal += ": User"
	}
	if act.Cid > 0 && act.Cid_ack == 0 {
		retVal += ": Device"
	}
	if act.Sid > 0 && act.Sid_ack == 0 {
		retVal += ": Software"
	}
	if act.Inform > 0 && act.Inform_ack == 0 {
		retVal += ": Inform"
	}
	if len(retVal) > 0 {
		retVal = "Pending" + retVal
	} else if act.Active == 1 {
		retVal = "Open but completed"
	} else {
		retVal = "Closed"
	}
	return retVal
}
