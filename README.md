# Textcrawl

A project to learn Go by building a MUD engine.

Current Status
--
- A basic telnet server. It does not understand any fancy display stuff.
- YAML loading of the dungeon. Only rooms and exits currently.
- A game engine that collects requests, then executes them as a batch.
- SQLite persistence of game state. Currently only loads.

Setup
--
Basically, just build it and run `rebuild.bash`. That will bootstrap the sqlite DBs.

