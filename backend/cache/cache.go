package cache

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func IsCacheValid(db *sql.DB, table_name string) bool {
	query := "SELECT ttl, updated_at FROM cacheentries WHERE table_name = $1"

	row := db.QueryRow(query, table_name)

	var ttlseconds int64
	var updated_at time.Time

	err := row.Scan(&ttlseconds, &updated_at)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Failed to retrieve rows: %w", err)
			return false
		}
		fmt.Println("Failed to scan row: %w", err)
		return false
	}

	ttl := time.Duration(ttlseconds) * time.Second

	return time.Since(updated_at) < ttl
}

func Default_TTL() int {
	return 60 * 60 * 24 * 100 // NOTE: 100 days in seconds if you wanna evict the thing do it manually
}

func UpdateCacheTimestamp(db *sql.DB, table_name string) (bool, error) {
	// first attempt
	updateQuery := "UPDATE cacheentries SET updated_at = NOW() WHERE table_name = $1"
	result, err := db.Exec(updateQuery, table_name)
	if err != nil {
		return false, fmt.Errorf("failed to update timestamp: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	// if does not exist create
	if rowsAffected == 0 {
		insertQuery := "INSERT INTO cacheentries (table_name, updated_at, ttl) VALUES ($1, NOW(), $2)"
		_, err := db.Exec(insertQuery, table_name, Default_TTL())
		if err != nil {
			return false, fmt.Errorf("failed to insert new cache entry: %w", err)
		}
		return true, nil
	}

	return true, nil
}
