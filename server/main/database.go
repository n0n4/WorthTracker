package main

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// DataAccess is our entryway for all database related functionality
// We want to avoid using the db directly in the rest of the program
// Instead, any needed operation should be a method on DataAccess

// Placing the DataAccess layer behind an interface is a small investment
// which makes it easier to e.g. build mock DataAccess types for testing
// or even more easily support database migrations later
type DataAccess interface {
	Close()
	Standup(context.Context) error
	// user methods
	AddUser(context.Context, string) error
	FindUserByName(context.Context, string) (*UserEntry, error)
	GetUsers(context.Context) (*[]UserEntry, error)
	// item methods
	AddItem(context.Context, int, string, string, int64) error
	UpdateItem(context.Context, int, int, string, string, int64) error
	DeleteItem(context.Context, int) error
	GetItemsByUser(context.Context, int) (*[]ItemEntry, error)
	FindItemById(context.Context, int) (*ItemEntry, error)
}

// DataAccessSQL is our actual DataAccess layer for this case
type DataAccessSQL struct {
	database *sql.DB
}

func OpenDataAccess(endpoint string) (DataAccess, error) {
	database, err := sql.Open("sqlite3", endpoint)
	da := DataAccess(DataAccessSQL{database: database})

	return da, err
}

func (da DataAccessSQL) Close() {
	da.database.Close()
}

const (
	standupSchema = `
CREATE TABLE IF NOT EXISTS users (
	uid  INTEGER PRIMARY KEY,
	name TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS items (
	id    INTEGER PRIMARY KEY,
	uid   INTEGER NOT NULL,
	name  TEXT,
	type  TEXT,
	value BIGINT
);
`
)

func (da DataAccessSQL) Standup(context context.Context) error {
	// create the database & its tables if it does not exist
	_, err := da.database.ExecContext(context, standupSchema)
	return err
}
