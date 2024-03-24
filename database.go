// database.go

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func connectToDatabase() (*sql.DB, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)
	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", dataSourceName)

	if err != nil {
		log.Fatal("Error opening database connection:", err)
	}
	// defer db.Close()

	// Set connection pool parameters
	db.SetMaxOpenConns(10) // Maximum number of open connections
	db.SetMaxIdleConns(5)  // Maximum number of idle connections

	// Ping the database to check if the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal("Error testing database connection:", err)
	}

	fmt.Println("Database connection successful")

	return db, nil
}
