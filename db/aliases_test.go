package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// You can now run your tests from the command line in the db directory with go test -v to get detailed results.

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *sql.DB {
	// Use a file-based in-memory DB which is cleaned up after the test.
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Create the aliases table. Based on other functions like SaveAliases,
	// 'mac' appears to be the primary key.
	createTableSQL := `
	CREATE TABLE aliases (
		mac TEXT PRIMARY KEY,
		alias TEXT NOT NULL,
		updated INTEGER
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create aliases table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// cleanTable removes all rows from the aliases table.
func cleanTable(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM aliases")
	if err != nil {
		t.Fatalf("Failed to clean aliases table: %v", err)
	}
}

// TestAddAliasPair contains all the test cases for the AddAliasPair function.
func TestAddAliasPair(t *testing.T) {
	// The application uses a global `db.Conn`. For testing, we should replace it
	// with a connection to a temporary test database.
	originalConn := Conn
	Conn = setupTestDB(t)
	defer func() { Conn = originalConn }()

	t.Run("1. Add a brand new pair", func(t *testing.T) {
		cleanTable(t, Conn)
		mac, alias := "newPairMac", "newPairAlias"

		if err := AddAliasPair(mac, alias); err != nil {
			t.Fatalf("AddAliasPair failed unexpectedly: %v", err)
		}

		// Assert: The pair should now exist in the table.
		var foundMac, foundAlias string
		err := Conn.QueryRow("SELECT mac, alias FROM aliases WHERE mac = ?", mac).Scan(&foundMac, &foundAlias)
		if err != nil {
			t.Fatalf("Failed to query for new pair: %v", err)
		}
		if foundMac != mac || foundAlias != alias {
			t.Errorf("Expected pair (%s, %s), but got (%s, %s)", mac, alias, foundMac, foundAlias)
		}
	})

	t.Run("2. Add a reversed new pair (should do nothing)", func(t *testing.T) {
		cleanTable(t, Conn)
		mac, alias := "newPairMac", "newPairAlias"
		// Pre-populate with the canonical pair.
		_, err := Conn.Exec("INSERT INTO aliases (mac, alias) VALUES (?, ?)", mac, alias)
		if err != nil {
			t.Fatalf("Test setup failed: %v", err)
		}

		// Act: Try to add the reversed pair.
		if err := AddAliasPair(alias, mac); err != nil {
			t.Fatalf("AddAliasPair failed unexpectedly: %v", err)
		}

		// Assert: No new rows should be added.
		var count int
		err = Conn.QueryRow("SELECT COUNT(*) FROM aliases").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count rows: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 row, but got %d", count)
		}
	})

	t.Run("3. Add an existing pair (should do nothing)", func(t *testing.T) {
		cleanTable(t, Conn)
		mac, alias := "06:E0:2D:92:71:73", "D8:FE:E3:00:57:68"
		_, err := Conn.Exec("INSERT INTO aliases (mac, alias) VALUES (?, ?)", mac, alias)
		if err != nil {
			t.Fatalf("Test setup failed: %v", err)
		}

		if err := AddAliasPair(mac, alias); err != nil {
			t.Fatalf("AddAliasPair failed unexpectedly: %v", err)
		}

		var count int
		err = Conn.QueryRow("SELECT COUNT(*) FROM aliases").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count rows: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 row, but got %d", count)
		}
	})

	t.Run("4. Add a reversed existing pair (should do nothing)", func(t *testing.T) {
		cleanTable(t, Conn)
		mac, alias := "06:E0:2D:92:71:73", "D8:FE:E3:00:57:68"
		_, err := Conn.Exec("INSERT INTO aliases (mac, alias) VALUES (?, ?)", mac, alias)
		if err != nil {
			t.Fatalf("Test setup failed: %v", err)
		}

		if err := AddAliasPair(alias, mac); err != nil {
			t.Fatalf("AddAliasPair failed unexpectedly: %v", err)
		}

		var count int
		err = Conn.QueryRow("SELECT COUNT(*) FROM aliases").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count rows: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 row, but got %d", count)
		}
	})

	t.Run("5. Add existing alias with a new mac", func(t *testing.T) {
		cleanTable(t, Conn)
		existingMac, masterAlias := "06:E0:2D:92:71:73", "D8:FE:E3:00:57:68"
		_, err := Conn.Exec("INSERT INTO aliases (mac, alias) VALUES (?, ?)", existingMac, masterAlias)
		if err != nil {
			t.Fatalf("Test setup failed: %v", err)
		}

		newMac := "existingAliasNewMac"
		if err := AddAliasPair(newMac, masterAlias); err != nil {
			t.Fatalf("AddAliasPair failed unexpectedly: %v", err)
		}

		// Assert: The new mac should now point to the master alias.
		var foundAlias string
		err = Conn.QueryRow("SELECT alias FROM aliases WHERE mac = ?", newMac).Scan(&foundAlias)
		if err != nil {
			t.Fatalf("Failed to query for new pair: %v", err)
		}
		if foundAlias != masterAlias {
			t.Errorf("Expected new mac to point to %s, but it points to %s", masterAlias, foundAlias)
		}
	})

	t.Run("6. Add existing master as mac with a new alias", func(t *testing.T) {
		cleanTable(t, Conn)
		existingMac, masterAlias := "06:E0:2D:92:71:73", "D8:FE:E3:00:57:68"
		_, err := Conn.Exec("INSERT INTO aliases (mac, alias) VALUES (?, ?)", existingMac, masterAlias)
		if err != nil {
			t.Fatalf("Test setup failed: %v", err)
		}

		newAlias := "existingAliasAlias"
		// Note: The original test had a `// FAILS` comment here. This test asserts
		// the behavior that seems correct based on the function's logic: the new
		// alias becomes a mac pointing to the existing master.
		if err := AddAliasPair(masterAlias, newAlias); err != nil {
			t.Fatalf("AddAliasPair failed unexpectedly: %v", err)
		}

		// Assert: The new alias should be a mac pointing to the master.
		var foundAlias string
		err = Conn.QueryRow("SELECT alias FROM aliases WHERE mac = ?", newAlias).Scan(&foundAlias)
		if err != nil {
			t.Fatalf("Failed to query for new pair: %v", err)
		}
		if foundAlias != masterAlias {
			t.Errorf("Expected new alias '%s' to point to master '%s', but it points to '%s'", newAlias, masterAlias, foundAlias)
		}
	})

	t.Run("7. Add existing mac with a new alias (chaining)", func(t *testing.T) {
		cleanTable(t, Conn)
		existingMac, masterAlias := "8E:1A:99:1F:20:16", "master1"
		_, err := Conn.Exec("INSERT INTO aliases (mac, alias) VALUES (?, ?)", existingMac, masterAlias)
		if err != nil {
			t.Fatalf("Test setup failed: %v", err)
		}

		newAlias := "existingMacsAlias"
		if err := AddAliasPair(existingMac, newAlias); err != nil {
			t.Fatalf("AddAliasPair failed unexpectedly: %v", err)
		}

		// Assert: The new alias should become a mac pointing to the ultimate master.
		var foundAlias string
		err = Conn.QueryRow("SELECT alias FROM aliases WHERE mac = ?", newAlias).Scan(&foundAlias)
		if err != nil {
			t.Fatalf("Failed to query for new pair: %v", err)
		}
		if foundAlias != masterAlias {
			t.Errorf("Expected new alias '%s' to point to master '%s', but it points to '%s'", newAlias, masterAlias, foundAlias)
		}
	})
}
