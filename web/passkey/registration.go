package passkey

import (
	"errors"
	"log"

	"github.com/gbsto/daisy/db"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

// handleError is a helper function to standardize error responses.
// It logs the internal error and sends a JSON response with a user-friendly message.
func handleError(c *fiber.Ctx, status int, userMsg string, internalErr error) error {
	// We log the internal error for debugging, but don't expose it to the client.
	if internalErr != nil {
		log.Printf("Error: %s - %v", userMsg, internalErr)
	} else {
		log.Printf("Error: %s", userMsg)
	}
	return c.Status(status).JSON(fiber.Map{"msg": userMsg})
}

// Register for passkey credentials
func BeginRegistration(c *fiber.Ctx) error {

	// Get username and passcode information from JSON sent here
	usrInfo := struct {
		Username string `json:"username"`
		Passcode string `json:"passcode"`
		Apicode  string `json:"apicode"`
	}{}

	if err := c.BodyParser(&usrInfo); err != nil {
		return handleError(c, fiber.StatusBadRequest, "Invalid request format", err)
	}

	// Confirm the request is from an authorized source
	ip := c.IP()
	if ips := c.IPs(); len(ips) > 0 {
		ip = ips[0]
	}
	if !db.IsApiCode(usrInfo.Apicode, ip) {
		return handleError(c, fiber.StatusUnauthorized, "Unauthorized request", nil)
	}

	// Check if user/passcode valid
	var cInfo credentialInfo
	cInfo.username = usrInfo.Username
	cInfo.passcode = usrInfo.Passcode
	if !cInfo.isValidUser() {
		// Log the failed attempt without leaking the passcode.
		log.Printf("Invalid user or passcode for user: %s", usrInfo.Username)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "Invalid user or passcode"})
	}

	// get the authID from the user profile, creating/adding it if it does not exist
	authID, err := cInfo.getAuthID()
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not process user identity", err)
	}

	// Get user information/credentials
	cInfo.authID = authID
	user, err := cInfo.getCredentials(false) // Load user and any existing credentials
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not retrieve user credentials", err)
	}

	// Start registration by building the options
	// We want to create a resident key on a platform authenticator (like Windows Hello or Touch ID)
	// and require user verification (biometrics/PIN). This provides a better user experience for logins.
	authSelection := protocol.AuthenticatorSelection{
		AuthenticatorAttachment: protocol.Platform,
		ResidentKey:             protocol.ResidentKeyRequirementRequired, // Use the modern 'residentKey' requirement
		UserVerification:        protocol.VerificationRequired,           // Require user verification (e.g., biometrics or PIN)
	}

	// Debug check: ensure the global webAuthn object is actually configured
	if webAuthn == nil {
		return handleError(c, fiber.StatusInternalServerError, "WebAuthn engine not initialized", nil)
	}

	options, session, err := webAuthn.BeginRegistration(
		user,
		webauthn.WithAuthenticatorSelection(authSelection),
	)
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not begin registration ceremony", err)
	}

	// Make a session key and store the sessionData values
	var sInfo sessionInfo
	sessionID, err := sInfo.genSessionID()
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not generate session ID", err)
	}
	sInfo.sessionID = sessionID
	sInfo.data = session
	if err := sInfo.saveSession(); err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not save session", err)
	}

	// Set the cookie
	c.Cookie(&fiber.Cookie{
		Name:     "sid",
		Value:    sessionID,
		Path:     "/api/passkey",
		MaxAge:   900, // 15 mins to register between requesting code, and getting the email
		Secure:   true,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	// options.publicKey now contains our registration options
	// return the options generated and place the session key in a sid cookie
	return c.Status(fiber.StatusOK).JSON(&options)
}

// **********************************************************************
// Browser Prompt for user authentication happens between these two steps
// **********************************************************************

func FinishRegistration(c *fiber.Ctx) error {

	// Get the session ID from cookie
	sid := c.Cookies("sid")
	c.ClearCookie("sid")
	if len(sid) == 0 {
		return handleError(c, fiber.StatusBadRequest, "Missing session ID", nil)
	}

	// Get the session data stored from the function above
	var sInfo sessionInfo
	sInfo.sessionID = sid
	session, err := sInfo.getSession()
	if err != nil {
		return handleError(c, fiber.StatusBadRequest, "Invalid or expired session", err)
	}
	// Also check if the session data is actually populated
	if len(session.UserID) == 0 {
		return handleError(c, fiber.StatusBadRequest, "Invalid or expired session", errors.New("session data not found for sid"))
	}

	var cInfo credentialInfo
	cInfo.authID = string(session.UserID)
	user, err := cInfo.getCredentials(false)
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not find user", err)
	}

	// Convert go Fiber context to net:http request
	httpReq, err := adaptor.ConvertRequest(c, true)
	if err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not process request", err)
	}

	// Finish registration by building the credentials
	credential, err := webAuthn.FinishRegistration(user, session, httpReq)
	if err != nil {
		// This error is often due to client-side issues (e.g., wrong authenticator, timeout).
		return handleError(c, fiber.StatusBadRequest, "Failed to finish registration", err)
	}

	//Store the credential object
	user.AddCredential(credential)
	cInfo.user = user
	if err := cInfo.saveCredentials(); err != nil {
		return handleError(c, fiber.StatusInternalServerError, "Could not save new credential", err)
	}

	// The credential has been successfully saved, now we can clean up the session.
	sInfo.deleteSession()

	cid := bytesToBase64String(credential.ID)
	c.Cookie(&fiber.Cookie{
		Name:     "cid",
		Value:    cid,
		Path:     "/",
		MaxAge:   3600, // 1 hr to log in
		Secure:   true,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"msg": "Registration Success"})
}
