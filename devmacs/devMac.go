package devmacs

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/gbsto/daisy/db"
	"github.com/gbsto/daisy/photos"
)

type DeviceInfo struct {
	Cid        int    `json:"cid"`        // Computer ID
	Name       string `json:"name"`       // Computer name
	Model      string `json:"model"`      // Computer model
	Kind       string `json:"kind"`       // Computer role (switch, desktop, router...) of the device
	KindName   string `json:"kindname"`   // User visible description of the kind
	Icon       string `json:"icon"`       // Name of svg icon for the role/kind
	Site       string `json:"site"`       //
	SiteName   string `json:"sitename"`   // The node's site
	Office     string `json:"office"`     //
	OfficeName string `json:"officename"` // User visible description of the office
	IsLinked   bool   `json:"islinked"`   // Does the computer have any MACs associated with it
	IsOnline   bool   `json:"isonline"`   // Is the device currently online
	Image      string `json:"image"`      // Photo of device
}

const deviceQuery = `SELECT D.cid, D.name, D.model, COALESCE(D.kind, '') AS kind, COALESCE(K.description, '') AS kindname, 
		site, COALESCE(S.description, "") AS sitename, Office, COALESCE(O.description, '') AS officeName,  
		COALESCE(I.icon,'') AS icon, COALESCE(D.image, 'missing.jpg') AS Image
		FROM devices D
		LEFT JOIN choices AS S ON S.field='SITE' AND S.code=D.SITE
		LEFT JOIN choices AS O ON O.field='OFFICE' AND O.code=D.OFFICE AND O.parent=D.site
		LEFT JOIN choices AS K ON K.field='KIND' AND K.code=D.kind
		LEFT JOIN icons AS I ON I.name=D.kind
		WHERE D.active=1 `

type MacInfo struct {
	Mid         int    `json:"mid"`         // MAC ID
	Mac         string `json:"mac"`         //
	MacName     string `json:"name"`        // Computer Name (not hostname)
	Cid         int    `json:"cid"`         //
	Parent      int    `json:"parent"`      //
	Created     string `json:"created"`     //
	Hostname    string `json:"hostname"`    //
	IP          string `json:"ip"`          //
	Kind        string `json:"kind"`        //
	Os          string `json:"os"`          //
	User        string `json:"user"`        //
	Site        string `json:"site"`        //
	Office      string `json:"office"`      //
	Location    string `json:"location"`    //
	Note        string `json:"note"`        //
	Scanned     int    `json:"scanned"`     //
	Vendor      string `json:"vendor"`      //
	IsOnline    bool   `json:"isonline"`    //
	Source      string `json:"source"`      //
	Intruder    bool   `json:"intruder"`    //
	Updated     int    `json:"updated"`     //
	Active      bool   `json:"active"`      //
	IsSolitary  bool   `json:"issolitary"`  //
	IsRandomMac bool   `json:"israndommac"` //
	IsIgnore    bool   `json:"isignore"`    //
	Icon        string `json:"icon"`
}

func (d *DeviceInfo) GetDevice(cid int) (item DeviceInfo, err error) {
	if cid < 1 {
		err = errors.New("invalid id")
		return
	}
	photoMap := photos.MapPhotos() //Map (name,true/false) of photos in the /public/images directory
	err = db.Conn.QueryRow(deviceQuery+` AND D.cid=?`, cid).Scan(&item.Cid, &item.Name, &item.Model, &item.Kind, &item.KindName,
		&item.Site, &item.SiteName, &item.Office, &item.OfficeName, &item.Icon, &item.Image)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Println(err)
		return
	}
	item.Image = photos.AddSuffixBeforeExtension(item.Image, "-sm")
	if len(item.Image) == 0 || !photoMap[item.Image] {
		item.Image = "missing-sm.jpg"
	}

	return
}

func (d *DeviceInfo) GetDevices(site string) (items []DeviceInfo, err error) {
	rows, err := db.Conn.Query(deviceQuery+` AND D.site=? ORDER BY D.Name`, site)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item DeviceInfo
		err = rows.Scan(&item.Cid, &item.Name, &item.Model, &item.Kind, &item.KindName, &item.Site, &item.SiteName, &item.Office, &item.OfficeName, &item.Icon, &item.Image)
		if err != nil {
			log.Println(err)
			return
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}
	return
}

// Find all the macs for this cid, and count the number of entries in the online table for each mac
// Get all the macs from the device_mac table. This table is filled from the audit program.
// So it contains what the computer's OS is telling 'audit' the network NICs are.
// Get all macs from the macs table for this device. This table is filled from the network probe (called laptops).
// The CID was assigned by device_mac table and by matching the network hostname to the device record name, or manually assigned by a user.
// Find which mac has records in the online table
// Find the max count in macs[]
func (d *DeviceInfo) GetDevicePrimaryMac(cid int) (retMac string, err error) {
	macs := make(map[string]int)
	maxC := 0
	query := `SELECT mac, count(*) AS cnt FROM online WHERE mac IN
		(SELECT mac FROM device_mac WHERE cid=? UNION SELECT mac FROM macs WHERE cid=? AND active=1)
		GROUP BY mac`
	rows, err := db.Conn.Query(query, cid, cid)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		mac := ""
		cnt := 0
		err = rows.Scan(&mac, &cnt)
		if err != nil {
			log.Println(err)
			return
		}
		if cnt > 0 {
			macs[mac] = cnt
			if cnt > maxC {
				maxC = cnt
				retMac = mac
			}
		}
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
	}
	return // TODO: return the macs[mac]count and combine all plots into one.
}


type DeviceMacs struct {
	Mac         string `json:"mac"`            // MAC  address
	Mid         int    `json:"mid"`            //    // If = 0, then it is not not in the MACS table, only the self-reporting table
	Hostname    string `json:"hostname"`       //
	Ip          string `json:"ip"`             //
	Kind        string `json:"kind"`           //
	KindName    string `json:"kindname"`       //
	Icon        string `json:"icon"`           //
	IsOnline    bool   `json:"isonline"`       // // Seen online in the last 15 minutes
	IsDeletable bool   `json:"isnotdeletable"` //
	FirstSeen   string `json:"firstseen"`      //
	LastSeen    string `json:"lastseen"`       //
}

const DeviceMacsQuery = `
		SELECT M.mid, M.mac, COALESCE(M.hostname, '') AS hostname, 
		COALESCE(M.ip, '') AS ip, 
		COALESCE(M.kind, '') AS kind,  COALESCE(K.description, '') AS kindname,
		COALESCE(I.icon, '') AS icon, 
		COALESCE(M.online, 0) AS online,
		strftime('%Y-%m-%d', M.created-?, 'unixepoch') AS firstseen,
		strftime('%Y-%m-%d', M.scanned-?, 'unixepoch') AS lastseen
		FROM macs M
		LEFT JOIN icons AS I ON I.name = M.kind
		LEFT JOIN choices AS K ON K.field='KIND' AND K.code=M.kind `

func GetMacInfoByCid(cid, tzoff int) (items []DeviceMacs, err error) {
	// First read all the macs for the computer that are in the self-reporting (audit.exe) table
	query := `
		SELECT COALESCE(M.mid, 0) AS mid, D.mac, COALESCE(M.hostname, '') AS hostname, 
		COALESCE(M.ip, '') AS ip, 
		COALESCE(M.kind, '') AS kind, COALESCE(K.description, '') AS kindname, 
		COALESCE(I.icon, '') AS icon, COALESCE(M.online, 0) as online
		FROM device_mac D
		LEFT JOIN macs M ON M.Mac=D.mac
		LEFT JOIN icons I ON I.name = M.kind
		LEFT JOIN choices AS K ON K.field='KIND' AND K.code=M.kind
		WHERE D.cid=?
		`
	rows, err := db.Conn.Query(query, cid)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item DeviceMacs
		err = rows.Scan(&item.Mid, &item.Mac, &item.Hostname, &item.Ip, &item.Kind, &item.KindName, &item.Icon, &item.IsOnline)
		if err != nil {
			log.Println(err)
			return
		}
		item.IsDeletable = false // from the self reporting table, so cannot delete these
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}

	//Second, find all cid macs in the Macs table that are not in the mac_device table
	query = DeviceMacsQuery + "WHERE M.cid=? AND M.mac NOT IN (SELECT mac FROM device_mac WHERE cid=?)"
	rows, err = db.Conn.Query(query, tzoff, tzoff, cid, cid)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item DeviceMacs
		err = rows.Scan(&item.Mid, &item.Mac, &item.Hostname, &item.Ip, &item.Kind, &item.KindName, &item.Icon, &item.IsOnline, &item.FirstSeen, &item.LastSeen)
		if err != nil {
			log.Println(err)
			return
		}
		item.IsDeletable = true
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}

	// Sort by mac alphabetically
	sort.Slice(items, func(i, j int) bool {
		return items[i].Mac < items[j].Mac
	})

	return
}

// Exclude randomized macs
func GetUnmatchedMacs(cid, tzoff int, site string) (items []DeviceMacs, err error) {
	query := DeviceMacsQuery + "WHERE M.cid IS NULL AND site=? AND M.israndomMac=0 AND M.mac NOT IN (SELECT mac FROM device_mac WHERE cid=?) AND M.kind NOT IN ('CELLPHONE', 'TABLET', 'WATCH', 'PTR')"
	rows, err := db.Conn.Query(query, tzoff, tzoff, site, cid)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item DeviceMacs
		err = rows.Scan(&item.Mid, &item.Mac, &item.Hostname, &item.Ip, &item.Kind, &item.KindName, &item.Icon, &item.IsOnline, &item.FirstSeen, &item.LastSeen)
		if err != nil {
			log.Println(err)
			return
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}
	return
}

type Associations struct {
	Mid           int    `json:"mid"`            // Alreday Attached Mid
	Mac           string `json:"mac"`            // Already Attached Mac
	Corr          int    `json:"corr"`           // Precentage Overlap: This is the number of AM/PM slots (48 15-minute intervals) the two devices are online together within the same AM or PM interval
	Pearson       int    `json:"pearson"`        // Pearsons number x100. Range -1 to +1. +1 is a linear relationship. O is no relationship. -1 is an oposite relationship, as one goes up the other goes down.
	Jaccard       int    `json:"jaccard"`        // Jaccard number x100. Range 0 (no overlap) to 1 (complete match)
	Slots         int    `json:"slots"`          // Number of AM and PM intervals where the two devices appear online within that 15-minute interval (not necessarly at the same time)
	OverLap       int    `json:"overlap"`        // Overlap is number of times the devices are online at the same time, within the same 15-minute intervals (slots)
	ComparedToMid int    `json:"comparedtomid"`  // MAC's ID compared against
	ComparedToMac string `json:"comparedtomac"`  // MAC compared against
	Recomendation string `json:"recommendation"` // Interpreted results
}

// Get a list of the device's MACs and their correlation data
func GetMacAssociations(cid, mid int) (items []Associations, err error) {
	query := `SELECT M.mid, M.mac, C.mac1, C.mac2, C.corr, C.jaccard, C.pearson, C.slots, C.overlap
		FROM macs M
		LEFT JOIN mac_correlation C ON (C.mac1 = M.mac OR C.mac2 = M.mac)
		AND (C.mac1 = (SELECT mac FROM macs WHERE mid=?)
				OR C.mac2 = (SELECT mac FROM macs WHERE mid=?))
		WHERE M.cid=?`
	rows, err := db.Conn.Query(query, mid, mid, cid)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item Associations
		var mac1 string
		var mac2 string
		err = rows.Scan(&item.Mid, &item.Mac, &mac1, &mac2, &item.Corr, &item.Jaccard, &item.Pearson, &item.Slots, &item.OverLap)
		if err != nil {
			log.Println(err)
			return
		}
		item.ComparedToMid = mid
		item.ComparedToMac = mac1
		if item.Mac == mac1 {
			item.ComparedToMac = mac2
		}
		item.Recomendation = recomendation(item)
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return
	}
	return
}

func recomendation(item Associations) string {

	DaysOverlap := 0
	query := `SELECT COUNT(DISTINCT A.date) AS DaysOverlap
		FROM online A
		JOIN online B ON B.date = A.date
		WHERE A.mac=? AND B.mac=?`
	err := db.Conn.QueryRow(query, item.Mac, item.ComparedToMac).Scan(&DaysOverlap)
	if err != nil {
		log.Println(err)
	}

	r := RecommendSameComputer(float64(item.Jaccard)/100.0, float64(item.Pearson)/100.0, float64(item.Corr), item.Slots, DaysOverlap)

	return fmt.Sprintf("%+v\n", r)
}

type MACRecommendation struct {
	Score      float64
	Decision   string
	Confidence string
	Reason     string
}

func RecommendSameComputer(jaccard, pearson, overlapPercent float64, counts, daysOverlap int) MACRecommendation {
	// a jaccard or overlapPercent of -1 means missing mac or not enough data points (< 10)

	// Validate input.
	if math.IsNaN(jaccard) ||
		math.IsNaN(pearson) ||
		math.IsNaN(overlapPercent) ||
		math.IsInf(jaccard, 0) ||
		math.IsInf(pearson, 0) ||
		math.IsInf(overlapPercent, 0) ||
		jaccard < 0 || jaccard > 1 ||
		pearson < -1 || pearson > 1 ||
		overlapPercent < 0 || overlapPercent > 100 ||
		counts < 0 || daysOverlap < 0 {

		return MACRecommendation{
			Decision:   "INSUFFICIENT DATA",
			Confidence: "None",
			Reason:     "No concurrent activity or no overlapping days.",
		}
	}

	if counts == 0 || daysOverlap == 0 {
		return MACRecommendation{
			Decision:   "INSUFFICIENT DATA",
			Confidence: "Very low",
			Reason:     "No concurrent activity or no overlapping days.",
		}
	}

	// Normalize the measurements to 0-1.
	overlap := overlapPercent / 100
	p := (pearson + 1) / 2

	// Base similarity score.
	// Jaccard and overlap are related, so avoid overweighting
	// either measurement. Pearson receives less weight because
	// similar schedules can produce high Pearson correlation.
	score := 100 * (0.45*jaccard + 0.15*p + 0.40*overlap)

	// Evidence from concurrent 15-minute time slots.
	countFactor := 1.0
	switch {
	case counts < 8:
		countFactor = 0.45
	case counts < 24:
		countFactor = 0.65
	case counts < 64:
		countFactor = 0.82
	}

	// Persistence across distinct days.
	// Repeated evidence is more convincing than a single day.
	dayFactor := 1.0
	switch {
	case daysOverlap < 2:
		dayFactor = 0.55
	case daysOverlap < 4:
		dayFactor = 0.70
	case daysOverlap < 7:
		dayFactor = 0.85
	}

	score *= countFactor * dayFactor
	score = math.Round(score*10) / 10

	result := MACRecommendation{
		Score:      score,
		Decision:   "UNCERTAIN",
		Confidence: "Moderate",
		Reason:     "Similarity evidence is inconclusive.",
	}

	// Evidence thresholds.
	switch {
	case counts < 8 || daysOverlap < 2:
		result.Decision = "INSUFFICIENT DATA"
		result.Confidence = "Very low"
		result.Reason = "Too little concurrent activity across days."

	case pearson >= 0.80 &&
		jaccard < 0.45 &&
		overlapPercent < 35:
		result.Decision = "SIMILAR SCHEDULE"
		result.Confidence = "Moderate"
		result.Reason = "Patterns correlate, but direct overlap is weak."

	case score >= 80 &&
		jaccard >= 0.70 &&
		overlapPercent >= 60 &&
		counts >= 64 &&
		daysOverlap >= 7:
		result.Decision = "SAME COMPUTER"
		result.Confidence = "High"
		result.Reason = "Strong similarity and concurrent activity repeated across days."

	case score >= 62 &&
		jaccard >= 0.45 &&
		overlapPercent >= 35 &&
		counts >= 24 &&
		daysOverlap >= 3:
		result.Decision = "PROBABLY SAME COMPUTER"
		result.Confidence = "Moderate"
		result.Reason = "Multiple indicators support a match across several days."

	case score < 30:
		result.Decision = "PROBABLY DIFFERENT COMPUTERS"
		result.Confidence = "Moderate"
		result.Reason = "The observed patterns provide little evidence of a match."

	default:
		result.Decision = "UNCERTAIN"
		result.Confidence = "Low to moderate"
		result.Reason = "More observations are needed to distinguish device identity from similar schedules."
	}

	return result
}

/*

Write a Go code function that receives 4 attributes: Jaccard similarity, Pearsons similarity, Overlap (in percent) and Counts, then determins a recommendation if the two MACs are from the same computer or not.
Consider the Overlap and Counts are measured in 15-minute intervals, where the Counts are the number of timelots that have both MACs online in AM or PM, and the Overlap is the number of 15-minute intervqals where they were concurrenly online.
include factors such as similar schedules for different devices.

For your MAC correlation system, I would use a weighted scoring function with confidence checks, rather than simply averaging Jaccard similarity, Pearson similarity, overlap, and counts.

The important distinction is that two MACs being online at the same time does not necessarily mean they belong to the same computer. Two different computers used by the same person, for example, may have very similar schedules.

Your measurements provide four useful signals:
- Jaccard similarity: How similar the two MACs' overall online time slots are, accounting for slots when either one was online.
- Pearson similarity: Whether their online/offline patterns move together. This can be high for devices with similar schedules, even when they aren't always online together.
- Overlap percentage: How frequently they are online concurrently. Its exact meaning depends on the denominator used to calculate it.
- Counts: The number of 15-minute time slots in which both MACs were online. This helps establish how much evidence supports the similarity scores.
There are 96 15-minute intervals per day, or 48 each for AM and PM. A count of 20 concurrent intervals represents five hours of simultaneous online activity.

I recommend three possible outcomes: same computer, probably same computer, and insufficient evidence / probably different computers. The function should also identify cases where high similarity may simply reflect similar schedules.
Go implementation
This function returns a score, a recommendation, and an explanation. The thresholds are starting points that you can tune using known MAC pairs from your network.

One assumption: I treat counts as the number of concurrent 15-minute intervals, and overlapPercent as a separate percentage measure. If your overlap percentage is actually calculated from the same count and observation period, those two inputs are not independent and should not receive full weight separately.




How I would interpret the results
Pattern	Likely interpretation
High Jaccard, high Pearson, high overlap, many counts	Strong candidate for the same computer
High Pearson, moderate or low Jaccard	Possibly two computers following the same schedule
High overlap percentage but very few counts	Too little evidence; avoid automatic matching
High Jaccard and overlap over many days	Stronger evidence of a shared device or closely linked interfaces
Low Pearson, low Jaccard, low overlap	Probably different computers
Important refinements for your network
1. Use observations spread across multiple days. A count of 100 intervals could represent a single day or several weeks. Those are not equally strong evidence. If possible, pass the number of distinct days as another input, or require evidence spanning several days before returning SAME COMPUTER.

2. Compare the actual online/offline sequences. Two computers used by the same person may have similar morning and afternoon patterns. If both MACs repeatedly switch online and offline in the same 15-minute intervals, that is more persuasive than merely having similar daily schedules. A further useful measure is the percentage of online/offline transitions that occur together.

3. Avoid treating Pearson as proof of identity. Pearson measures how closely patterns vary together; it does not establish that the devices are physically the same computer. That's why the function requires strong Jaccard and overlap evidence for its strongest recommendation.

4. Calibrate against known pairs. Collect examples of MACs you know belong to the same computer and MACs belonging to different computers. The best thresholds for your shelter's network will depend on how frequently devices remain online, sleep, disconnect, or move between networks.

With DaysOverlap added, I would use the number of distinct days with concurrent activity to measure how persistent the relationship is, rather than relying on the total number of matching 15-minute intervals alone.

Your clarification also matters: OverlapPercent represents the proportion of observed 15-minute time slots in which both MACs were online concurrently. Because that overlaps with the information represented by Counts, the scoring should avoid giving the same evidence too much weight.

I'll keep the function's three outcomes for likely identity, while using DaysOverlap to increase confidence only when the concurrent activity is spread across multiple days.


How DaysOverlap affects the result
Overlapping days	Effect
1	Insufficient evidence for a reliable match
2–3	Similarity score is reduced
4–6	Moderate evidence of persistence
7 or more	Full day-based scoring factor; still requires strong similarity
These are initial thresholds, not statistically calibrated probabilities. The returned score is a ranking score, not a probability that the MACs belong to the same computer.

One important caveat: this version assumes OverlapPercent is a valid measure of concurrent online activity. If Counts is the numerator used to calculate that same percentage, the two measurements are correlated. Also, DaysOverlap should count distinct calendar days, not AM and PM separately.

For your network, I would eventually distinguish a same physical computer recommendation from a likely related device recommendation, because two computers used on the same schedule can still look highly correlated.

*/
