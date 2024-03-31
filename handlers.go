package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

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

	// Prepare the SQL statement with placeholders
	stmt, err := db.Prepare("SELECT id, name, age, email FROM users WHERE id = ? OR name = ? OR email = ?")
	if err != nil {
		handleDBError(w, err, "Failed to prepare SQL statement")
		return
	}
	defer stmt.Close()

	// Execute the prepared statement with parameters
	err = stmt.QueryRow(userID, userID, userID).Scan(&user.ID, &user.Name, &user.Age, &user.Email)

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
		log.Println("Failed to decode request body:", err)
		return
	}

	// Begin a new transaction
	tx, err := db.Begin()
	if err != nil {
		handleDBError(w, err, "Failed to begin transaction")
		return
	}

	// Prepare the SQL statement with placeholders
	stmt, err := tx.Prepare("INSERT INTO users (name, age, email) VALUES (?, ?, ?)")
	if err != nil {
		handleDBError(w, err, "Failed to prepare SQL statement")
		tx.Rollback() // Rollback the transaction on error
		return
	}
	defer stmt.Close()

	// Insert user data into the database
	// Execute the prepared statement with parameters
	result, err := stmt.Exec(user.Name, user.Age, user.Email)
	if err != nil {
		handleDBError(w, err, "Failed to execute SQL statement")
		tx.Rollback() // Rollback the transaction on error
		return
	}

	// Check the number of rows affected by the insert operation
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		handleDBError(w, err, "Failed to check rows affected after insert")
		tx.Rollback() // Rollback the transaction on error
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"success": false, "message": "No rows affected after insert"}`, http.StatusInternalServerError)
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

	// Prepare the SQL statement with placeholders
	stmt, err := tx.Prepare("UPDATE users SET name = ?, age = ?, email = ? WHERE id = ?")
	if err != nil {
		handleDBError(w, err, "Failed to prepare SQL statement")
		tx.Rollback() // Rollback the transaction on error
		return
	}
	defer stmt.Close()

	// Execute the UPDATE query to update the user in the database
	result, err := stmt.Exec(user.Name, user.Age, user.Email, userID)
	if err != nil {
		handleDBError(w, err, `{"success": false, "message": "message": "Failed to update user in database"}`)
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

	stmt, err := tx.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		handleDBError(w, err, "Failed to prepare SQL statement")
		tx.Rollback() // Rollback the transaction on error
		return
	}
	defer stmt.Close()

	// Query the database to get the user by ID
	result, err := stmt.Exec(userID)
	if err != nil {
		handleDBError(w, err, `{"success": false, "message": "Failed to delete user from database"}`)
		tx.Rollback() // Rollback the transaction on error
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

	// Prepare the SQL statement with placeholders
	stmt, err := tx.Prepare(query)

	if err != nil {
		handleDBError(w, err, "Failed to prepare SQL statement")
		tx.Rollback() // Rollback the transaction on error
		return
	}
	defer stmt.Close()

	// Execute the UPDATE query to update the user in the database
	_, err = stmt.Exec(args...)

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

func handleDBError(w http.ResponseWriter, err error, errorMessage string) {
	http.Error(w, errorMessage, http.StatusInternalServerError)
	log.Println(errorMessage+":", err)
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	log.Println("Health started...")
	json.NewEncoder(w).Encode("Hello from Golang Demo Service")
	log.Println("get all user completed...")
}
