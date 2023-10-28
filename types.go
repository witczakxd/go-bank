package main

import (
	"math/rand"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	Number int64 `json:"number"`
	Token string `json:"token"`
}

type LoginRequest struct {
	Number int64 `json:"number"`
	Password string `json:"password"`
}

type TransferRequest struct {
	ToAccount int `json:"to_account"`
	Amount int `json:"amount"`
} 

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Password string `json:"password"`
}

type Account struct {
	ID int `json:"id"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Number int64 `json:"number"`
	EncryptedPassword string `json:"-"`
	Balance int64 `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *Account) validatePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword), []byte(password)) == nil
}

func NewAccount (firstName, lastName,password string) (*Account,error) {
	encpw,err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		FirstName: firstName,
		LastName: lastName,
		EncryptedPassword: string(encpw),
		Number: int64(rand.Intn(10000000)),
		Balance: 0,
		CreatedAt: time.Now().UTC(),
	},nil
}
