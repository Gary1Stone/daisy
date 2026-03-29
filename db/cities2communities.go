package db

import (
	"log"
	"math"
)

// Scan the communities table and match a community to the nearest city
// and update city_id in the communities table.

type communityLocation struct {
	RecId     int
	Longitude float64
	Latitude  float64
}

func RunScanCommunities() {
	var coms []communityLocation
	query := "SELECT community_id, longitude, latitude FROM communities"
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var com communityLocation
		err := rows.Scan(&com.RecId, &com.Longitude, &com.Latitude)
		if err != nil {
			log.Println(err)
			return
		}
		coms = append(coms, com)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}
	for _, com := range coms {
		cityId := scan4NearestCity(com)
		if cityId > 0 {
			query = "UPDATE communities SET city_id=? WHERE community_id=?"
			_, err := Conn.Exec(query, cityId, com.RecId)
			if err != nil {
				log.Println(err)
			}
		}
	}

}

func scan4NearestCity(loc communityLocation) int {
	curSinLat := math.Sin(loc.Latitude * math.Pi / 180)
	curCosLat := math.Cos(loc.Latitude * math.Pi / 180)
	curCosLon := math.Cos(loc.Longitude * math.Pi / 180)
	curSinLon := math.Sin(loc.Longitude * math.Pi / 180)
	var dist float64
	var cityId int
	query := `
		SELECT city_id, ? * sin_lat + ? * cos_lat * (cos_lon * ? + sin_lon * ?) as distance 
		FROM cities ORDER BY distance DESC LIMIT 1 
	`
	err := Conn.QueryRow(query, curSinLat, curCosLat, curCosLon, curSinLon).Scan(&cityId, &dist)
	if err != nil {
		log.Println(err)
		return 0
	}
	return cityId
}
