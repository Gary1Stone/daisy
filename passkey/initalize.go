package passkey

import (
	"fmt"
	"log"
	"os"

	"github.com/go-webauthn/webauthn/webauthn"
	_ "github.com/mattn/go-sqlite3"
)

var webAuthn *webauthn.WebAuthn

func init() {
	var err error
	proto := os.Getenv("PROTO")
	host := os.Getenv("HOST")
	origin := fmt.Sprintf("%s://%s", proto, host)
	wconfig := &webauthn.Config{
		RPDisplayName: "Daisy",          // Display Name for your site
		RPID:          host,             // Generally the FQDN for your site
		RPOrigins:     []string{origin}, // The origin URLs allowed for WebAuthn
	}
	if webAuthn, err = webauthn.New(wconfig); err != nil {
		log.Printf("[FATA] %s", err)
	}
}
