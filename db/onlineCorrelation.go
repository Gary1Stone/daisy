package db

import (
	"log"
	"math"
	"math/bits"
	"sort"
	"sync"
)

/**************************************************************************
 *
 * Mac correlation table
 *
 * This runs every time new data comes in from the probe(s)
 * Find pairs of mac addresses with how much online time overlap
 * NOTE: Values in the database have been multiplied by 100
 * and stored as integers
 *
 * Online Overlap (corr):
 * 0 no Overlap
 * <1% small overlap...
 *
 * Pearson:
 * > 0.90	    extremely likely same device/user
 * 0.80–0.90	strong similarity
 * 0.65–0.80	possible
 * < 0.50	    weak
 * < -0.6       high advoidance
 *
 * Jaccard:
 * > 0.85	    almost certainly same device or user
 * 0.70–0.85	strong
 * 0.50–0.70	moderate
 * < 0.40	    weak
 *
 * Best production signal:
 * flag if:
 *     Jaccard > 0.75
 * AND Pearson > 0.80
 * This combination is extremely reliable for:
 * - MAC randomization detection
 * - same-phone detection
 * - carried-together devices
 * ************************************************************************/

var (
	correlationMutex     sync.Mutex
	isCorrelationRunning bool
)

// Get the MAC addresses online information for each site
func BuildMacCorrelationTable(site string) {
	correlationMutex.Lock()
	if isCorrelationRunning {
		correlationMutex.Unlock()
		return
	}
	isCorrelationRunning = true
	correlationMutex.Unlock()

	defer func() {
		correlationMutex.Lock()
		isCorrelationRunning = false
		correlationMutex.Unlock()
	}()

	if site == "" {
		log.Println("No site provided for correlation")
		return
	}

	// 1. Load data using the logic that respects aliases.
	onlineInfo, err := getOnlineHistoryInfo(site)
	if err != nil {
		log.Printf("Error calculating correlations for site %s: %v.", site, err)
		return
	}
	if len(onlineInfo) == 0 {
		return // No data, no correlations.
	}

	// 2. Collect unique dates and MACs, and group data by MAC.
	dates, macs, byMac := organizeOnlineInfo(onlineInfo)

	// 3. Open a transaction to save results iteratively.
	tx, err := Conn.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return
	}
	defer tx.Rollback()

	// 4. Wipe the table for the site first.
	if _, err := tx.Exec("DELETE FROM mac_correlation WHERE site=?", site); err != nil {
		log.Println("Error deleting from mac_correlation:", err)
		return
	}

	// 5. Prepare statement for inserting/updating correlation results.
	query := `
		INSERT INTO mac_correlation (mac1, mac2, corr, pearson, jaccard, site, slots, overlap)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(mac1, mac2) DO UPDATE SET corr = excluded.corr,
		pearson = excluded.pearson, jaccard = excluded.jaccard, site = excluded.site`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println("Error preparing statement:", err)
		return
	}
	defer stmt.Close()

	// Pre-allocate two vectors to reuse memory, preventing millions of allocations
	vecSize := len(dates) * 96
	vec1 := make([]float64, vecSize)
	vec2 := make([]float64, vecSize)

	// 6. Compute and save correlations for all pairs iteratively.
	for i := range macs {
		mac1 := macs[i]
		fillVector(mac1, dates, byMac, vec1)

		for j := i + 1; j < len(macs); j++ {
			mac2 := macs[j]
			fillVector(mac2, dates, byMac, vec2)

			// Pearson correlation from vectors.
			pearsonScore := pearson(vec1, vec2)

			// Overlap-like correlation calculated more directly.
			overlapScore, slots, overlap := calculateOverlap(mac1, mac2, dates, byMac)

			// Jaccard similarity score.
			jaccardScore := calculateJaccard(mac1, mac2, dates, byMac)

			// Save result for this pair
			corrInt := int(overlapScore)
			if overlapScore == -1 {
				corrInt = -1
			}
			pearsonInt := int(pearsonScore * 100)
			jaccardInt := int(jaccardScore * 100)
			if jaccardScore == -1 {
				jaccardInt = -1
			}

			_, err := stmt.Exec(mac1, mac2, corrInt, pearsonInt, jaccardInt, site, slots, overlap)
			if err != nil {
				log.Printf("Error inserting/updating correlation for %v vs %v: %v. Rolling back.", mac1, mac2, err)
				return // The deferred Rollback will execute.
			}
		}
	}

	// 7. Commit the transaction.
	if err := tx.Commit(); err != nil {
		log.Printf("Error saving correlation results for site %s: %v", site, err)
	}
}

func organizeOnlineInfo(onlineInfo []Online) ([]int, []string, map[string]map[int]Online) {
	dateSet := make(map[int]struct{})
	macSet := make(map[string]struct{})
	byMac := make(map[string]map[int]Online)

	for _, item := range onlineInfo {
		dateInt := int(item.Date)
		dateSet[dateInt] = struct{}{}
		macSet[item.Mac] = struct{}{}

		if _, ok := byMac[item.Mac]; !ok {
			byMac[item.Mac] = make(map[int]Online)
		}
		byMac[item.Mac][dateInt] = item
	}

	dates := make([]int, 0, len(dateSet))
	for date := range dateSet {
		dates = append(dates, date)
	}
	sort.Ints(dates)

	macs := make([]string, 0, len(macSet))
	for mac := range macSet {
		macs = append(macs, mac)
	}
	sort.Strings(macs)

	return dates, macs, byMac
}

// fillVector populates a pre-allocated feature vector for a single MAC address.
func fillVector(mac string, dates []int, byMac map[string]map[int]Online, vec []float64) {
	// Zero out the reused vector
	for i := range vec {
		vec[i] = 0
	}

	history, macExists := byMac[mac]
	if !macExists {
		return // All zeros
	}

	for i, day := range dates {
		if r, ok := history[day]; ok {
			slots := vec[i*96 : (i+1)*96]
			bitmapToSlots(r.Am, r.Pm, slots)
		}
	}
}

func bitmapToSlots(am, pm int64, dst []float64) {
	for i := range 48 {
		if (am>>i)&1 != 0 {
			dst[i] = 1
		}
	}
	for i := range 48 {
		if (pm>>i)&1 != 0 {
			dst[i+48] = 1
		}
	}
}

func calculateOverlap(mac1, mac2 string, dates []int, byMac map[string]map[int]Online) (float64, int, int) {
	var sameTime, overlap int
	history1, ok1 := byMac[mac1]
	history2, ok2 := byMac[mac2]
	if !ok1 || !ok2 {
		return -1, 0, 0
	}

	for _, date := range dates {
		d1, ok1 := history1[date]
		d2, ok2 := history2[date]

		if !ok1 || !ok2 {
			continue
		}

		if d1.Am > 0 && d2.Am > 0 {
			sameTime++
			overlap += bits.OnesCount64(uint64(d1.Am & d2.Am))
		}
		if d1.Pm > 0 && d2.Pm > 0 {
			sameTime++
			overlap += bits.OnesCount64(uint64(d1.Pm & d2.Pm))
		}
	}

	const minSameTime = 10
	if sameTime < minSameTime {
		return -1, sameTime, overlap
	}

	if sameTime == 0 {
		return 0, sameTime, overlap
	}

	// TODO: QUESTION: Do we simply want to store the overlap count instead of the percentage? It may be more useful in the after-analyis
	return (float64(overlap) / float64(sameTime*48)) * 100.0, sameTime, overlap
}

func calculateJaccard(mac1, mac2 string, dates []int, byMac map[string]map[int]Online) float64 {
	var intersection, union int
	var sameTime int // for min threshold
	history1, ok1 := byMac[mac1]
	history2, ok2 := byMac[mac2]
	if !ok1 || !ok2 {
		return -1
	}

	for _, date := range dates {
		d1, ok1 := history1[date]
		d2, ok2 := history2[date]

		if !ok1 && !ok2 {
			continue
		}

		if d1.Am > 0 && d2.Am > 0 {
			sameTime++
		}
		intersection += bits.OnesCount64(uint64(d1.Am & d2.Am))
		union += bits.OnesCount64(uint64(d1.Am | d2.Am))

		if d1.Pm > 0 && d2.Pm > 0 {
			sameTime++
		}
		intersection += bits.OnesCount64(uint64(d1.Pm & d2.Pm))
		union += bits.OnesCount64(uint64(d1.Pm | d2.Pm))
	}

	const minSameTime = 10
	if sameTime < minSameTime {
		return -1
	}

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func pearson(x, y []float64) float64 {
	n := len(x)
	if n == 0 || len(y) != n {
		return 0
	}

	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := range n {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	num := float64(n)*sumXY - sumX*sumY
	den := math.Sqrt((float64(n)*sumX2 - sumX*sumX) * (float64(n)*sumY2 - sumY*sumY))

	if den == 0 {
		return 0
	}
	return num / den
}

// Ensure to eliminate macs coveed by alias, or not active, and is not identified as solitary
func getOnlineHistoryInfo(site string) ([]Online, error) {
	var items []Online
	query := `
		SELECT COALESCE(A.alias, O.mac) as mac, O.date, O.am, O.pm
		FROM online O
		JOIN macs M ON O.mac=M.mac
		LEFT JOIN aliases A ON O.mac = A.alias
		WHERE M.active=1 AND M.isSolitary=0 AND M.isIgnore=0 AND M.site=? AND (O.am > 0 OR O.pm > 0 )
		ORDER BY O.date`
	rows, err := Conn.Query(query, site)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Online
		if err := rows.Scan(&item.Mac, &item.Date, &item.Am, &item.Pm); err != nil {
			log.Println(err)
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	scrubbed, err := scrubOnline(items) // remove duplicate macs
	if err != nil {
		return nil, err
	}

	return scrubbed, nil
}
