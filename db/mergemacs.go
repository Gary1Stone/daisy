package db

// The two macs are not same device, set the macs table isSolitary field to true (1)
func SetIsSolitaryDevices(mac1, mac2 string) error {
	tx, err := Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE macs SET isSolitary=1 WHERE mac IN (?, ?)", mac1, mac2)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func SetIsIgnoreDevices(mac1, mac2 string) error {
	tx, err := Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec("UPDATE macs SET isIgnore=1 WHERE mac IN (?, ?)", mac1, mac2)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Now We need to determine if we merge online history permantaly, or on the fly?

// I wonder if it is possible to create a virtual table that combines online with alias, replacing mac with the alias

// var MacMergeRunning = false

// // CompareOnlineHistory scans the macs table for hostnames that have multiple MAC addresses.
// // It generates a map where keys are the hostnames and values are slices of the corresponding MAC addresses.
// // This is useful for identifying potential devices with multiple network interfaces.
// func CompareOnlineHistory() {
// 	if MacMergeRunning {
// 		return
// 	}
// 	MacMergeRunning = true
// 	defer func() {
// 		MacMergeRunning = false
// 	}()
// 	hostsMap, err := findSameHosts(14) // minimum number of days that must be recorded
// 	if err != nil {
// 		log.Printf("Error finding hosts with same name: %v", err)
// 		return
// 	}

// 	for _, macs := range hostsMap {
// 		if len(macs) < 2 {
// 			continue // No pairs to compare, proceed to next iteration
// 		}

// 		// Pre-fetch history for all MACs of this hostname to avoid repetitive DB calls.
// 		histories, err := getOnlineHistory(macs)
// 		if err != nil {
// 			log.Println("Could not get histories.", err)
// 			continue
// 		}

// 		// Use a Disjoint Set Union (DSU) data structure to group matching MACs.
// 		parent := make(map[string]string)
// 		for _, mac := range macs {
// 			parent[mac] = mac // Each MAC is its own parent initially.
// 		}

// 		var find func(string) string // Recursive, so need to define it first
// 		find = func(i string) string {
// 			if parent[i] == i {
// 				return i
// 			}
// 			parent[i] = find(parent[i])
// 			return parent[i]
// 		}

// 		union := func(i, j string) {
// 			rootI := find(i)
// 			rootJ := find(j)
// 			if rootI != rootJ {
// 				parent[rootI] = rootJ
// 			}
// 		}

// 		// Compare all unique pairs and union them if they match.
// 		for i := range macs {
// 			for j := i + 1; j < len(macs); j++ {
// 				mac1, mac2 := macs[i], macs[j]
// 				if !isOverlapping(histories[mac1], histories[mac2]) {
// 					union(mac1, mac2)
// 				}
// 			}
// 		}

// 		// Group MACs by their root parent to form the final sets.
// 		sets := make(map[string][]string)
// 		for _, mac := range macs {
// 			root := find(mac)
// 			sets[root] = append(sets[root], mac)
// 		}

// 		for _, macs := range sets {
// 			// Get the oldest mac and make it the default for all operations going forward
// 			if len(macs) > 1 {
// 				oldestMac, err := getOldestMac(macs)
// 				if err != nil {
// 					log.Println(err)
// 					continue
// 				}

// 				// Merge the macs online histories into one history for the actual mac
// 				histories, err := getOnlineHistory(macs)
// 				if err != nil {
// 					log.Println(err)
// 					continue
// 				}

// 				// Remove oldestMac from this slice of macs
// 				for i, mac := range macs {
// 					if mac == oldestMac {
// 						macs = append(macs[:i], macs[i+1:]...)
// 						break
// 					}
// 				}

// 				// Add the oldestMac and it's alternative macs to the alias table
// 				var aliases Aliases
// 				err = aliases.SetAlias(oldestMac, macs)
// 				if err != nil {
// 					log.Printf("Error writing aliases: %v", err)
// 				}

// 				// Get all the macs histories and update oldestMac
// 				err = writeActualHistory(oldestMac, mergeHistories(oldestMac, histories))
// 				if err != nil {
// 					log.Printf("Error writing merged history for %s: %v", oldestMac, err)
// 					continue
// 				}

// 				// Merge the mac records into one record
// 				err = mergeMacRecords(oldestMac, macs)
// 				if err != nil {
// 					log.Printf("Error writing merged record for %s: %v", oldestMac, err)
// 					continue
// 				}

// 				// Remove the other macs from the mac table and delete their histories
// 				err = DeleteMacAndHistory(macs)
// 				if err != nil {
// 					log.Printf("Error deleting mac records: %v", err)
// 				}
// 			}
// 		}
// 	}
// }

// // Minimum of 14 days of meshing data needed before this analysis can be done
// func findSameHosts(minimumDaysMeshing int) (map[string][]string, error) {
// 	query := `
// 		SELECT  m.hostname, m.mac
// 		FROM macs m
// 		LEFT JOIN online o ON o.mac = m.mac
// 		WHERE m.hostname IN (
// 			SELECT hostname
// 			FROM macs
// 			WHERE hostname IS NOT NULL AND hostname != '' AND SUBSTR(mac, 2, 1) NOT IN ('2', '6', 'A', 'E')
// 			GROUP BY hostname
// 			HAVING COUNT(hostname) > 1
// 		)
// 		GROUP BY m.hostname, m.mac
// 		HAVING COUNT(o.mac) > ?
// 		ORDER BY m.hostname, m.mac
// 	`
// 	// Use the following query to eliminate combining any MAC addresses that are randomized (such as Apple iPhones and watches).
// 	// 		WHERE hostname IS NOT NULL AND hostname != '' AND SUBSTR(mac, 2, 1) NOT IN ('2', '6', 'A', 'E')
// 	rows, err := Conn.Query(query, minimumDaysMeshing)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	// Use a map to group MAC addresses by hostname.
// 	hostsMap := make(map[string][]string)

// 	for rows.Next() {
// 		var hostname, mac string
// 		err := rows.Scan(&hostname, &mac)
// 		if err != nil {
// 			return nil, err // Stop processing on scan error
// 		}
// 		hostsMap[hostname] = append(hostsMap[hostname], mac)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return hostsMap, nil
// }

// // Get the online histories for the list of macs
// func getOnlineHistory(macs []string) (map[string][]Online, error) {
// 	if len(macs) == 0 {
// 		return nil, nil
// 	}
// 	allHistories := make(map[string][]Online, len(macs))
// 	var histories []Online

// 	// Convert macs to []interface{} for query
// 	args := make([]any, len(macs))
// 	for i, mac := range macs {
// 		args[i] = mac
// 	}

// 	// Build placeholders and query
// 	questionMarks := strings.Repeat("?,", len(macs))
// 	questionMarks = strings.TrimSuffix(questionMarks, ",")
// 	query := fmt.Sprintf(`SELECT mac, date, am, pm FROM online WHERE mac IN (%s) ORDER BY date`, questionMarks)
// 	rows, err := Conn.Query(query, args...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var history Online
// 		err := rows.Scan(&history.Mac, &history.Date, &history.Am, &history.Pm)
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}
// 		histories = append(histories, history)
// 		allHistories[history.Mac] = histories //make map[mac] hold all the histories for each mac
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	// If no rows were found, the map will be empty.
// 	// We can return nil to indicate no duplicates were found.
// 	if len(allHistories) == 0 {
// 		return nil, nil
// 	}
// 	return allHistories, nil
// }

// // no overlapping online times
// func isOverlapping(hist1, hist2 []OnlineHistory) bool {
// 	// Create maps for quick lookup of history by date.
// 	oldestMap := make(map[int]OnlineHistory)
// 	for _, h := range hist1 {
// 		oldestMap[h.Date] = h
// 	}

// 	// Create map that holds all the other macs's date and AM/PM info
// 	hist2Map := make(map[int]OnlineHistory)
// 	for _, h := range hist2 {
// 		hist2Map[h.Date] = h
// 	}

// 	// Iterate through the dates from the first history.
// 	for date, oldest := range oldestMap {
// 		// Check if the same date exists in the second history.
// 		if h2, found := hist2Map[date]; found {
// 			if oldest.Mac == h2.Mac { // Skip the oldestMac, that is what we are comparing it against
// 				continue
// 			}
// 			amOverlap := (oldest.Am & h2.Am) != 0 // Use bitwise AND to check for overlapping "on" times.
// 			pmOverlap := (oldest.Pm & h2.Pm) != 0 // If the result is non-zero, at least one bit is set in both, indicating an overlap.
// 			if amOverlap || pmOverlap {
// 				return true // Overlap detected
// 			}
// 		}
// 	}
// 	return false // No overlap of online times
// }

// // Return the oldest mac as the actual mac
// func getOldestMac(macs []string) (string, error) {
// 	var oldMac string
// 	if len(macs) == 0 {
// 		return "", nil
// 	}
// 	if len(macs) == 1 {
// 		return macs[0], nil
// 	}
// 	// Convert macs to []interface{}
// 	args := make([]any, len(macs))
// 	for i, mac := range macs {
// 		args[i] = mac
// 	}

// 	// Build placeholders and query
// 	questionMarks := strings.Repeat("?,", len(macs))
// 	questionMarks = strings.TrimSuffix(questionMarks, ",")
// 	query := fmt.Sprintf(`SELECT mac FROM macs WHERE mac IN (%s) ORDER BY created ASC, mac ASC LIMIT 1`, questionMarks)
// 	err := Conn.QueryRow(query, args...).Scan(&oldMac)
// 	if err != nil {
// 		return "", err
// 	}
// 	return oldMac, nil
// }

// func mergeHistories(oldestMac string, histories map[string][]OnlineHistory) []OnlineHistory {
// 	merged := make(map[int]OnlineHistory)

// 	// Add all items from the oldestMac history to the new date map
// 	for _, h1 := range histories[oldestMac] {
// 		merged[h1.Date] = h1
// 	}

// 	// delete the oldestMac entry in the histories map
// 	delete(histories, oldestMac)

// 	// Add or merge items from the remaining history
// 	for _, hist := range histories {
// 		for _, h2 := range hist {
// 			if h1, ok := merged[h2.Date]; ok {
// 				h1.Am |= h2.Am // Date exists, so merge with bitwise OR (the | is bitwise OR,  0+0=0, 0+1=1, 1+0=1, 1+1=1)
// 				h1.Pm |= h2.Pm
// 				h1.Mac = oldestMac
// 				merged[h2.Date] = h1
// 			} else {
// 				h2.Mac = oldestMac
// 				merged[h2.Date] = h2 // Date is new, so just add it.
// 			}
// 		}
// 	}

// 	// Convert the map back to a slice.
// 	result := make([]OnlineHistory, 0, len(merged))
// 	for _, h := range merged {
// 		result = append(result, h)
// 	}
// 	return result
// }

// // Merge macs history
// func writeActualHistory(oldestMac string, actualHistory []OnlineHistory) error {
// 	tx, err := Conn.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback() // Rollback on error, no-op on success
// 	query := `INSERT INTO online (mac, date, am, pm) VALUES (?, ?, ?, ?)
// 		ON CONFLICT(mac, date) DO UPDATE SET am=excluded.am, pm=excluded.pm`

// 	for _, h := range actualHistory {
// 		_, err := tx.Exec(query, oldestMac, h.Date, h.Am, h.Pm)
// 		if err != nil {
// 			return fmt.Errorf("failed to upsert history for mac %s on date %d: %w", oldestMac, h.Date, err)
// 		}
// 	}
// 	return tx.Commit()
// }

// // MergeMacRecords merges data from macB into macA.
// // It fills empty string fields in macA with values from macB and takes the latest 'Scanned' time.
// // It then updates the database record for oldestMac and returns the merged struct.
// func mergeMacRecords(oldestMac string, macs []string) error {
// 	var macA MacInfo
// 	err := macA.GetMacByMac(oldestMac, 0)
// 	if err != nil {
// 		return err
// 	}

// 	var items MacInfo
// 	err = items.GetMacsByMacList(macs, 0)
// 	if err != nil {
// 		return err
// 	}

// 	for _, macB := range items.Macs {
// 		// Use reflection to merge string fields.
// 		valA := reflect.ValueOf(&macA).Elem()
// 		valB := reflect.ValueOf(macB)

// 		for i := 0; i < valA.NumField(); i++ {
// 			fieldA := valA.Field(i)
// 			// Ensure we are dealing with a string field that can be set.
// 			if fieldA.Kind() == reflect.String && fieldA.CanSet() {
// 				if fieldA.String() == "" {
// 					fieldB := valB.Field(i)
// 					fieldA.SetString(fieldB.String())
// 				}
// 			}
// 		}

// 		// Handle non-string fields manually.
// 		if macA.Scanned < macB.Scanned {
// 			macA.Scanned = macB.Scanned
// 		}
// 	}
// 	return macA.SetMac()
// }

// func DeleteMacAndHistory(macs []string) error {
// 	tx, err := Conn.Begin()
// 	if err != nil {
// 		return fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback() // Rollback on error, no-op on success

// 	for _, mac := range macs {
// 		// Delete from the online history table first.
// 		if _, err := tx.Exec("DELETE FROM online WHERE mac = ?", mac); err != nil {
// 			return fmt.Errorf("failed to delete from online table for mac %s: %w", mac, err)
// 		}

// 		// Delete from the main macs table.
// 		if _, err := tx.Exec("DELETE FROM macs WHERE mac = ?", mac); err != nil {
// 			return fmt.Errorf("failed to delete from macs table for mac %s: %w", mac, err)
// 		}
// 		log.Printf("Successfully deleted MAC %s and its history.", mac)
// 	}
// 	return tx.Commit()
// }

// // About
// // The second hexadecimal digit of a MAC address changes for randomization. If this digit is a 2, 6, A, or E, it indicates a randomized (locally administered) MAC address, as the operating system sets the "locally administered" bit in the first byte to '1'. The other 46 bits are then randomized.
// // Locally Administered (LA) bit: This bit is part of the first byte (the first octet) of the MAC address.
// // Universal vs. Local: A '0' in this bit means the address is universally administered (assigned by the manufacturer), while a '1' means it is locally administered (assigned by the device or network).
// // How it works: By setting this bit to '1' and randomizing the remaining 46 bits, the device creates a valid, yet non-unique, MAC address for its network connection.
// // Example: A randomized MAC address might start with 02:XX:XX:XX:XX:XX or A2:XX:XX:XX:XX:XX, whereas a standard, non-randomized address would not start with a 2, 6, A, or E in the second positio

// // For a standard, non-private MAC address, the last six hexadecimal characters (numbers 0-9 and letters A-F) are changed, while the first six are the manufacturer's identifier. However, iPhones use a feature called "Private Wi-Fi Address" to assign a different, randomized MAC address for each network. This randomization affects all 12 hexadecimal characters to provide enhanced privacy.
// // Standard MAC address
// // Structure: A 12-character hexadecimal number (e.g., 00:1A:2B:3C:4D:5E).
// // Manufacturer Identifier: The first six characters (00:1A:2B) are the Organizationally Unique Identifier (OUI) and are constant for all of Apple's devices.
// // Device Identifier: The last six characters (3C:4D:5E) are the device-specific part and are unique to that particular piece of hardware.
// // iPhone "Private Wi-Fi Address" (default)
// // Dynamic Assignment: The iPhone generates a unique, randomized MAC address for each Wi-Fi network you connect to.
// // Randomized Characters: This randomization affects all 12 hexadecimal characters in the MAC address.
// // Persistence: The randomized address is persistent for that network, meaning it stays the same for every future connection to that same network, but changes if you "forget" the network and reconnect.

// // CONCEPT
// // 1) Pick the oldest mac to be the actual (earliest created timestamp) [ok]
// // 2) Check if maclist table has matching entries, else add one
// // 3) Merge all mac histories into one mac history [ok]
// // 4) Write the new history
// // 4) Check if either mac has USER data filled in within the macs table [ok]
// // 5) If newer mac has data, and older is missing some, merge data into the old one [ok]
// // 6) Delete newer mac from online and from mac table, keeping it in the maclist table [ok]
