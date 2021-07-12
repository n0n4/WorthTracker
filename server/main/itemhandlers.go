package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"unicode/utf8"
)

type itemHandlers struct {
	da DataAccess
}

type InvalidItemNameError struct {
	Name   string
	Reason string
}

func (err *InvalidItemNameError) Error() string {
	return err.Name + " is an invalid name: " + err.Reason
}

type InvalidItemTypeError struct {
	Type   string
	Reason string
}

func (err *InvalidItemTypeError) Error() string {
	return err.Type + " is an invalid item type: " + err.Reason
}

type ItemDoesNotExistError struct {
	Id int
}

func (err *ItemDoesNotExistError) Error() string {
	return "The item with id '" + strconv.Itoa(err.Id) + "' does not exist."
}

type InvalidItemValueError struct {
	Reason string
}

func (err *InvalidItemValueError) Error() string {
	return "Invalid item value: " + err.Reason
}

// Helper method to validate item inputs (e.g. name length, item type is asset or liability, etc)
func validateItem(name string, itemType string, value int64) error {
	// ensure the name is properly sized
	namelen := utf8.RuneCountInString(name)
	if namelen <= 1 {
		return &InvalidItemNameError{Name: name, Reason: "Must be longer than 1 character."}
	} else if namelen >= 200 {
		return &InvalidItemNameError{Name: name, Reason: "Must be shorter than 200 characters."}
	}

	// ensure the itemType is valid
	if itemType != ItemTypeAsset && itemType != ItemTypeLiability {
		return &InvalidItemTypeError{Type: itemType, Reason: "Must be " + ItemTypeAsset + " or " + ItemTypeLiability}
	}

	// ensure the value is not negative
	if value < 0 {
		return &InvalidItemValueError{Reason: "Must be greater than zero."}
	}

	return nil
}

// Performs validation on item inputs and then tries to add the new item to the database
func AddItem(da DataAccess, name string, itemType string, username string, value int64) error {
	// run standard validation
	if err := validateItem(name, itemType, value); err != nil {
		return err
	}

	// find the user and verify they exist
	user, err := FindUserByName(da, username)
	if err != nil {
		return err
	}

	// try to add the new item
	return da.AddItem(context.Background(), user.Id, name, itemType, value)
}

// Performs validation on item inputs and then tries to update the existing item
func UpdateItem(da DataAccess, id int, name string, itemType string, username string, value int64) error {
	// verify the item id already exists
	item, err := da.FindItemById(context.Background(), id)
	if err != nil || item == nil {
		return &ItemDoesNotExistError{Id: id}
	}

	// run standard validation
	if err := validateItem(name, itemType, value); err != nil {
		return err
	}

	// find the user and verify they exist
	user, err := FindUserByName(da, username)
	if err != nil {
		return err
	}

	// try to update the item
	return da.UpdateItem(context.Background(), id, user.Id, name, itemType, value)
}

// Tries to delete an item
func DeleteItem(da DataAccess, id int) error {
	// try to delete the item
	return da.DeleteItem(context.Background(), id)
}

type ItemList struct {
	Username       string
	Items          *[]ItemEntry
	NetWorth       int64
	AssetTotal     int64
	LiabilityTotal int64
}

// Gets all of the items for a given user, and calculates certain analytics
func GetItems(da DataAccess, username string) (*ItemList, error) {
	// find the user and verify they exist
	user, err := FindUserByName(da, username)
	if err != nil {
		return nil, err
	}

	// try to get their items
	items, err := da.GetItemsByUser(context.Background(), user.Id)
	if err != nil {
		return nil, err
	}

	// calculate net worth, asset total, liability total
	var net, asset, liability int64
	for i := range *items {
		if (*items)[i].Type == ItemTypeAsset {
			net += (*items)[i].Value
			asset += (*items)[i].Value
		} else if (*items)[i].Type == ItemTypeLiability {
			net -= (*items)[i].Value
			liability += (*items)[i].Value
		}
	}

	return &ItemList{Username: username, Items: items, NetWorth: net, AssetTotal: asset, LiabilityTotal: liability}, nil
}

type getItemsRequest struct {
	Username string
}

type addItemRequest struct {
	Username string
	Name     string
	ItemType string
	Value    int64
}

type updateItemRequest struct {
	Id       int
	Username string
	Name     string
	ItemType string
	Value    int64
}

type deleteItemRequest struct {
	Id int
}

// Handles the incoming http requests for the item API
func (ih itemHandlers) ItemRequestHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	switch request.Method {
	case http.MethodOptions:
		return
	case http.MethodPost:
		// try to add a new item
		var addRequest addItemRequest
		err := json.NewDecoder(request.Body).Decode(&addRequest)
		if err != nil {
			fmt.Println("Failed to decode add item request: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		err = AddItem(ih.da, addRequest.Name, addRequest.ItemType, addRequest.Username, addRequest.Value)
		if err != nil {
			fmt.Println("Failed to add item: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	case http.MethodPut:
		// try to update an existing item
		var updateRequest updateItemRequest
		err := json.NewDecoder(request.Body).Decode(&updateRequest)
		if err != nil {
			fmt.Println("Failed to update item: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		err = UpdateItem(ih.da, updateRequest.Id, updateRequest.Name, updateRequest.ItemType, updateRequest.Username, updateRequest.Value)
		if err != nil {
			fmt.Println("Failed to update item: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		// try to delete an existing item
		var deleteRequest deleteItemRequest
		err := json.NewDecoder(request.Body).Decode(&deleteRequest)
		if err != nil {
			fmt.Println("Failed to delete item: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("Received delete request for item " + strconv.Itoa(deleteRequest.Id))

		err = DeleteItem(ih.da, deleteRequest.Id)
		if err != nil {
			fmt.Println("Failed to delete item: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(writer, "Invalid request method.", 405)
	}
}

func (ih itemHandlers) ItemListRequestHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	switch request.Method {
	case http.MethodOptions:
		return
	case http.MethodPost:
		// return a json of this user's items
		// decode the request
		var getRequest getItemsRequest
		err := json.NewDecoder(request.Body).Decode(&getRequest)
		if err != nil {
			fmt.Println("Failed to decode get items request: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		// look up the items
		fmt.Println("Received get items request for '" + getRequest.Username + "'")
		itemList, err := GetItems(ih.da, getRequest.Username)
		if err != nil {
			fmt.Println("Failed to get items: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		// respond with the item list
		json.NewEncoder(writer).Encode(itemList)
	default:
		http.Error(writer, "Invalid request method.", 405)
	}
}

func (ih itemHandlers) ItemDeleteRequestHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	switch request.Method {
	case http.MethodOptions:
		return
	case http.MethodPost:
		// try to delete an existing item
		var deleteRequest deleteItemRequest
		err := json.NewDecoder(request.Body).Decode(&deleteRequest)
		if err != nil {
			fmt.Println("Failed to delete item: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("Received delete request for item " + strconv.Itoa(deleteRequest.Id))

		err = DeleteItem(ih.da, deleteRequest.Id)
		if err != nil {
			fmt.Println("Failed to delete item: " + err.Error())
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	default:
		http.Error(writer, "Invalid request method.", 405)
	}
}
