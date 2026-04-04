package db

import (
	"bytes"
	"database/sql"
	"errors"
	"log"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gbsto/daisy/util"
	"github.com/jordan-wright/email"
)

type Emails struct {
	ID        int    `json:"id"`
	Timestamp string `json:"timestamp"`
	Function  string `json:"function"`
	Uid       int    `json:"uid"`
	User      string `json:"user"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Template  string `json:"template"`
	Param1    string `json:"param1"` // User's First Name
	Param2    string `json:"param2"` // One Time Password
	Status    string `json:"status"`
	Sent      int    `json:"sent"`
}

func SendOneTimePassword(uid int) error {
	usr, err := GetProfile(SYS_PROFILE.Uid, uid)
	if err != nil {
		return err
	}
	_, err = mail.ParseAddress(usr.User)
	if err != nil {
		return err
	}
	// Generate a new one time password
	usr.Otp = util.GetRandomPassword()
	usr.Pwd_reset = 1
	if err := usr.UpdateRecord(usr.Uid); err != nil {
		log.Println(err)
		return err
	}
	var mail Emails
	mail.Function = "OTP" //Type of email to be sent
	mail.Uid = usr.Uid
	mail.User = usr.User
	mail.Body = ""
	mail.Subject = "DAISY One Time Password"
	mail.Template = "email.html"
	mail.Param1 = usr.First
	mail.Param2 = usr.Otp
	mail.Status = "2SEND" //Waiting to be sent
	mail.Sent = 0         //First try at sending

	err = queueEmail(mail)
	if err != nil {
		log.Println(err)
		return err
	} else {
		go transmitEmail()
	}
	return nil
}

func transmitEmail() {
	type TemplateInfo struct {
		Param1 string
		Param2 string
	}

	//Get the first queued email
	var info TemplateInfo
	stillMore2Go := true
	toSend, err := getEmailFromQueue(SYS_PROFILE.Uid)
	if err != nil {
		stillMore2Go = false
	}

	for stillMore2Go {
		info.Param1 = toSend.Param1
		info.Param2 = toSend.Param2

		//Configure to use email template file
		var body bytes.Buffer

		// Get path to ./web/views/templatename.html
		workingDir, err := os.Getwd()
		if err != nil {
			workingDir = "."
		}
		templateFile := filepath.Join(workingDir, "web", "views", toSend.Template)
		if _, err := os.Stat(templateFile); os.IsNotExist(err) {
			log.Println(err)
			return
		}
		t, err := template.ParseFiles(templateFile)
		if err != nil {
			log.Println(err)
		}
		t.Execute(&body, info)

		//Send the email
		err = SendAnEmail(toSend.User, os.Getenv("EMAIL"), toSend.Subject, body.String())
		if err != nil {
			log.Println(err)
			toSend.Status = "ERROR"
			toSend.Sent += 1
		} else {
			toSend.Status = "SENT"
			toSend.Sent = int(time.Now().UTC().Unix())
		}

		//Save the status
		err = updateEmailInQueue(toSend)
		if err != nil {
			log.Println(err)
		}

		//Wait 3 seconds and check for any more emails to send
		time.Sleep(3 * time.Second)
		toSend, err = getEmailFromQueue(SYS_PROFILE.Uid)
		if err != nil {
			stillMore2Go = false
		}
	}
}

// port 465 is SSL, port 587 is TLS. 465 worked in Java. 587 works in Go.
// How to ensure the connection is open: from PowerShell or VScode terminal run:
// Test-NetConnection smtp.gmail.com -Port 587
// Real-NetConnection smtp.gmail.com -Port 465
func SendAnEmail(to, from, subject, body string) error {
	e := email.NewEmail()
	e.From = from
	e.To = []string{to}
	e.Subject = from
	e.HTML = []byte(body)
	return e.Send(os.Getenv("EMAILSERVER")+":"+os.Getenv("EMAILPORT"), smtp.PlainAuth("", os.Getenv("EMAIL"), os.Getenv("EMAILPWD"), os.Getenv("EMAILSERVER")))
}

// Insert email request into queue
func queueEmail(email Emails) error {
	// Check email address is valid
	_, err := mail.ParseAddress(email.User)
	if err != nil {
		return err
	}
	// Check this is a current active user
	usr, err := GetUserByEmail(email.User)
	if err != nil {
		return err
	}
	if usr.Uid <= 0 || usr.Active != 1 {
		return errors.New("invalid user")
	}
	// Check if user wants email notifications
	// if usr.Notify != 1 && (email.Function != "OTP" || email.Function != "SOFTWARE") { // Enhance later for different type of notifications
	// 	return errors.New("email notifications are off")
	// }
	// Save email request to database table (queue)
	query := `
		INSERT INTO emails (
		function, uid, user, subject, body, template, param1, param2, status, sent
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = Conn.Exec(query, email.Function, email.Uid, email.User, email.Subject,
		email.Body, email.Template, email.Param1, email.Param2, email.Status, email.Sent)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Return first email not sent from queue
// Will only try sending an email three times
// Skip One-Time-Passwords (OTP) that have expired (15 minutes old)
func getEmailFromQueue(curUid int) (Emails, error) {
	var email Emails
	tzoff := GetTzoff(curUid)
	query := `
		SELECT id, function, uid, user, subject, body, status, sent, 
		strftime('%Y-%m-%d %H:%M', timestamp-0, 'unixepoch') AS timestamp, template, param1, param2 
		FROM emails 
		WHERE sent<3 AND (
			(function = 'OTP' AND timestamp > CAST(strftime('%s', 'now', '-15 minutes') AS INTEGER))
			OR
			(function != 'OTP' OR function IS NULL) -- Keeps other functions without time check
		)
		ORDER BY sent ASC 
		LIMIT 1
	`
	err := Conn.QueryRow(query, tzoff).Scan(&email.ID, &email.Function, &email.Uid,
		&email.User, &email.Subject, &email.Body, &email.Status, &email.Sent,
		&email.Timestamp, &email.Template, &email.Param1, &email.Param2)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println(err)
		}
	}
	return email, err
}

// Update email queue
func updateEmailInQueue(email Emails) error {
	query := `
		UPDATE emails SET 
		function=?, uid=?, user=?, subject=?, body=?, status=?, sent=?, template=?, param1=?, param2=? 
		WHERE ID=?
	`
	_, err := Conn.Exec(query, email.Function, email.Uid, email.User, email.Subject,
		email.Body, email.Status, email.Sent, email.Template, email.Param1,
		email.Param2, email.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}

func emailNewSoftwareList(hostname string, swlist []string) error {
	var mail Emails
	mail.Function = "SOFTWARE" //Type of email to be sent
	mail.Uid = SYS_PROFILE.Uid
	mail.User = SYS_PROFILE.User
	mail.Body = ""
	mail.Subject = "DAISY Software Installation Notifiction"
	mail.Template = "emailsoftware.html"
	mail.Param1 = hostname
	mail.Param2 = strings.Join(swlist, ", ")
	mail.Status = "2SEND" //Waiting to be sent
	mail.Sent = 0         //First try at sending
	err := queueEmail(mail)
	if err != nil {
		log.Println(err)
		return err
	} else {
		go transmitEmail()
	}
	return nil
}

// Send new iSaw PIN to the iSaw user
func EmailNewPin(emailAddr, pin string) error {
	// Check email address is valid
	_, err := mail.ParseAddress(emailAddr)
	if err != nil {
		log.Println(err)
		return err
	}
	// Check this is a current active user
	usr, err := GetUserByEmail(emailAddr)
	if err != nil {
		log.Println(err)
		return err
	}
	if usr.Uid <= 0 || usr.Active != 1 {
		log.Println("invalid user")
		return errors.New("invalid user")
	}
	var mail Emails
	mail.Function = "PIN" //Type of email to be sent
	mail.Uid = usr.Uid
	mail.User = emailAddr
	mail.Body = ""
	mail.Subject = "iSAW PIN Reset Request"
	mail.Template = "emailpin.html"
	mail.Param1 = usr.First
	mail.Param2 = pin
	mail.Status = "2SEND" //Waiting to be sent
	mail.Sent = 0         //First try at sending
	err = queueEmail(mail)
	if err != nil {
		log.Println(err)
		return err
	} else {
		go transmitEmail()
	}
	return nil
}
