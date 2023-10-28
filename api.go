package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHttpHandleFunc(s.handleLogin))

	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))

	router.HandleFunc("/account/{id}", withJWTAuth(makeHttpHandleFunc(s.handleAccountID),s.store))

	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleGetAccount(w, r)
	}
	if r.Method == http.MethodPost {
		return s.handleCreateAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleAccountID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == http.MethodGet {
		return s.handleGetAccountByID(w, r)
	}
	if r.Method == http.MethodDelete {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	account, err := s.store.GetAccountByNumber(int(req.Number))
	if err != nil {
		return err
	}

	if !account.validatePassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}


	token,err := createJWT(account)
	if err != nil {
		return err
	}

	resp := LoginResponse{
		Number: account.Number,
		Token: token,
	}

	return WriteJson(w, http.StatusOK, resp)
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJson(w, http.StatusOK, accounts)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	return WriteJson(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}
	defer r.Body.Close()

	account,err := NewAccount(createAccountReq.FirstName, createAccountReq.LastName, createAccountReq.Password)
	if err != nil {
		return err
	}

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJson(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}

	return WriteJson(w, http.StatusOK, nil)
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	tramsferReq := new(TransferRequest)
	if err := json.NewDecoder(r.Body).Decode(tramsferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	return WriteJson(w, http.StatusOK, tramsferReq)
}

func createJWT(account *Account) (string, error) {
	claims := jwt.MapClaims{
		"exp":           time.Now().Add(time.Hour * 24).Unix(),
		"accountNumber": account.Number,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func withJWTAuth(handlerFunc http.HandlerFunc,s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("withJWTAuth")

		tokenString := r.Header.Get("x-jwt-token")

		token, err := validateJWT(tokenString)
		if err != nil {
			WriteJson(w, http.StatusUnauthorized, ApiError{Error: "permission denied"})
			return
		}

		if !token.Valid {
			WriteJson(w, http.StatusUnauthorized, ApiError{Error: "permission denied"})
			return
		}

		userID,err := getID(r)
		if err != nil {
			WriteJson(w, http.StatusUnauthorized, ApiError{Error: "permission denied"})
			return
		}

		account,err := s.GetAccountByID(userID)
		if err != nil {
			WriteJson(w, http.StatusUnauthorized, ApiError{Error: "permission denied"})
			return
		}

		if account.Number != int64(token.Claims.(jwt.MapClaims)["accountNumber"].(float64)) {
			WriteJson(w, http.StatusUnauthorized, ApiError{Error: "permission denied"})
			return
		}


		handlerFunc(w, r)
	}

}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJson(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid id: %s", idStr)
	}
	return id, nil
}
