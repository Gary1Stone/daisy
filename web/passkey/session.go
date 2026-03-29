package passkey

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gbsto/daisy/db"

	"github.com/go-webauthn/webauthn/webauthn"
)

type sessionInfo struct {
	sessionID string
	data      *webauthn.SessionData
}

// Save session information
// Requires sessionID and data be set before invoking
func (s *sessionInfo) saveSession() error {
	if len(s.sessionID) < 10 {
		return errors.New("invalid session id")
	}
	if len(s.data.UserID) == 0 {
		return errors.New("invalid webauthn.SessionData")
	}
	credParamsBytes, err := json.Marshal(s.data.CredParams)
	if err != nil {
		log.Println("Error marshaling CredParams:", err)
		return err
	}

	var params []any
	params = append(params, s.sessionID)           // string primary Key
	params = append(params, string(s.data.UserID)) // []byte  //=auth_id
	params = append(params, s.data.Challenge)      // string
	params = append(params, s.data.RelyingPartyID) // string
	now := time.Now().UTC()
	futureTime := now.Add(15 * time.Minute)
	params = append(params, futureTime.Unix())       // expiry date
	params = append(params, s.data.UserVerification) // string: required, preferred, discouraged, ""
	params = append(params, string(credParamsBytes))
	query := `
	INSERT INTO sessions (
			session_id, auth_id, challenge,
			relying_party_id, expires, user_verification, cred_params
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(session_id) DO UPDATE SET
		auth_id = excluded.auth_id,
		challenge = excluded.challenge,
		relying_party_id = excluded.relying_party_id,
		expires = excluded.expires,
		user_verification = excluded.user_verification,
		cred_params = excluded.cred_params;
	`
	_, err = db.Conn.Exec(query, params...)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Delete the session information
// Requires sessionID be set before invoking
func (s *sessionInfo) deleteSession() {
	query := "DELETE FROM sessions WHERE session_id=?"
	_, err := db.Conn.Exec(query, s.sessionID)
	if err != nil {
		log.Println(err)
	}
	purgeExpiredSessions()
}

// Purge expired sessions
func purgeExpiredSessions() {
	query := "DELETE FROM sessions WHERE expires < strftime('%s', 'now')"
	_, err := db.Conn.Exec(query)
	if err != nil {
		log.Println(err)
	}
}

// Gets the session infomation
// Requires sessionID be set before invoking
func (s *sessionInfo) getSession() (webauthn.SessionData, error) {
	var wasd webauthn.SessionData //w.a.s.d. = Web Authn Session Data
	var userID string
	var unixTimestamp int64
	var credParamsStr string
	query := "SELECT auth_id, challenge, relying_party_id, expires, user_verification, cred_params FROM sessions WHERE session_id=?"
	rows, err := db.Conn.Query(query, s.sessionID)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return wasd, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&userID, &wasd.Challenge, &wasd.RelyingPartyID, &unixTimestamp, &wasd.UserVerification, &credParamsStr)
		if err != nil {
			log.Println(err)
			return wasd, err
		}
		wasd.UserID = []byte(userID)
		wasd.Expires = time.Unix(unixTimestamp, 0).UTC()
		if len(credParamsStr) > 0 {
			if err := json.Unmarshal([]byte(credParamsStr), &wasd.CredParams); err != nil {
				return wasd, fmt.Errorf("error unmarshaling cred_params: %w", err)
			}
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return wasd, err
	}
	return wasd, nil
}

// genSessionID: Creates a 32 character random string for the sessionID
// Concept - save a bunch of random strings to a, 'available' database table,
// Remove any that are already in use. If the table gets low,
// add a bunch more before trying to fetch the next row.
func (s *sessionInfo) genSessionID() (string, error) {
	sid := ""
	cnt := 0
	// Count how many rows left
	query := "SELECT count(*) FROM sids"
	err := db.Conn.QueryRow(query).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	if cnt < 25 {
		s.addMoreSids()
	}
	// Get next available session id
	query = "SELECT sid FROM sids LIMIT 1"
	err = db.Conn.QueryRow(query).Scan(&sid)
	if err != nil {
		log.Println(err)
		return "", err
	}
	//remove it from the available list
	query = "DELETE FROM sids WHERE sid=?"
	_, err = db.Conn.Exec(query, sid)
	if err != nil {
		log.Println(err)
	}
	return sid, nil
}

func (s *sessionInfo) addMoreSids() {
	var sids []string
	for i := 0; i < 200; i++ {
		sid, err := genID(32)
		if err != nil {
			log.Println(err)
			continue
		}
		sids = append(sids, sid)
	}
	// Miracle of 1 insert for 200 values
	placeholders := make([]string, len(sids))
	values := make([]any, len(sids))
	for i, id := range sids {
		placeholders[i] = "(?)"
		values[i] = id
	}
	// Put the new sids into database sids (available sid) table
	query := fmt.Sprintf("INSERT INTO sids (sid) VALUES %s", strings.Join(placeholders, ","))
	_, err := db.Conn.Exec(query, values...)
	if err != nil {
		log.Println(err)
	}
	//Now remove any that are used already
	query = `
		DELETE FROM sids
		WHERE EXISTS (SELECT 1
			FROM sessions B
			WHERE sids.sid = B.session_id )
	`
	_, err = db.Conn.Exec(query)
	if err != nil {
		log.Println(err)
	}
}
