package schedule

import (
	"log"
	"time"

	"github.com/gbsto/daisy/db"
)

func StartServer() {
	go everyQuarterHour() // Run every 15 minutes
	go atStartUp()        // Run at startup
}

func atStartUp() {
	// Nothing here yet
}

// This starts a scheduler that runs every 15 minutes.
// and do the database backup every day at 2 AM or thereabouts
// It runs indefinitely until the program is stopped.
func everyQuarterHour() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	lastRunDate := ""

	for now := range ticker.C {

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
