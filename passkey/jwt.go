package passkey

import (
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gbsto/daisy/db"

	"github.com/golang-jwt/jwt/v5"
)

// Build the JSON Web Token and have it expire in 60 days
// When saving the cookie, it is set to expire in 60 days
// Which gives several weeks to just log in to reset the token
// Otherwise the user has to re-register with emailed passcode
func CreateJWTToken(jwtInfo db.Logins) (string, time.Time, error) {
	permissions, err := db.GetUidPermissions(jwtInfo.Uid)
	if err != nil {
		log.Println("cannot get permissions", err)
	}
	exp := time.Now().Add(time.Hour * 24 * 60)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = exp.UTC().Unix()
	claims["curUid"] = jwtInfo.Uid
	claims["user"] = jwtInfo.User
	claims["ip"] = jwtInfo.Ip // We kick out user if their IP changed.
	claims["session"] = jwtInfo.Session
	claims["fullname"] = jwtInfo.Fullname
	claims["permissions"] = permissions
	claims["credential_id"] = jwtInfo.Credential_id
	claims["timezone"] = jwtInfo.Timezone
	claims["tzoff"] = jwtInfo.Tzoff
	jwt, err := token.SignedString(getSecret())
	if err != nil {
		log.Println("error in signed string", err)
		return "", exp, err
	}
	return jwt, exp, nil
}

// Check the JSON Web Token (JWT) to ensure it is valid
func DecodeJwtToken(tokenString string) (db.Logins, bool, error) {
	var jwtInfo db.Logins
	expired := true
	jwtInfo.Uid = 0
	if len(tokenString) == 0 {
		return jwtInfo, expired, errors.New("no cookie")
	}

	//Decode/validate the cookie
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		//Check ALGorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid jwt signing algorithm")
		}
		return getSecret(), nil
	})
	if err != nil {
		return jwtInfo, expired, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if len(claims) < 9 {
			return jwtInfo, expired, err
		}
		//Check the expiration
		if float64(time.Now().UTC().Unix()) > claims["exp"].(float64) {
			return jwtInfo, expired, errors.New("session expired")
		}
		expired = false
		//Read the claims
		jwtInfo.Uid = int(claims["curUid"].(float64))
		jwtInfo.User = claims["user"].(string)
		jwtInfo.Fullname = claims["fullname"].(string)
		jwtInfo.Credential_id = claims["credential_id"].(string)
		jwtInfo.Ip = claims["ip"].(string)
		jwtInfo.Session = claims["session"].(string)
		jwtInfo.Permissions = claims["permissions"].(string)
		jwtInfo.Timezone = claims["timezone"].(string)
		jwtInfo.Tzoff = int(claims["tzoff"].(float64))
	}
	return jwtInfo, expired, nil
}

// getSecret retrieves the session secret key value and caches it
// For speed, ensures the value is only set once with only one OS call
var (
	secret     []byte
	secretOnce sync.Once
)

func getSecret() []byte {
	secretOnce.Do(func() {
		secret = []byte(os.Getenv("SECRET"))
	})
	return secret
}
