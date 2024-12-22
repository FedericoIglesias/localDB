package main

import (
	"encoding/json"
	"fmt"

	"github.com/FedericoIglesias/local_db/local_db"
)

const version = "0.0.1"

type Address struct {
	City   string
	Number json.Number
}

type User struct {
	Name     string
	LastName string
	Age      json.Number
	Address  Address
}

func main() {

	fmt.Printf("DB Versión %s \n ", version)

	dir := "./"

	db, err := local_db.New(dir, nil)

	if err != nil {
		fmt.Println("is a feature")
	}

	person := []User{
		{"Fede", "Iglesias", "28", Address{"Ponferrada", "7"}},
		{"Remo", "Pepe", "10", Address{"Bs.As.", "10"}},
	}

	for _, value := range person {
		db.Write("users", value.Name, User{
			Name:     value.Name,
			LastName: value.LastName,
			Age:      value.Age,
			Address:  value.Address,
		})
	}

	records, err := db.ReadAll("users")

	if err != nil {
		fmt.Println("Is another feature")
	}

	fmt.Println(records)

	allUsers := []User{}

	for _, f := range records {
		personFound := User{}
		if err := json.Unmarshal([]byte(f), &personFound); err != nil {
			fmt.Println(err)
		}
		allUsers = append(allUsers, personFound)
	}
	fmt.Println(allUsers)

	// if err = db.Delete("user","Fede"); err != nil{
	// 	fmt.Println("I swear is a feature")
	// }

}
