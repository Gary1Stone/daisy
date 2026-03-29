package passkey

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gbsto/daisy/db"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type credentialInfo struct {
	//	uid          int
	username     string
	passcode     string
	authID       string
	credentialID string
	user         PasskeyUser
}

// Return the User ID (uid), fullname, mins since last update, if in the database, otherwise 0
// Requires the username be set before invoking
func (c *credentialInfo) getUid() (int, string, int, error) {
	uid := 0 // User ID
	fullname := ""
	mins := 0 // minutes since last updated time
	query := "SELECT uid, fullname, ((strftime('%s', 'now') - last_updated_time)/60) AS mins FROM profiles WHERE active=1 AND user=?"
	err := db.Conn.QueryRow(query, c.username).Scan(&uid, &fullname, &mins)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return uid, fullname, 0, err
	}
	return uid, fullname, mins, nil
}

// Return the authID (authorization user ID) and if no authID
// available, create one and save to the user profile
// Requires the username be set before invoking
func (c *credentialInfo) getAuthID() (string, error) {
	authID := ""
	uid := 0
	query := "SELECT uid, COALESCE(auth_id, '') FROM profiles WHERE active=1 AND user=?"
	err := db.Conn.QueryRow(query, c.username).Scan(&uid, &authID)
	if err != nil {
		log.Println(err)
		return authID, err
	}
	if len(authID) > 0 {
		return authID, nil
	}
	authID, err = c.genAuthID()
	if err != nil {
		log.Println(err)
		return authID, err
	}
	query = "UPDATE profiles SET auth_id=? WHERE uid=?"
	_, err = db.Conn.Exec(query, authID, uid)
	if err != nil {
		log.Println(err)
		return authID, err
	}
	return authID, nil
}

// Check if the user entered username and passcode is correct
// Requires the username and passcode be set before invoking
func (c *credentialInfo) isValidUser() bool {
	uid := 0
	query := "SELECT uid FROM profiles WHERE active=1 AND user=? AND otp=?"
	err := db.Conn.QueryRow(query, c.username, c.passcode).Scan(&uid)
	if err != nil {
		return false
	}
	return uid > 0
}

// Save the user's credentials
// Requires the user (PasskeyUser) be set before invoking
func (c *credentialInfo) saveCredentials() error {
	type clientDataJSON struct {
		Type        string `json:"type"`
		Challenge   string `json:"challenge"`
		Origin      string `json:"origin"`
		CrossOrigin bool   `json:"crossOrigin"`
	}
	var clientData clientDataJSON
	WebAuthnCredentials := c.user.WebAuthnCredentials()
	for _, creds := range WebAuthnCredentials {
		// It's crucial to re-initialize the params slice for each credential to prevent errors.
		var params []any
		params = append(params, string(c.user.WebAuthnID())) //auth_id column
		params = append(params, c.user.WebAuthnName())
		params = append(params, c.user.WebAuthnDisplayName())
		c.credentialID = bytesToBase64String(creds.ID)
		params = append(params, bytesToBase64String(creds.ID))
		params = append(params, bytesToBase64String(creds.PublicKey))
		//Make a varable slice to hold the different transport methods "usb,nfc,ble,smart-card,hybrid,internal"
		transportStrings := make([]string, len(creds.Transport))
		// Fill in the slice
		for i, transport := range creds.Transport {
			transportStrings[i] = string(transport)
		}
		// Join the []string into a single string with a comma separator
		transportsString := strings.Join(transportStrings, ",")
		params = append(params, transportsString)
		params = append(params, bytesToBase64String(creds.Attestation.AuthenticatorData))
		params = append(params, creds.AttestationType)
		params = append(params, bytesToBase64String(creds.Attestation.ClientDataHash))
		params = append(params, string(creds.Attestation.ClientDataJSON))
		// Unmarshal the JSON into the struct
		err := json.Unmarshal(creds.Attestation.ClientDataJSON, &clientData)
		if err != nil {
			log.Println("Error unmarshaling JSON:", err)
		}
		params = append(params, clientData.Challenge)
		params = append(params, creds.Attestation.PublicKeyAlgorithm) // -7 for "ES256" and -257 for "RS256"
		params = append(params, creds.Authenticator.Attachment)
		params = append(params, bytesToBase64String(creds.Authenticator.AAGUID))
		params = append(params, creds.Authenticator.SignCount)
		if creds.Authenticator.CloneWarning {
			params = append(params, 1)
		} else {
			params = append(params, 0)
		}
		if creds.Flags.BackupEligible {
			params = append(params, 1)
		} else {
			params = append(params, 0)
		}
		if creds.Flags.BackupState {
			params = append(params, 1)
		} else {
			params = append(params, 0)
		}
		if creds.Flags.UserPresent {
			params = append(params, 1)
		} else {
			params = append(params, 0)
		}
		if creds.Flags.UserVerified {
			params = append(params, 1)
		} else {
			params = append(params, 0)
		}
		query := `
		INSERT INTO credentials (
			auth_id, webAuthnName, webAuthnDisplayName,
			credentials_id, PublicKey, Transport,
			AuthenticatorData, AttestationType,
			ClientDataHash, ClientDataJSON, Challenge, PublicKeyAlgorithm,
			Attachment, AAGUID,	
			SignCount, CloneWarning,
			BackupEligible, BackupState, UserPresent, UserVerified
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(credentials_id) DO UPDATE SET
			auth_id = excluded.auth_id, 
			webAuthnName = excluded.webAuthnName, 
			webAuthnDisplayName = excluded.webAuthnDisplayName,
			credentials_id = excluded.credentials_id, 
			PublicKey = excluded.PublicKey, 
			Transport = excluded.Transport,
			AuthenticatorData = excluded.AuthenticatorData, 
			AttestationType = excluded.AttestationType,
			ClientDataHash = excluded.ClientDataHash, 
			ClientDataJSON = excluded.ClientDataJSON, 
			Challenge = excluded.Challenge, 
			PublicKeyAlgorithm = excluded.PublicKeyAlgorithm,
			Attachment = excluded.Attachment, 
			AAGUID = excluded.AAGUID,	
			SignCount = excluded.SignCount, 
			CloneWarning = excluded.CloneWarning,
			BackupEligible = excluded.BackupEligible, 
			BackupState = excluded.BackupState, 
			UserPresent = excluded.UserPresent, 
			UserVerified = excluded.UserVerified;
		`
		_, err = db.Conn.Exec(query, params...)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// Deletes existing user credentials
// Requires the credentialID be set before invoking
func (c *credentialInfo) deleteCredentials() {
	query := "DELETE FROM credentials WHERE credentials_id=?"
	_, err := db.Conn.Exec(query, c.credentialID)
	if err != nil {
		log.Println(err)
	}
}

// Purge any credentials older than 30 days
func (c *credentialInfo) purgeCredentials() {
	query := "DELETE FROM credentials WHERE cast((strftime('%s', 'now') - created) / 86400 AS INTEGER)>30"
	_, err := db.Conn.Exec(query, c.credentialID)
	if err != nil {
		log.Println(err)
	}

	// delete any credentials that are not linked to an active user profile (cleanup orphaned credentials)
	query = `
	DELETE FROM credentials
	WHERE auth_id NOT IN (SELECT auth_id FROM profiles WHERE active=1)
	`
	_, err = db.Conn.Exec(query)
	if err != nil {
		log.Println(err)
	}

	// delete any credentials for users that have not logged in for over 90 days (cleanup old credentials)
	query = `
	DELETE FROM credentials
	WHERE auth_id IN (
		SELECT auth_id FROM profiles 
		WHERE active=1 AND ((strftime('%s', 'now') - last_updated_time) / 86400) > 90
	)
	`
	_, err = db.Conn.Exec(query)
	if err != nil {
		log.Println(err)
	}

	// Delete all but the most recent 5 credentials for each user (cleanup old credentials but keep some history)
	query = `
	DELETE FROM credentials
	WHERE credentials_id NOT IN (
		SELECT credentials_id FROM (
			SELECT credentials_id FROM credentials WHERE auth_id IN (SELECT auth_id FROM profiles WHERE active=1)
			ORDER BY created DESC
			LIMIT 5
		)
	)
	`
	_, err = db.Conn.Exec(query)
	if err != nil {
		log.Println(err)
	}

}

// Reads the user's credentials from the database
// Requires the authID be set before invoking
func (c *credentialInfo) getCredentials(isNew bool) (PasskeyUser, error) {
	user := &User{
		ID:          []byte(c.authID),
		DisplayName: c.authID,
		Name:        c.authID,
	}
	fullname := ""
	query := "SELECT fullname FROM profiles WHERE active=1 AND auth_id=?"
	err := db.Conn.QueryRow(query, c.authID).Scan(&fullname)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, nil
		}
		return user, err
	}
	user.DisplayName = fullname
	user.Name = fullname
	if isNew {
		return user, nil
	}
	var (
		// creds is now created inside the loop to avoid pointer issues.
		webAuthID         string // string to []byte
		webAuthnName      string
		webAuthnDispName  string
		credentials_id    string // base64 to []byte
		PublicKey         string // base64 to []byte
		Transport         string // comma seperated list to slice
		AuthenticatorData string // base64 to []byte
		ClientDataHash    string // base64 to []byte
		ClientDataJSON    string // to []byte
		Attachment        string // platform or cross-platform
		AAGUID            string // base64 to []byte
		CloneWarning      int    // int to bool
		BackupEligible    int    // int to bool
		BackupState       int    // int to bool
		UserPresent       int    // int to bool
		UserVerified      int    // int to bool
	)
	query = `
		SELECT 
		auth_id, webAuthnName, webAuthnDisplayName,
		credentials_id, PublicKey, Transport,
		AuthenticatorData, AttestationType,
		ClientDataHash, ClientDataJSON, PublicKeyAlgorithm,
		Attachment, AAGUID,	
		SignCount, CloneWarning,
		BackupEligible, BackupState, UserPresent, UserVerified
		FROM credentials WHERE auth_id=?
	`
	rows, err := db.Conn.Query(query, c.authID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error querying for credentials with authID %s: %v", c.authID, err)
		return user, err
	}
	defer rows.Close()

	credCount := 0
	for rows.Next() {
		credCount++
		var cred webauthn.Credential // Create a new credential for each row.
		err := rows.Scan(&webAuthID, &webAuthnName, &webAuthnDispName,
			&credentials_id, &PublicKey, &Transport, &AuthenticatorData,
			&cred.AttestationType, &ClientDataHash, &ClientDataJSON,
			&cred.Attestation.PublicKeyAlgorithm, &Attachment, &AAGUID,
			&cred.Authenticator.SignCount, &CloneWarning, &BackupEligible, &BackupState,
			&UserPresent, &UserVerified)
		if err != nil {
			log.Printf("Error scanning credential row: %v", err)
			return user, err
		}
		//Fill the user and credential Info
		user.ID = []byte(webAuthID)
		// We already have the user's name from the profiles table, which is the source of truth.
		// We don't need to overwrite it with the per-credential display names.
		cred.ID, _ = base64StringToBytes(credentials_id)
		cred.PublicKey, _ = base64StringToBytes(PublicKey)
		Transports := strings.Split(Transport, ",")
		retrievedTransports := make([]protocol.AuthenticatorTransport, len(Transports))
		for i, transport := range Transports {
			retrievedTransports[i] = protocol.AuthenticatorTransport(transport)
		}
		cred.Transport = retrievedTransports
		cred.Attestation.AuthenticatorData, _ = base64StringToBytes(AuthenticatorData)
		cred.Attestation.ClientDataHash, _ = base64StringToBytes(ClientDataHash)
		cred.Attestation.ClientDataJSON = []byte(ClientDataJSON)
		if Attachment == "platform" {
			cred.Authenticator.Attachment = "platform"
		} else {
			cred.Authenticator.Attachment = "cross-platform"
		}
		cred.Authenticator.AAGUID, _ = base64StringToBytes(AAGUID)
		cred.Authenticator.CloneWarning = isTrue(CloneWarning)
		cred.Flags.BackupEligible = isTrue(BackupEligible)
		cred.Flags.BackupState = isTrue(BackupState)
		cred.Flags.UserPresent = isTrue(UserPresent)
		cred.Flags.UserVerified = isTrue(UserVerified)
		user.AddCredential(&cred)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error after iterating credential rows: %v", err)
		return user, err
	}

	log.Printf("Found %d credentials for authID: %s", credCount, c.authID)
	if credCount == 0 {
		log.Printf("No credentials found for user %s. This will trigger a discoverable credential login (QR code).", user.WebAuthnName())
	}

	return user, nil
}

// genAuthID: Creates a 64 character random string (authID) representing the user's ID
// Concept - save a bunch of random strings to a, 'available' database table,
// Remove any that are already in use. If the table gets low,
// add a bunch more before trying to fetch the next row.
func (c *credentialInfo) genAuthID() (string, error) {
	cid := ""
	cnt := 0
	// Count how many rows left
	query := "SELECT count(*) FROM cids"
	err := db.Conn.QueryRow(query).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	if cnt < 25 {
		c.addMoreCids()
	}
	// Get next available credentials id
	query = "SELECT cid FROM cids LIMIT 1"
	err = db.Conn.QueryRow(query).Scan(&cid)
	if err != nil {
		log.Println(err)
		return "", err
	}
	//remove it from the available list
	query = "DELETE FROM cids WHERE cid=?"
	_, err = db.Conn.Exec(query, cid)
	if err != nil {
		log.Println(err)
	}
	return cid, nil
}

func (c *credentialInfo) addMoreCids() {
	var cids []string
	for i := 0; i < 200; i++ {
		sid, err := genID(32)
		if err != nil {
			log.Println(err)
			continue
		}
		cids = append(cids, sid)
	}
	// Miracle of 1 insert for 200 values
	placeholders := make([]string, len(cids))
	values := make([]any, len(cids))
	for i, id := range cids {
		placeholders[i] = "(?)"
		values[i] = id
	}
	// Execute the SINGLE insert query
	query := fmt.Sprintf("INSERT INTO cids (cid) VALUES %s", strings.Join(placeholders, ","))
	_, err := db.Conn.Exec(query, values...)
	if err != nil {
		log.Println(err)
	}
	//Now remove any that are used already
	query = `
		DELETE FROM cids
		WHERE EXISTS (SELECT 1
			FROM credentials B
			WHERE cids.cid = B.credentials_id )
	`
	_, err = db.Conn.Exec(query)
	if err != nil {
		log.Println(err)
	}
}

// IsCredentials returns true/false for a given credential ID
func IsCredentials(credential_id string) bool {
	cnt := 0
	query := "SELECT count(*) FROM credentials WHERE credentials.credentials_id=?"
	db.Conn.QueryRow(query, credential_id).Scan(&cnt)
	return cnt > 0
}

func GetUserInfoFromCredentials(credentials_id string) (string, string, error) {
	user := ""
	fullname := ""
	query := `
		SELECT A.user, A.fullname FROM profiles A
		LEFT JOIN credentials B ON B.auth_id=A.auth_id
		WHERE B.credentials_id=?
	`
	err := db.Conn.QueryRow(query, credentials_id).Scan(&user, &fullname)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		return user, fullname, err
	}
	return user, fullname, nil
}
