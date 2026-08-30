package webserver

import (
	"crypto/tls"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gbsto/daisy/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/natefinch/lumberjack.v2"
)

func StartServer(daisyLogger *lumberjack.Logger) {
	// Constants
	domain := os.Getenv("HOST")
	httpsPort := os.Getenv("PORT")

	// Logging
	log.SetOutput(io.MultiWriter(os.Stderr, daisyLogger))

	// Working Directories
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}
	certDir := filepath.Join(workingDir, "certs")           // ./certs
	publicDir := filepath.Join(workingDir, "web", "public") // ./web/public
	viewsDir := filepath.Join(workingDir, "web", "views")   // ./web/views

	// Ensure certsCacheDir directory exists
	if _, err := os.Stat(certDir); os.IsNotExist(err) {
		err = os.MkdirAll(certDir, 0700)
		if err != nil {
			log.Println("failed to create certs directory:", err)
			return
		}
	} else if err != nil {
		log.Printf("failed to inspect certificate cache %q: %v", certDir, err)
	}

	// Initialize GoFiber html template engine
	engine := html.New(viewsDir, ".html")

	// Create GoFiber app
	app := fiber.New(fiber.Config{
		Views:              engine,
		ServerHeader:       "Daisy",
		AppName:            "Daisy App v2026.08.28",
		EnableIPValidation: true,
	})
	server := app.Server()
	server.MaxRequestBodySize = 5 * 1024 * 1024 // Allow images up to 5MBytes to be uploaded, default is normally 4MB

	// Give external access to the public folder
	app.Static("/", publicDir)
	app.Use(recover.New())

	// Logger captures all traffic and potential errors
	app.Use(logger.New(logger.Config{
		Output: daisyLogger,
	}))

	app.Use(addHitCounter())
	addProtection(app)

	// https: Certificate manager
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache(certDir),
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
	app.Use(SecureOnly(httpsPort))

	// Register all your specific application routes
	routes(app)

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

	// Logger captures all traffic and potential errors
	app.Use(logger.New(logger.Config{
		Output: daisyLogger,
	}))

	// Start server on HTTPS port 443
	// Remember to open ports 443 and 80 in the windows firewall
	// And open ports 587 and 465 for sending email as well
	// And set port forwarding up on your ISP modem/router/wifi
	ln, err := tls.Listen("tcp", httpsPort, cfg)
	if err != nil {
		panic(err)
	}

	// Start server
	defer db.Conn.Close()
	log.Fatal(app.Listener(ln))
}

// ATTACKS: Adding the catch-all middleware AFTER routes()
// meaning if user asks for a page that does not exist, kick them out.
// app.Use(func(c *fiber.Ctx) error {
// 	// Determine the originator's IP address, even through multiple proxies
// 	ip := c.IP()
// 	ips := c.IPs() // If multiple IPs, use the first one
// 	if len(ips) > 0 {
// 		ip = ips[0]
// 	}
// 	// Record the attack
// 	db.RecordAttack(ip, c.Method(), c.Path(), c.Get("User-Agent"))
// 	// Set the status code to 404 Not Found
// 	c.Status(fiber.StatusNotFound)
// 	return c.Render("404", fiber.Map{ // HTML template is named "404.html"
// 		"Path": c.Path(),
// 	})
// })
