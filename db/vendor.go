package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/logocomune/maclookup-go"
)

// Collect MACs that need vendor updates
func PopulateVendors() error {
	var macsToUpdate []struct {
		Mac    string
		Vendor string
	}
	query := "SELECT mac FROM macs WHERE vendor='' AND isRandomMac=0"
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var mac string
		err := rows.Scan(&mac)
		if err != nil {
			log.Println("Error scanning MAC during vendor population:", err)
			return err
		}

		vendor, err := findVendor(mac)
		if err != nil {
			log.Println("Error finding vendor for MAC", mac, ":", err)
			return err
		}

		if vendor == "" {
			vendor = "Unknown" // prevent unknown vendor from hitting the API again
		}

		macsToUpdate = append(macsToUpdate, struct {
			Mac    string
			Vendor string
		}{Mac: mac, Vendor: vendor})
	}

	// Check for any errors during row iteration.
	if err = rows.Err(); err != nil {
		log.Println("Error iterating over MAC rows:", err)
		return err
	}

	// if none found, return
	if len(macsToUpdate) == 0 {
		//log.Println("No MACs found requiring vendor updates.")
		return nil
	}

	// Perform a single batched UPDATE for all collected MACs.
	tx, err := Conn.Begin()
	if err != nil {
		log.Println("Error starting transaction for vendor update:", err)
		return err
	}
	defer tx.Rollback() // Rollback on error, no-op on commit.

	var updateQuery strings.Builder
	updateQuery.WriteString("UPDATE macs SET vendor = CASE mac ")
	params := make([]any, 0, 2*len(macsToUpdate))

	for _, item := range macsToUpdate {
		fmt.Fprintf(&updateQuery, "WHEN ? THEN ? ")
		params = append(params, item.Mac, item.Vendor)
	}
	updateQuery.WriteString("END WHERE mac IN (")

	// Add MACs to the IN clause.
	for i, item := range macsToUpdate {
		updateQuery.WriteString("?")
		params = append(params, item.Mac)
		if i < len(macsToUpdate)-1 {
			updateQuery.WriteString(", ")
		}
	}
	updateQuery.WriteString(")")

	_, err = tx.Exec(updateQuery.String(), params...)
	if err != nil {
		log.Println("Error executing batched vendor update:", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Println("Error committing vendor update transaction:", err)
		return err
	}
	// log.Printf("Successfully updated vendors for %d MACs.", len(macsToUpdate))
	return nil
}

// Search the table for revious matches on MAC addresses
// A typical OUI (vendor prefix) is 3 bytes (8 characters with colons).
func findVendor(mac string) (string, error) {
	// The query finds the longest matching prefix for the given MAC address.
	query := `SELECT Vendor, MacPrefix FROM vendors
		WHERE ? LIKE MacPrefix || '%'
		ORDER BY LENGTH(MacPrefix) DESC LIMIT 1`
	vendor := ""
	macPrefix := ""
	err := Conn.QueryRow(query, mac).Scan(&vendor, &macPrefix)
	if err != nil {
		if err == sql.ErrNoRows {
			vendor, err = macAPI(mac)
			if err != nil {
				log.Println("Vendor lookup failed", err)
				vendor = ""
			}
			return vendor, nil
		}
		return "", err
	}
	return vendor, nil
}

var lastRequest = time.Now() // rate limit the macAPI

func macAPI(mac string) (string, error) {
	if mac == "" {
		return "", nil
	}

	log.Println("Had to fetch missing mac vendor")
	client := maclookup.New()
	mac_api_key := os.Getenv("MAC_API_KEY")
	client.WithAPIKey(mac_api_key)

	if time.Since(lastRequest) < time.Second {
		time.Sleep(time.Second - time.Since(lastRequest))
	}
	lastRequest = time.Now()

	recvd, err := client.Lookup(mac)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// Save Unknown entry into database to prevent subsequent searches.
	if !recvd.Found {
		recvd.Company = "Unknown"
		if len(mac) > 8 {
			recvd.MacPrefix = mac[0:8]
		} else {
			recvd.MacPrefix = mac
		}
		recvd.BlockType = "MA-L"
		recvd.Updated = time.Now().Format("2006-01-02 15:04:05")
		recvd.IsPrivate = false
	}

	err = addNewVendor(recvd.Company, recvd.MacPrefix, recvd.BlockType, recvd.Updated, recvd.IsPrivate)
	if err != nil {
		log.Println(err)
	}

	return recvd.Company, nil
}

func addNewVendor(vendor, macPrefix, blockType, updated string, isPrivate bool) error {
	// Add the colons to the MacPrefix
	if !strings.Contains(macPrefix, ":") {
		var res strings.Builder
		for i := 0; i < len(macPrefix); i++ {
			if i > 0 && i%2 == 0 {
				res.WriteByte(':')
			}
			res.WriteByte(macPrefix[i])
		}
		macPrefix = strings.ToUpper(res.String())
	}

	query := `INSERT INTO vendors (MacPrefix, vendor, BlockType, LastUpdate, Private) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(MacPrefix) DO UPDATE SET vendor=excluded.vendor, BlockType=excluded.BlockType, LastUpdate=excluded.LastUpdate, Private=excluded.Private`
	_, err := Conn.Exec(query, macPrefix, vendor, blockType, updated, isPrivate)
	if err != nil {
		log.Printf("Error adding new vendor %s: %v", macPrefix, err)
	}
	return err
}
