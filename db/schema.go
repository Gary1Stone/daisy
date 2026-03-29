package db

import (
	"os"
	"path/filepath"
	"strings"
)

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
		schema.WriteString(sql + ";\n\n")
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
