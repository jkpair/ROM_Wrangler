package romdb

import (
	"database/sql"

	"github.com/kurlmarx/romwrangler/internal/scraper"
	"github.com/kurlmarx/romwrangler/internal/systems"
)

// GetByHash retrieves cached game info by SHA1 hash.
func (rdb *DB) GetByHash(sha1 string) (*scraper.GameInfo, bool) {
	var info scraper.GameInfo
	var system string
	var region, serial, description, publisher, year sql.NullString

	err := rdb.db.QueryRow(
		`SELECT name, system, region, serial, description, publisher, year, source
		 FROM game_cache WHERE sha1 = ?`, sha1,
	).Scan(&info.Name, &system, &region, &serial, &description, &publisher, &year, &info.Source)

	if err != nil {
		return nil, false
	}

	info.System = systems.SystemID(system)
	info.Region = region.String
	info.Serial = serial.String
	info.Description = description.String
	info.Publisher = publisher.String
	info.Year = year.String

	return &info, true
}

// Put stores game info in the cache.
func (rdb *DB) Put(sha1 string, info *scraper.GameInfo) error {
	_, err := rdb.db.Exec(
		`INSERT OR REPLACE INTO game_cache
		 (sha1, name, system, region, serial, description, publisher, year, source, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		sha1, info.Name, string(info.System), info.Region, info.Serial,
		info.Description, info.Publisher, info.Year, info.Source,
	)
	return err
}
