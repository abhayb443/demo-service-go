package main

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
)

func registerRoutes(router *mux.Router, db *sql.DB, middleware mux.MiddlewareFunc) {
	// Define routes
	router.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) { getUsers(w, r, db) }).Methods("GET")
	router.HandleFunc("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) { getUser(w, r, db) }).Methods("GET")
	router.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) { createUser(w, r, db) }).Methods("POST")
	router.HandleFunc("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) { updateUser(w, r, db) }).Methods("PUT")
	router.HandleFunc("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) { deleteUser(w, r, db) }).Methods("DELETE")
	router.HandleFunc("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) { patchUser(w, r, db) }).Methods("PATCH")

	// Apply middleware to all routes
	router.Use(middleware)
}
