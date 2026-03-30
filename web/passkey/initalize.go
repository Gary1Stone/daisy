package passkey

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/go-webauthn/webauthn/webauthn"
)

var webAuthn *webauthn.WebAuthn

func StartWebAuthn() {
	var err error
	proto := os.Getenv("PROTO")
	hostPort := os.Getenv("HOST")

	// The RPID MUST be a valid domain string and MUST NOT include a scheme or port.
	// We extract the host from the HOST environment variable in case it contains a port.
	rpID, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		rpID = hostPort // Fallback if no port is present (e.g., "example.com")
	}

	origin := fmt.Sprintf("%s://%s", proto, hostPort)

	wconfig := &webauthn.Config{
		RPDisplayName: "Daisy",          // Display Name for your site
		RPID:          rpID,             // Must be the domain only (e.g., "localhost" or "example.com")
		RPOrigins:     []string{origin}, // The origin URLs allowed for WebAuthn
	}

	if webAuthn, err = webauthn.New(wconfig); err != nil {
		log.Fatalf("Failed to create WebAuthn instance: %v (RPID: %s, Origin: %s)", err, rpID, origin)
	}
}
