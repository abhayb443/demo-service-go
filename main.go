// main.go

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	_ "github.com/go-sql-driver/mysql"
)

// User struct represents a user in our system
type User struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Age   int    `json:"age,omitempty"`
	Email string `json:"email,omitempty"`
}

func main() {
	// Initialize the router
	router := mux.NewRouter()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Connect to the database
	db, err := connectToDatabase()
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	// Register routes
	registerRoutes(router, db, loggingMiddleware)

	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":8000", router))
}
