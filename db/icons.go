package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type iconMap struct {
	sync.RWMutex
	theMap map[string]string
}

var icons iconMap

func (i *iconMap) get(name string) string {
	i.RLock()
	defer i.RUnlock()
	if icon, ok := i.theMap[strings.ToUpper(name)]; ok {
		return icon
	}
	return ""
}

// Clear any existing map, then load all icons at once while locked
func (i *iconMap) loadIcons() {
	rows, err := Conn.Query("SELECT name, icon FROM icons")
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	// Lock the mutex so reads cannot happen while it is being loaded
	i.Lock()
	defer i.Unlock()

	// Clear the map before loading new icons
	i.theMap = make(map[string]string)

	for rows.Next() {
		var name, icon string
		err := rows.Scan(&name, &icon)
		if err != nil {
			log.Println(err)
			continue
		} else {
			i.theMap[strings.ToUpper(name)] = icon
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}
}

func FindIconNameByName(name string) string {
	if len(icons.theMap) == 0 {
		icons.loadIcons()
	}
	return icons.get(name)
}

func GetKindUsingParentCid(cid int) string {
	kind := ""
	err := Conn.QueryRow("SELECT kind FROM devices WHERE cid=?", cid).Scan(&kind)
	if err != nil {
		log.Println(err)
	}
	return FindIconNameByName(kind)
}

// Get all the icons in the icons table
// Get all the icons in the choices table
func FindMissingExtraIcons() {

	type IconInfo struct {
		Code           string // the reference code to look up the icon
		Description    string
		Filename       string // the SVG filename
		IsFileThere    bool   // does the Filename have a match in the svg icon directory
		IsChoicesThere bool   // Does the choices table have the icon code in it
	}

	icons := make(map[string]IconInfo)

	rows, err := Conn.Query("SELECT name, description, icon FROM icons ORDER BY name")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var ico IconInfo
		err := rows.Scan(&ico.Code, &ico.Description, &ico.Filename)
		if err != nil {
			log.Println(err)
			continue
		} else {
			ico.IsFileThere = false
			ico.IsChoicesThere = false
			icons[ico.Code] = ico
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}

	// Check if all the codes are in the choices table
	rows, err = Conn.Query("SELECT code FROM choices ORDER BY code")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		err := rows.Scan(&code)
		if err != nil {
			log.Println(err)
			continue
		} else {
			if item, exists := icons[code]; exists {
				item.IsChoicesThere = true
				icons[code] = item
				//} else {
				//	fmt.Println("Not-icons: ", code)
			}
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}

	// Get all the icons in the svg directory
	executable, err := os.Executable()
	if err != nil {
		log.Println(err)
	}
	serverDir := filepath.Dir(executable)

	svgMap := make(map[string]bool)
	dirPath := filepath.Join(serverDir, "web", "public", "icons")
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Println(err)
	}
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			ext := filepath.Ext(name)
			nameWithoutExt := strings.TrimSuffix(name, ext)
			svgMap[nameWithoutExt] = true
		}
	}

	for _, item := range icons {
		if svgMap[item.Filename] {
			item.IsFileThere = true
			icons[item.Code] = item
		} else {
			fmt.Println("missing: ", item.Filename)
		}
	}

	//Print the result
	for _, item := range icons {
		if !item.IsChoicesThere {
			fmt.Println("Choice missing:", item.Code)
		}
		if !item.IsFileThere {
			fmt.Println("File missing", item.Code)
		}
	}
	fmt.Println("Done")

}

/* Results 

Choice missing: SOFTWARE
Choice missing: TROUBLE
Choice missing: HISTORY
Choice missing: CARE
Choice missing: BACKUP
Choice missing: COMMENT
Choice missing: GROUP
Choice missing: REQUEST
Choice missing: STATUS
Choice missing: ASSET
Choice missing: DRIVETYPE
Choice missing: PARENT
Choice missing: USER
Choice missing: IMPACT
Choice missing: CORES
Choice missing: TICKET
Choice missing: TREE
Choice missing: SITE
Choice missing: YEARS
Choice missing: MAKE
Choice missing: OS
Choice missing: OFFICE
Choice missing: CLONE
Choice missing: GEOFENCE
*/