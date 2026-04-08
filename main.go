package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func seedAccount(store Storage, fname, lname, pw string) *Account {
	acc, err := NewAccount(fname, lname, pw)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("New account seeded =>", acc.Number)
	return acc
}

func seedAccounts(s Storage) {
	seedAccount(s, "Arijit", "Singh", "password123")
	seedAccount(s, "Krish", "Tester", "password123")
}

func main() {
	// 1. Create a custom terminal flag
	seed := flag.Bool("seed", false, "seed the database with dummy data")
	flag.Parse()

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

	// 2. Only run the seeder if the flag was passed!
	if *seed {
		fmt.Println("Seeding the database with dummy accounts...")
		seedAccounts(store)
	}

	server := newAPIServer(":3000", store)
	server.Run()
}
