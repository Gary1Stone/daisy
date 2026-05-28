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
	// Redirect standard logger output to daisyLogger for centralized logging, including autocert messages
	log.SetOutput(io.MultiWriter(os.Stderr, daisyLogger))
	// certFile := "cert.pem"
	// keyFile := "key.pem"

	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	// get absolute paths
	// certFile = filepath.Join(workingDir, "certs", certFile)
	// keyFile = filepath.Join(workingDir, "certs", keyFile)
	publicDir := filepath.Join(workingDir, "web", "public")
	viewsDir := filepath.Join(workingDir, "web", "views")
	certCacheDir := filepath.Join(workingDir, "certs")

	// Ensure certsCacheDir directory exists
	if _, err := os.Stat(certCacheDir); os.IsNotExist(err) {
		err = os.MkdirAll(certCacheDir, 0700)
		if err != nil {
			log.Println("failed to create certs directory:", err)
			return
		}
	}

	// Initialize GoFiber html template engine
	engine := html.New(viewsDir, ".html")

	// Create GoFiber app
	app := fiber.New(fiber.Config{
		Views:              engine,
		ServerHeader:       "Daisy",
		AppName:            "Daisy App v2026.04.04",
		EnableIPValidation: true,
	})

	// Allow images up to 5MBytes to be uploaded, default is normally 4MB
	server := app.Server()
	server.MaxRequestBodySize = 5 * 1024 * 1024

	// Give external access to the public folder
	// where javascript, css, images,... are stored
	// app.Static("/", dir+"/public")
	app.Static("/", publicDir)
	app.Use(recover.New())

	// Move logger up so it captures all traffic and potential errors
	app.Use(logger.New(logger.Config{
		Output: daisyLogger,
	}))

	app.Use(addHitCounter())
	addProtection(app)

	// https: Certificate manager
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
	port := os.Getenv("PORT")
	if port == "" || len(port) < 2 || port[0] != ':' {
		port = ":8443" // Default to 8443
	}
	app.Use(SecureOnly(port))

	// Register all your specific application routes
	routes(app)

	log.Println("Daisy Web Server starting...")

	// Remember to open ports 8443 and 80 in the windows firewall
	// And open ports 587 and 465 for sending email as well
	// And set port forwarding up on your ISP modem/router/wifi
	ln, err := tls.Listen("tcp", port, cfg)
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
