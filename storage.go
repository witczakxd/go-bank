package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountByID(int) (*Account, error)
	GetAccounts() ([]*Account, error)
	GetAccountByNumber(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres password=gobank dbname=postgres sslmode=disable port=5433"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) init() error {
	return s.CreateAccountTable()
}

func (s *PostgresStore) CreateAccountTable() error {
	query := "CREATE TABLE IF NOT EXISTS accounts (id serial PRIMARY KEY, first_name varchar(255), last_name varchar(255), number bigint,encrypted_password  varchar(255),balance bigint,created_at timestamp);"
	
	_,err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateAccount(a *Account) error {
	query := "INSERT INTO accounts ( first_name, last_name, number,encrypted_password, balance, created_at) VALUES ($1,$2,$3,$4,$5,$6)"

	_,err := s.db.Exec(query, a.FirstName, a.LastName, a.Number,a.EncryptedPassword ,a.Balance, a.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	_,err := s.db.Exec("DELETE FROM accounts WHERE id = $1", id)
	return err
}

func (s *PostgresStore) GetAccountByNumber(number int) (*Account, error) {
	rows,err := s.db.Query("SELECT * FROM accounts WHERE number = $1", number)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account not found")
} 

func (s *PostgresStore) UpdateAccount(a *Account) error {
	return nil
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	rows,err := s.db.Query("SELECT * FROM accounts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account not found")
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows,err := s.db.Query("SELECT * FROM accounts")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}

	for rows.Next() {
		account,err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanIntoAccount (rows *sql.Rows) (*Account, error) {
	account := &Account{}
	err := rows.Scan(&account.ID, &account.FirstName, &account.LastName, &account.Number,&account.EncryptedPassword, &account.Balance, &account.CreatedAt)
	if err != nil {
		return nil, err
	}
	return account, nil
}