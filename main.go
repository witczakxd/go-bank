package main

import (
	"flag"
	"fmt"
	"log"
)

func seedAccount(store Storage,fname,lname,pw string) *Account {
	acc,err := NewAccount(fname,lname,pw)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(acc); err != nil {
		log.Fatal(err)
	}

	fmt.Println("created account", acc.Number)

	return acc
} 

func seedAccounts(s Storage) {
	seedAccount(s,"a","b","c")
}

func main() {
	seed:= flag.Bool("seed",false,"seed the database")
	flag.Parse()


	store,err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.init(); err != nil {
		log.Fatal(err)
	}

	if *seed {
		fmt.Println("seeding database")
		//seed
		seedAccounts(store)
	}

	server := NewAPIServer(":3000",store)

	server.Run()
}