package db

import (
	"log"
	"strings"
	"sync"
)

type droplistMap struct {
	sync.RWMutex
	theMap map[string]Droplist
}

var droplists droplistMap

type Droplist struct {
	Label    string
	Id       string
	Name     string
	Title    string
	ReadOnly bool
	ErrMsg   string
	Action   string
}

func (d *droplistMap) get(field string) Droplist {
	d.RLock()
	defer d.RUnlock()
	if item, ok := d.theMap[strings.ToUpper(field)]; ok {
		return item
	}
	return Droplist{}
}

func (d *droplistMap) loadDroplists() {
	query := "SELECT field, label, id, name, title, errmsg, action FROM droplists"
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	// Do the locking after the query returns the rows, keeping the locking time the shortest
	d.Lock()
	defer d.Unlock()
	d.theMap = make(map[string]Droplist, 21) // There are 21 entries in the database

	for rows.Next() {
		var item Droplist
		var field string
		err := rows.Scan(&field, &item.Label, &item.Id, &item.Name,
			&item.Title, &item.ErrMsg, &item.Action)
		if err != nil {
			log.Println(err)
			continue
		} else {
			d.theMap[strings.ToUpper(field)] = item
		}
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
	}
}

// Return the droplist info from the cache
func GetDroplistInfo(field string) Droplist {
	if len(droplists.theMap) == 0 {
		droplists.loadDroplists()
	}
	return droplists.get(field)
}
