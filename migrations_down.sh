#!/usr/bin/zsh
DB_PATH="./akrasia.db"

goose -dir sql/schema sqlite3 $DB_PATH down
