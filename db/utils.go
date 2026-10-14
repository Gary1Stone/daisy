package db

import (
	"database/sql"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
)

// LogHits logs the hit count at the end of a given 15-minute interval.
// The database adds the integer timestamp (UTC) automatically
func LogHits(count uint64) {
	cnt := int64(count)
	if cnt < 0 {
		cnt = math.MaxInt64
	}
	query := "INSERT INTO hits (hits) VALUES (?)"
	_, err := Conn.Exec(query, cnt)
	if err != nil {
		log.Println(err)
	}
}

// Return the User's Time Zone offset in minutes
func GetTzoff(curUid int) int {
	// we want only the most recent entry
	if curUid < 1 {
		curUid = SYS_PROFILE.Uid
	}
	var tzoff int = 0
	query := "SELECT tzoff FROM logins WHERE uid=? AND success=1 ORDER BY timestamp DESC LIMIT 1"
	err := Conn.QueryRow(query, curUid).Scan(&tzoff) // If user never signed in, no logs to look at, so ignore err (no rows returned)
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}
	return tzoff
}

// Reduce the point count to match the width of the SVG plot
func BucketDownsample(data []int, pixelWidth int) []int {
	n := len(data)
	if n == 0 || pixelWidth <= 0 {
		return nil
	}
	if pixelWidth >= n {
		return data
	}

	bucketSize := int(math.Ceil(float64(n) / float64(pixelWidth)))
	result := make([]int, 0, pixelWidth)

	for i := 0; i < n; i += bucketSize {
		end := i + bucketSize
		if end > n {
			end = n
		}
		max := data[i]
		for _, v := range data[i:end] {
			if v > max {
				max = v
			}
		}
		result = append(result, max)
	}
	return result
}

func WrapSearchTerm(str string) string {
	return "%" + str + "%"
}

func GetSqlVersion() string {
	var version string
	err := Conn.QueryRow("SELECT sqlite_version()").Scan(&version)
	if err != nil {
		return "unknown"
	}
	return version
}

func ExportSchema() error {
	rows, err := Conn.Query(`
		SELECT sql
		FROM sqlite_master
		WHERE sql IS NOT NULL
		AND type IN ('table','index','trigger','view')
		ORDER BY type, name;
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var schema strings.Builder
	for rows.Next() {
		var sql string
		if err := rows.Scan(&sql); err != nil {
			return err
		}
		schema.WriteString(sql)
		schema.WriteString(";\n\n")
	}
	if err = rows.Err(); err != nil {
		log.Println(err)
		return err
	}
	// Build the path to the db directory
	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}
	schemaFilePath := filepath.Join(workDir, "db", "schema.sql")
	err = os.WriteFile(schemaFilePath, []byte(schema.String()), 0644)
	return err
}

// Foreign Key handing!
// Go does not support setting ints and strings to nil
// Database foreign keys (usually ints) absolutley require being able
// to set a child table's foreign keys to null.
// ForeignKey converts zero-equivalent values (empty strings, non-positive numbers)
// to nil so they can be inserted as NULL into the database.
func ForeignKey(value any) any {
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
