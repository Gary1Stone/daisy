package db

import (
	"database/sql"
	"math"
	"sync"
)

type searchInfo struct {
	latitude  float64
	longitude float64
	curSinLat float64
	curCosLat float64
	curCosLon float64
	curSinLon float64
	minLat    float64
	maxLat    float64
	minLon    float64
	maxLon    float64
}

// Returns city_id and community_id
func SearchNearestPlace(longitude, latitude float64) (int, int) {
	//	start := time.Now()
	// Confirm long/lat is valid
	if longitude > 180.0 || longitude < -180.0 ||
		latitude > 90.0 || latitude < -90.0 {
		return 0, 0
	}

	// fast search computations
	var searchData searchInfo
	searchData.latitude = latitude
	searchData.longitude = longitude
	searchData.curSinLat = math.Sin(latitude * math.Pi / 180)
	searchData.curCosLat = math.Cos(latitude * math.Pi / 180)
	searchData.curCosLon = math.Cos(longitude * math.Pi / 180)
	searchData.curSinLon = math.Sin(longitude * math.Pi / 180)

	//set a bounding rectangle to search within
	searchRadiusDegrees := 1.5 //1 degree latitude = ~111 km always, 1 degree longitude = 111 km × cos(latitude)
	searchData.minLat = latitude - searchRadiusDegrees
	searchData.maxLat = latitude + searchRadiusDegrees
	searchData.minLon = longitude - searchRadiusDegrees
	searchData.maxLon = longitude + searchRadiusDegrees

	//Basic handling for longitude wrapping
	if searchData.minLon < -180 {
		searchData.minLon += 360
	}
	if searchData.maxLon > 180 {
		searchData.maxLon -= 360
	}
	// Clamp latitude to the valid range [-90, 90] Poles don't wrap
	if searchData.minLat < -90 {
		searchData.minLat = -90
	}
	if searchData.maxLat > 90 {
		searchData.maxLat = 90
	}

	var cityId int = 0
	var communityId int
	var wg sync.WaitGroup
	wg.Add(1)
	go searchNearestCity(searchData, &cityId, &wg)
	wg.Add(1)
	go searchNearestCommunity(searchData, &communityId, &wg)
	wg.Wait()
	//	log.Println("SearchTime: " + time.Since(start).String())
	return cityId, communityId
}

func searchNearestCity(searchData searchInfo, cityId *int, wg *sync.WaitGroup) {
	defer wg.Done()
	var query string
	var args []any
	// Check if the longitude range crosses the antimeridian
	if searchData.minLon > searchData.maxLon {
		// Handle wrap-around case -- Use OR for wrap-around
		query = `
			SELECT city_id, ? * sin_lat + ? * cos_lat * (cos_lon * ? + sin_lon * ?) as distance
			FROM cities
			WHERE latitude BETWEEN ? AND ?
			AND (longitude >= ? OR longitude <= ?)
			ORDER BY distance DESC
			LIMIT 1
		`
		args = []any{searchData.curSinLat, searchData.curCosLat, searchData.curCosLon, searchData.curSinLon, searchData.minLat, searchData.maxLat, searchData.minLon, searchData.maxLon}
	} else {
		// Normal case (no wrap-around) -- Use BETWEEN for normal case
		query = `
			SELECT city_id, ? * sin_lat + ? * cos_lat * (cos_lon * ? + sin_lon * ?) as distance
			FROM cities
			WHERE latitude BETWEEN ? AND ?
			AND longitude BETWEEN ? AND ?
			ORDER BY distance DESC
			LIMIT 1
		`
		args = []any{searchData.curSinLat, searchData.curCosLat, searchData.curCosLon, searchData.curSinLon, searchData.minLat, searchData.maxLat, searchData.minLon, searchData.maxLon}
	}
	*cityId = searchForNearst(query, args...)
	if *cityId > 0 {
		return
	}
	//log.Println("fallback used for city search")
	query = `
		SELECT city_id, ? * sin_lat + ? * cos_lat * (cos_lon * ? + sin_lon * ?) as distance
		FROM cities ORDER BY distance DESC LIMIT 1
	`
	args = []any{searchData.curSinLat, searchData.curCosLat, searchData.curCosLon, searchData.curSinLon}
	*cityId = searchForNearst(query, args...)
}

func searchForNearst(query string, args ...any) int {
	var recId int
	var dist float64
	err := Conn.QueryRow(query, args...).Scan(&recId, &dist)
	if err == sql.ErrNoRows {
		return 0
	} else if err != nil {
		return 0
	}
	return recId
}

// func searchNearestCommunity(searchData searchInfo, communityId *int, cityId *int) {
func searchNearestCommunity(searchData searchInfo, communityId *int, wg *sync.WaitGroup) {
	defer wg.Done()
	var query string
	var args []any
	//Check if the longitude range crosses the antimeridian
	if searchData.minLon > searchData.maxLon {
		// Handle wrap-around case -- Use OR for wrap-around
		query = `
			SELECT community_id, ? * sin_lat + ? * cos_lat * (cos_lon * ? + sin_lon * ?) as distance
			FROM communities
			WHERE latitude BETWEEN ? AND ?
			AND (longitude >= ? OR longitude <= ?)
			ORDER BY distance DESC
			LIMIT 1
		`
		args = []any{searchData.curSinLat, searchData.curCosLat, searchData.curCosLon, searchData.curSinLon, searchData.minLat, searchData.maxLat, searchData.minLon, searchData.maxLon}
	} else {
		// Normal case (no wrap-around) -- Use BETWEEN for normal case
		query = `
			SELECT community_id, ? * sin_lat + ? * cos_lat * (cos_lon * ? + sin_lon * ?) as distance
			FROM communities
			WHERE latitude BETWEEN ? AND ?
			AND longitude BETWEEN ? AND ?
			ORDER BY distance DESC
			LIMIT 1
		`
		args = []any{searchData.curSinLat, searchData.curCosLat, searchData.curCosLon, searchData.curSinLon, searchData.minLat, searchData.maxLat, searchData.minLon, searchData.maxLon}
	}
	*communityId = searchForNearst(query, args...)
	if *communityId > 0 {
		return
	}
	//	log.Println("fallback used for community search")
	query = `
		SELECT community_id, ? * sin_lat + ? * cos_lat * (cos_lon * ? + sin_lon * ?) as distance 
		FROM communities ORDER BY distance DESC LIMIT 1
	`
	args = []any{searchData.curSinLat, searchData.curCosLat, searchData.curCosLon, searchData.curSinLon}
	*communityId = searchForNearst(query, args...)
}
