package middleware

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/web/passkey"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/helmet/v2"
)

func AddProtection(app *fiber.App) {
	app.Use(helmet.New()) // Set some security headers on all responses

	// Prevent frequent pings (DoS attacks)
	app.Use(limiter.New(limiter.Config{
		Max:               120,
		Expiration:        10 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))

	// Prevent cross-site reference        PREVENTS AJAX CALLS FROM WORKING BECAUSE NO COOKIE IS RETURNED
	// app.Use(csrf.New(csrf.Config{
	// 	KeyLookup:      "header:X-Csrf-Token", // You can also use "form:csrf" for form submissions
	// 	CookieName:     "csrf_token",
	// 	Expiration:     36000,
	// 	CookieSecure:   true, // Set it to true if using HTTPS
	// 	CookieSameSite: "Lax",
	// 	CookieHTTPOnly: true,
	// 	CookieDomain:   "daisy.hopto.org",
	// }))

	// Confirm its not a banned IP address
	app.Use(func(c *fiber.Ctx) error {
		ip := c.IP()
		ips := c.IPs()
		if len(ips) > 0 {
			ip = ips[0]
		}

		if db.IsBanned(ip) {
			log.Printf("The user at %v is banned\n", ip)
			return c.Status(fiber.StatusForbidden).Redirect("banned.html")
		}
		return c.Next()
	})
}

// Check the JSON Web Token (JWT) to ensure it is valid
func CheckToken(c *fiber.Ctx) error {
	// Get the cookie off the request
	cookieName := os.Getenv("JWT")
	tokenString := c.Cookies(cookieName)
	if len(tokenString) == 0 {
		log.Println("tokenString has zero length")
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}

	jwtInfo, expired, err := passkey.DecodeJwtToken(tokenString)
	if err != nil || expired {
		log.Println("jwtInfo has expired or an error", err)
		c.ClearCookie(cookieName)
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}

	if expired || jwtInfo.Uid < 1 || len(jwtInfo.Session) < 6 || len(jwtInfo.Ip) < 6 {
		log.Println("jwtInfo has expired or userid=0 or session id to small, or ip to small")
		c.ClearCookie(cookieName)
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}

	oldSession, oldIP, err := db.GetLastSessionByUid(jwtInfo.Uid)
	if err != nil {
		log.Println(err)
	}

	if oldSession != jwtInfo.Session || oldIP != jwtInfo.Ip {
		log.Println("Session terminated")
		c.ClearCookie(cookieName)
		return c.Status(fiber.StatusOK).Redirect("index.html")
	}

	// Attach to the request
	c.Locals("curUid", fmt.Sprintf("%v", jwtInfo.Uid))
	c.Locals("user", fmt.Sprintf("%v", jwtInfo.User))
	c.Locals("fullname", fmt.Sprintf("%v", jwtInfo.Fullname))
	c.Locals("permissions", fmt.Sprintf("%v", jwtInfo.Permissions))
	c.Locals("timezone", fmt.Sprintf("%v", jwtInfo.Timezone))
	c.Locals("tzoff", fmt.Sprintf("%v", jwtInfo.Tzoff))
	return c.Next()
}
