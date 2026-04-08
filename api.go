package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func newAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))

	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountByID)))

	router.HandleFunc("/transfer", makeHTTPHandleFunc(s.handleTransfer))

	log.Println("JSON API running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	acc, err := s.store.GetAccountByNumber(int(req.Number))
	if err != nil {
		return err
	}

	// 1. Compare the Hashed Database password to the Raw password
	if err := bcrypt.CompareHashAndPassword([]byte(acc.EncryptedPassword), []byte(req.Password)); err != nil {
		// If the math fails, reject them violently!
		return fmt.Errorf("not authenticated: invalid password")
	}

	// 2. If the math succeeded, the password is correct! Let's generate their token.
	token, err := createJWT(acc)
	if err != nil {
		return err
	}

	// 3. Instead of echoing the request, send them their brand new token!
	return WriteJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

// GET/account
// return accounts
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := getID(r)

		if err != nil {
			return err
		}

		// AUTHORIZATION CHECK: Prevent User A from reading User B's account!
		if err := matchJWT(r, id); err != nil {
			return err
		}

		account, err := s.store.GetAccountByID(id)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, account)

	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("mehtod not allowed %s", r.Method)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}

	account, err := NewAccount(createAccountReq.FirstName, createAccountReq.LastName, createAccountReq.Password)
	if err != nil {
		return err
	}
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	// Get the ID from URL
	targetID, err := getID(r)

	if err != nil {
		return err
	}

	//  AUTHORIZATION CHECK (Using helper!)
	if err := matchJWT(r, targetID); err != nil {
		return err
	}

	if err := s.store.DeleteAccount(targetID); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int{"deleted": targetID})
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()

	transferReq := new(TransferRequest)

	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, transferReq)
}

func createJWT(account *Account) (string, error) {
	claims := jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
		"userID":        account.ID,
	}
	// Build the token using the HS256 algorithm
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Fetch your server's secret from .env
	secret := os.Getenv("JWT_SECRET")
	// Cryptographically sign the token
	return token.SignedString([]byte(secret))
}

func withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")

		token, err := validateJWT(tokenString)

		if err != nil {
			WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token"})
			return
		}

		// 1. Grab the payload (claims) out of the validated token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

			// 2. Put the claims into the Request Context
			ctx := context.WithValue(r.Context(), "user_claims", claims)

			// 3. Create a cloned HTTP Request with the new Context attached
			r = r.WithContext(ctx)

			// 4. Pass the new request down to the actual handler (like handleTransfer)
			handlerFunc(w, r)
			return
		}
		WriteJSON(w, http.StatusForbidden, ApiError{Error: "Permission Denied"})
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	token, err := jwt.Parse(tokenString,
		func(token *jwt.Token) (any, error) {
			// cast to []byte
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	return token, nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func getClaimsFromContext(r *http.Request) (jwt.MapClaims, error) {
	claims, ok := r.Context().Value("user_claims").(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("no claims found in context or wrong type")
	}
	return claims, nil
}

// we need to make our handler functions that return an err to http.HandlerFunc type which dont
// we could have kept our handler functions with the same function type
// but we would have needed to handle errors in each explicitly making it complex
// so we instead change our function signature and handle the err at one place
type apiFunc func(http.ResponseWriter, *http.Request) error // type of our handler function

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			// handle the error
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id %s", idStr)
	}

	return id, nil
}

func matchJWT(r *http.Request, targetAccountID int) error {
	claims, err := getClaimsFromContext(r)
	if err != nil {
		return err
	}

	tokenID := int(claims["userID"].(float64))

	if tokenID != targetAccountID {
		return fmt.Errorf("permission denied: you do not own account %d", targetAccountID)
	}

	return nil
}
