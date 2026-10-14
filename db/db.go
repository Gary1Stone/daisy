package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Make the database connection globally available
var Conn *sql.DB
var SYS_PROFILE SysUser

type SysUser struct {
	Uid  int    `json:"uid"`
	User string `json:"user"`
}

const (
	PROFILE_TABLE  int = 2
	SOFTWARE_TABLE int = 3
	REQUEST_TABLE  int = 4
	DEVICE_TABLE   int = 5
	ACTION_TABLE   int = 6
)

func StartServer() {
	connectToDatabase()
	buildAdminCache()
	setSystemProfile()
}

// Create connection to the database
func connectToDatabase() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Println("WARNING: DB_URL environment variable not set.")
		dbURL = "daisy.db"
	}

	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}
	databaseFilePath := filepath.Join(workDir, "db", dbURL)

	Conn, err = sql.Open("sqlite", databaseFilePath)
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
		log.Println("WARNING: Empty Database!")
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

// Initialize one user to be the Global system profile
func setSystemProfile() {
	sysUser := os.Getenv("SYS_USER")
	if sysUser == "" {
		log.Panic("FATAL: Environment variable SYS_USER not set in .env file")
	}
	err := Conn.QueryRow("SELECT uid, user FROM profiles WHERE user=? AND active=1 LIMIT 1", sysUser).Scan(&SYS_PROFILE.Uid, &SYS_PROFILE.User)
	if err != nil || SYS_PROFILE.User == "" {
		log.Panicf("FATAL: System User %v is not configured in the database profiles table. Error: %v", sysUser, err)
	}
}

// TwoAmBackup schedules a backup of the database every day at 2 AM.
// It uses VACUUM INTO to safely backup the database even in WAL mode.
func TwoAmBackup() error {
	// Ensure backups directory exists
	backupDir := os.Getenv("BACKUP_DIR")
	if backupDir == "" {
		// Get the base file dir
		workingDir, err := os.Getwd()
		if err != nil {
			workingDir = "."
		}
		// Set the backup path
		backupDir := filepath.Join(workingDir, "backups")
		// Create directory if it does not exist
		if stat, err := os.Stat(backupDir); os.IsNotExist(err) {
			os.Mkdir(backupDir, 0755)
		} else {
			if !stat.IsDir() {
				log.Printf("ERROR: Cannot create backup directory")
				return err
			}
		}
	}

	// Define backup filename with timestamp
	filename := fmt.Sprintf("daisy.%s", time.Now().Format("Mon"))
	backupPath := filepath.Join(backupDir, filename)

	// Check if a backup for today exists, if so, delete it
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Remove(backupPath); err != nil {
			log.Println("Error deleting previous backup:", err)
			return err
		}
	}
	log.Println("Starting database backup...")

	// VACUUM INTO creates a consistent backup without locking the DB for writes
	_, err := Conn.Exec("VACUUM INTO ?", backupPath)
	if err != nil {
		log.Printf("ERROR: Database backup failed: %v", err)
		return err
	} else {
		log.Printf("Database successfully backed up to: %s", backupPath)
	}

	return nil
}
