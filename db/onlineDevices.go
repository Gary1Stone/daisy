package db

import (
	"fmt"
	"log"
	"math/bits"
	"time"
)

const slotsPerDay = 96 // 24 hours * 4 slots per hour

// get which devices were online for a given date (YYYYMMDD)
func GetOnlineDay(tzoff int, yyyymmdd string) ([]Online, error) {
	input := yyyymmdd + " 00:01" // Convert yyyymmdd hh:nn to time
	layout := "20060102 15:04"
	dayUTC, err := time.Parse(layout, input) // Parse without timezone (naive time)
	if err != nil {
		dayUTC = time.Now().UTC()
	}
	dayUTC = dayUTC.Add(time.Duration(tzoff) * time.Second) // Apply offset: UTC = local + offset
	return GetDayOnlineDevices(tzoff, dayUTC)
}

// get the online history for all the devices in the 3 days for timezone shifting
func GetDayOnlineDevices(tzoff int, dayUTC time.Time) ([]Online, error) {
	prevDay, currDay, nextDay := gen3Days(dayUTC)
	startIdx := genStartindex(tzoff)                               // Pre-calculate timezone offset index
	onlineInfo, err := getWhatWasOnline(prevDay, currDay, nextDay) // This has been scrubbed
	if err != nil {
		log.Println("Error fetching online data:", err)
		return nil, err
	}

	if len(onlineInfo) == 0 {
		return nil, nil
	}

	// get system history in range in a date indexed map
	sysHist, err := getSysHistoryInRange(int64(prevDay), int64(nextDay)) // Already handled scrubbing
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup of device data
	deviceData := make(map[string]*Online, len(onlineInfo))
	for i := range onlineInfo {
		key := fmt.Sprintf("%s_%d", onlineInfo[i].Mac, onlineInfo[i].Date)
		deviceData[key] = &onlineInfo[i]
	}

	// func Helper to fill a section for a given three-days slot buffer
	fillSection := func(dayInt int, mac string, offset int, threeDaysOfSlots *[slotsPerDay * 3]int) {
		// Check device data
		var am, pm uint64 // initalize to 0, offline by default
		key := fmt.Sprintf("%s_%d", mac, dayInt)
		if dev, ok := deviceData[key]; ok {
			am = uint64(dev.Am)
			pm = uint64(dev.Pm)
		}
		// Check system data
		var sysAM, sysPM uint64
		if sys, ok := sysHist[int64(dayInt)]; ok {
			sysAM = uint64(sys.Am)
			sysPM = uint64(sys.Pm)
		}
		// Fill directly into the provided array slice
		fillSlots(threeDaysOfSlots[offset:offset+slotsPerDay], am, pm, sysAM, sysPM)
	}

	// Eliminate devices that were not active in the user's timezone for the current day
	activeMacs := make([]Online, 0, len(onlineInfo))

	// Iterate over the days we have data for and adjust for timezone
	// We have up to three (days) entries of macs in the list, but we only need one for analysis
	processedMacs := make(map[string]bool) // Track processed macs to avoid duplicates

	for i := range onlineInfo {
		if processedMacs[onlineInfo[i].Mac] {
			continue
		}
		processedMacs[onlineInfo[i].Mac] = true

		var threeDaysOfSlots [slotsPerDay * 3]int // Buffer for the three days of slots
		fillSection(prevDay, onlineInfo[i].Mac, 0, &threeDaysOfSlots)
		fillSection(currDay, onlineInfo[i].Mac, slotsPerDay, &threeDaysOfSlots)
		fillSection(nextDay, onlineInfo[i].Mac, slotsPerDay*2, &threeDaysOfSlots)
		// Copy the relevant window to the item's slots
		copy(onlineInfo[i].Slots, threeDaysOfSlots[startIdx:startIdx+slotsPerDay])

		// Check if the device was active in any of the slots for the current day
		isActive := false
		for j := range slotsPerDay {
			if onlineInfo[i].Slots[j] == 1 {
				isActive = true
				break
			}
		}
		if isActive {
			activeMacs = append(activeMacs, onlineInfo[i])
		}
	}
	return activeMacs, nil
}

func dayKey(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
}

// Helper function to create YYYYMMDD int value
func gen3Days(dayUTC time.Time) (prevDay, currDay, nextDay int) {
	prevDay = dayKey(dayUTC.AddDate(0, 0, -1))
	currDay = dayKey(dayUTC)
	nextDay = dayKey(dayUTC.AddDate(0, 0, 1))
	return prevDay, currDay, nextDay
}

// Calculate 3-day wide array's start index to convert to user's local day
func genStartindex(tzoff int) int {
	startIdx := slotsPerDay + (tzoff / 900) // Correct calculation is with plus (+) sign
	if startIdx < 0 {
		startIdx = 0
	} else if startIdx+slotsPerDay > 3*slotsPerDay {
		startIdx = 3*slotsPerDay - slotsPerDay
	}
	return startIdx
}

// Optimized to write directly to the destination slice to avoid allocations
func fillSlots(day []int, am, pm, sysAM, sysPM uint64) {
	for i := range 48 {
		if (sysAM>>uint(i))&1 == 0 {
			day[i] = -1
		} else {
			day[i] = 0
		}
		if (sysPM>>uint(i))&1 == 0 {
			day[i+48] = -1
		} else {
			day[i+48] = 0
		}
	}
	// set device-online (1) by iterating set bits
	for bitsSet := am; bitsSet != 0; bitsSet &= bitsSet - 1 {
		i := bits.TrailingZeros64(bitsSet)
		day[i] = 1
	}
	for bitsSet := pm; bitsSet != 0; bitsSet &= bitsSet - 1 {
		i := bits.TrailingZeros64(bitsSet)
		day[48+i] = 1
	}
}

// Fetch System (Host) status and all device online data for the 3 days
func getWhatWasOnline(prevDay, currDay, nextDay int) ([]Online, error) {
	items := make([]Online, 0)
	query := `SELECT O.mac, O.date, O.am, O.pm, O.host FROM onlinehistory O 
		LEFT JOIN macs M ON O.mac=M.mac
		WHERE date IN (?, ?, ?)
		ORDER BY  M.intruder DESC, M.name
	`
	rows, err := Conn.Query(query, prevDay, currDay, nextDay)
	if err != nil {
		log.Println("Error querying online data:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item Online
		err := rows.Scan(&item.Mac, &item.Date, &item.Am, &item.Pm, &item.Host)
		if err != nil {
			log.Println("Error scanning online data:", err)
			continue
		}
		item.Slots = make([]int, 96) // prefills with zeros
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println("Error iterating online data:", err)
		return nil, err
	}
	return scrubOnline(items)
}
