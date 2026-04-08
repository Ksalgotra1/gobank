package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"gobank/internal/api"
	"gobank/internal/models"
	"gobank/internal/storage"
)

func seedAccount(store storage.Storage, fname, lname, pw string) *models.Account {
	acc, err := models.NewAccount(fname, lname, pw)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("New account seeded =>", acc.Number)
	return acc
}

func seedAccounts(s storage.Storage) {
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

	store, err := storage.NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	// 2. Only run the seeder if the flag was passed!
	if *seed {
		fmt.Println("Seeding the database with dummy accounts...")
		seedAccounts(store)
	}

	server := api.NewServer(":3000", store)
	server.Run()
}
