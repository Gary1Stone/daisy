package online

import (
	"fmt"
	"math/bits"
	"time"
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

// // get all the online history for the computers from newest first to oldest last
func handleTimezones(tzoff int, onlineInfo []Online) ([]Online, error) {
	if len(onlineInfo) == 0 {
		return onlineInfo, nil
	}

	minT := yyyymmddUTC2time(onlineInfo[len(onlineInfo)-1].Date).AddDate(0, 0, -2)
	maxT := yyyymmddUTC2time(onlineInfo[0].Date).AddDate(0, 0, 2)

	minDate := int64(dayKey(minT))
	maxDate := int64(dayKey(maxT))

	sysHist, err := getSysHistoryInRange(minDate, maxDate)
	if err != nil {
		return nil, err
	}

	deviceData := make(map[int64]*Online, len(onlineInfo))
	for i := range onlineInfo {
		deviceData[onlineInfo[i].Date] = &onlineInfo[i]
	}

	startIdx := genStartindex(tzoff)
	var threeDaysOfSlots [slotsPerDay * 3]int

	for i := range onlineInfo {
		item := &onlineInfo[i]

		if len(item.Slots) != slotsPerDay {
			item.Slots = make([]int, slotsPerDay)
		}

		prevDay, currDay, nextDay := gen3Days(
			yyyymmddUTC2time(item.Date),
		)

		fillSection := func(dayInt, offset int) {
			var am, pm, sysAM, sysPM uint64

			if dev, ok := deviceData[int64(dayInt)]; ok {
				am, pm = uint64(dev.Am), uint64(dev.Pm)
			}

			if sys, ok := sysHist[int64(dayInt)]; ok {
				sysAM, sysPM = uint64(sys.Am), uint64(sys.Pm)
			}

			fillSlots(
				threeDaysOfSlots[offset:offset+slotsPerDay],
				am, pm, sysAM, sysPM,
			)
		}

		fillSection(prevDay, 0)
		fillSection(currDay, slotsPerDay)
		fillSection(nextDay, slotsPerDay*2)

		copy(item.Slots, threeDaysOfSlots[startIdx:startIdx+slotsPerDay])
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
