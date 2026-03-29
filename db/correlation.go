package db

import (
	"fmt"
	"log"
	"math/bits"
	"sort"
	"strings"
	"time"
)

// Find pairs of mac addresses with almost no online time overlap

type PairInfo struct {
	Hostname1 string `json:"Hostname1"`
	Hostname2 string `json:"Hostname2"`
	Mid1      int    `json:"Mid1"`
	Mid2      int    `json:"Mid2"`
	Mac1      string `json:"Mac1"`
	Mac2      string `json:"Mac2"`
	Name1     string `json:"Name1"`
	Name2     string `json:"Name2"`
	Corr      int    `json:"Corr"`
	Jaccard   int    `json:"Jaccard"`
	Pearson   int    `json:"Pearson"`
	Slots     int    `json:"Slots"`
	Overlap   int    `json:"Overlap"`
}

func GetMacCorrelations(isCorrelatedOnly bool, macList []string) ([]PairInfo, error) {
	var results []PairInfo
	if len(macList) == 0 {
		return nil, nil
	}

	// Build the IN clause to filter at the database level
	placeholders := make([]string, len(macList))
	args := make([]any, len(macList)*2)
	for i, mac := range macList {
		placeholders[i] = "?"
		args[i] = mac
		args[i+len(macList)] = mac
	}
	inClause := strings.Join(placeholders, ",")

	corTxt := ""
	if isCorrelatedOnly {
		corTxt = "AND corr >= 0 "
	}
	sqlTxt := `SELECT C.mac1, C.mac2, C.corr, M.hostname, M.name, M.mid, N.hostname, N.name, N.mid
		FROM mac_correlation C
		JOIN macs M ON C.mac1=M.mac
		JOIN macs N ON C.mac2=N.mac
		WHERE C.mac1 < C.mac2 AND C.mac1 IN (%s) AND C.mac2 IN (%s) %s
		ORDER BY M.name, M.hostname, C.corr`
	query := fmt.Sprintf(sqlTxt, inClause, inClause, corTxt)
	rows, err := Conn.Query(query, args...)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pair PairInfo
		if err := rows.Scan(&pair.Mac1, &pair.Mac2, &pair.Corr, &pair.Hostname1, &pair.Name1, &pair.Mid1, &pair.Hostname2, &pair.Name2, &pair.Mid2); err != nil {
			log.Println(err)
			continue
		}
		results = append(results, pair)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return results, nil
}

type MacCorrelationFilter struct {
	Jaccard   int  `json:"Jaccard"`   // Jaccard number x100
	Pearson   int  `json:"Pearson"`   // Pearsons number x100
	Fixed     bool `json:"Fixed"`     // false=Ignore, true=Fixed Pairs Only
	Random    bool `json:"Random"`    // false=Ignore, true=Random Pairs Only
	Hostnames bool `json:"Hostnames"` // false=Ignore Hostnames, true=Hostnames must match
	Jsign     bool `json:"Jsign"`     // false=less than, true=greater than
	Psign     bool `json:"Psign"`     // false=less than, true=greater than
}

// Find devices that their online time matches almost perfectly
func GetMacCorrelation(filter MacCorrelationFilter) ([]PairInfo, error) {
	var results []PairInfo
	jsign := "<"
	if filter.Jsign {
		jsign = ">"
	}
	psign := "<"
	if filter.Psign {
		psign = ">"
	}

	var query strings.Builder
	query.WriteString(`SELECT C.mac1, C.mac2, M.hostname, M.name, 
		M.mid, N.hostname, N.name, N.mid, C.corr, C.jaccard, C.pearson
		FROM mac_correlation C
		JOIN macs M ON C.mac1=M.mac
		JOIN macs N ON C.mac2=N.mac
		WHERE M.isSolitary=0 AND N.isSolitary=0 AND M.isIgnore=0 AND N.isIgnore=0`)
	query.WriteString(` AND C.jaccard `)
	query.WriteString(jsign)
	query.WriteString(` ? AND C.pearson `)
	query.WriteString(psign)
	query.WriteString(` ?`)

	// Okay, chat gpt tells me if the overlap is less than 2 timeslots, then these could be the same device.
	//	query.WriteString(` AND C.overlap <= 2`)
	/*
		Minimum overlap slots     ≥ 10
		Pearsons similarity       ≥ 0.90
		Jaccard similarity        ≥ 0.70
		Simultaneous occurrences  ≤ 2

		Prevent impossible merges:
		If two MACs appear simultaneously often, reject merging them
		simultaneous_count > 3 → different device
		If both MACs have native macs (non-random), be more conservative.
	*/

	if filter.Fixed {
		query.WriteString(` AND (M.isRandomMac=0 AND N.isRandomMac=0)`)
	} else if filter.Random {
		query.WriteString(` AND (M.isRandomMac=1 AND N.isRandomMac=1)`)
	}
	if filter.Hostnames {
		query.WriteString(` AND M.hostname=N.hostname`)
	}
	rows, err := Conn.Query(query.String(), filter.Jaccard, filter.Pearson)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pair PairInfo
		if err := rows.Scan(&pair.Mac1, &pair.Mac2, &pair.Hostname1, &pair.Name1, &pair.Mid1, &pair.Hostname2, &pair.Name2, &pair.Mid2, &pair.Corr, &pair.Jaccard, &pair.Pearson); err != nil {
			log.Println(err)
			continue
		}
		results = append(results, pair)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return results, nil
}

// Find random macs that have minimal slot overlap but proceed in a time-series sequence.
// MAC1 ↔ MAC2
// MAC2 ↔ MAC3
// MAC3 ↔ MAC4
// Device 1 = {MAC1, MAC2, MAC3, MAC4}
func MacRotations() ([]string, error) {
	// 1. Get all online history for random macs, grouped by mac in memory.
	type macHistory struct {
		Date int
		Am   int64
		Pm   int64
	}
	allHistory := make(map[string][]macHistory)
	// Get min/max dates for random, active, non-solitary/ignored MACs
	query := `
		SELECT O.mac, O.date, O.am, O.pm
		FROM onlinehistory O
		JOIN macs M ON O.mac = M.mac
		WHERE M.isRandomMac=1 AND M.isSolitary=0 AND M.isIgnore=0 AND (O.am > 0 OR O.pm > 0)
		ORDER BY O.mac, O.date
	`
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mac string
		var h macHistory
		if err := rows.Scan(&mac, &h.Date, &h.Am, &h.Pm); err != nil {
			log.Println("Error scanning online history for mac rotations:", err)
			continue
		}
		allHistory[mac] = append(allHistory[mac], h)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error iterating online history for mac rotations:", err)
		return nil, err
	}

	// Helper to convert YYYYMMDD to days since a fixed epoch.
	var epoch = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	dateToDays := func(yyyymmdd int) int64 {
		t := time.Date(yyyymmdd/10000, time.Month((yyyymmdd%10000)/100), yyyymmdd%100, 0, 0, 0, 0, time.UTC)
		return int64(t.Sub(epoch).Hours() / 24)
	}

	// 2. Create precise time ranges based on 15-minute slots.
	type macSlotRange struct {
		Mac       string
		StartSlot int64
		EndSlot   int64
	}
	slotRanges := make([]macSlotRange, 0, len(allHistory))

	for mac, history := range allHistory {
		if len(history) == 0 {
			continue
		}

		// Find start slot from the first day's record
		firstDayHistory := history[0]
		days := dateToDays(firstDayHistory.Date)
		var startSlotOffset int
		if firstDayHistory.Am != 0 {
			startSlotOffset = bits.TrailingZeros64(uint64(firstDayHistory.Am))
		} else { // It must be in Pm since we filtered for am>0 or pm>0
			startSlotOffset = 48 + bits.TrailingZeros64(uint64(firstDayHistory.Pm))
		}
		startSlot := days*96 + int64(startSlotOffset)

		// Find end slot from the last day's record
		lastDayHistory := history[len(history)-1]
		days = dateToDays(lastDayHistory.Date)
		var endSlotOffset int
		if lastDayHistory.Pm != 0 {
			endSlotOffset = 48 + (63 - bits.LeadingZeros64(uint64(lastDayHistory.Pm)))
		} else { // It must be in Am
			endSlotOffset = 63 - bits.LeadingZeros64(uint64(lastDayHistory.Am))
		}
		endSlot := days*96 + int64(endSlotOffset)

		slotRanges = append(slotRanges, macSlotRange{
			Mac:       mac,
			StartSlot: startSlot,
			EndSlot:   endSlot,
		})
	}

	// 3. Sort ranges by their start slot to optimize sequence detection.
	sort.Slice(slotRanges, func(i, j int) bool {
		return slotRanges[i].StartSlot < slotRanges[j].StartSlot
	})

	// 4. Find sequences with minimal gap/overlap.

	type Gaps struct {
		Mac1        string
		Mac2        string
		GapDuration int
	}
	gaps := make([]Gaps, 0)
	chains := make([]string, 0)
	const maxGapSlots = 3 * 96     // 3 days worth of 15-min slots
	const maxOverlapSlots = 2 * 96 // 2 days worth of 15-min slots

	slotToTime := func(slot int64) time.Time {
		days := slot / 96
		slotInDay := slot % 96
		return epoch.Add(time.Duration(days)*24*time.Hour + time.Duration(slotInDay)*15*time.Minute)
	}

	for i := 0; i < len(slotRanges); i++ {
		a := slotRanges[i]
		for j := i + 1; j < len(slotRanges); j++ {
			b := slotRanges[j]

			slotDiff := b.StartSlot - a.EndSlot

			// Since the list is sorted by StartSlot, if b's start is too far ahead, subsequent ones will be too.
			if slotDiff > maxGapSlots {
				break
			}

			// Check for continuity: -maxOverlapSlots <= Gap <= maxGapSlots
			if slotDiff >= -maxOverlapSlots {
				startA := slotToTime(a.StartSlot).Format("2006-01-02 15:04")
				endA := slotToTime(a.EndSlot).Format("2006-01-02 15:04")
				startB := slotToTime(b.StartSlot).Format("2006-01-02 15:04")
				endB := slotToTime(b.EndSlot).Format("2006-01-02 15:04")
				gapDuration := (time.Duration(slotDiff) * 15 * time.Minute).Round(time.Minute)
				if gapDuration.Minutes() >= -31.0 && gapDuration.Minutes() <= 31.0 {
					if gapDuration.Minutes() < 0.0 {
						gaps = append(gaps, Gaps{Mac1: b.Mac, Mac2: a.Mac, GapDuration: int(-gapDuration.Minutes())})
					} else {
						gaps = append(gaps, Gaps{Mac1: a.Mac, Mac2: b.Mac, GapDuration: int(gapDuration.Minutes())})
					}
				}
				chains = append(chains, fmt.Sprintf("%s (%s to %s) → %s (%s to %s) [Gap: %v]", a.Mac, startA, endA, b.Mac, startB, endB, gapDuration))
			}
		}
	}
	// Reduce list to mac pairs where the gap is less than two timeslots (+/- 30 mins) since we have negative gaps.
	_, err = Conn.Exec("DELETE FROM chains")
	if err != nil {
		log.Println(err)
	}
	for i := range gaps {
		_, err = Conn.Exec("INSERT INTO chains (mac1, mac2, gap) VALUES (?, ?, ?)", gaps[i].Mac1, gaps[i].Mac2, gaps[i].GapDuration)
		if err != nil {
			log.Println(err)
		}
	}

	return chains, nil
}

// Multi-nic Detection
/*
Simple Multi-NIC Detection Rule

A practical rule set:

Jaccard ≥ 0.85
Piccard ≥ 0.95
simultaneous_ratio ≥ 0.80


If all are true:

→ very likely same physical device (multi NIC)
*/
