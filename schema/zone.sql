CREATE TABLE IF NOT EXISTS object (
	   id TEXT PRIMARY KEY NOT NULL,
	   attributes TEXT NOT NULL,
	   title TEXT NOT NULL,
	   description TEXT NOT NULL,
	   location TEXT NOT NULL,
	   flags INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS actor (
	   id TEXT PRIMARY KEY NOT NULL,
	   obj_id TEXT NOT NULL,
	   stats TEXT NOT NULL,
	   FOREIGN KEY (obj_id) REFERENCES object (id)	   
);
