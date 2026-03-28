package main

import (
	"log"
	"os"

	"github.com/gbsto/daisy/web"
	"github.com/joho/godotenv"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Load the environment variables
func init() {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("FATAL: .env file not found. Please create one in the current working directory.")
		} else {
			log.Fatal("FATAL: Failed to load .env file. Error:", err.Error())
		}
	}
}

func main() {
	webLogger := setupLogFiles()
	defer webLogger.Close()

	web.StartServer()

	log.Println("Starting Daisy")

}

// Use lumberjack for automatic log rotation and truncation.
func setupLogFiles() *lumberjack.Logger {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "error.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,  // number of old log files to keep
		MaxAge:     28, // days to keep old log files
		Compress:   false,
	})
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	// Configure a separate logger for web requests.
	webLogger := &lumberjack.Logger{
		Filename:   "web.log",
		MaxSize:    10, // megabytes
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   false,
	}
	return webLogger
}
