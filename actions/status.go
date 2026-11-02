package actions

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
