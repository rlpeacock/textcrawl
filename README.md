# Textcrawl

A project to learn Go by building a MUD engine.

Current Status
--

- A basic telnet server. It does not understand any fancy display stuff.
- YAML loading of the dungeon. Only rooms and exits currently.
- A game engine that collects requests, then executes them as a batch.
- Requests are handed off to an embedded Lua interpreter for game logic.
- SQLite persistence of game state. Currently only loads.

Next Steps
--

- Implement picking up and dropping objects in Lua
- Save dirty objects at end of turn
- MOB animation
- MOB spawning
- Formalize the Lua API

I might call it done at that point. I'm not actually a MUD player so not sure I have much interest in building out the actual game mechanics, but we'll see.
