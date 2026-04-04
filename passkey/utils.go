package passkey

import (
	"crypto/rand"
	"encoding/base64"
)

// Database Helper functions
// Generate a web-safe random character string of set length
func genID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// Convert int to boolean where 1=true, 0=false
func isTrue(val int) bool {
	return val == 1
}

// Convert []byte to base64 string using base64 encoding
func bytesToBase64String(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Convert base64 string back to []byte
func base64StringToBytes(base64Str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(base64Str)
}
