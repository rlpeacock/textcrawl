#!/usr/bin/env bash

set -e

echo "removing old data"
rm world/*.dat 2>/dev/null || true
echo "creating starter databases"
sqlite3 world/0.dat < schema/zone.sql
cp world/0.dat world/1.dat
sqlite3 world/1.dat < schema/1.sql
sqlite3 world/player.dat < schema/player.sql
sqlite3 world/player.dat < schema/test_player.sql
echo "done"