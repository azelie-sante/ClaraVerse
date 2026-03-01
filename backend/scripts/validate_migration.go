package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

func main() {
	// Read configuration
	sqlitePath := getEnv("SQLITE_PATH", "./model_capabilities.db")
	mysqlDSN := getEnv("MYSQL_DSN", "")

	if mysqlDSN == "" {
		log.Fatal("❌ MYSQL_DSN environment variable required")
	}

	log.Println("🔍 Validating migration...")
	log.Printf("   SQLite: %s", sqlitePath)
	log.Printf("   MySQL:  %s\n", maskDSN(mysqlDSN))

	// Open databases
	sqliteDB, err := sql.Open("sqlite", sqlitePath)
	if err != nil {
		log.Fatalf("❌ Failed to open SQLite: %v", err)
	}
	defer sqliteDB.Close()

	mysqlDB, err := sql.Open("mysql", mysqlDSN+"?parseTime=true")
	if err != nil {
		log.Fatalf("❌ Failed to open MySQL: %v", err)
	}
	defer mysqlDB.Close()

	// Test connections
	if err := sqliteDB.Ping(); err != nil {
		log.Fatalf("❌ SQLite connection failed: %v", err)
	}
	if err := mysqlDB.Ping(); err != nil {
		log.Fatalf("❌ MySQL connection failed: %v", err)
	}

	// Tables to validate
	tables := []string{
		"providers",
		"models",
		"model_aliases",
		"provider_model_filters",
		"model_capabilities",
		"model_refresh_log",
	}

	allMatch := true
	for _, table := range tables {
		var sqliteCount, mysqlCount int

		// Get SQLite count
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		err := sqliteDB.QueryRow(query).Scan(&sqliteCount)
		if err != nil {
			if strings.Contains(err.Error(), "no such table") {
				log.Printf("   ⚠️  %s: Table doesn't exist in SQLite (skipping)", table)
				continue
			}
			log.Printf("   ❌ %s: Failed to query SQLite: %v", table, err)
			allMatch = false
			continue
		}

		// Get MySQL count
		err = mysqlDB.QueryRow(query).Scan(&mysqlCount)
		if err != nil {
			log.Printf("   ❌ %s: Failed to query MySQL: %v", table, err)
			allMatch = false
			continue
		}

		// Compare
		if sqliteCount != mysqlCount {
			log.Printf("   ❌ %s: SQLite=%d, MySQL=%d (MISMATCH)", table, sqliteCount, mysqlCount)
			allMatch = false
		} else if sqliteCount > 0 {
			log.Printf("   ✅ %s: %d records match", table, mysqlCount)
		} else {
			log.Printf("   ℹ️  %s: 0 records (empty table)", table)
		}
	}

	// Summary
	fmt.Println("\n" + strings.Repeat("=", 60))
	if allMatch {
		log.Println("✅ Validation PASSED - All record counts match!")
	} else {
		log.Println("❌ Validation FAILED - Some counts don't match")
		log.Println("   Review errors above and re-run migration if needed")
		os.Exit(1)
	}
	log.Println(strings.Repeat("=", 60))
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func maskDSN(dsn string) string {
	parts := strings.Split(dsn, "@")
	if len(parts) < 2 {
		return dsn
	}
	userPass := strings.Split(parts[0], ":")
	if len(userPass) < 2 {
		return dsn
	}
	return userPass[0] + ":***@" + parts[1]
}
