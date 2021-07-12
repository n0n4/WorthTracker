package main

import (
	"context"
	"database/sql"
)

const (
	ItemTypeAsset     = "Asset"
	ItemTypeLiability = "Liability"
	insertItemCommand = `
INSERT INTO items (uid, name, type, value) VALUES ($1, $2, $3, $4)
`
	updateItemCommand = `
REPLACE INTO items VALUES ($1, $2, $3, $4, $5)
`
	deleteItemCommand = `
DELETE FROM items WHERE id = $1
`
	getItemsCommand = `
SELECT * FROM items WHERE uid = $1
`
	findItemByIdCommand = `
SELECT * FROM items WHERE id = $1
`
)

type ItemEntry struct {
	Id    int
	Uid   int
	Name  string
	Type  string
	Value int64
}

func (da DataAccessSQL) AddItem(context context.Context, userid int, name string, itemType string, value int64) error {
	_, err := da.database.ExecContext(context, insertItemCommand, userid, name, itemType, value)
	return err
}

func (da DataAccessSQL) UpdateItem(context context.Context, id int, userid int, name string, itemType string, value int64) error {
	_, err := da.database.ExecContext(context, updateItemCommand, id, userid, name, itemType, value)
	return err
}

func (da DataAccessSQL) DeleteItem(context context.Context, id int) error {
	_, err := da.database.ExecContext(context, deleteItemCommand, id)
	return err
}

func (da DataAccessSQL) GetItemsByUser(context context.Context, userid int) (*[]ItemEntry, error) {
	rows, err := da.database.QueryContext(context, getItemsCommand, userid)
	// make sure to clean up rows when we're finished
	defer func() {
		rows.Close()
	}()

	items := make([]ItemEntry, 0)
	if err == sql.ErrNoRows {
		return &items, nil
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
		var id, uid int
		var name, itemType string
		var value int64
		err = rows.Scan(&id, &uid, &name, &itemType, &value)
		if err != nil {
			return &items, err
		}

		items = append(items, ItemEntry{Id: id, Uid: uid, Name: name, Type: itemType, Value: value})
	}

	return &items, nil
}

func (da DataAccessSQL) FindItemById(context context.Context, id int) (*ItemEntry, error) {
	rows, err := da.database.QueryContext(context, findItemByIdCommand, id)
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
		var id, uid int
		var name, itemType string
		var value int64
		err = rows.Scan(&id, &uid, &name, &itemType, &value)
		if err != nil {
			return nil, err
		}

		return &ItemEntry{Id: id, Uid: uid, Name: name, Type: itemType, Value: value}, nil
	}

	return nil, nil
}
