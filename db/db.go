package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Make the database connection globally available
var Conn *sql.DB

const (
	PROFILE_TABLE  int = 2
	SOFTWARE_TABLE int = 3
	REQUEST_TABLE  int = 4
	DEVICE_TABLE   int = 5
	ACTION_TABLE   int = 6
)

func init() {
	loadEnvVariables()
	connectToDatabase()
	buildAdminCache()
}

// Load the environment variables
func loadEnvVariables() {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			log.Fatal("FATAL: .env file not found. Please create one in the current working directory.")
		} else {
			log.Fatal("FATAL: Failed to load .env file. Error:", err.Error())
		}
	}
}

// Create connection to the database
func connectToDatabase() {

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("FATAL: DB_URL environment variable not set.")
	}

	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}
	databaseFilePath := filepath.Join(workDir, dbURL)

	Conn, err = sql.Open("sqlite3", databaseFilePath)
	if err != nil {
		// This error is unlikely here, but we handle it just in case.
		log.Fatalf("FATAL: Error preparing database connection: %v", err)
	}

	// Set connection pool settings before first use.
	Conn.SetMaxIdleConns(64) // Default is 2
	Conn.SetMaxOpenConns(64) // Default is 0 (unlimited)

	// Ping verifies the connection is alive.
	err = Conn.Ping()
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to the database: %v", err)
	}

	// Turn on Write-ahead-logging (WAL) for speed.
	// This needs to be done once per database file, but executing it on every connection is safe.
	_, err = Conn.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Fatalf("FATAL: Failed to set WAL journal mode: %v", err)
	}

	// Set a busy timeout to prevent SQLITE_BUSY errors on concurrent writes. 5 seconds is a reasonable default.
	_, err = Conn.Exec("PRAGMA busy_timeout = 5000;")
	if err != nil {
		log.Fatalf("FATAL: Failed to set busy_timeout pragma: %v", err)
	}

	// Check if the database is empty
	var count int
	query := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`
	err = Conn.QueryRow(query).Scan(&count)
	if err != nil {
		log.Fatalf("FATAL: Failed to check if the database is empty: %v", err)
	}

	// Create tables
	if count == 0 {
		_, err = Conn.Exec(schemaSQL)
		if err != nil {
			return
		}
	}

	// Only export the schema if we are in a development environment.
	// We detect this by checking if the 'db' source directory exists.
	// The 'db' folder is not included in production deployments.
	if info, err := os.Stat(filepath.Join(workDir, "db")); err == nil && info.IsDir() {
		err = ExportSchema()
		if err != nil {
			log.Println("ERROR: Failed to export database schema " + err.Error())
		}
	}
}

// Close closes the database connection. It's safe to call this multiple times.
func Close() error {
	if Conn != nil {
		log.Println("Closing database connection...")
		return Conn.Close()
	}
	return nil
}

// Foreign Key handing!
// Go does not support setting ints and strings to nil
// Database foreign keys (usually ints) absolutley require being able
// to set a child table's foreign keys to null.
// ForeignKey converts zero-equivalent values (empty strings, non-positive numbers)
// to nil so they can be inserted as NULL into the database.
func foreignKey(value any) any {
	switch v := value.(type) {
	case string:
		if len(v) == 0 {
			return nil
		}
	case int:
		if v <= 0 {
			return nil
		}
	case int64:
		if v <= 0 {
			return nil
		}
	case float64:
		if v == 0.0 {
			return nil
		}
	}
	return value
}

// TwoAmBackup schedules a backup of the database every day at 2 AM.
// It uses VACUUM INTO to safely backup the database even in WAL mode.
func TwoAmBackup() {
	for {
		now := time.Now()
		// Calculate next 2:00 AM
		next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		// Sleep until 2 AM
		time.Sleep(next.Sub(now))

		// Remove unused photos from the images directory
		RemoveOldPhotos()

		// Ensure backups directory exists
		backupDir := os.Getenv("BACKUP_DIR")
		if backupDir == "" {
			backupDir = "./backups"
		}
		if _, err := os.Stat(backupDir); os.IsNotExist(err) {
			err := os.Mkdir(backupDir, 0755)
			if err != nil {
				log.Println("Error creating backup directory:", err)
				continue
			}
		}

		// Define backup filename with timestamp
		filename := fmt.Sprintf("daisy.%s", time.Now().Format("Mon"))
		backupPath := filepath.Join(backupDir, filename)

		// Check if a backup for today exists, if so, delete it
		if _, err := os.Stat(backupPath); err == nil {
			if err := os.Remove(backupPath); err != nil {
				log.Println("Error deleting previous backup:", err)
				continue
			}
		}
		log.Println("Starting database backup...")

		// VACUUM INTO creates a consistent backup without locking the DB for writes
		_, err := Conn.Exec("VACUUM INTO ?", backupPath)
		if err != nil {
			log.Printf("ERROR: Database backup failed: %v", err)
		} else {
			log.Printf("Database successfully backed up to: %s", backupPath)
		}
	}
}

func GetSqlVersion() string {
	version, _, _ := sqlite3.Version()
	return version
}
