package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestNewAccount verifies that the NewAccount factory correctly assigns variables
// and securely hashes the password using bcrypt.
func TestNewAccount(t *testing.T) {
	// 1. Arrange: Define our inputs
	firstName := "Arijit"
	lastName := "Singh"
	plainTextPassword := "supersecret123"

	// 2. Act: Call the function we want to test
	acc, err := NewAccount(firstName, lastName, plainTextPassword)

	// 3. Assert (Using Testify): Did it behave exactly as we expect?
	
	// Ensure no error was thrown during account creation
	assert.Nil(t, err)

	// Ensure the returned account is not nil
	assert.NotNil(t, acc)

	// Ensure the names mapped correctly
	assert.Equal(t, firstName, acc.FirstName)
	assert.Equal(t, lastName, acc.LastName)

	// SECURITY TESTS:
	
	// Ensure the password didn't save as plain text
	assert.NotEqual(t, plainTextPassword, acc.EncryptedPassword)
	
	// Check that the bcrypt hash is mathematically valid
	fmt.Println("Hash generated:", acc.EncryptedPassword)
	
	err = bcrypt.CompareHashAndPassword([]byte(acc.EncryptedPassword), []byte(plainTextPassword))
	assert.Nil(t, err, "The encrypted password should mathematically match the plain text password")
}
