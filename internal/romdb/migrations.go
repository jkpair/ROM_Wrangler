package romdb

const schemaSQL = `
CREATE TABLE IF NOT EXISTS game_cache (
	sha1       TEXT PRIMARY KEY,
	name       TEXT NOT NULL,
	system     TEXT NOT NULL,
	region     TEXT,
	serial     TEXT,
	description TEXT,
	publisher  TEXT,
	year       TEXT,
	source     TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_game_cache_system ON game_cache(system);
CREATE INDEX IF NOT EXISTS idx_game_cache_name ON game_cache(name);
`
