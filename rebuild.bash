#!/usr/bin/env bash

set -e

rm world/*.dat
sqlite3 world/0.dat < schema/zone.sql
cp world/0.dat world/1.dat
sqlite3 world/1.dat < schema/1.sql
