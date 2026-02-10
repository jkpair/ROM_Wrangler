package romdb

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database for ROM cache.
type DB struct {
	db *sql.DB
}

// DefaultPath returns the default cache database path.
func DefaultPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "romwrangler", "cache.db")
}

// Open opens or creates the cache database.
func Open(path string) (*DB, error) {
	if path == "" {
		path = DefaultPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Enable WAL mode for better concurrent access
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		sqlDB.Close()
		return nil, err
	}

	rdb := &DB{db: sqlDB}
	if err := rdb.migrate(); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return rdb, nil
}

func (rdb *DB) migrate() error {
	_, err := rdb.db.Exec(schemaSQL)
	return err
}

// Close closes the database.
func (rdb *DB) Close() error {
	return rdb.db.Close()
}
