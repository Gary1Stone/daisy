package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/schedule"
	"github.com/gbsto/daisy/web"
	"github.com/gbsto/daisy/web/passkey"
	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	loadEnvVariables()
}

func main() {
	daisyLogger := setupLogFiles()
	defer daisyLogger.Close()
	log.Println("Starting Daisy...")
	db.StartServer()             // Start the database server
	schedule.StartServer()       // Schedule 2 am backups and photos directory cleanups
	passkey.StartWebAuthn()      // Configure Passkey for use
	web.StartServer(daisyLogger) // Start the webserver. This is a blocking never-ending loop
}

// Load the environment variables
func loadEnvVariables() {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("FATAL: .env file not found. Please create one in the current working directory.")
		} else {
			log.Fatal("FATAL: Failed to load .env file. Error:", err.Error())
		}
	}
}

// Use lumberjack for automatic log rotation and truncation.
func setupLogFiles() *lumberjack.Logger {
	// Get the base file dir
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	//Set the logs path
	logPath := filepath.Join(workingDir, "logs")

	//Create directory if it does not exist
	if stat, err := os.Stat(logPath); os.IsNotExist(err) {
		os.Mkdir(logPath, 0755)
	} else {
		if !stat.IsDir() {
			log.Printf("ERROR: Cannot begin due to a file named 'logs' in the home: [%v] directory. Remove it please.", logPath)
			panic("Ending")
		}
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   filepath.Join("logs", "errors.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 3,  // number of old log files to keep
		MaxAge:     28, // days to keep old log files
		Compress:   false,
	})
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	// Configure a separate logger for web requests.
	daisyLogger := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "daisy.log"),
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   false,
	}
	return daisyLogger
}
