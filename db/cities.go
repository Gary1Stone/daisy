package db

import (
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
)

// Lookup the nearest city information and put it into dto
func getCity(curSinLat, curCosLat, curCosLon, curSinLon float64, dto *Logins, wg *sync.WaitGroup) {
	defer wg.Done()
	var dist float64
	query := `
		SELECT city, city_id, country, country_code, state, 
		? * sin_lat + ? * cos_lat * (cos_lon * ? + sin_lon * ?) as distance 
		FROM cities ORDER BY distance DESC LIMIT 1 
	`
	err := Conn.QueryRow(query, curSinLat, curCosLat, curCosLon, curSinLon).Scan(&dto.City, &dto.City_id, &dto.Country, &dto.Country_code, &dto.State, &dist)
	if err != nil {
		log.Println(err)
	}
}

// Lookup the nearest community and put it into dto
func getCommunity(curSinLat, curCosLat, curCosLon, curSinLon float64, dto *Logins, wg *sync.WaitGroup) {
	defer wg.Done()
	var dist float64
	query := `
		SELECT community, community_id, 
		? * sin_lat + ? * cos_lat * (cos_lon * ? + sin_lon * ?) as distance 
		FROM communities ORDER BY distance DESC LIMIT 1
	`
	err := Conn.QueryRow(query, curSinLat, curCosLat, curCosLon, curSinLon).Scan(&dto.Community, &dto.Community_id, &dist)
	if err != nil {
		log.Println(err)
	}
}

// Determine if the user is within their geofence, returns true/false and distance (km) to home location
func IsInsideGeoFence(geoFence string, geoRadius int, userLat float64, userLon float64) (bool, int) {
	if geoRadius <= 0 || len(geoFence) < 3 {
		return false, -1 // the -1 indicates the user is not geoFenced
	}
	fenceCoords := strings.Split(geoFence, ",")
	if len(fenceCoords) != 2 {
		return false, 0
	}
	fenceLat, err := strconv.ParseFloat(fenceCoords[0], 64)
	if err != nil {
		return false, 0
	}
	fenceLon, err := strconv.ParseFloat(fenceCoords[1], 64)
	if err != nil {
		return false, 0
	}
	dist := HowFar(userLat, userLon, fenceLat, fenceLon)
	if geoRadius <= dist {
		return true, dist
	}
	return false, dist
}

// HowFar (Haversine) finds the distance between two points (lat/log) on the earth's surface
// in Kilometers which is why the 1.609344 factor in included
func HowFar(lat1, lon1, lat2, lon2 float64) int {
	if lat1 < -90 || lat1 > 90 || lon1 < -180 || lon1 > 180 || lat2 < -90 || lat2 > 90 || lon2 < -180 || lon2 > 180 {
		return -1
	}
	lat1 = math.Max(lat1, lat2)
	lat2 = math.Min(lat1, lat2)
	lon1 = math.Max(lon1, lon2)
	lon2 = math.Min(lon1, lon2)
	const rad float64 = 0.017453292519943295 // Convert degrees to radians = (math.Pi / 180)
	const toKm float64 = 6370.693485653058   // Convert to Km (60 * 1.1515 * 1.609344 * 180/math.Pi)
	var dist float64 = math.Sin(lat1*rad)*math.Sin(lat2*rad) + math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Cos(math.Abs(lon1-lon2)*rad)
	if math.IsNaN(dist) {
		return -1
	}
	var retVal int = int(math.Round(math.Acos(dist) * toKm))
	return retVal
}
