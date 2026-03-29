package db

import (
	"database/sql"
	"log"
)

type Aliases struct {
	Mac     string `json:"mac"`     // Device's actual mac address
	Alias   string `json:"alias"`   // Host's mac
	Updated int64  `json:"updated"` // last updated counter
}

func SaveAliases(items []Aliases) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := Conn.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer tx.Rollback() // Defer Rollback. It's a no-op if Commit succeeds.

	query := `
		INSERT INTO aliases (mac, alias, updated) VALUES (?, ?, ?)
		ON CONFLICT(mac) DO UPDATE SET
		mac=excluded.mac, alias=excluded.alias, updated=excluded.updated`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println("Error preparing statement:", err)
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		_, err := stmt.Exec(item.Mac, item.Alias, item.Updated)
		if err != nil {
			log.Printf("Error saving aliases to database for %v. Rolling back transaction. Error: %v", item, err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}

	if err = cleanAliases(); err != nil {
		log.Println("Error cleaning aliases after committing transaction:", err)
		return err
	}
	return nil
}

// Scrub the aliases table
func cleanAliases() error {
	tx, err := Conn.Begin()
	if err != nil {
		log.Println("Error")
		return err
	}
	defer tx.Rollback()

	//1. Remove any rows where mac and alias are identical, as these are self-references invalid entries.
	_, err = tx.Exec("DELETE FROM aliases WHERE mac=alias")
	if err != nil {
		return err
	}

	//2. Remove any entries where any alias column entries appears in the mac column, this leads to unwanted chaining of finding the master alias
	_, err = tx.Exec("DELETE FROM aliases WHERE mac IN (SELECT alias FROM aliases)")
	if err != nil {
		return err
	}

	//3. Remove any entries where any mac column entries appears in the alias column, this leads to a mac being both a master and a slave
	_, err = tx.Exec("DELETE FROM aliases WHERE alias IN (SELECT mac FROM aliases)")
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetAliases(isNewOnly bool) ([]Aliases, error) {
	items := make([]Aliases, 0)
	query := `SELECT mac, alias, updated FROM aliases`
	if isNewOnly {
		query += ` WHERE updated=-1`
	}
	rows, err := Conn.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Aliases
		err := rows.Scan(&item.Mac, &item.Alias, &item.Updated)
		if err != nil {
			log.Println(err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// We have two variables passed to a AddAliasPair(mac, alias string) function.
// These two variables hold Mac Addresses.
// The alias is the master in a sqlite aliases table.
// alias < mac always !!!!!!!!!!!!!!!!
// The aliases table has two columns, mac and alias.
// - If the mac/alias pair is already in the aliases table, return.
// - Swap mac and alias values. If the swapped mac/alias pair is already in the aliases table, return.
// - Find if either mac or alias is anywhere in the table and which column(s) it may be in.
// - If the alias is in the aliases table's alias column, insert mac/alias. Return.
// - If the alias is in the aliases table's mac column, then find the true master alias for that alias by getting the alias column where mac column is the alias. Repeat this sequence for the new pair.
// - If the mac is in the aliases table's alias column, then swap mac and alias. Insert mac/alias. Return.
// - If the mac is in the aliases table's mac column, then find its true alias for that mac, Get the new master alias, Repeat this sequence with this new pair as well.
func AddAliasPair(mac, alias string) error {
	tx, err := Conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback on error, no-op on Commit

	// If exact pair exists, return
	var exists int
	err = tx.QueryRow("SELECT 1 FROM aliases WHERE mac = ? AND alias = ?", mac, alias).Scan(&exists)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}

	// If swapped pair exists, return
	err = tx.QueryRow("SELECT 1 FROM aliases WHERE mac = ? AND alias = ?", alias, mac).Scan(&exists)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return err
	}

	// Resolve both to their true masters
	macMaster, err := findMaster(tx, mac)
	if err != nil {
		return err
	}

	aliasMaster, err := findMaster(tx, alias)
	if err != nil {
		return err
	}

	// If macMaster is itself a master, swap mac and alias
	isMacAMaster, err := isMaster(tx, macMaster)
	if err != nil {
		return err
	}
	if isMacAMaster {
		macMaster, aliasMaster = aliasMaster, macMaster
	}

	// Re-resolve aliasMaster to ensure it is the true master
	aliasMaster, err = findMaster(tx, aliasMaster)
	if err != nil {
		return err
	}

	// If both resolve to same master, nothing to do
	if macMaster == aliasMaster {
		return nil
	}

	// Insert canonical pair: non-master → master
	_, err = tx.Exec("INSERT OR IGNORE INTO aliases(mac, alias) VALUES(?, ?)", macMaster, aliasMaster)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing add alias pair transaction:", err)
		return err
	}

	err = cleanAliases()
	if err != nil {
		return err
	}

	return nil
}

type querier interface {
	QueryRow(query string, args ...any) *sql.Row
}

func isMaster(q querier, x string) (bool, error) {
	var dummy int
	err := q.QueryRow("SELECT 1 FROM aliases WHERE alias = ?", x).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func findMaster(q querier, x string) (string, error) {
	cur := x
	for {
		var next string
		err := q.QueryRow("SELECT alias FROM aliases WHERE mac = ?", cur).Scan(&next)
		if err == sql.ErrNoRows {
			return cur, nil // reached true master
		}
		if err != nil {
			return "", err
		}
		cur = next
	}
}
