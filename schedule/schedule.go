package schedule

import (
	"log"
	"time"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/webserver"
)

func StartServer() {
	go atStartUp() // Run at startup
}

// Calculate the end time of the current 15-minute interval for the first log.
// Wait until the quarter hour to start the scheduler
func atStartUp() {
	now := time.Now()
	firstIntervalEndTime := now.Truncate(15 * time.Minute).Add(15 * time.Minute)
	initialDelay := firstIntervalEndTime.Sub(now)

	// Wait until the end of the current 15-minute interval
	time.Sleep(initialDelay)

	// Perform the first log operation
	webserver.ResetHits()
	go everyQuarterHour() // Run every 15 minutes forever
}

// This starts a scheduler that runs every 15 minutes.
// and does the database backup every day at 2 AM or thereabouts
// It runs indefinitely until the program is stopped.
func everyQuarterHour() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	lastRunDate := ""

	for now := range ticker.C {
		webserver.ResetHits()

		// Check if it's currently the 2 AM hour
		if now.Hour() == 2 {
			today := now.Format("2006-01-02")
			if today != lastRunDate {
				lastRunDate = today

				// Remove unused photos from the images directory
				if err := db.RemoveOldPhotos(); err != nil {
					log.Printf("ERROR: Failed to remove old photos: %v", err)
				}

				// Backup the database
				if err := db.TwoAmBackup(); err != nil {
					log.Printf("ERROR: Failed to backup database: %v", err)
				}
			}
		}
	}
}
