package webserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func BellTest() {
	certFile := "cert.pem"
	keyFile := "key.pem"

	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	// get absolute paths
	certFile = filepath.Join(workingDir, "certs", certFile)
	keyFile = filepath.Join(workingDir, "certs", keyFile)
	publicDir := filepath.Join(workingDir, "public")

	// Ensure certs directory exists
	if err := os.MkdirAll(filepath.Join(workingDir, "certs"), 0755); err != nil {
		log.Fatalf("failed to create certs directory: %v", err)
	}

	ports := []string{"80", "8443", "8080", "443"}
	if len(os.Args) > 1 {
		ports = os.Args[1:]
	}

	m := &autocert.Manager{
		Cache:      autocert.DirCache("certs"), // Store certificates in "certs" folder
		Prompt:     autocert.AcceptTOS,         // Automatically accept Let's Encrypt TOS
		HostPolicy: autocert.HostWhitelist("daisy.hopto.org"),
	}

	handler := http.NewServeMux()
	// Serve the static index.html web page for the root path
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(publicDir, "index.html"))
	})

	servers := make([]*http.Server, 0, len(ports))
	errCh := make(chan error, len(ports))
	var wg sync.WaitGroup

	for _, port := range ports {
		srv := &http.Server{
			Addr:              ":" + port,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		}

		// Handle Let's Encrypt HTTP-01 challenges on port 80
		if port == "80" {
			srv.Handler = m.HTTPHandler(handler)
		}

		// Configure TLS if the port is for HTTPS
		if port == "443" || port == "8443" {
			srv.TLSConfig = m.TLSConfig()
		}

		servers = append(servers, srv)
		wg.Add(1)
		go func(s *http.Server, p string) {
			defer wg.Done()

			var err error
			if p == "443" || p == "8443" {
				log.Printf("listening on %s (HTTPS)", s.Addr)
				err = s.ListenAndServeTLS("", "") // No local cert files needed
			} else {
				log.Printf("listening on %s (HTTP)", s.Addr)
				err = s.ListenAndServe()
			}

			if err != nil && err != http.ErrServerClosed {
				errCh <- fmt.Errorf("server %s: %w", s.Addr, err)
			}
		}(srv, port)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	select {
	case err := <-errCh:
		log.Fatalf("server error: %v", err)
	case sig := <-stop:
		log.Printf("shutdown signal: %s", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error on %s: %v", srv.Addr, err)
		}
	}

	wg.Wait()
	log.Println("shutdown complete")
}
