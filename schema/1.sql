INSERT INTO thing (id, attributes, title, description, location, flags)
VALUES ('T1', '1:1,2:2,3:3', 'tin knife', 'a flimsy tin knife suitable for spreading butter, but only if it''s warm', 'R1', 0);

INSERT INTO thing (id, attributes, title, description, location, flags)
VALUES ('T2', '100:100,10:10,20:20', 'a man', 'a non-descript man.', 'R1', 0);

INSERT INTO thing (id, attributes, title, description, location, flags)
VALUES ('T3', '3:3,2:2,1:1', 'a rusty bucket', 'An old rusty bucket. Probably wouldn''t hold water.', 'R1', 0);


INSERT INTO actor (id, thing_id, stats) VALUES ('A1', 'T2', '18:18,18:18,18:18,18:18,18:18,18:18');
