package passkey

import (
	"log"
	"os"
	"strconv"

	"github.com/gbsto/daisy/emails"
	"github.com/gbsto/daisy/register"

	"github.com/gofiber/fiber/v2"
)

// Determines if a passcode (2FA) is needed for registration, or just a login
func RequestCode(c *fiber.Ctx) error {
	reply := struct {
		Msg string `json:"msg"`
	}{
		Msg: "goLogin",
	}

	usr := struct {
		Username string `json:"username"`
		Apicode  string `json:"apicode"`
	}{}

	if err := c.BodyParser(&usr); err != nil {
		reply.Msg = "can't get user name"
		log.Println(reply.Msg, err)
		return c.Status(fiber.StatusBadRequest).JSON(reply)
	}

	//Confirm the request is from an authorized source
	ip := c.IP()
	ips := c.IPs()
	if len(ips) > 0 {
		ip = ips[0]
	}
	var r register.RegisterCode
	r.Apicode = usr.Apicode
	r.Ip = ip
	if !r.IsApiCode() {
		reply.Msg = "Unknown user"
		return c.Status(fiber.StatusBadRequest).JSON(reply)
	}

	// Confirm the user has credentials in the database
	var cInfo credentialInfo
	cInfo.username = usr.Username
	uid, _, mins, err := cInfo.getUid() // mins = minutes since last updated
	howOften, err := strconv.Atoi(os.Getenv("EMAILFREQUENCY"))
	if err != nil {
		howOften = 15
	}
	if err != nil || uid == 0 || mins < howOften {
		reply.Msg = "Unknown user"
		if mins < howOften && uid > 0 {
			reply.Msg = "Only one request every 15 minutes please"
		}
		log.Println(reply.Msg, usr.Username, err)
		return c.Status(fiber.StatusBadRequest).JSON(reply)
	}

	// If user already has credentials, from previous login
	// they only need to log in, not register
	isDirectlogin := false
	cookieName := os.Getenv("JWT")
	tokenString := c.Cookies(cookieName) // the long lifespan (30 day) cookie
	if tokenString != "" {
		jwtInfo, expired, err := DecodeJwtToken(tokenString) // Ignore token's expiry, its set to 7 days
		isDirectlogin = err == nil && jwtInfo.User != "" && IsCredentials(jwtInfo.Credential_id)
		if !expired && isDirectlogin {
			reply.Msg = "goHome"
		}
	}

	if !isDirectlogin {
		reply.Msg = "passcode emailed"
		var e emails.Emails
		if err := e.SendOneTimePassword(uid); err != nil {
			log.Println(err)
			reply.Msg = "passcode email error"
		}
	}

	return c.Status(fiber.StatusOK).JSON(reply)
}
