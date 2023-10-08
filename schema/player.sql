CREATE TABLE IF NOT EXISTS player (
	   username TEXT PRIMARY KEY,
	   password TEXT NOT NULL,
	   actor_id TEXT NOT NULL UNIQUE,
	   active boolean
);
