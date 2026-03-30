package util

import (
	"log"
	"math/rand"
	"strconv"
	"time"
)

// Return a substring
func Substr(input string, start int, length int) string {
	inputLength := len(input)
	if start < 0 || length <= 0 || inputLength == 0 || start > inputLength {
		return ""
	}
	if start+length > inputLength {
		length = inputLength - start
	}
	asRunes := []rune(input[start : start+length])
	return string(asRunes)
}

// Convert yyyy-mm-dd to Unix timestamp in seconds UTC since Jan 1, 1970
func ToUnixSeconds(yyyymmdd string) int64 {
	t, err := time.Parse("2006-01-02", yyyymmdd)
	if err != nil {
		log.Println(err)
		return time.Now().UTC().Unix()
	}
	return t.Unix()
}

// Builds a random password containg upper/lower/numbers no symbols
func GetRandomPassword() string {
	const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const lower = "abcdefghijklmnopqrstuvwxyz"
	const numbers = "0123456789"
	pwd := make([]byte, 10)
	rnd := rand.New(rand.NewSource(int64(time.Now().UTC().Unix())))
	pwd[0] = upper[rnd.Intn(len(upper))]
	pwd[1] = lower[rnd.Intn(len(lower))]
	pwd[2] = numbers[rnd.Intn(len(numbers))]
	pwd[3] = upper[rnd.Intn(len(upper))]
	pwd[4] = lower[rnd.Intn(len(lower))]
	pwd[5] = upper[rnd.Intn(len(upper))]
	pwd[6] = upper[rnd.Intn(len(upper))]
	pwd[7] = numbers[rnd.Intn(len(numbers))]
	pwd[8] = lower[rnd.Intn(len(lower))]
	pwd[9] = upper[rnd.Intn(len(upper))]
	rand.Shuffle(len(pwd), func(i, j int) {
		pwd[i], pwd[j] = pwd[j], pwd[i]
	})
	return string(pwd)
}

// Calculate how long the alert was opened
// Duration is the elapsed seconds,
// Returns days/hrs/mins string
func CalcDuration(opened, closed int64) string {
	response := ""
	now := time.Now().UTC().Unix()
	duration := now - opened
	if closed > 1000 {
		duration = closed - opened
	}
	numberOfDays := duration / 86400
	numberOfHours := (duration % 86400) / 3600
	numberOfMinutes := ((duration % 86400) % 3600) / 60
	if numberOfDays > 0 {
		response += strconv.FormatInt(numberOfDays, 10) + " days"
	} else if numberOfHours > 0 {
		response += strconv.FormatInt(numberOfHours, 10) + " hrs"
	} else {
		response += strconv.FormatInt(numberOfMinutes, 10) + " mins"
	}
	return response
}
