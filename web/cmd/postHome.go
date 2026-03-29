package cmd

import (
	"log"
	"os"

	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/web/passkey"

	"github.com/gofiber/fiber/v2"
)

func PostHome(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	var recvd struct {
		Task     string  `json:"task"`
		Aid      int     `json:"aid"`
		Tzoff    int     `json:"tzoff"`    // Timezone offset in seconds
		Lon      float64 `json:"lon"`      // Longitude
		Lat      float64 `json:"lat"`      // Latitude
		Ip       string  `json:"ip"`       // User IP address. May not be there
		Sec      int64   `json:"sec"`      // Seconds since epoch
		Err      bool    `json:"err"`      // Was there an error getting the lon/lat
		Source   int     `json:"source"`   // 0=Browser, 1=GeoLoc, 2=Extreme
		Uid      int     `json:"uid"`      // Other person's uid to ack
		Timezone string  `json:"timezone"` // Timezone name of the user
	}

	if err := c.BodyParser(&recvd); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusOK).SendString("CRITICAL SERVER ERROR!")
	}
	var response string
	switch recvd.Task {
	case "get_alerts": // NOTE: this is used by HOME and PROFILE pages
		if recvd.Uid == 0 {
			recvd.Uid = user.Uid
		}
		if recvd.Aid > 0 {
			db.AckAlert(user.Uid, recvd.Uid, recvd.Aid, true, true, true)
			db.AckAction(user.Uid, recvd.Aid, true, true, true)
		}
		response = ctrls.GetAlertButtons(recvd.Uid)
	case "save_lon_lat":
		ip := c.IP()
		ips := c.IPs()
		if len(ips) > 0 {
			ip = ips[0]
		}
		cookieName := os.Getenv("JWT")
		tokenString := c.Cookies(cookieName) // Get the JWT cookie
		if len(tokenString) > 0 {
			loginInfo, _, err := passkey.DecodeJwtToken(tokenString)
			if err != nil {
				log.Println(err)
			}
			loginInfo.Longitude = recvd.Lon
			loginInfo.Latitude = recvd.Lat
			loginInfo.Tzoff = recvd.Tzoff
			loginInfo.Ip = ip
			loginInfo.Success = 1
			loginInfo.Timezone = recvd.Timezone
			err = db.SaveLogin(&loginInfo)
			if err != nil {
				log.Println(err)
			}
		}
		response = "ok"
	}
	return c.Status(fiber.StatusOK).SendString(response)
}
