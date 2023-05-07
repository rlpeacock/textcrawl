INSERT INTO object (id, attributes, title, description, room, flags)
VALUES ('1', '1:1,1:1,1:1', 'tin knife', 'a flimsy tin knife suitable for spreading butter, but only if it''s warm', '1', 0);

INSERT INTO object (id, attributes, title, description, room, flags)
VALUES ('2', '100:100,10:10,10:10', 'a man', 'a non-descript man.', '1', 0);

INSERT INTO actor (id, obj_id, stats) VALUES ('1', '2', '18:18,18:18,18:18,18:18,18:18,18:18');
