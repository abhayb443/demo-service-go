// main.go

package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// User struct represents a user in our system
type User struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Age   int    `json:"age,omitempty"`
	Email string `json:"email,omitempty"`
}

var users []User

func main() {
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/v1/users", getUsers).Methods("GET")
	router.HandleFunc("/v1/users/{id}", getUser).Methods("GET")
	router.HandleFunc("/v1/users", createUser).Methods("POST")
	router.HandleFunc("/v1/users/{id}", updateUser).Methods("PUT")
	router.HandleFunc("/v1/users/{id}", deleteUser).Methods("DELETE")
	router.HandleFunc("/v2/users/{id}", patchUser).Methods("PATCH")

	log.Fatal(http.ListenAndServe(":8000", router))
}

// Get all users
func getUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("get all user started...")
	json.NewEncoder(w).Encode(users)
	log.Println("get all user completed...")
}

// Get single user by ID
func getUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	log.Println("get specific user with id started...")
	for _, item := range users {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			log.Println("get specific user with id completed...")
			return
		}
	}
	json.NewEncoder(w).Encode(&User{})
}

// Create a new user
func createUser(w http.ResponseWriter, r *http.Request) {
	log.Println("creating user...")

	var user User
	_ = json.NewDecoder(r.Body).Decode(&user)
	users = append(users, user)

	log.Println("user creation complete...")

	// Encode the updated user as JSON and send it in the response
	json.NewEncoder(w).Encode(user)
}

// Update user by ID
func updateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("updating user...")

	// Parse URL parameters
	params := mux.Vars(r)

	// Iterate over the users
	for index, item := range users {
		// Check if the user ID matches the parameter
		if item.ID == params["id"] {
			// Log the found user
			log.Printf("user found: %+v", item)

			// Remove the user from the slice
			users = append(users[:index], users[index+1:]...)

			// Decode the request body into a new user struct
			var user User
			err := json.NewDecoder(r.Body).Decode(&user)
			if err != nil {
				log.Printf("error decoding request body: %v", err)
				http.Error(w, "failed to decode request body", http.StatusBadRequest)
				return
			}

			// Update the user ID
			user.ID = params["id"]

			// Append the updated user to the slice
			users = append(users, user)

			// Log the updated user
			log.Printf("user updated: %+v", user)

			// Encode the updated user as JSON and send it in the response
			json.NewEncoder(w).Encode(user)

			log.Println("user update complete.")
			return
		}
	}

	// If the user is not found, log and return the list of users
	log.Println("user not found.")
	json.NewEncoder(w).Encode(users)
}

// Delete user by ID
func deleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	log.Println("delete user started...")

	for index, item := range users {
		log.Println("checking if user exists...")
		if item.ID == params["id"] {
			log.Println("user found for deletion...")
			users = append(users[:index], users[index+1:]...)
			log.Println("delete user completed...")
			break
		}
	}
	// Encode the updated user as JSON and send it in the response
	json.NewEncoder(w).Encode(users)
}

// Define a function to handle PATCH requests for updating users
func patchUser(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters to get the user ID
	params := mux.Vars(r)

	log.Println("Patch user with id started...")
	// Find the user in the slice by ID
	for index, item := range users {
		if item.ID == params["id"] {
			log.Println("Patch user id found...")
			// Decode the request body into a map to get the updates
			var updates map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&updates)
			if err != nil {
				log.Printf("Error decoding request body: %v", err)
				http.Error(w, "Failed to decode request body", http.StatusBadRequest)
				return
			}

			log.Println("Patch applying in progress...")
			// Apply updates to the user
			for key, value := range updates {
				switch key {
				case "name":
					log.Println("Patching name started...")
					users[index].Name = value.(string)
					log.Println("Patching name completed...")
				case "age":
					log.Println("Patching age started...")
					users[index].Age = int(value.(float64))
					log.Println("Patching age completed...")
				case "email":
					log.Println("Patching email started...")
					users[index].Email = value.(string)
					log.Println("Patching email completed...")
				}
			}

			log.Println("Patching user completed...")
			// Encode the updated user as JSON and send it in the response
			json.NewEncoder(w).Encode(users[index])
			return
		}
	}

	// If the user is not found, return an error response
	http.Error(w, "User not found", http.StatusNotFound)
}
