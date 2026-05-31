package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/passkey"

	"github.com/gofiber/fiber/v2"
)

// There may be a long lived jwt token if the user signed in previously (30 day token)
// There may be a short lived cid token if the user registered (1 hr token)
// If neither, then the user has start over
func GetLogin(c *fiber.Ctx) error {
	var usrInfo db.Logins
	usrInfo.User = ""
	usrInfo.Fullname = ""

	cid := c.Cookies("cid")
	if len(cid) > 0 {
		usrInfo.User, usrInfo.Fullname, _ = passkey.GetUserInfoFromCredentials(cid)
	}

	if len(usrInfo.User) == 0 {
		cookieName := os.Getenv("JWT")
		tokenString := c.Cookies(cookieName) // the long lifespan cookie
		if len(tokenString) > 0 {
			usrInfo, _, _ = passkey.DecodeJwtToken(tokenString)
		}
	}

	return c.Render("login", fiber.Map{
		"fullname": usrInfo.Fullname,
		"user":     usrInfo.User,
	})
}

// Logout and then serve the user the index page
func GetLogout(c *fiber.Ctx) error {
	curUid, err := strconv.Atoi(fmt.Sprint(c.Locals("curUid")))
	if err != nil {
		curUid = 0
	}
	db.EndSession(curUid)
	cookieName := os.Getenv("JWT")
	c.ClearCookie(cookieName)

	return c.Render("index", fiber.Map{
		"username": "",
		"nextpage": "registration.html",
	})
}
