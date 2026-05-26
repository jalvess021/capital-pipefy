package database

import "database/sql"

type Database interface {
	Ping() error
	Close() error
	DB() *sql.DB
}