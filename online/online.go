package online

import (
	"fmt"
	"time"
)

type Online struct {
	Mac     string `json:"Mac"`  // Mac ID unique
	Date    int64  `json:"Date"` // Unix timestamp UTC (Mac + Date = Primary Key in the form of YYYYMMDD)
	Am      int64  `json:"Am"`   // Morning (midnight to noon) Binary
	Pm      int64  `json:"Pm"`   // Afternoon (noon to midnight) Binary
	SysAm   int64  `json:"-"`    // System availability, shifted to the same local day
	SysPm   int64  `json:"-"`
	Host    bool   `json:"Host"`    // Is the entry from the host (network scanning) server
	Updated int64  `json:"Updated"` // Monotonic counter of last update for table replication, filled in by db trigger only
}

const slotsPerDay = 96

// The date column in the online table is YYYYMMDD, and this generate it.
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

// Debugging utility
func PrintBitmap(mac string, bmp []int) {
	fmt.Print(mac, " ")
	for i := range bmp {
		fmt.Print(bmp[i])
	}
	fmt.Print("\n")
}

// Shift the UTC (database) time to the user timezone.
// tzoff = users timezone offset from UTC
func shiftTimezones(tzoff int, onlineInfo []Online) ([]Online, error) {
	if len(onlineInfo) == 0 {
		return onlineInfo, nil
	}

	// need three days of data to cover all possible timezone shifts
	minT := yyyymmddUTC2time(onlineInfo[len(onlineInfo)-1].Date).AddDate(0, 0, -2)
	maxT := yyyymmddUTC2time(onlineInfo[0].Date).AddDate(0, 0, 2)

	// dayKey is an int64 in the form of YYYYMMDD, such as 20161007
	minDate := int64(dayKey(minT))
	maxDate := int64(dayKey(maxT))

	// system History is the network probe's online/offline status
	sysHist, err := getSysHistoryInRange(minDate, maxDate)
	if err != nil {
		return nil, err
	}

	deviceData := make(map[int64]Online, len(onlineInfo))
	for i := range onlineInfo {
		deviceData[onlineInfo[i].Date] = onlineInfo[i]
	}

	startIdx := genStartindex(tzoff)

	for i := range onlineInfo {
		item := &onlineInfo[i]

		prevDay, currDay, nextDay := gen3Days(
			yyyymmddUTC2time(item.Date),
		)

		var deviceDays, systemDays [3]Online
		for dayIndex, day := range [3]int{prevDay, currDay, nextDay} {
			if dev, ok := deviceData[int64(day)]; ok {
				deviceDays[dayIndex] = dev
			}
			if sys, ok := sysHist[int64(day)]; ok {
				systemDays[dayIndex] = sys
			}
		}

		shiftOnlineBits(item, deviceDays, systemDays, startIdx)
	}

	return onlineInfo, nil
}

func shiftOnlineBits(item *Online, deviceDays, systemDays [3]Online, startIdx int) {
	var deviceAM, devicePM, systemAM, systemPM uint64

	for slot := range slotsPerDay {
		sourceSlot := startIdx + slot
		dayIndex, daySlot := sourceSlot/slotsPerDay, sourceSlot%slotsPerDay
		bitIndex := daySlot
		deviceBits, systemBits := uint64(deviceDays[dayIndex].Am), uint64(systemDays[dayIndex].Am)
		if daySlot >= 48 {
			bitIndex -= 48
			deviceBits, systemBits = uint64(deviceDays[dayIndex].Pm), uint64(systemDays[dayIndex].Pm)
		}

		mask := uint64(1) << uint(bitIndex)
		destinationMask := uint64(1) << uint(slot%48)
		if deviceBits&mask != 0 {
			if slot < 48 {
				deviceAM |= destinationMask
			} else {
				devicePM |= destinationMask
			}
		}
		if systemBits&mask != 0 {
			if slot < 48 {
				systemAM |= destinationMask
			} else {
				systemPM |= destinationMask
			}
		}
	}

	item.Am, item.Pm = int64(deviceAM), int64(devicePM)
	item.SysAm, item.SysPm = int64(systemAM), int64(systemPM)
}
