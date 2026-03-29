package db

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"
)

type Action struct {
	Aid                 int    `json:"aid"`                        //Action ID
	Action              string `json:"action"`                     //Action taken
	Originator          int    `json:"originator" db:"originator"` //UID of Who did it (Creator)
	OriginatorGroup     string `json:"originatorgroup" db:"originatorgroup"`
	OriginatorName      string `json:"originatorname" db:"originatorname"`
	OriginatorEmail     string `json:"originatoremail" db:"originatoremail"`
	Opened              string `json:"opened" db:"opened"`
	OpenedInt           int64  `json:"openedint" db:"openedint"`
	Cid                 int    `json:"cid" db:"cid"`
	Cid_ack             int    `json:"cid_ack" db:"cid_ack"`
	Devicename          string `json:"devicename" db:"devicename"`
	Devicetype          string `json:"devicetype" db:"devicetype"`
	DeviceIcon          string `json:"deviceicon"`
	Sid                 int    `json:"sid" db:"sid"`
	Sid_ack             int    `json:"sid_ack" db:"sid_ack"`
	Softwarename        string `json:"softwarename" db:"softwarename"`
	Impact              int    `json:"impact" db:"impact"`
	Report              string `json:"report"`
	Notes               string `json:"notes" db:"notes"`
	Active              int    `json:"active" db:"active"`
	Closed_by           int    `json:"closed_by" db:"closed_by"`
	Color               string `json:"color" db:"color"`
	Icon                string `json:"icon" db:"icon"`
	Localtime           string `json:"localtime"`
	Closedtime          string `json:"closedtime"`
	ClosedInt           int64  `json:"closed"`
	NowGMT              string `json:"nowgmt"`
	Gid                 int    `json:"gid"`
	AssignedGroupName   string `json:"assignedGroupName"`
	Uid                 int    `json:"uid"`
	AssignedUserName    string `json:"assignedusername"`
	Uid_ack             int    `json:"uid_ack"`
	AssignedUserNameAck string `json:"assignedusernameack"` // Who acknowledged the notification
	Inform_gid          int    `json:"inform_gid"`          // Group ID of who needs to be informed
	Inform              int    `json:"inform"`              // User ID of who needs to be informed
	Inform_ack          int    `json:"inform_ack"`
	InformGroupName     string `json:"informgroupname"`
	InformUserName      string `json:"informusername"`
	InformUserNameAck   string `json:"informusernameack"` // Who acknowledged the inform notification
	CidUserNameAck      string `json:"cidusernameack"`    // Who acknowledged the Computer notification
	SidUserNameAck      string `json:"sidusernameack"`    // Who acknowledged the Software notification
	ActionDescription   string `json:"actiondescription"`
	Wlog                int    `json:"wlog"` // Worklog items are closed by who (>0) or active (0)
	Image               string `json:"image"`
	Trouble             int    `json:"trouble"` // Trouble classification Code
	TroubleDescription  string `json:"troubledescription"`
}

type ActionFilter struct {
	Task               string `json:"task"` // Not Used in filtering, only for post/get control
	Action             string `json:"action"`
	DevType            string `json:"type"`
	Site               string `json:"site"`
	Office             string `json:"office"`
	Aid                int    `json:"aid"`
	Active             int    `json:"active"`
	Gid                int    `json:"gid"`
	Uid                int    `json:"uid"`
	Sid                int    `json:"sid"`
	Cid                int    `json:"cid"`
	Impact             int    `json:"impact"`
	Pending            int    `json:"pending"` // -1 disables search for pending alerts,
	IncludeBlankOption bool   `json:"includeBlankOption"`
	DefaultToCurUser   bool   `json:"defaultToCurUser"`
	IncludeOtherOption bool   `json:"includeOtherOption"`
	Page               int    `json:"page"` // Used to limit the number of items returned in a search
}

type Actions struct {
	Id          int    `json:"id"`          // Action ID
	Name        string `json:"name"`        // Action Name
	Description string `json:"description"` // Action user visible description
	Color       string `json:"color"`       // Color (probably not used)
	Priority    int    `json:"priority"`    // Action Priority
	Icon        string `json:"icon"`        // Action icon
	Is_device   int    `json:"is_device"`   // Is this action on a device
}

// Return struct of the code/descriptions
func GetActionCodes(onlyDevices bool) ([]Actions, error) {
	var items []Actions
	query := "SELECT id, name, description, color, priority, icon, is_device FROM icons "
	if onlyDevices {
		query += "WHERE is_device=1 "
	}
	query += "ORDER BY description"
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		var dto Actions
		err := rows.Scan(&dto.Id, &dto.Name, &dto.Description, &dto.Color, &dto.Priority, &dto.Icon, &dto.Is_device)
		if err != nil {
			log.Println(err)
		} else {
			items = append(items, dto)
		}
	}
	return items, nil
}

// Pending=-1 disables search for pending alerts, because it is not used for the queue functionality
// It is only used to show open (pending) alerts on other screens (device & software)

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

// When a new struct is created, strings are set to zero length ("") and ints are set to 0 automatically
func (A *Action) Init() {
	A.Active = 1
}

// get one Action log entry
func GetAction(curUid, aid int) (*Action, error) {
	filter := new(ActionFilter) //create in heap to share with other functions
	filter.Aid = aid
	filter.Active = -1 //Flag to say don't care if open or closed
	actions, err := filter.GetActions(curUid)
	if err != nil {
		log.Println(err)
		return actions[0], err
	}
	if len(actions) == 0 {
		log.Println("ERROR: No Record Found!")
		return actions[0], err
	}
	addActionLock(curUid, aid)
	return actions[0], nil
}

func readActionTable(curUid, page int, whereClause string, params ...any) ([]*Action, error) {
	acts := make([]*Action, 0, 50) // Preallocate space for 50 items
	tzoff := GetTzoff(curUid)
	var query strings.Builder
	query.WriteString(`
		SELECT 
			A.aid, A.action, coalesce(A.originator, 0) originator, coalesce(A.cid, 0) cid, A.cid_ack, coalesce(A.sid, 0) sid, 
			A.sid_ack, A.gid, coalesce(A.uid, 0) uid, A.uid_ack, coalesce(A.inform, 0) inform, A.inform_ack, 
			coalesce(A.impact, 0) impact, A.notes, A.active, coalesce(A.closed_by, 0) closed_by, 
			coalesce(G.color, '') colour, coalesce(G.icon,'') icon, 
			A.opened as openedInt, A.closed AS closedInt, 
			strftime('%Y-%m-%d %H:%M', A.opened-?, 'unixepoch') AS opentime, 
			strftime('%Y-%m-%d %H:%M', A.opened-?, 'unixepoch') AS localtime, 
			strftime('%Y-%m-%d %H:%M', A.closed-?, 'unixepoch') AS closedtime, 
			strftime('%s', 'now') AS nowGMT, 
			coalesce(B.fullname, '') OriginatorName, coalesce(B.user, '') OriginatorEmail, coalesce(C.name, '') softwarename, 
			coalesce(D.name, '') devicename, coalesce(D.type, '') devicetype, 
			coalesce(E.fullname, '') UserNameAffected, 
			coalesce(F.fullname, '') UserNameInform, 
			coalesce(G.description, '') ActionDescription, 
			coalesce(H.description, '') GroupName,
			coalesce(J.description, '') InformGroupName,
			A.wlog, coalesce(A.inform_gid, 0) inform_gid, coalesce(D.image, '') image,
			coalesce(K.fullname, '') AssignedUserNameAck, 
			coalesce(L.fullname, '') InformUserNameAck,
			coalesce(A.report, '') report,
			coalesce(M.fullname, '') CidUserNameAck,
			coalesce(N.fullname, '') SidUserNameAck, 
			coalesce(A.trouble, 0) trouble,
			coalesce(P.description, '') troubledescription
		FROM action_log A 
		LEFT JOIN profiles B ON A.originator = B.uid 
		LEFT JOIN software C ON A.sid = C.sid 
		LEFT JOIN devices D ON A.cid = D.cid 
		LEFT JOIN profiles E ON A.uid = E.uid 
		LEFT JOIN profiles F ON A.inform = F.uid 
		LEFT JOIN icons G ON A.action = G.name 
		LEFT JOIN choices H ON A.gid = H.code AND H.field='GROUP' 
		LEFT JOIN choices J ON A.inform_gid = J.code AND J.field='GROUP'
		LEFT JOIN profiles K ON A.uid_ack = K.uid 
		LEFT JOIN profiles L ON A.inform_ack = L.uid 
		LEFT JOIN profiles M ON A.cid_ack = M.uid 
		LEFT JOIN profiles N ON A.sid_ack = N.uid
		LEFT JOIN choices P ON A.trouble = P.code AND P.field='TROUBLE' 
	`)
	query.WriteString(whereClause)
	query.WriteString("GROUP BY A.aid ")
	query.WriteString("ORDER BY A.impact DESC, A.opened DESC ")
	if page >= 0 {
		query.WriteString("LIMIT 50 OFFSET ")
		query.WriteString(strconv.Itoa(page * 50))
	}

	// Prepend the time zone conversion values to the original slice
	prependValues := []any{tzoff, tzoff, tzoff}
	params = append(prependValues, params...)

	//Perform the query
	//	log.Println(query.String())
	rows, err := Conn.Query(query.String(), params...)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	defer rows.Close()

	//Get the results into a slice of structs
	for rows.Next() {
		var act Action
		err := rows.Scan(&act.Aid, &act.Action, &act.Originator, &act.Cid, &act.Cid_ack,
			&act.Sid, &act.Sid_ack, &act.Gid, &act.Uid, &act.Uid_ack, &act.Inform,
			&act.Inform_ack, &act.Impact, &act.Notes, &act.Active, &act.Closed_by,
			&act.Color, &act.Icon, &act.OpenedInt, &act.ClosedInt,
			&act.Opened, &act.Localtime, &act.Closedtime, &act.NowGMT,
			&act.OriginatorName, &act.OriginatorEmail, &act.Softwarename, &act.Devicename,
			&act.Devicetype, &act.AssignedUserName, &act.InformUserName, &act.ActionDescription,
			&act.AssignedGroupName, &act.InformGroupName, &act.Wlog, &act.Inform_gid, &act.Image,
			&act.AssignedUserNameAck, &act.InformUserNameAck, &act.Report, &act.CidUserNameAck,
			&act.SidUserNameAck, &act.Trouble, &act.TroubleDescription)
		if err != nil {
			log.Println(err)
		} else {
			act.DeviceIcon = GetIcon(act.Devicetype)
			acts = append(acts, &act)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
	}
	return acts, err
}

func GetAllActionableActions(curUid int) ([]*Action, error) {
	var params []any
	whereClause := "WHERE A.active=1 AND A.action IN ('BROKEN', 'DIED', 'LOST', 'CARE', 'REQUEST') "
	return readActionTable(curUid, -1, whereClause, params...)
}

// Get the count of the outstanding (not acknowledged) action log items for the current user
func CountPendingTickets(CurUid int) int {
	cnt := 0
	query := `
		SELECT count(*) FROM action_log 
		WHERE uid=? AND active=1 AND uid_ack=0 
		AND action IN ('BROKEN', 'DIED', 'LOST', 'CARE', 'REQUEST')
	`
	err := Conn.QueryRow(query, CurUid, CurUid).Scan(&cnt)
	if err != nil {
		log.Println(err)
	}
	return cnt
}

// Fetch the action log entries for user or software or computer screens
func (filter *ActionFilter) GetActions(curUid int) ([]*Action, error) {
	whereClause, params, err := filter.buildActionWhereClause()
	if err != nil {
		log.Println(err)
		return []*Action{}, err
	}
	return readActionTable(curUid, filter.Page, whereClause, params...)
}

func (filter *ActionFilter) buildActionWhereClause() (string, []any, error) {
	var params []any
	var clause strings.Builder
	// Get the individual action
	if filter.Aid > 0 {
		params = append(params, filter.Aid)
		clause.WriteString("AND A.aid=? ")
	}
	if filter.Active >= 0 { //Any negative number means don't filter by active
		params = append(params, filter.Active)
		clause.WriteString("AND A.active=? ")
	}
	if filter.Gid > 0 {
		params = append(params, filter.Gid)
		params = append(params, filter.Gid)
		clause.WriteString("AND (A.gid=? ")
		clause.WriteString("OR A.inform_gid=?) ")
	}
	if filter.Uid > 0 {
		params = append(params, filter.Uid)
		clause.WriteString("AND ((A.uid=? ")
		if filter.Pending == 0 {
			clause.WriteString("AND A.uid_ack=0 ")
		}
		params = append(params, filter.Uid)
		clause.WriteString(") OR (A.inform=? ")
		if filter.Pending == 0 {
			clause.WriteString("AND A.inform_ack=0 ")
		}
		clause.WriteString(")) ")
	}
	if filter.Sid > 0 {
		params = append(params, filter.Sid)
		clause.WriteString("AND A.sid=? ")
		if filter.Pending == 0 { //open/pending alerts
			clause.WriteString("AND sid_ack=0 ")
		}
	} else if filter.Sid < 0 { // -1 flag = get all software installs/removes
		clause.WriteString("AND A.sid>0 ")
		if filter.Pending == 0 { // open/pending alerts
			clause.WriteString("AND sid_ack=0 ")
		}
	}
	if filter.Cid > 0 {
		params = append(params, filter.Cid)
		clause.WriteString("AND A.cid=? ")
		if filter.Pending == 0 {
			clause.WriteString("AND A.cid_ack=0 ")
		}
	}
	if filter.Impact > 0 {
		params = append(params, filter.Impact)
		clause.WriteString("AND A.impact=? ")
	}
	if len(filter.Action) > 0 {
		params = append(params, filter.Action)
		clause.WriteString("AND A.action=? ")
	}
	if len(filter.DevType) > 0 {
		params = append(params, filter.DevType)
		clause.WriteString("AND D.type=? ")
	}
	if len(filter.Office) > 0 {
		params = append(params, filter.Office)
		clause.WriteString("AND D.office=? ")
	}
	if len(filter.Site) > 0 {
		params = append(params, filter.Site)
		clause.WriteString("AND D.site=? ")
	}
	// Remove leading "AND ", replace it with "WHERE "
	whereClause := clause.String()
	if len(whereClause) > 5 {
		whereClause = "WHERE " + whereClause[4:]
	}
	return whereClause, params, nil
}

// Add a new action to the action log, checking if it is open or closed
func (act *Action) AddAction(curUid int) error {
	// INSTALL/REMOVE actions can set the opened time, otherwise set it to now
	if act.OpenedInt == 0 {
		act.OpenedInt = time.Now().UTC().Unix() // GMT time in seconds
	}
	//Set the group IDs for the user and inform
	act.Gid = GetGid(act.Uid)
	//	act.Inform_gid, _ = GetGid(act.Inform)
	//If action is to be closed, set the closed time and ack(s)
	if act.Active == 0 {
		act.ClosedInt = act.OpenedInt // closed when it was opened
		act.Closed_by = curUid
		if act.Uid > 0 {
			act.Uid_ack = curUid
		}
		if act.Cid > 0 {
			act.Cid_ack = curUid
		}
		if act.Sid > 0 {
			act.Sid_ack = curUid
		}
	}
	//TODO: Add email inform person if it was closes, then inform_ack=curUid
	return act.insertAction()
}

// Insert new Action
func (act *Action) insertAction() error {
	query := `
		INSERT INTO action_log (
			opened, impact, active, originator, gid, cid, sid, uid, 
			inform, inform_gid, cid_ack, sid_ack, uid_ack, inform_ack, action, notes,
			closed, closed_by, report, trouble
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	`
	result, err := Conn.Exec(query, act.OpenedInt, act.Impact, act.Active, foreignKey(act.Originator),
		act.Gid, foreignKey(act.Cid), foreignKey(act.Sid), foreignKey(act.Uid),
		foreignKey(act.Inform), act.Inform_gid, act.Cid_ack, act.Sid_ack, act.Uid_ack, act.Inform_ack, act.Action,
		act.Notes, act.ClosedInt, foreignKey(act.Closed_by), act.Report, act.Trouble)
	if err != nil {
		log.Println(err)
		return err
	}
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		act.Aid = 0
		return err
	}
	act.Aid = int(lastInsertID)
	return nil
}

// Update existing actions
func (act *Action) updateAction(curUid int) error {
	if isActionLocked(curUid, act.Aid) {
		return errors.New("record was updated by someone else before you tried to save")
	}
	query := `
		UPDATE action_log SET
		uid_ack=?, inform_ack=?, cid_ack=?, sid_ack=?, 
		active=?, closed_by=?, closed=?, wlog=?, report=?, trouble=?
		WHERE aid=?
	`
	_, err := Conn.Exec(query, act.Uid_ack, act.Inform_ack, act.Cid_ack,
		act.Sid_ack, act.Active, foreignKey(act.Closed_by),
		act.ClosedInt, act.Wlog, act.Report, act.Trouble, act.Aid)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Fetch the action log open issues
func GetDeviceIssues(curUid int, devType string) ([]*Action, error) {
	var params []any
	params = append(params, devType)
	clause := "WHERE A.cid>0 AND cid_ack=0 AND D.type=? AND D.active=1 "
	clause += "AND A.action IN ('BROKEN', 'CARE', 'LOST', 'DIED', 'REQUEST') "
	acts, err := readActionTable(curUid, -1, clause, params...)
	if err != nil {
		log.Println(err)
	}
	return acts, err
}

type installedSoftware struct {
	Id           int    `json:"id" db:"id"`                     // Software ID
	Sid          int    `json:"sid" db:"sid"`                   // Software ID
	Name         string `json:"name" db:"name"`                 // Software Package Name
	ScanDate     string `json:"scandate" db:"scandate"`         // Last scan date
	IsTracked    bool   `json:"istracked" db:"istracked"`       // Is it also in the action log with the INSTALL action
	PreInstalled bool   `json:"preinstalled" db:"preinstalled"` // Does this computer's software come with it's own OEM licence?
}

// build a slice of software names
func GetInstalledSoftware(curUid, cid int) ([]installedSoftware, error) {
	items := make([]installedSoftware, 0)
	if cid < 1 {
		return items, nil
	}
	tzoff := GetTzoff(curUid) //Time Zone Offest in minutes
	// Get all the manually tracked softare on each computer
	tracked, err := GetManuallyTrackedSoftwareOnComputers(cid)
	if err != nil {
		log.Println(err)
	}
	// Make a MAP so dont have to scan through array every time
	trackSid := make(map[int]bool)
	for _, trk := range tracked {
		trackSid[trk.Sid] = true
	}
	// get all the software for this computer in inventory
	query := `
		SELECT id, name, coalesce(sid, 0) as sid, preinstalled,
		strftime('%Y-%m-%d %H:%M', scandate-?, 'unixepoch') AS scanDate
		FROM sw_inv
		WHERE cid=?
		ORDER BY sid desc, name asc
	`
	rows, err := Conn.Query(query, tzoff, cid)
	if err != nil {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	var item installedSoftware
	for rows.Next() {
		err := rows.Scan(&item.Id, &item.Name, &item.Sid, &item.PreInstalled, &item.ScanDate)
		if err != nil {
			log.Println(err)
		} else {
			item.IsTracked = trackSid[item.Sid]
			items = append(items, item)
		}
	}
	return items, rows.Err()
}

type installed struct {
	ScanDate string `json:"scandate" db:"scandate"` // Last Scan Date
	Cid      int    `json:"cid" db:"cid"`           // Computer ID for link
	Name     string `json:"name" db:"name"`         // Unique name for the computer
	Model    string `json:"model" db:"model"`       // Computer model
	Icon     string `json:"icon" db:"-"`            // Icon for the computer
}

// MAP for software page listing all the computers it is installed on
func GetInstalledComputers(curUid, sid int) ([]installed, error) {
	items := make([]installed, 0)
	if sid < 1 {
		return items, nil
	}
	tzoff := GetTzoff(curUid) //Time Zone Offest in minutes
	query := `
		SELECT strftime('%Y-%m-%d %H:%M', scandate-?, 'unixepoch') AS scanDate, 
		coalesce(A.cid, 0) as cid, B.name, B.model, C.icon FROM sw_inv A 
		LEFT JOIN devices B ON B.cid=A.cid
		LEFT JOIN icons C ON B.type=C.name
		WHERE A.sid=? AND B.active=1 
		ORDER By B.name
		`
	rows, err := Conn.Query(query, tzoff, sid)
	if err != nil {
		log.Println(err)
		return items, err
	}
	defer rows.Close()
	var dto installed
	for rows.Next() {
		err := rows.Scan(&dto.ScanDate, &dto.Cid, &dto.Name, &dto.Model, &dto.Icon)
		if err != nil {
			log.Println(err)
		} else {
			items = append(items, dto)
		}
	}
	return items, rows.Err()
}

func AckAction(curUid, aid int, isUid, isCid, isSid bool) error {
	params := []any{}
	query := "UPDATE action_log SET "
	fields := map[string]bool{
		"uid_ack": isUid,
		"cid_ack": isCid,
		"sid_ack": isSid,
	}
	var setClauses []string
	for field, isSet := range fields {
		if isSet {
			setClauses = append(setClauses, field+"=?")
			params = append(params, curUid) // Always append curUid for each set field
		}
	}
	if len(setClauses) > 0 {
		query += strings.Join(setClauses, ", ")
	} else {
		return nil
	}
	query += " WHERE aid=?"
	params = append(params, aid) // Append aid at the end, after the SET clause
	_, err := Conn.Exec(query, params...)
	if err != nil {
		log.Println(err)
	}
	return err
}
