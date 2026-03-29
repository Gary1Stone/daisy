package db

import (
	"log"
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

// Return the mif-icon from the cache for speed
func GetIcon(name string) string {
	if len(icons.theMap) == 0 {
		icons.loadIcons()
	}
	return icons.get(name)
}
