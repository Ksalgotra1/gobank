package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file immediately when the program starts
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.init(); err != nil {
		log.Fatal(err)
	}

	server := newAPIServer(":3000", store)
	server.Run()
}
