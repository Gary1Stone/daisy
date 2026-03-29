package db

import (
	"log"

	"github.com/gbsto/daisy/util"
)

// remove photos in the images directory that are no longer in the database
func RemoveOldPhotos() {
	// Get all the photos in the images directory
	existingPhotos := util.MapPhotos()
	// Keep this application's photos
	existingPhotos["daisy.png"] = false
	existingPhotos["wknc-network.png"] = false
	existingPhotos["missing-sm.jpg"] = false
	existingPhotos["missing.jpg"] = false
	query := "SELECT image FROM devices WHERE image IS NOT NULL ORDER BY image"
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var photo string
		err := rows.Scan(&photo)
		if err != nil {
			log.Println(err)
		} else {
			if _, exists := existingPhotos[photo]; exists {
				// If the photo exists in the map, mark it as not needing to be deleted
				existingPhotos[photo] = false
			} else {
				log.Println("Missing photo: ", photo)
			}
			// Check for small version of the photo
			smallPhoto := util.AddSuffixBeforeExtension(photo, "-sm")
			if _, exists := existingPhotos[smallPhoto]; exists {
				// If the small version exists in the map, mark it as not needing to be deleted
				existingPhotos[smallPhoto] = false
			} else {
				log.Println("Missing small photo: ", smallPhoto)
			}
		}
	}

	for photo, toDelete := range existingPhotos {
		if toDelete {
			log.Println("Removing unused photo:", photo)
			util.DeletePhoto(photo)
		}
	}
}
