package storage

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"gobank/internal/models"
)

type Storage interface {
	CreateAccount(*models.Account) error
	DeleteAccount(int) error
	UpdateAccount(*models.Account) error
	GetAccounts() ([]*models.Account, error)
	GetAccountByID(int) (*models.Account, error)
	GetAccountByNumber(int) (*models.Account, error)
	Transfer(amount int, fromAccountID int, toAccountID int) error
	Init() error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	connStr := fmt.Sprintf("host=%s user=postgres password=mysecretpassword dbname=postgres sslmode=disable", host)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil

}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	query := `create table if not exists account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		encrypted_password varchar(100),
		balance integer,
		created_at timestamp
	)`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(acc *models.Account) error {
	query := `
	INSERT INTO account (first_name, last_name, number, encrypted_password, balance, created_at)
	VALUES($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.Query(
		query,
		acc.FirstName,
		acc.LastName,
		acc.Number,
		acc.EncryptedPassword,
		acc.Balance,
		acc.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(*models.Account) error {
	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	_, err := s.db.Query("DELETE FROM ACCOUNT WHERE id = $1", id)

	return err
}

func (s *PostgresStore) Transfer(amount int, fromAccountID int, toAccountID int) error {
	// 1. Begin the ACID Transaction!
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Helper: If anything goes wrong below, ROLLBACK the transaction securely!
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 2. Check if the sender has enough funds AND lock their row from concurrent modification (FOR UPDATE)
	var balance int
	err = tx.QueryRow("SELECT balance FROM account WHERE id = $1 FOR UPDATE", fromAccountID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("sender account not found: %v", err)
	}

	if balance < amount {
		err = fmt.Errorf("insufficient funds") // This triggers the defer rollback!
		return err
	}

	// 3. Subtract funds from the Sender
	_, err = tx.Exec("UPDATE account SET balance = balance - $1 WHERE id = $2", amount, fromAccountID)
	if err != nil {
		return err
	}

	// 4. Add funds to the Receiver
	res, err := tx.Exec("UPDATE account SET balance = balance + $1 WHERE id = $2", amount, toAccountID)
	if err != nil {
		return err
	}
	
	// Double check the receiver actually exists before committing
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		err = fmt.Errorf("receiver account not found") // Triggers rollback!
		return err
	}

	// 5. Success! Commit the changes permanently. 
	err = tx.Commit()
	return err
}

func (s *PostgresStore) GetAccounts() ([]*models.Account, error) {
	rows, err := s.db.Query("SELECT * FROM ACCOUNT")

	if err != nil {
		return nil, err
	}

	accounts := []*models.Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil

}

func (s *PostgresStore) GetAccountByID(id int) (*models.Account, error) {
	rows, err := s.db.Query("SELECT * FROM ACCOUNT WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStore) GetAccountByNumber(number int) (*models.Account, error) {
	rows, err := s.db.Query("SELECT * FROM ACCOUNT WHERE number = $1", number)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account with number %d not found", number)
}

func scanIntoAccount(rows *sql.Rows) (*models.Account, error) {

	account := new(models.Account)
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.EncryptedPassword,
		&account.Balance,
		&account.CreatedAt,
	)

	return account, err
}
