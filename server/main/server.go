package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Starting WorthTracker server on port 3000...")

	// load database endpoint
	dataEndpoint := loadDatabaseEndpoint()

	// open database access
	dataAccess, err := OpenDataAccess(dataEndpoint)
	// if we failed to open the database, abort
	if err != nil {
		log.Panic(err)
	}
	// ensure the database connection is closed when
	// the server shuts down
	defer dataAccess.Close()

	// ensure database is setup
	err = dataAccess.Standup(context.Background())
	// if we failed to standup the database, abort
	if err != nil {
		log.Panic(err)
	}

	// setup http handlers
	userHandlers := userHandlers{da: dataAccess}
	http.HandleFunc("/api/user", userHandlers.UserRequestHandler)

	itemHandlers := itemHandlers{da: dataAccess}
	http.HandleFunc("/api/item", itemHandlers.ItemRequestHandler)
	http.HandleFunc("/api/itemlist", itemHandlers.ItemListRequestHandler)
	http.HandleFunc("/api/itemdelete", itemHandlers.ItemDeleteRequestHandler)

	fs := http.FileServer(http.Dir("./../client/dist"))
	http.Handle("/", fs)

	// begin running the server
	log.Panic(http.ListenAndServe(":3000", nil))
}

func loadDatabaseEndpoint() string {
	// try to load the database endpoint from a file
	buffer, err := ioutil.ReadFile("database.txt")
	if err != nil {
		fmt.Println("Could not load database endpoint from database.txt")
		log.Panic(err)
	}

	return string(buffer)
}
