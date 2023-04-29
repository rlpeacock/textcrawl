CREATE TABLE IF NOT EXISTS object (
	   id INTEGER PRIMARY KEY ASC,
	   attributes TEXT NOT NULL,
	   title TEXT NOT NULL,
	   description TEXT NOT NULL,
	   room INTEGER NOT NULL,
	   flags INTEGER
);

CREATE TABLE IF NOT EXISTS actor (
	   id INTEGER PRIMARY KEY ASC,
	   obj_id INTEGER NOT NULL,
	   stats TEXT NOT NULL
);
