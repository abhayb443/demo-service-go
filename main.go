// main.go

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
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
	// Initialize the router
	router := mux.NewRouter()

	// Connect to the database
	db, err := connectToDatabase()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	// Register routes
	registerRoutes(router, db)

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8000", router))
}

func connectToDatabase() (*sql.DB, error) {
	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", "root:Fastrack@123@tcp(localhost:3306)/userData")
	if err != nil {
		return nil, err
	}

	// Ping the database to check if the connection is successful
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to the database!")
	return db, nil
}

func registerRoutes(router *mux.Router, db *sql.DB) {
	// Define routes
	router.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) { getUsers(w, r, db) }).Methods("GET")
	router.HandleFunc("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) { getUser(w, r, db) }).Methods("GET")
	router.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) { createUser(w, r, db) }).Methods("POST")
	router.HandleFunc("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) { updateUser(w, r, db) }).Methods("PUT")
	router.HandleFunc("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) { deleteUser(w, r, db) }).Methods("DELETE")
	router.HandleFunc("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) { patchUser(w, r, db) }).Methods("PATCH")
}

func handleDBError(w http.ResponseWriter, err error, errorMessage string) {
	http.Error(w, errorMessage, http.StatusInternalServerError)
	log.Println(errorMessage+":", err)
}

// Get all users
func getUsers(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("get all users started...")

	// Query the database to get all users
	rows, err := db.Query("SELECT id, name, age, email FROM users")
	if err != nil {
		http.Error(w, `{"success": false, "message": "Failed to fetch users from database"}`, http.StatusInternalServerError)
		log.Println("Failed to fetch users from database:", err)
		return
	}
	defer rows.Close() // defer rows.Close() immediately after rows assignment

	var users []User

	// Iterate over the rows and scan data into User structs
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Age, &user.Email)
		if err != nil {
			http.Error(w, `{"success": false, "message": "Failed to fetch users from database"}`, http.StatusInternalServerError)
			log.Println("Failed to fetch users from database:", err)
			return
		}
		users = append(users, user)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		http.Error(w, `{"success": false, "message": "Failed to fetch users from database"}`, http.StatusInternalServerError)
		log.Println("Failed to fetch users from database:", err)
		return
	}

	// Encode users as JSON and send response
	json.NewEncoder(w).Encode(users)

	log.Println("get all users completed...")
}

// Get single user by ID
func getUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	params := mux.Vars(r)
	userID := params["id"]

	log.Println("get specific user with id started...")

	// Query the database to get the user by ID
	var user User

	err := db.QueryRow("SELECT id, name, age, email FROM users WHERE id = ? OR name = ? OR email = ?", userID, userID, userID).Scan(&user.ID, &user.Name, &user.Age, &user.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"success": false, "message": "User not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"success": false, "message": "Failed to fetch user from database"}`, http.StatusInternalServerError)
			log.Println("Failed to fetch user from database:", err)
		}
		return
	}

	// Encode the user as JSON and send response
	json.NewEncoder(w).Encode(user)
	log.Println("get specific user with id completed...")
}

// Create a new user
func createUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("creating user...")

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		http.Error(w, `{"success": false, "message": "Failed to decode request body"}`, http.StatusBadRequest)
		return
	}

	// Begin a new transaction
	tx, err := db.Begin()
	if err != nil {
		handleDBError(w, err, "Failed to begin transaction")
		return
	}

	// Insert user data into the database
	_, err = tx.Exec("INSERT INTO users (name, age, email) VALUES (?, ?, ?)", user.Name, user.Age, user.Email)
	if err != nil {
		http.Error(w, `{"success": false, "message": "Failed to insert into database"}`, http.StatusInternalServerError)
		tx.Rollback() // Rollback the transaction on error
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		handleDBError(w, err, "Failed to commit transaction")
		return
	}

	w.WriteHeader(http.StatusCreated)
	log.Println("user creation complete...")
	// Encode the updated user as JSON and send it in the response
	json.NewEncoder(w).Encode(user)
}

// Update user by ID
func updateUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("updating user...")

	// Parse URL parameters
	params := mux.Vars(r)
	userID := params["id"]

	// Decode the request body into a new user struct
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("error decoding request body: %v", err)
		http.Error(w, `{"success": false, "message": "failed to decode request body"}`, http.StatusBadRequest)
		return
	}

	// Update the user ID
	user.ID = userID

	// Begin a new transaction
	tx, err := db.Begin()
	if err != nil {
		handleDBError(w, err, "Failed to begin transaction")
		return
	}

	// Execute the UPDATE query to update the user in the database
	result, err := tx.Exec("UPDATE users SET name = ?, age = ?, email = ? WHERE id = ?", user.Name, user.Age, user.Email, userID)
	if err != nil {
		handleDBError(w, err, `{"success": false, "message": "Failed to update user in database"}`)
		tx.Rollback() // Rollback the transaction on error
		return
	}

	// Check the number of rows affected by the update operation
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		handleDBError(w, err, `{"success": false, "message": "Failed to check rows affected after update"}`)
		return
	}

	// If no rows were affected, it means the user with the given ID doesn't exist
	if rowsAffected == 0 {
		http.Error(w, `{"success": false, "message": "User with ID `+userID+` doesn't exist"}`, http.StatusNotFound)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		handleDBError(w, err, "Failed to commit transaction")
		return
	}

	// Respond with the updated user
	json.NewEncoder(w).Encode(user)

	log.Println("user update complete.")
}

// Delete user by ID
func deleteUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	params := mux.Vars(r)
	userID := params["id"]
	log.Println("delete user started...")

	// Begin a new transaction
	tx, err := db.Begin()
	if err != nil {
		handleDBError(w, err, "Failed to begin transaction")
		return
	}

	// Query the database to get the user by ID
	result, err := tx.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		handleDBError(w, err, `{"success": false, "message": "Failed to delete user from database"}`)
		return
	}

	// Check the number of rows affected by the delete operation
	numRowsAffected, err := result.RowsAffected()
	if err != nil {
		handleDBError(w, err, `{"success": false, "message": "Failed to delete user from database"}`)
		tx.Rollback() // Rollback the transaction on error
		return
	}

	// If no rows were affected, it means the user with the given ID doesn't exist
	if numRowsAffected == 0 {
		http.Error(w, `{"success": false, "message": "User not found"}`, http.StatusNotFound)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		handleDBError(w, err, "Failed to commit transaction")
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User with ID %s deleted successfully", userID)

	log.Println("delete user completed...")
}

// Define a function to handle PATCH requests for updating users
func patchUser(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse URL parameters to get the user ID
	params := mux.Vars(r)
	userID := params["id"]

	log.Println("Patch user with id started...")

	// Decode the request body into a map to get the updates
	var updates map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, `{"success": false, "message": "Failed to decode request body"}`, http.StatusBadRequest)
		return
	}

	// Construct the SQL query to update the user based on the provided updates
	var query string
	var args []interface{}
	query = "UPDATE users SET "
	for key, value := range updates {
		query += key + "=?, "
		args = append(args, value)
		log.Println(args)
	}
	// Remove the trailing comma and space
	log.Println(query)
	query = query[:len(query)-2]
	log.Println(query)
	query += " WHERE id = ?"
	log.Println(query)
	args = append(args, userID)
	log.Println(args)

	// Begin a new transaction
	tx, err := db.Begin()
	if err != nil {
		handleDBError(w, err, "Failed to begin transaction")
		return
	}

	// Execute the UPDATE query to update the user in the database
	_, err = tx.Exec(query, args...)
	if err != nil {
		handleDBError(w, err, `{"success": false, "message": "Failed to update user in database"}`)
		tx.Rollback() // Rollback the transaction on error
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		handleDBError(w, err, "Failed to commit transaction")
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User with ID %s patched successfully", userID)

	log.Println("Patching user completed...")
}
