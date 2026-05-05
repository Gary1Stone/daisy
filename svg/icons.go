package svg

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	iconMap = make(map[string]string)
	iconMu  sync.RWMutex
)

const defaultIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none" /><path d="M14 3v4a1 1 0 0 0 1 1h4" /><path d="M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2" /><path d="M12 17v.01" /><path d="M12 14a1.5 1.5 0 1 0 -1.14 -2.474" /></svg>`

// GetIcon returns the svg content from the iconMap or a default icon if not found
// while populating the cache with only the used icons
func GetIcon(iconName string) string {

	if iconName == "" {
		return defaultIcon
	}

	// Get the icon from iconMap, if it is there
	iconMu.RLock()
	if icon, ok := iconMap[iconName]; ok {
		iconMu.RUnlock()
		return icon
	}
	iconMu.RUnlock()

	filename := iconName + ".svg"
	iconPath := filepath.Join("web", "public", "icons", filename)

	// Get the icon from the icon directory and add it to the iconMap[]
	data, err := os.ReadFile(iconPath)
	if err != nil {
		log.Println("Error reading icon file:", filename, err)
		return defaultIcon
	}

	// Clean comments and set currentColor for theme support
	processedSvg, err := processSvg(data)
	if err != nil {
		log.Println("Error processing SVG:", filename, err)
		return string(data)
	}
	icon := string(processedSvg)

	// Save the icon and then return it
	iconMu.Lock()
	iconMap[iconName] = icon
	iconMu.Unlock()
	return icon
}

// processSvg cleans comments and sets color to currentColor in a single pass
func processSvg(data []byte) ([]byte, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.Comment:
			// Skip comments
			continue
		case xml.StartElement:
			hasColorAttr := false
			for i, attr := range t.Attr {
				if attr.Name.Local == "fill" || attr.Name.Local == "stroke" {
					hasColorAttr = true // Found a color attribute, regardless of value
					if attr.Value != "none" {
						t.Attr[i].Value = "currentColor"
					}
				}
			}

			// If it's the root <svg> tag and absolutely no color info was found,
			// add fill="currentColor" so it respects the Pico theme.
			if t.Name.Local == "svg" && !hasColorAttr {
				t.Attr = append(t.Attr, xml.Attr{Name: xml.Name{Local: "fill"}, Value: "currentColor"})
			}
			token = t
		}

		if err := encoder.EncodeToken(token); err != nil {
			return nil, err
		}
	}

	encoder.Flush()
	return buf.Bytes(), nil
}

// UPDATE ICONS SET icon='eye.svg' WHERE icon='mif-eye';
// UPDATE ICONS SET icon='' WHERE icon='mif-checkmark';
// UPDATE ICONS SET icon='tag.svg' WHERE icon='mif-tag';
// UPDATE ICONS SET icon='' WHERE icon='mif-heart-broken';
// UPDATE ICONS SET icon='' WHERE icon='mif-steps';
// UPDATE ICONS SET icon='' WHERE icon='mif-stethoscope';
// UPDATE ICONS SET icon='' WHERE icon='mif-wrench';
// UPDATE ICONS SET icon='' WHERE icon='mif-copy';
// UPDATE ICONS SET icon='' WHERE icon='mif-apps';
// UPDATE ICONS SET icon='' WHERE icon='mif-comment';
// UPDATE ICONS SET icon='' WHERE icon='mif-display';
// UPDATE ICONS SET icon='' WHERE icon='mif-laptop';
// UPDATE ICONS SET icon='' WHERE icon='mif-tablet';
// UPDATE ICONS SET icon='' WHERE icon='mif-printer';
// UPDATE ICONS SET icon='' WHERE icon='mif-phone';
// UPDATE ICONS SET icon='' WHERE icon='mif-cast';
// UPDATE ICONS SET icon='' WHERE icon='mif-volume-medium';
// UPDATE ICONS SET icon='' WHERE icon='mif-flow-tree';
// UPDATE ICONS SET icon='' WHERE icon='mif-news';
// UPDATE ICONS SET icon='' WHERE icon='mif-file-binary';
// UPDATE ICONS SET icon='' WHERE icon='mif-fire';
// UPDATE ICONS SET icon='' WHERE icon='mif-calendar';
// UPDATE ICONS SET icon='user.svg' WHERE icon='mif-user';
// UPDATE ICONS SET icon='' WHERE icon='mif-map2';
// UPDATE ICONS SET icon='' WHERE icon='mif-room';
// UPDATE ICONS SET icon='' WHERE icon='mif-users';
// UPDATE ICONS SET icon='' WHERE icon='mif-hammer';
// UPDATE ICONS SET icon='' WHERE icon='mif-question';
// UPDATE ICONS SET icon='' WHERE icon='mif-location-city';
// UPDATE ICONS SET icon='' WHERE icon='mif-calculator2';
// UPDATE ICONS SET icon='' WHERE icon='mif-cabinet';
// UPDATE ICONS SET icon='' WHERE icon='mif-windows';
// UPDATE ICONS SET icon='' WHERE icon='mif-my-location';
// UPDATE ICONS SET icon='' WHERE icon='mif-star-half';
