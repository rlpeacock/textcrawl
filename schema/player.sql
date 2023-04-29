CREATE TABLE IF NOT EXISTS player (
	   username TEXT PRIMARY KEY,
	   password TEXT NOT NULL,
	   actor_id INTEGER UNIQUE,
	   active boolean
);
