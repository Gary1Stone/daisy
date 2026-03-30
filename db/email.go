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
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gbsto/daisy/util"
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
	mail.Template = "./views/email.html"
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
		toAddress := toSend.User
		info.Param1 = toSend.Param1
		info.Param2 = toSend.Param2

		//Configure to use email template file
		var body bytes.Buffer

		// Get path to ./web/views/templatename.html
		workingDir, err := os.Getwd()
		if err != nil {
			workingDir = "."
		}

		// BAD GARY - should not store directory seperators in database emails table.
		// get everything after the last / in the template name
		templateName := toSend.Template
		templateName = templateName[strings.LastIndex(templateName, "/")+1:]
		toSend.Template = filepath.Join(workingDir, "web", "views", templateName)

		t, err := template.ParseFiles(toSend.Template)
		if err != nil {
			log.Println(err)
		}
		t.Execute(&body, info)

		//Build the sendmail information
		emailFrom := os.Getenv("EMAIL")
		emailServer := os.Getenv("EMAILSERVER")
		emailPort := os.Getenv("EMAILPORT") //port 465 is SSL, port 587 is TLS. 465 worked in Java. 587 works in Go.
		emailPassword := os.Getenv("EMAILPWD")
		auth := smtp.PlainAuth("", emailFrom, emailPassword, emailServer)
		header := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
		msg := "Subject: " + toSend.Subject + "\n" + header + "\n\n" + body.String()
		emailServer += ":" + emailPort

		//Send the email
		err = smtp.SendMail(emailServer, auth, emailFrom, []string{toAddress}, []byte(msg))
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
	var query strings.Builder
	query.WriteString("INSERT INTO emails (function, uid, user, subject, body, ")
	query.WriteString("template, param1, param2, status, sent) ")
	query.WriteString("VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	_, err = Conn.Exec(query.String(), email.Function, email.Uid, email.User, email.Subject,
		email.Body, email.Template, email.Param1, email.Param2, email.Status, email.Sent)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Return first email not sent from queue
// Will only try sending an email three times
func getEmailFromQueue(curUid int) (Emails, error) {
	var email Emails
	tzoff := GetTzoff(curUid)
	var query strings.Builder
	query.WriteString("SELECT id, function, uid, user, subject, body, status, sent, ")
	query.WriteString("strftime('%Y-%m-%d %H:%M', timestamp - ")
	query.WriteString(strconv.Itoa(tzoff))
	query.WriteString(", 'unixepoch') AS timestamp, ")
	query.WriteString("template, param1, param2 ")
	query.WriteString("FROM emails WHERE sent<3 ORDER BY sent ASC LIMIT 1")
	err := Conn.QueryRow(query.String()).Scan(&email.ID, &email.Function, &email.Uid,
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
	var query strings.Builder
	query.WriteString("UPDATE emails SET ")
	query.WriteString("function=?, uid=?, user=?, ")
	query.WriteString("subject=?, body=?, status=?, sent=?, ")
	query.WriteString("template=?, param1=?, param2=? ")
	query.WriteString("WHERE ID=?")
	_, err := Conn.Exec(query.String(), email.Function, email.Uid, email.User, email.Subject,
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
	mail.Template = "./views/emailsoftware.html"
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
	mail.Template = "./views/emailpin.html"
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
