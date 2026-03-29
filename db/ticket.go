package db

import (
	"database/sql"
	"errors"
	"log"
	"strings"
)

type Ticket struct {
	Task       string `json:"task"`       // Not Used, only for post/get control
	Aid        int    `json:"aid"`        // Ticket ID (Action ID)
	Cid        int    `json:"cid"`        // Device = need searchable picklist
	Cid_ack    int    `json:"cid_ack"`    // Device Highlight
	Sid        int    `json:"sid"`        // Software Package
	Sid_ack    int    `json:"sid-ack"`    // Software Highlight
	Trouble    int    `json:"trouble"`    // Trouble select
	Report     string `json:"report"`     // Note - user comment
	Impact     int    `json:"impact"`     // Impact Select List
	Gid        int    `json:"gid"`        // Assigned Group
	Uid        int    `json:"uid"`        // Assigned User
	Uid_ack    int    `json:"uid_ack"`    // Assigned Acknowledged - Auto
	Inform_gid int    `json:"inform_gid"` // Notify Group
	Inform     int    `json:"inform"`     // Notify Person
	Inform_ack int    `json:"inform_ack"` // Notify Acknowledge
	Cmd        string `json:"cmd"`        // Action Log command
	Log        string `json:"log"`        // Action Log entry
	OldGid     int    `json:"oldgid"`     //
	OldUid     int    `json:"olduid"`     //
	OldGroup   string `json:"oldgroup"`   //
	OldUser    string `json:"olduser"`    //
	Informs    []int  `json:"informs"`    // Array of inform values
}

// Work Log
type Wlog struct {
	Task      string `json:"task"`      // Ignore, its for the front end
	Wid       int    `json:"wid"`       // Work log unique identifier
	Timestamp string `json:"timestamp"` // Time (int64) when it occured
	Aid       int    `json:"aid"`       // Action log unique identifier
	Uid       int    `json:"uid"`       // User ID of person assigned the alert
	Cid       int    `json:"cid"`       // Computer ID of affected computer
	Cmd       string `json:"cmd"`       // Action Log command
	Note      string `json:"note"`      // Action Log notes
	Fullname  string `json:"fullname"`  // Fullname of the person assigned
	OldUid    int    `json:"olduid"`    // Track record of who the alert was assigned to
	OldGid    int    `json:"oldgid"`    // Track record of the group the alert was assigned to
	NewUid    int    `json:"newuid"`    // Who the alert was newly assigned to
	NewGid    int    `json:"newgid"`    // Group newly assigned the alert
	Ack       int    `json:"uid_ack"`   // Flag to set / unset the four acknowlege checkboxes
}

// Small fast routine to get only the modifiable fields on a ticket (Alert)
func GetTicket(curUid, aid int) (Ticket, error) {
	var tkt Ticket
	var query strings.Builder
	addActionLock(curUid, aid)
	query.WriteString(`
		SELECT aid, 
			coalesce(cid, 0) cid, cid_ack, 
			coalesce(sid, 0) sid, sid_ack,
			coalesce(gid, 0) gid, coalesce(uid, 0) uid, uid_ack, 
			coalesce(inform_gid, 0) inform_gid, coalesce(inform, 0) inform, inform_ack, 
			coalesce(impact, 0) impact, coalesce(trouble, 0) trouble, coalesce(report, '') report
		FROM action_log
		WHERE aid=?
	`)
	rows, err := Conn.Query(query.String(), aid)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&tkt.Aid, &tkt.Cid, &tkt.Cid_ack,
			&tkt.Sid, &tkt.Sid_ack,
			&tkt.Gid, &tkt.Uid, &tkt.Uid_ack,
			&tkt.Inform_gid, &tkt.Inform, &tkt.Inform_ack,
			&tkt.Impact, &tkt.Trouble, &tkt.Report)
		if err != nil {
			log.Println(err)
			return tkt, err
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return tkt, err
}

func SetTicket(curUid int, tkt *Ticket) error {
	if isActionLocked(curUid, tkt.Aid) {
		return errors.New("the ticket was updated by someone else before you could save")
	}
	AddLog(curUid, tkt)
	query := `
		UPDATE action_log SET cid=?, cid_ack=?, sid=?, sid_ack=?,
		gid=?, uid=?, uid_ack=?, inform_gid=?, inform=?, inform_ack=?, 
		impact=?, trouble=?, report=? 
		WHERE aid=?
	`
	_, err := Conn.Exec(query, foreignKey(tkt.Cid), tkt.Cid_ack,
		foreignKey(tkt.Sid), tkt.Sid_ack, foreignKey(tkt.Gid),
		foreignKey(tkt.Uid), tkt.Uid_ack, foreignKey(tkt.Inform_gid),
		foreignKey(tkt.Inform), tkt.Inform_ack, tkt.Impact,
		tkt.Trouble, tkt.Report, tkt.Aid)
	if err != nil {
		log.Println(err)
		return err
	}
	updateAlerts(tkt.Aid, tkt.Informs)
	return nil
}

// Update the list of informs (alerts) for a ticket
// Add any new UIDs, remove any no longer in the new list
func updateAlerts(aid int, newList []int) {

	// Use a map to store old UIDs (no search of a slice is needed)
	oldUids := make(map[int]int)
	query := "SELECT uid, id FROM alerts WHERE aid=? AND wait=1 AND ack IS NULL"
	rows, err := Conn.Query(query, aid)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var uid, id int
		if err := rows.Scan(&uid, &id); err != nil {
			log.Println(err)
			continue
		}
		oldUids[uid] = id
	}

	// Start a transaction
	tx, err := Conn.Begin()
	if err != nil {
		log.Println(err)
		return
	}

	// Compare newList with oldUids map
	for _, uid := range newList {
		if _, exists := oldUids[uid]; exists {
			// If uid exists in oldUids, remove it from the map (to track leftover old UIDs)
			delete(oldUids, uid)
		} else if uid > 0 { // add the new uid
			_, err := Conn.Exec("INSERT INTO alerts (aid, uid, gid, ack, wait) VALUES (?,?,(SELECT gid FROM profiles WHERE uid=?),NULL,?)", aid, uid, uid, 1)
			if err != nil {
				log.Println(err)
				tx.Rollback()
				return
			}
		}
	}

	// Any remaing old uids should be deleted, they should no longer be in the DB list
	for _, id := range oldUids {
		_, err := Conn.Exec("DELETE FROM alerts WHERE id=?", id)
		if err != nil {
			log.Println(err)
			tx.Rollback()
			return
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Println(err)
	}
}

// Add an entry to the work log
func AddLog(curUid int, tkt *Ticket) error {
	if len(tkt.Log) == 0 {
		return nil
	}
	const query = "INSERT INTO wlog (aid, uid, cid, cmd, notes) VALUES (?,?,?,?,?)"
	_, err := Conn.Exec(query, foreignKey(tkt.Aid), foreignKey(curUid),
		foreignKey(tkt.Cid), tkt.Cmd, tkt.Log)
	if err != nil {
		log.Println(err)
		return err
	}
	if strings.ToUpper(tkt.Cmd) == "CLOSED" {
		closeTicket(curUid, tkt)
	}
	return nil
}

// Update the action_log to set the wlog field to the UID of who closed it
// This should be done once and only once
// action_log.wlog=0 means still open
func closeTicket(curUid int, tkt *Ticket) error {
	if tkt.Cid > 0 && tkt.Cid_ack == 0 {
		tkt.Cid_ack = curUid
	}
	if tkt.Sid > 0 && tkt.Sid_ack == 0 {
		tkt.Sid_ack = curUid
	}
	if (tkt.Gid > 0 || tkt.Uid > 0) && tkt.Uid_ack == 0 {
		tkt.Uid_ack = curUid
	}
	if (tkt.Inform_gid > 0 || tkt.Inform > 0) && tkt.Inform_ack == 0 {
		tkt.Inform_ack = curUid
	}
	const query = `UPDATE action_log SET active=0, closed=strftime('%s', 'now'),
		 wlog=?, closed_by=?, cid_ack=?, sid_ack=?, uid_ack=?, inform_ack=?
		WHERE aid=?`
	_, err := Conn.Exec(query, curUid, curUid, tkt.Cid_ack, tkt.Sid_ack, tkt.Uid_ack, tkt.Inform_ack, tkt.Aid)
	if err != nil {
		log.Println(err)
	}
	sendClosureEmail(tkt)
	return err
}

func sendClosureEmail(tkt *Ticket) error {
	log.Println("TODO: Need to send closure email to ", tkt.Inform)
	return nil
}

func GetLogs(curUid, aid int) ([]Wlog, error) {
	var items []Wlog
	tzoff := GetTzoff(curUid)
	const query = `
	SELECT coalesce(A.aid, 0) AS aid, 
		strftime('%Y-%m-%d %H:%M', A.timestamp-?, 'unixepoch') AS timestamp,
		coalesce(A.uid, 0) AS uid, coalesce(A.cid, 0) AS cid, 
		A.cmd, A.notes, B.fullname 
	FROM wlog A 
	LEFT JOIN profiles B ON A.uid=B.uid 
	WHERE A.aid=? ORDER BY A.timestamp
	`
	rows, err := Conn.Query(query, tzoff, aid)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Wlog
		err := rows.Scan(&item.Aid, &item.Timestamp, &item.Uid, &item.Cid, &item.Cmd, &item.Note, &item.Fullname)
		if err != nil {
			log.Println(err)
		} else {
			items = append(items, item)
		}
	}
	return items, nil
}
