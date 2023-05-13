INSERT INTO thing (id, attributes, title, description, location, flags)
VALUES ('C1', '1:1,1:1,1:1', 'tin knife', 'a flimsy tin knife suitable for spreading butter, but only if it''s warm', 'R1', 0);

INSERT INTO thing (id, attributes, title, description, location, flags)
VALUES ('C2', '100:100,10:10,10:10', 'a man', 'a non-descript man.', 'R1', 0);

INSERT INTO actor (id, thing_id, stats) VALUES ('A1', 'C2', '18:18,18:18,18:18,18:18,18:18,18:18');
