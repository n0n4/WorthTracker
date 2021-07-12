package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"unicode/utf8"
)

type userHandlers struct {
	da DataAccess
}

type newUserRequest struct {
	Name string
}

type InvalidUserNameError struct {
	Name   string
	Reason string
}

func (err *InvalidUserNameError) Error() string {
	return err.Name + " is an invalid name: " + err.Reason
}

type UserDoesNotExistError struct {
	Name *string
	Uid  *int
}

func (err *UserDoesNotExistError) Error() string {
	if err.Name != nil {
		return "The user '" + *err.Name + "' does not exist."
	} else if err.Uid != nil {
		return "The user with id '" + strconv.Itoa(*err.Uid) + "' does not exist."
	}
	return "The user could not be found."
}

// Find a user given their username
func FindUserByName(da DataAccess, name string) (*UserEntry, error) {
	// find the user and verify they exist
	user, err := da.FindUserByName(context.Background(), name)
	if err != nil || user == nil {
		return nil, &UserDoesNotExistError{Name: &name, Uid: nil}
	}
	return user, nil
}

// Perform validation and add a new user
func AddUser(da DataAccess, name string) error {
	// ensure the name is properly sized
	namelen := utf8.RuneCountInString(name)
	if namelen <= 1 {
		return &InvalidUserNameError{Name: name, Reason: "Must be longer than 1 character."}
	} else if namelen >= 64 {
		return &InvalidUserNameError{Name: name, Reason: "Must be shorter than 64 characters."}
	}

	// ensure the name is unique
	user, err := da.FindUserByName(context.Background(), name)
	if user != nil {
		return &InvalidUserNameError{Name: name, Reason: "There is already a user by that name."}
	} else if err != nil {
		return err
	}

	// try to add the new user
	return da.AddUser(context.Background(), name)
}

// Handle http requests for the user API
func (uh userHandlers) UserRequestHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	switch request.Method {
	case http.MethodOptions:
		return
	case http.MethodGet:
		// return a json of the users
		users, err := uh.da.GetUsers(context.Background())
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(users)
	case http.MethodPost:
		// add a user and return success or failure
		var userRequest newUserRequest
		err := json.NewDecoder(request.Body).Decode(&userRequest)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println("Got new user request for name '" + userRequest.Name + "'")

		err = AddUser(uh.da, userRequest.Name)
		if err != nil {
			fmt.Println("Failed to add user: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		// respond with the new user info
		user, err := uh.da.FindUserByName(context.Background(), userRequest.Name)

		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(user)
	default:
		http.Error(writer, "Invalid request method.", 405)
	}
}
