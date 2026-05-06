package cmd

import (
	"errors"
	"html/template"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gbsto/daisy/colors"
	"github.com/gbsto/daisy/ctrls"
	"github.com/gbsto/daisy/passkey"
	"github.com/gbsto/daisy/svg"
	"github.com/gbsto/daisy/util"

	"github.com/gbsto/daisy/db"

	"github.com/gofiber/fiber/v2"
)

type userInfo struct {
	Uid         int
	Fullname    string
	Email       string
	Permissions util.Permissions
	IsAdmin     bool
	Timezone    string
	Tzoff       int
}

func GetHome(c *fiber.Ctx) error {
	user, err := extractUserInfo(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Redirect("index.html")
	}

	// I tried using go routines for the following, but the overhead causes it to be 5ms slower.
	// Build the last login message: when, from where and how far
	msg, err := getLastLoginMsg(user.Uid)
	if err != nil {
		msg = ""
	}

	// Generate list of computers/devices assigned to this user
	assigned, err := getAssignedDevices(user.Uid)
	if err != nil {
		assigned = ""
	}

	// Count the number of pending tickets this person has
	btnColour := colors.Secondary
	btnLabel := "Tickets"
	cnt := db.CountPendingTickets(user.Uid)
	if cnt > 0 {
		btnColour = colors.Warning
		btnLabel = strconv.Itoa(cnt) + " Ticket"
		if cnt > 1 {
			btnLabel += "s"
		}
	}

	return c.Render("home", fiber.Map{
		"title":        template.HTML("&#127809; Daisy"),
		"fullName":     user.Fullname,
		"isAdmin":      user.IsAdmin,
		"lastLoginMsg": template.HTML(msg),
		"firstName":    user.Fullname[:strings.Index(user.Fullname, " ")],
		"assigned":     template.HTML(assigned),
		"btnColour":    btnColour,
		"btnLabel":     btnLabel,
		"myAlerts":     template.HTML(ctrls.GetAlertButtons(user.Uid)),
		"menu":         template.HTML(svg.GetIcon("menu")),
		"home":         template.HTML(svg.GetIcon("home")),
		"ticket":       template.HTML(svg.GetIcon("ticket")),
		"devices":      template.HTML(svg.GetIcon("devices")),
		"software":     template.HTML(svg.GetIcon("software")),
		"profiles":     template.HTML(svg.GetIcon("profiles")),
		"reports":      template.HTML(svg.GetIcon("reports")),
		"control":      template.HTML(svg.GetIcon("control")),
		"network":      template.HTML(svg.GetIcon("network")),
		"settings":     template.HTML(svg.GetIcon("settings")),
		"about":        template.HTML(svg.GetIcon("about")),
		"logout":       template.HTML(svg.GetIcon("logout")),
		"user":         template.HTML(svg.GetIcon("user")),
		"wizard":       template.HTML(svg.GetIcon("wizard")),
		"bell":         template.HTML(svg.GetIcon("bell")),
		"eye":          template.HTML(svg.GetIcon("eye")),
		"check":        template.HTML(svg.GetIcon("check")),
		"tag":          template.HTML(svg.GetIcon("tag")),
		"wrench":       template.HTML(svg.GetIcon("wrench")),
		"broken":       template.HTML(svg.GetIcon("broken")),
		"stethoscope":  template.HTML(svg.GetIcon("stethoscope")),
		"footprint":    template.HTML(svg.GetIcon("footprint")),
		"copy":         template.HTML(svg.GetIcon("copy")),
	})
}

// Return a list of devices this person is assigned
func getAssignedDevices(curUid int) (string, error) {
	var msg strings.Builder
	items, err := db.GetAssignedDevices(curUid, curUid)
	if err != nil {
		log.Println("No assigned devices")
		return "", err
	}

	missing := db.GetMissingDevices()
	var i int = 0
	var cnt int = 0

	for _, item := range items {
		msg.WriteString("<p>")
		msg.WriteString(item.Name)
		msg.WriteString(" ")
		msg.WriteString(item.Model)
		msg.WriteString(" ")
		i = sort.SearchInts(missing, item.Cid)
		if i < len(missing) && item.Cid == missing[i] {
			msg.WriteString("<span class='mif-search mif-1x fg-red' title='Not seen recently'></span>")
		}
		msg.WriteString("</p>")
		cnt++
	}
	//if more than 4 items in the list, then add scroll bar
	var str = msg.String()
	if cnt > 4 {
		str = "<div style='overflow-y: scroll; height: 12rem; '>" + str + "</div>"
	}
	return str, err
}

// Build the last Login Message: when, from where and how far
func getLastLoginMsg(curUid int) (string, error) {
	item, err := db.GetLastLogin(curUid)
	if err != nil {
		log.Println(err)
	}

	if len(item.Timestamp) == 0 {
		return "<p>Welcome</p>", nil
	}
	var msg strings.Builder
	msg.WriteString("<p>Your last access was:</p><p>")
	msg.WriteString(item.Weekday)
	msg.WriteString(" ")
	msg.WriteString(item.Timestamp)
	msg.WriteString("</p><p>From ")

	if len(item.City) > 0 {
		msg.WriteString(item.City)
		if len(item.State) > 0 {
			msg.WriteString(" ")
			msg.WriteString(item.State)
		}
		if len(item.Country) > 0 {
			msg.WriteString(" ")
			msg.WriteString(item.Country)
		}
	} else {
		msg.WriteString(item.Ip)
	}
	msg.WriteString("</p><p>Roughly ")
	msg.WriteString(strconv.Itoa(item.Distance))
	msg.WriteString(" Km from ")
	msg.WriteString(item.Home)
	msg.WriteString("</p>")
	return msg.String(), nil
}

// ExtractUserInfo extracts Uid, fullname, email, and permissions from the context locals.
// It must have been processed through the middware func CheckToken() to set the c.Locals values!
func extractUserInfo(c *fiber.Ctx) (userInfo, error) {
	var user userInfo

	// Extract Uid
	curUidStr, ok := c.Locals("curUid").(string)
	if !ok || curUidStr == "" {
		return userInfo{}, errors.New("missing or invalid curUid in context locals")
	}
	curUid, err := strconv.Atoi(curUidStr)
	if err != nil || curUid < 1 {
		return userInfo{}, errors.New("curUid is not a valid positive integer")
	}
	user.Uid = curUid

	// Extract fullname
	fullname, ok := c.Locals("fullname").(string)
	if !ok || fullname == "" {
		return userInfo{}, errors.New("missing or invalid fullname in context locals")
	}
	user.Fullname = fullname

	// Extract permissions
	permissionsStr, ok := c.Locals("permissions").(string)
	if !ok || permissionsStr == "" {
		return userInfo{}, errors.New("missing or invalid permissions in context locals")
	}
	user.Permissions.GetPermissions(permissionsStr)
	user.IsAdmin = user.Permissions.Admin.Read

	// Extract email
	user.Email, ok = c.Locals("user").(string)
	if !ok || user.Email == "" {
		return userInfo{}, errors.New("missing or invalid email in context locals")
	}

	// Extract timezone
	user.Timezone, ok = c.Locals("timezone").(string)
	if !ok || user.Timezone == "" {
		return userInfo{}, errors.New("missing or invalid timezone in context locals")
	}

	// Extract timezone offset
	tzoffStr, ok := c.Locals("tzoff").(string)
	if !ok || tzoffStr == "" {
		return userInfo{}, errors.New("missing or invalid tzoff in context locals")
	}
	user.Tzoff, err = strconv.Atoi(tzoffStr)
	if err != nil {
		return userInfo{}, errors.New("tzoff is not a valid integer")
	}

	return user, nil
}

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
