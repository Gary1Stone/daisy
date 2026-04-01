package util

import (
	"log"
	"os"
	"path/filepath"
)

// Read in an svg file from the ./web/icons directory

func GetIcon(filename string) string {
	defaultSvg := `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="icon icon-tabler icons-tabler-outline icon-tabler-file-unknown"><path stroke="none" d="M0 0h24v24H0z" fill="none" /><path d="M14 3v4a1 1 0 0 0 1 1h4" /><path d="M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2" /><path d="M12 17v.01" /><path d="M12 14a1.5 1.5 0 1 0 -1.14 -2.474" /></svg>`

	if filename == "" {
		return defaultSvg
	}

	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err)
		wd = "."
	}
	filepath := filepath.Join(wd, "web", "icons", filename)

	// Check if the file/directory is there
	if _, err := os.Stat(filepath); err != nil {
		if os.IsNotExist(err) {
			log.Println(err)
			return defaultSvg
		}
		log.Println(err)
		return defaultSvg
	}
	
	// File exists, Read the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Println(err)
		return defaultSvg
	}
	return string(data)

}
