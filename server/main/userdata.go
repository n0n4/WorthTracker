package main

import (
	"context"
	"database/sql"
)

const (
	insertUserCommand = `
INSERT INTO users (name) VALUES ($1);	
`
	findUserCommand = `
SELECT * FROM users WHERE name = $1
`
	getUsersCommand = `
SELECT * FROM users
`
)

type UserEntry struct {
	Id   int
	Name string
}

func (da DataAccessSQL) AddUser(context context.Context, username string) error {
	_, err := da.database.ExecContext(context, insertUserCommand, username)
	return err
}

func (da DataAccessSQL) FindUserByName(context context.Context, username string) (*UserEntry, error) {
	rows, err := da.database.QueryContext(context, findUserCommand, username)
	// make sure to clean up rows when we're finished
	defer func() {
		rows.Close()
	}()

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	// process the rows into UserEntries
	for rows.Next() {
		// check for errors
		err = rows.Err()
		if err != nil {
			return nil, err
		}

		// scan the next row
		var uid int
		var uname string
		err = rows.Scan(&uid, &uname)
		if err != nil {
			return nil, err
		}

		// return the first user we find (there should only be one...)
		return &UserEntry{Id: uid, Name: uname}, nil
	}

	return nil, nil
}

func (da DataAccessSQL) GetUsers(context context.Context) (*[]UserEntry, error) {
	rows, err := da.database.QueryContext(context, getUsersCommand)
	// make sure to clean up rows when we're finished
	defer func() {
		rows.Close()
	}()

	users := make([]UserEntry, 0)
	if err == sql.ErrNoRows {
		return &users, nil
	} else if err != nil {
		return nil, err
	}

	// process the rows into UserEntries
	for rows.Next() {
		// check for errors
		err = rows.Err()
		if err != nil {
			return nil, err
		}

		// scan the next row
		var uid int
		var uname string
		err = rows.Scan(&uid, &uname)
		if err != nil {
			return &users, err
		}

		users = append(users, UserEntry{Id: uid, Name: uname})
	}

	return &users, nil
}
