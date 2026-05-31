package passkey

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gbsto/daisy/db"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func BeginLogin(c *fiber.Ctx) error {

	usr := struct {
		Username string  `json:"username"`
		Tzoff    int     `json:"tzoff"`
		Lon      float64 `json:"lon"`
		Lat      float64 `json:"lat"`
		Timezone string  `json:"timezone"`
	}{}

	if err := c.BodyParser(&usr); err != nil {
		return handleError(c, fiber.StatusBadRequest, "Invalid request format", err)
	}

	// Ensure user is in the database
	var cInfo credentialInfo
	cInfo.purgeCredentials() // Remove old unwanted credentials (30 day limit, last 5)
	cInfo.username = usr.Username
	var err error
	uid, fullname, _, err := cInfo.getUid()
	if err != nil || uid == 0 {
		return handleError(c, fiber.StatusUnauthorized, "Unknown user", err)
	}

	// Get or Create 64 character random string as the user's authentication ID
	authID, err := cInfo.getAuthID()
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not get user's authentication ID", err)
	}

	// Get the user from the database
	log.Printf("Attempting to get credentials for authID: %s", authID)
	cInfo.authID = authID
	user, err := cInfo.getCredentials(false)
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not find user", err)
	}

	// Process the user's login information
	// We prefer user verification (biometrics/PIN) for logins.
	// The webAuthn.BeginLogin function will automatically populate the `allowCredentials`
	// option from the credentials stored in the `user` object. If the `user` object
	// has no credentials, it will initiate a discoverable credential login.
	options, session, err := webAuthn.BeginLogin(
		user,
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not begin login ceremony", err)
	}

	// Make a new session key and store the sessionData values
	// The login could occur days after registration,
	// so the cookie with the orignal SID may not be there
	// Ditto for the database
	var sInfo sessionInfo
	sessionID, err := sInfo.genSessionID()
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not generate session ID", err)
	}

	// Save the session (sessionID, *session)
	sInfo.sessionID = sessionID
	sInfo.data = session
	if err = sInfo.saveSession(); err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not save session", err)
	}

	// Session ID values
	sidCookieValue := struct {
		Sessionid string  `json:"sessionid"`
		Username  string  `json:"username"`
		Fullname  string  `json:"fullname"`
		Uid       int     `json:"uid"`
		Tzinfo    int     `json:"tzinfo"`
		Lon       float64 `json:"lon"`
		Lat       float64 `json:"lat"`
		Timezone  string  `json:"timezone"`
	}{
		Sessionid: sessionID,
		Username:  usr.Username,
		Fullname:  fullname,
		Uid:       uid,
		Tzinfo:    usr.Tzoff,
		Lon:       usr.Lon,
		Lat:       usr.Lat,
		Timezone:  usr.Timezone,
	}
	jsonByteSid, err := json.Marshal(sidCookieValue)
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not create session cookie", err)
	}

	// Set the cookie
	c.Cookie(&fiber.Cookie{
		Name:     "sid",
		Value:    string(jsonByteSid),
		Path:     "/api/passkey",
		MaxAge:   3600,
		Secure:   true,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	// return the options generated with the session key
	// options.publicKey contain the registration options
	return c.Status(fiber.StatusOK).JSON(options)
}

// **********************************************************************
// Browser Prompt for user authentication happens between these two steps
// **********************************************************************

func FinishLogin(c *fiber.Ctx) error {

	usr := struct {
		Sessionid string  `json:"sessionid"`
		Username  string  `json:"username"`
		Fullname  string  `json:"fullname"`
		Uid       int     `json:"uid"`
		Tzinfo    int     `json:"tzinfo"`
		Lon       float64 `json:"lon"`
		Lat       float64 `json:"lat"`
		Timezone  string  `json:"timezone"`
	}{}

	// Get the sessionID key from cookie
	jsonStr := c.Cookies("sid")
	err := json.Unmarshal([]byte(jsonStr), &usr)
	if err != nil || len(usr.Sessionid) == 0 {
		return handleError(c, fiber.StatusBadRequest, "Missing or invalid session ID", err)
	}

	// Get the session data stored from the BeginLogin function above
	var sInfo sessionInfo
	sInfo.sessionID = usr.Sessionid
	session, err := sInfo.getSession()
	if err != nil {
		return handleError(c, fiber.StatusBadRequest, "Invalid or expired session", err)
	}

	// Get the user with their credentials
	var cInfo credentialInfo
	cInfo.authID = string(session.UserID)
	user, err := cInfo.getCredentials(false)
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not find user", err)
	}

	// Convert goFiber context to net/http request object
	httpReq, err := adaptor.ConvertRequest(c, true)
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not process request", err)
	}
	// Explicitly set the URL scheme and host for WebAuthn origin validation.
	// Fiber's adaptor.ConvertRequest might not fully populate http.Request.URL
	// in a way that webauthn expects, especially with non-standard ports or proxies.
	// We use the configured RPOrigins to ensure consistency.
	if webAuthn != nil && len(webAuthn.Config.RPOrigins) > 0 {
		parsedOrigin, parseErr := url.Parse(webAuthn.Config.RPOrigins[0])
		if parseErr == nil {
			httpReq.URL.Scheme = parsedOrigin.Scheme
			httpReq.URL.Host = parsedOrigin.Host
			httpReq.Host = parsedOrigin.Host // Also set the Host header for good measure
		} else {
			log.Printf("WARNING: Could not parse RPOrigin '%s': %v", webAuthn.Config.RPOrigins[0], parseErr)
		}
	}

	//Finish login
	credential, err := webAuthn.FinishLogin(user, session, httpReq)
	if err != nil {
		return handleError(c, fiber.StatusBadRequest, "Failed to finish login", err)
	}

	// Registration successful, now clear the session cookie.
	// Must specify the path used when the cookie was created.
	c.ClearCookie("sid", "/api/passkey")

	// Handle possible cloned authenticator,
	// by removing both cookies and credentials/session in the database
	if credential.Authenticator.CloneWarning {
		cInfo.credentialID = bytesToBase64String(credential.ID)
		cInfo.deleteCredentials()
		sInfo.deleteSession()
		return handleError(c, fiber.StatusConflict, "Security issue: cloned authenticator detected", errors.New("cloned authenticator"))
	}

	// If login was successful, update the credential object
	user.UpdateCredential(credential)
	cInfo.user = user
	if err := cInfo.saveCredentials(); err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not update credential", err)
	}

	// Remove session info from database
	sInfo.sessionID = usr.Sessionid
	sInfo.deleteSession()

	// Create JWT cookie
	ip := c.IP()
	ips := c.IPs()
	if len(ips) > 0 {
		ip = ips[0]
	}
	var loginInfo db.Logins
	loginInfo.Uid = usr.Uid
	loginInfo.User = usr.Username
	loginInfo.Fullname = usr.Fullname
	loginInfo.Tzoff = usr.Tzinfo
	loginInfo.Longitude = usr.Lon
	loginInfo.Latitude = usr.Lat
	loginInfo.Ip = ip
	loginInfo.Session, _ = genID(32)
	loginInfo.Success = 1
	loginInfo.Timezone = usr.Timezone
	loginInfo.Credential_id = bytesToBase64String(credential.ID)
	token, _, err := CreateJWTToken(loginInfo)
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not create authentication token", err)
	}

	// Set the Auth cookie
	cookieName := os.Getenv("JWT")
	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 60), // 60 day JWT expiry
		Secure:   true,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	// save login info
	err = db.SaveLogin(&loginInfo)
	if err != nil {
		log.Println(err)
	}

	// reply to the user
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"msg": "Login Success"})
}
