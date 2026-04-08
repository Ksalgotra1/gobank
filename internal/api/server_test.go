package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gobank/internal/models"
)

// 1. Create a Fake Database Interface (Mocking)
type MockStore struct{}

func (m *MockStore) CreateAccount(*models.Account) error { return nil }
func (m *MockStore) DeleteAccount(int) error             { return nil }
func (m *MockStore) UpdateAccount(*models.Account) error { return nil }
func (m *MockStore) GetAccountByID(int) (*models.Account, error) { return nil, nil }
func (m *MockStore) GetAccountByNumber(int) (*models.Account, error) { return nil, nil }
func (m *MockStore) Transfer(amount int, fromAccountID int, toAccountID int) error { return nil }
func (m *MockStore) Init() error                         { return nil }

// This is the function we actually care about testing today!
func (m *MockStore) GetAccounts() ([]*models.Account, error) {
	fakeAccounts := []*models.Account{
		{ID: 1, FirstName: "Arijit", LastName: "Singh", Number: 554433, Balance: 0, CreatedAt: time.Now()},
		{ID: 2, FirstName: "Krish", LastName: "Tester", Number: 112233, Balance: 0, CreatedAt: time.Now()},
	}
	return fakeAccounts, nil
}

func TestHandleGetAccount(t *testing.T) {
	// 2. Set up our completely fake internal state
	mockStore := &MockStore{}
	server := NewServer(":3000", mockStore)

	// 3. Create a fake HTTP Request (just like Postman does!)
	req, err := http.NewRequest("GET", "/account", nil)
	assert.Nil(t, err)

	// 4. Create a fake HTTP Response Recorder (to catch the server's reply)
	rr := httptest.NewRecorder()

	// 5. Fire the request directly into our function
	handler := http.HandlerFunc(makeHTTPHandleFunc(server.handleAccount))
	handler.ServeHTTP(rr, req)

	// 6. Assertions! If these fail, our API is broken.
	assert.Equal(t, http.StatusOK, rr.Code, "Expected HTTP Status 200")

	// 7. Verify the JSON payload natively
	var accounts []*models.Account
	err = json.NewDecoder(rr.Body).Decode(&accounts)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(accounts), "Expected exactly 2 mock accounts")
	assert.Equal(t, "Arijit", accounts[0].FirstName)
}
