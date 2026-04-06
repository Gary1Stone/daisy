package glyphs

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ConvertGlyphs() {
	// Structs matching the SVG Font format found in metro-ui-core.svg
	type Glyph struct {
		Name string `xml:"glyph-name,attr"`
		D    string `xml:"d,attr"`
	}

	type Font struct {
		Glyphs []Glyph `xml:"glyph"`
	}

	type Defs struct {
		Font Font `xml:"font"`
	}

	type SVG struct {
		Defs Defs `xml:"defs"`
	}

	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	glyphsCorePath := filepath.Join(workingDir, "glyphs", "glyph.svg")
	outputIconsDir := filepath.Join(workingDir, "glyphs", "icons")

	// Check if the metro-ui-core.svg file exists
	if _, err := os.Stat(glyphsCorePath); os.IsNotExist(err) {
		log.Printf("metro-ui-core.svg not found at %s. Skipping glyph conversion.", glyphsCorePath)
		return
	} else if err != nil {
		log.Printf("Error checking for metro-ui-core.svg: %v", err)
		return
	}

	// Read the contents of the file
	fileContents, err := os.ReadFile(glyphsCorePath)
	if err != nil {
		log.Printf("Error reading metro-ui-core.svg: %v", err)
		return
	}

	var svg SVG
	err = xml.Unmarshal(fileContents, &svg)
	if err != nil {
		log.Printf("Error unmarshaling SVG: %v", err)
		return
	}

	// Create the output directory if it doesn't exist
	if _, err := os.Stat(outputIconsDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputIconsDir, 0755)
		if err != nil {
			log.Printf("Error creating output directory %s: %v", outputIconsDir, err)
			return
		}
	}

	log.Printf("Converting glyphs from %s to individual SVG files in %s", glyphsCorePath, outputIconsDir)

	for _, glyph := range svg.Defs.Font.Glyphs {
		// Skip glyphs without names or paths (like the space character)
		if glyph.Name == "" || glyph.D == "" {
			continue
		}

		filename := glyph.Name + ".svg"
		filepath := filepath.Join(outputIconsDir, filename)

		// Scale font coordinates (1024) down to 24x24 (24/1024 = 0.0234375).
		// We also keep the Y-flip transform to correct the font's coordinate system.
		svgContent := fmt.Sprintf(
			`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><g transform="scale(0.0234375) matrix(1 0 0 -1 0 960)"><path d="%s" fill="currentColor" /></g></svg>`,
			glyph.D,
		)

		// Write the content to the new SVG file
		err = os.WriteFile(filepath, []byte(svgContent), 0644)
		if err != nil {
			log.Printf("Error writing SVG file %s: %v", filepath, err)
		} else {
			log.Printf("Successfully created %s", filepath)
		}
	}
	log.Println("Glyph conversion complete.")
}

// Search through this project's files to find all refernces of <span class='mif-****'></span>
func FindMifIconReferences() map[string]bool {
	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	iconMap := make(map[string]bool)
	// This regex captures the name part after 'mif-'
	re := regexp.MustCompile(`mif-([a-zA-Z0-9-]+)`)

	err = filepath.Walk(workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories that contain generated assets or dependencies
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "icons" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only scan files that likely contain UI or logic code
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".go" || ext == ".html" || ext == ".js" || ext == ".tmpl" {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip files that can't be read
			}

			matches := re.FindAllStringSubmatch(string(content), -1)
			for _, match := range matches {
				if len(match) > 1 {
					iconMap[match[1]] = true
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error scanning for mif references: %v", err)
	}

	log.Printf("Found %d unique MIF icon references in the project.", len(iconMap))

	for key := range iconMap {
		//fmt.Printf("%s.svg", key)

		sourceIconsDir := filepath.Join(workingDir, "glyphs", "icons")
		destinationIconsDir := filepath.Join(workingDir, "web", "public", "icons")

		sourcePath := filepath.Join(sourceIconsDir, fmt.Sprintf("%s.svg", key))
		destinationPath := filepath.Join(destinationIconsDir, fmt.Sprintf("%s.svg", key))

		// Check if the source file exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			log.Println("Source file does not exist:", fmt.Sprintf("%s.svg", key))
			continue
		}
		// Check if the destination directory exists
		if _, err := os.Stat(destinationIconsDir); os.IsNotExist(err) {
			err = os.MkdirAll(destinationIconsDir, 0755)
			if err != nil {
				log.Printf("Error creating output directory %s: %v", destinationIconsDir, err)
				continue
			}
		}
		// Copy the file
		err = os.Rename(sourcePath, destinationPath)
	}
	return iconMap
}

// Convert the metro.json file into the icon svg files

func Json2Svg() {

	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	jsonFile := filepath.Join(workingDir, "glyphs", "metro.json")
	outputIconsDir := filepath.Join(workingDir, "glyphs", "icons")

	// Check if the metro-ui-core.svg file exists
	if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
		log.Printf("metro.json not found at %s. Skipping conversion.", jsonFile)
		return
	} else if err != nil {
		log.Printf("Error checking for metro.json: %v", err)
		return
	}

	// Read the contents of the file
	fileContents, err := os.ReadFile(jsonFile)
	if err != nil {
		log.Printf("Error reading metro.json: %v", err)
		return
	}

	var selection struct {
		Icons []struct {
			Icon struct {
				Paths []string `json:"paths"`
			} `json:"icon"`
			Properties struct {
				Name string `json:"name"`
			} `json:"properties"`
		} `json:"icons"`
	}

	if err := json.Unmarshal(fileContents, &selection); err != nil {
		log.Printf("Error parsing metro.json: %v", err)
		return
	}

	os.MkdirAll(outputIconsDir, 0755)

	for _, item := range selection.Icons {
		name := item.Properties.Name
		if name == "" || len(item.Icon.Paths) == 0 {
			continue
		}

		svgPath := filepath.Join(outputIconsDir, name+".svg")
		d := strings.Join(item.Icon.Paths, " ")
		svgContent := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 1024 1024"><path d="%s" fill="currentColor" /></svg>`, d)

		os.WriteFile(svgPath, []byte(svgContent), 0644)
	}
}
