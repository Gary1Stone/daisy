package wigle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"time"
)

// WigleResponse defines the structure for the top-level API response from Wigle.
type WigleResponse struct {
	Success      bool          `json:"success"`
	TotalResults int           `json:"totalResults"`
	Results      []WigleResult `json:"results"`
	Message      string        `json:"message"`
}

// WigleResult defines the structure for an individual network result from Wigle.
type WigleResult struct {
	Trilat  float64 `json:"trilat"`
	Trilong float64 `json:"trilong"`
	Ssid    string  `json:"ssid"`
	NetID   string  `json:"netid"` // BSSID
}

var dailyLimitReached = false
var dailyLimitOccurred = time.Now()

// GetWiFiLocationFromWigle queries the Wigle.net API to find the geographic coordinates of a WiFi access point.
// It uses the SSID and BSSID (netid) for the query. Providing an approximate lat/lon helps narrow the search.
func GetWiFiLocationFromWigle(ssid, bssid string, lat, lon float64) (float64, float64, error) {
	apiName := os.Getenv("WIGLE_API_NAME")
	apiToken := os.Getenv("WIGLE_API_TOKEN")

	if apiName == "" || apiToken == "" {
		return 0, 0, errors.New("wigle api credentials not set in environment")
	}

	if dailyLimitReached {
		if time.Since(dailyLimitOccurred) < 24*time.Hour {
			return 0, 0, errors.New("wigle daily Limit Reached")
		}
		dailyLimitReached = false
		dailyLimitOccurred = time.Now()
	}

	// Construct the URL with query parameters
	apiURL := "https://api.wigle.net/api/v2/network/search"
	params := url.Values{}
	params.Add("netid", bssid)
	if ssid != "" {
		params.Add("ssid", ssid)
	}
	// Add a bounding box to narrow the search. The provided lat/lon has an
	// accuracy of 10km. We'll create a search box with a ~15km radius to be safe.
	const searchRadiusKm = 15.0
	const kmPerDegreeLat = 111.0 // Approximate km per degree of latitude.

	latRange := searchRadiusKm / kmPerDegreeLat
	lonRange := searchRadiusKm / (kmPerDegreeLat * math.Cos(lat*math.Pi/180))
	params.Add("latrange1", fmt.Sprintf("%f", lat-latRange))
	params.Add("latrange2", fmt.Sprintf("%f", lat+latRange))
	params.Add("longrange1", fmt.Sprintf("%f", lon-lonRange))
	params.Add("longrange2", fmt.Sprintf("%f", lon+lonRange))

	req, err := http.NewRequest("GET", apiURL+"?"+params.Encode(), nil)
	if err != nil {
		return 0, 0, fmt.Errorf("error creating Wigle request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(apiName, apiToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("error performing wigle request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("error reading wigle response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Wigle API Error: Status %d, Body: %s", resp.StatusCode, string(body))
		if resp.StatusCode == http.StatusTooManyRequests { // Status 429
			dailyLimitReached = true
			dailyLimitOccurred = time.Now()
		}
		return 0, 0, fmt.Errorf("wigle api returned non-200 status: %s", resp.Status)
	}

	var wigleResp WigleResponse
	if err := json.Unmarshal(body, &wigleResp); err != nil {
		return 0, 0, fmt.Errorf("error unmarshalling Wigle response: %w", err)
	}

	if !wigleResp.Success {
		return 0, 0, fmt.Errorf("wigle api call failed: %s", wigleResp.Message)
	}

	if len(wigleResp.Results) == 0 {
		return 0, 0, errors.New("no results found from Wigle API for the given BSSID")
	}

	// Return the coordinates from the first result
	firstResult := wigleResp.Results[0]
	log.Printf("Wigle API: Found location for BSSID %s: Lat %f, Lon %f", bssid, firstResult.Trilat, firstResult.Trilong)
	return firstResult.Trilat, firstResult.Trilong, nil
}

//	curl -i -H 'Accept:application/json' -u AID26bdc3b4eaef52afa1fc290ebbb52c47:01677d6090d05a074ce14d54db197d3f --basic https://api.wigle.net/api/v2/profile/user
