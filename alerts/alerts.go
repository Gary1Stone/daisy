package alerts

import (
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gbsto/daisy/actions"
	"github.com/gbsto/daisy/db"
)

// The alerts table is a subtable of the Action log.
// It is used when alerts (notifications) need to be given to users
type AlertTable struct {
	Id   int `json:"id"`   // Row ID from the table
	Aid  int `json:"aid"`  // Action_Log ID //:  -1,  // Get Any ticket
	Uid  int `json:"uid"`  // User ID //:  uid, // Just this user
	Gid  int `json:"gid"`  // User's group ID //:  -1,  // Get Any group
	Ack  int `json:"ack"`  // User ID of person acknowledging the inform (alert) //:  0,   // Not Acknowledged
	Wait int `json:"wait"` // 0=do not wait (default), 1=wait for closure of the ticket, changes to 0 when ticket is closed //: 0,   // Not waiting
}

type AlertInfo struct {
	AlertTable
	Action       string `json:"action"`       // Action (CLAIMING, SIGHTING...)
	Notes        string `json:"notes"`        // Description of the alert
	Cid          int    `json:"cid"`          // Computer ID
	DeviceName   string `json:"devicename"`   // Computer Name
	DeviceType   string `json:"devicetype"`   // Type of Device - Desktops, Laptops, Tablets
	DeviceIcon   string `json:"deviceicon"`   // Icon for the device
	Sid          int    `json:"sid"`          // Software ID
	SoftwareName string `json:"softwarename"` // Name of the software
	Uid_ack      int    `json:"uid_ack"`      // User Acknowledge the completion of the ticket
	ActionIcon   string `json:"actionicon"`   // Icon for the action
	Fullname     string `json:"fullname"`     // User's full name
}

type Alert struct {
	AlertTable
	Cid_ack int `json:"cid_ack"`
	Sid_ack int `json:"sid_ack"`
	Sid     int `json:"sid"`
	Cid     int `json:"cid"`
}

// Filter is based on Alert struct where:
// >0 means specific one, <0 means all, =0 means is null (or 0 for GID and WAIT)
func (a *Alert) buildAlertWhereClause() (string, []any) {
	var whereClause strings.Builder
	params := []any{}

	// Only get alerts that are not waiting for closure
	switch a.Wait {
	case 0:
		whereClause.WriteString("WHERE A.wait=0 ")
	case 1:
		whereClause.WriteString("WHERE A.wait=1 ")
	default:
		whereClause.WriteString("WHERE A.wait>=0 ")
	}
	// AID is mandatory
	if a.Aid > 0 {
		whereClause.WriteString("AND A.aid=? ")
		params = append(params, a.Aid)
	} else if a.Aid < 0 {
		whereClause.WriteString("AND A.aid>0 ")
	}
	// UID is optional (can assign only to a group)
	if a.Uid > 0 {
		whereClause.WriteString("AND A.uid=? ")
		params = append(params, a.Uid)
	} else if a.Uid == 0 {
		whereClause.WriteString("AND A.uid IS NULL ")
	} else if a.Uid < 0 {
		whereClause.WriteString("AND A.uid>0 ")
	}
	// GID is mandatory but not forgien key driven so can be 0
	if a.Gid >= 0 {
		whereClause.WriteString("AND A.gid=? ")
		params = append(params, a.Gid)
	} else if a.Gid < 0 {
		whereClause.WriteString("AND A.gid>=0 ")
	}
	// ACK is optional but can be null
	if a.Ack > 0 {
		whereClause.WriteString("AND A.ack=? ")
		params = append(params, a.Ack)
	} else if a.Ack == 0 {
		whereClause.WriteString("AND A.ack IS NULL ")
	} else if a.Ack < 0 {
		whereClause.WriteString("AND A.ack>0 ")
	}
	return whereClause.String(), params
}

// Read the AlertDetails
func (a *Alert) GetAlerts() ([]*AlertInfo, error) {
	alerts := make([]*AlertInfo, 0, 1000)
	whereClause, params := a.buildAlertWhereClause()
	whereClause += "AND I.is_device=1 "
	query := `
		SELECT A.id, coalesce(A.aid, 0), coalesce(A.Uid, 0), A.gid, coalesce(A.ack, 0), A.wait,
		B.action, coalesce(B.Cid, 0), coalesce(B.notes, ''), coalesce(B.sid, 0), coalesce(S.name, ''),
		D.name, D.type, I.icon, coalesce(B.uid_ack, 0), J.icon, P.fullname
		FROM alerts A
		LEFT JOIN action_log B on A.aid=B.aid
		LEFT JOIN devices D on D.cid=B.cid
		LEFT JOIN icons I ON D.type=I.name
		LEFT JOIN icons J ON B.action=J.name
		LEFT JOIN software S ON B.sid=S.sid 
		LEFT JOIN profiles P ON P.uid=A.uid
	`
	query += whereClause
	query += "LIMIT 1000 "
	rows, err := db.Conn.Query(query, params...)
	if err != nil {
		log.Println(err)
		if err == sql.ErrNoRows {
			return alerts, nil
		}
		return alerts, err
	}
	defer rows.Close()
	for rows.Next() {
		var alert AlertInfo
		err := rows.Scan(&alert.Id, &alert.Aid, &alert.Uid, &alert.Gid,
			&alert.Ack, &alert.Wait, &alert.Action, &alert.Cid, &alert.Notes,
			&alert.Sid, &alert.SoftwareName, &alert.DeviceName, &alert.DeviceType, &alert.DeviceIcon,
			&alert.Uid_ack, &alert.ActionIcon, &alert.Fullname)
		if err != nil {
			log.Println(err)
		} else {
			alerts = append(alerts, &alert)
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return alerts, err
	}
	return alerts, nil
}

func (a *Alert) SetAlert() error {
	if a.Id < 1 {
		log.Println("ID required to set alert")
		return errors.New("ID required to set alert")
	}
	var query = "UPDATE alerts SET aid=?, uid=?, ack=?, wait=? WHERE id=?"
	_, err := db.Conn.Exec(query, db.ForeignKey(a.Aid), db.ForeignKey(a.Uid), db.ForeignKey(a.Ack), a.Wait, a.Id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// curUid, aid, active int, alerts []alerts.AlertInfo
func (a *Alert) AddAlert() {
	query := "INSERT INTO alerts (aid, uid, gid, ack, wait) VALUES (?,?,?,?,?)"
	result, err := db.Conn.Exec(query, db.ForeignKey(a.Aid), db.ForeignKey(a.Uid), a.Gid, db.ForeignKey(a.Ack), a.Wait)
	if err != nil {
		log.Println(err)
		return
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		a.Aid = 0
		return
	}
	a.Id = int(lastInsertID)
}

// At ticket closing, run this to activate waiting alerts
func (a *Alert) RemoveAlertWait() error {
	if a.Aid < 1 {
		log.Println("AID required to remove alert")
		return errors.New("aid required to remove alert")
	}
	query := "UPDATE alerts set wait=0 WHERE aid=?"
	_, err := db.Conn.Exec(query, db.ForeignKey(a.Aid))
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Acknowedge the alert for the action log item
// Action_log Foreign keys: originator, cid, sid, uid, inform, closed_by
func (a *Alert) AckAlert(curUid, uid, aid int, uid_ack, cid_ack, sid_ack bool) error {
	if curUid < 1 || uid < 1 || aid < 1 {
		log.Println("no alert to ack")
		return errors.New("no alert to ack")
	}

	// Ack this alert
	query := "UPDATE alerts SET ack=? WHERE uid=? and aid=?"
	_, err := db.Conn.Exec(query, db.ForeignKey(curUid), db.ForeignKey(uid), db.ForeignKey(aid))
	if err != nil {
		log.Println(err)
		return err
	}

	// Check if there are any remaining pending alerts for this aid
	query = "SELECT count(*) FROM alerts WHERE aid=? AND ack IS NULL"
	cnt := 0
	err = db.Conn.QueryRow(query, db.ForeignKey(aid)).Scan(&cnt)
	if err != nil {
		log.Println(err)
		return err
	}
	if cnt > 0 {
		return nil
	}
	// To not get here means device and software highlights stay on until everyone ack'd their alerts.
	// Now check if the action can be closed
	act, err := actions.GetAction(curUid, aid)
	if err != nil {
		return err
	}
	//Check if action requires work log to be closed
	isWlogClosedNeeded := false
	switch act.Action {
	case "BROKEN", "CARE", "LOST", "DIED", "REQUEST":
		isWlogClosedNeeded = true
	}
	// If this user is acknowledging this user's UID alert
	// if uid_ack && act.Uid == curUid {
	// 	if !isWlogClosedNeeded {
	// 		act.Uid_ack = curUid
	// 	} else if act.Wlog > 0 {
	// 		act.Uid_ack = curUid
	// 	}
	// }
	// If this user is acknowledging another user's UID alert
	// if filterUid > 0 && (uid_ack && act.Uid == filterUid) {
	// 	if !isWlogClosedNeeded {
	// 		act.Uid_ack = curUid
	// 	} else if act.Wlog > 0 {
	// 		act.Uid_ack = curUid
	// 	}
	// }
	// If this user is ack their INFORM alert
	// if uid_ack && act.Inform == curUid {
	// 	act.Inform_ack = curUid
	// }
	// If this user is acknowledging another user's INFORM alert
	// if filterUid > 0 && (uid_ack && act.Inform == filterUid) {
	// 	act.Inform_ack = curUid
	// }

	//if this user is acking the device alert
	if cid_ack && act.Cid > 0 {
		act.Cid_ack = curUid
	}
	//if the user is acking the software alert
	if sid_ack && act.Sid > 0 {
		act.Sid_ack = curUid
	}
	//Check all pending acks are done
	isAllAcked := false
	if (act.Uid_ack > 0 || act.Uid == 0) &&
		(act.Cid_ack > 0 || act.Cid == 0) &&
		(act.Sid_ack > 0 || act.Sid == 0) &&
		(act.Inform_ack > 0 || act.Inform == 0) {
		isAllAcked = true
	}
	//Check if ready to close
	isReadyToClose := false
	if isAllAcked && !isWlogClosedNeeded {
		isReadyToClose = true
	} else if isAllAcked && isWlogClosedNeeded && act.Wlog > 0 {
		isReadyToClose = true
	}
	//Set the closing info
	if isReadyToClose {
		act.Closed_by = curUid
		act.ClosedInt = time.Now().UTC().Unix()
		act.Active = 0
		act.Wlog = curUid
	}
	err = act.UpdateAction(curUid)
	if err != nil {
		return err
	}
	if sid_ack && act.Sid > 0 {
		a.Sid = act.Sid
		go a.setSidColor()
	}
	if cid_ack && act.Cid > 0 {
		a.Cid = act.Cid
		go a.setCidColor()
	}
	if (uid_ack && act.Uid == curUid) || (uid_ack && act.Inform == curUid) {
		a.Uid = act.Uid
		go a.setUidColor()
	}
	return nil
}

// Find the next color for the software table
// Set the software color to the last open action's color or blank
func (a *Alert) setSidColor() error {
	if a.Sid < 1 {
		log.Println("sid is mandatory")
		return errors.New("sid is mandatory")
	}
	query := `
		SELECT B.color FROM action_log A
		LEFT JOIN icons B ON A.action=B.name
		WHERE sid=? AND sid_ack=0 
		ORDER BY A.opened DESC
		LIMIT 1
	`
	color := ""
	err := db.Conn.QueryRow(query, a.Sid).Scan(&color)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	query = "UPDATE software SET color=? WHERE sid=?"
	_, err = db.Conn.Exec(query, color, a.Sid)
	if err != nil {
		log.Println(err)
	}
	return nil
}

// Get the next device color and update the device table
func (a *Alert) setCidColor() error {
	if a.Cid < 1 {
		log.Println("cid is mandatory")
		return errors.New("cid is mandatory")
	}
	query := `
		SELECT B.color FROM action_log A
		LEFT JOIN icons B ON A.action=B.name
		WHERE cid=? AND cid_ack=0
		ORDER BY A.opened DESC
		LIMIT 1
	`
	color := ""
	err := db.Conn.QueryRow(query, a.Cid).Scan(&color)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	_, err = db.Conn.Exec("UPDATE devices SET color=? WHERE cid=?", color, a.Cid)
	if err != nil {
		log.Println(err)
	}
	return nil
}

// Check if this user has any remaining open actions in the action log and set the color in the profile table
// Set the profile color to the last open action's color or blank. UID is mandatory
func (a *Alert) setUidColor() error {
	if a.Uid < 1 {
		log.Println("uid is mandatory")
		return errors.New("uid is mandatory")
	}
	query := `
		SELECT B.color FROM action_log A
		LEFT JOIN icons B ON A.action=B.name
		WHERE (uid=? AND uid_ack=0) OR (inform=? AND inform_ack=0)
		ORDER BY A.opened DESC
		LIMIT 1
	`
	color := ""
	err := db.Conn.QueryRow(query, a.Uid, a.Uid).Scan(&color)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return err
	}
	_, err = db.Conn.Exec("UPDATE profiles SET color=? WHERE uid=?", color, a.Uid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
