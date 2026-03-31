package cmd

import (
	"log"
	"os"

	"github.com/gbsto/daisy/web/passkey"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

// Check the Auth cookie to see if we can get the user details
func GetIndex(c *fiber.Ctx) error {
	var jwtInfo db.Logins
	expired := false
	var err error
	nextpage := "registration.html"
	cookieName := os.Getenv("JWT")
	tokenString := c.Cookies(cookieName) // the long lifespan cookie
	if len(tokenString) > 0 {
		jwtInfo, expired, err = passkey.DecodeJwtToken(tokenString)
		if err != nil {
			log.Println(err)
		}
		// If no user, show register. if has creds and expired, show login
		if len(jwtInfo.User) > 0 && passkey.IsCredentials(jwtInfo.Credential_id) {
			if expired {
				nextpage = "login.html"
			} else {
				nextpage = "home.html"
			}
		}
	}

	return c.Render("index2", fiber.Map{
		"username": jwtInfo.Fullname,
		"nextpage": nextpage,
	})
}
