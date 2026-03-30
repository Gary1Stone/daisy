package web

import (
	"crypto/tls"
	"log"
	"os"
	"path/filepath"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/util"
	"github.com/gbsto/daisy/web/middleware"
	"github.com/gbsto/daisy/web/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"

	"golang.org/x/crypto/acme/autocert"
)

func StartServer() {

	// Get the current directory's full path
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalf("ERROR: Cannot find the current working directory")
	}

	// Initialize GoFiber html template engine
	viewsDir :=filepath.Join(dir, "web", "views")
	engine := html.New(viewsDir, ".html")

	// Create GoFiber app
	app := fiber.New(fiber.Config{
		Views:              engine,
		ServerHeader:       "Daisy",
		AppName:            "Daisy App v2026.03.29",
		EnableIPValidation: true,
	})

	// Allow images up to 5MBytes to be uploaded, default is normally 4MB
	server := app.Server()
	server.MaxRequestBodySize = 5 * 1024 * 1024


	// Give external access to the public folder
	// where javascript, css, images,... are stored
	// app.Static("/", dir+"/public")
	app.Static("/", filepath.Join(dir, "web", "public"))
	app.Use(recover.New())
	app.Use(middleware.AddHitCounter())
	middleware.AddProtection(app)

	// https: Certificate manager
	certCacheDir := filepath.Join(dir, "web", "certs")
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("daisy.hopto.org"),
		Cache:      autocert.DirCache(certCacheDir),
	}

	// TLS Config
	// Get Certificate from Let's Encrypt
	cfg := &tls.Config{
		GetCertificate: m.GetCertificate,
		NextProtos: []string{
			"http/1.1", "acme-tls/1",
		},
	}

	// Middleware to enforce HTTPS
	app.Use(middleware.SecureOnly())

	// Register all your specific application routes
	routes.Routes(app)

	// ATTACKS: Adding the catch-all middleware AFTER routes.Routes()
	// meaning if user asks for a page that does not exist, kick them out.
	app.Use(func(c *fiber.Ctx) error {
		// Determine the originator's IP address, even through multiple proxies
		ip := c.IP()
		ips := c.IPs() // If multiple IPs, use the first one
		if len(ips) > 0 {
			ip = ips[0]
		}
		// Record the attack
		db.RecordAttack(ip, c.Method(), c.Path(), c.Get("User-Agent"))
		// Set the status code to 404 Not Found
		c.Status(fiber.StatusNotFound)
		return c.Render("404", fiber.Map{ // HTML template is named "404.html"
			"Path": c.Path(),
		})
	})

	//Set up error logging directory
	logFile := util.CheckLogsDirectoryExists()
	defer logFile.Close()

	app.Use(logger.New(logger.Config{
		Output: logFile,
	}))

	log.Println("SQLite Version:", db.GetSqlVersion())
	log.Println("Daisy Web Server starting...")

	// Start server on HTTPS port 443
	// Remember to open ports 443 and 80 in the windows firewall
	// And open ports 587 and 465 for sending email as well
	// And set port forwarding up on your ISP modem/router/wifi
	ln, err := tls.Listen("tcp", ":443", cfg)
	if err != nil {
		panic(err)
	}

	// Start server
	defer db.Conn.Close()
	log.Fatal(app.Listener(ln))
}
