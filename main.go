package main

import (
	"log"
	"os"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/schedule"
	"github.com/gbsto/daisy/web"
	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	loadEnvVariables()
	setupLogFiles()
}

func main() {
	log.Println("Starting Daisy")
	db.StartServer()
	schedule.StartServer()
	web.StartServer()
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
func setupLogFiles() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "error.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,  // number of old log files to keep
		MaxAge:     28, // days to keep old log files
		Compress:   false,
	})
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}
