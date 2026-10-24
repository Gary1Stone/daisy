package online

import (
	"fmt"
	"math/bits"
	"time"

	"github.com/gbsto/daisy/macs"
)

const slotsPerDay = 96

func dayKey(t time.Time) int {
	return t.Year()*10000 + int(t.Month())*100 + t.Day()
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

// Helper function to create YYYYMMDD int value
func gen3Days(dayUTC time.Time) (prevDay, currDay, nextDay int) {
	prevDay = dayKey(dayUTC.AddDate(0, 0, -1))
	currDay = dayKey(dayUTC)
	nextDay = dayKey(dayUTC.AddDate(0, 0, 1))
	return prevDay, currDay, nextDay
}

// func convertYYYYMMDDToDate(yyyymmdd int) string {
func formatDate(dateInt int, tzoff int) string {
	// Extract year, month, and day
	year := dateInt / 10000
	month := (dateInt / 100) % 100
	day := dateInt % 100

	// Create date in UTC
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	// Apply timezone offset (in seconds)
	t = t.Add(time.Duration(tzoff) * time.Second)

	return t.Format("Monday 02 Jan 2006")
}

// Convert yyyymmdd to time
func yyyymmddUTC2time(yyyymmdd int64) time.Time {
	y := int(yyyymmdd / 10000)
	m := int((yyyymmdd % 10000) / 100)
	d := int(yyyymmdd % 100)
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

// get all the online history for the computers from newest first to oldest last
func handleTimezones(tzoff int, onlineInfo []macs.Online) ([]macs.Online, error) {
	// Calculate date range for system history
	// Add buffer for timezone shifts (at least 1 day)
	maxDateVal := onlineInfo[0].Date
	minDateVal := onlineInfo[len(onlineInfo)-1].Date

	// Convert to time to subtract/add days safely
	minT := yyyymmddUTC2time(minDateVal).AddDate(0, 0, -2)
	maxT := yyyymmddUTC2time(maxDateVal).AddDate(0, 0, 2)

	minDate := int64(dayKey(minT))
	maxDate := int64(dayKey(maxT))

	// get system history in range
	sysHist, err := getSysHistoryInRange(minDate, maxDate)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup of device data
	deviceData := make(map[int64]*macs.Online, len(onlineInfo))
	for i := range onlineInfo {
		deviceData[onlineInfo[i].Date] = &onlineInfo[i]
	}

	// Pre-calculate timezone offset index
	startIdx := genStartindex(tzoff)

	// Helper to fill a section for a given three-days slot buffer
	// defined outside the loop to avoid reallocating the closure each iteration.
	fillSection := func(dayInt int, offset int, threeDaysOfSlots *[slotsPerDay * 3]int) {
		// Check device data
		var am, pm uint64
		if dev, ok := deviceData[int64(dayInt)]; ok {
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

	// Iterate over the days we have data for and adjust for timezone
	for i := range onlineInfo {
		d := yyyymmddUTC2time(onlineInfo[i].Date)
		if len(onlineInfo[i].Slots) != slotsPerDay {
			onlineInfo[i].Slots = make([]int, slotsPerDay) // 96 elements
		}
		prevDay, currDay, nextDay := gen3Days(d)

		var threeDaysOfSlots [slotsPerDay * 3]int

		fillSection(prevDay, 0, &threeDaysOfSlots)
		fillSection(currDay, slotsPerDay, &threeDaysOfSlots)
		fillSection(nextDay, slotsPerDay*2, &threeDaysOfSlots)

		// Copy the relevant window to the item's slots
		copy(onlineInfo[i].Slots, threeDaysOfSlots[startIdx:startIdx+slotsPerDay])
	}
	return onlineInfo, nil
}

// Debugging utility
func PrintBitmap(mac string, bmp []int) {
	fmt.Print(mac, " ")
	for i := range bmp {
		fmt.Print(bmp[i])
	}
	fmt.Print("\n")
}
