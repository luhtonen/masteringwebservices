package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var database *sql.DB

// API is Rest API struct
type API struct {
	Message string `json:"message"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"username"`
	Email string `json:"email"`
	First string `json:"first"`
	Last  string `json:"last"`
}

func Hello(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	name := urlParams["user"]
	HelloMessage := "Hello, " + name

	message := API{HelloMessage}
	output, err := json.Marshal(message)

	if err != nil {
		fmt.Println("Something went wrong!", err.Error())
	}

	fmt.Fprintf(w, string(output))
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	newUser := User{}
	newUser.Name = r.FormValue("user")
	newUser.Email = r.FormValue("email")
	newUser.First = r.FormValue("first")
	newUser.Last = r.FormValue("last")
	output, err := json.Marshal(newUser)
	fmt.Println(string(output))
	if err != nil {
		fmt.Println("Something went wrong", err.Error())
	}

	query := "INSERT INTO users set user_nickname='" + newUser.Name +
		"', user_first='" + newUser.First +
		"', user_last='" + newUser.Last + "', user_email='" + newUser.Email + "'"

	q, err := database.Exec(query)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(q)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	urlParams := mux.Vars(r)
	id := urlParams["id"]
	readUser := User{}
	err := database.QueryRow("select * from users where user_id=?", id).
		Scan(&readUser.ID, &readUser.Name, &readUser.First, &readUser.Last, &readUser.Email)
	switch {
	case err == sql.ErrNoRows:
		fmt.Fprint(w, "No such user")
	case err != nil:
		log.Fatal(err.Error())
	default:
		output, _ := json.Marshal(readUser)
		fmt.Fprintf(w, string(output))
	}
}

func main() {
	db, err := sql.Open("mysql", "gosocnet:gosocnet@/social_network")
	if err != nil {
		fmt.Println("cannot connec to database", err.Error())
		os.Exit(1)
	}
	database = db

	routes := mux.NewRouter()
	routes.HandleFunc("/api/{user:[0-9]+}", Hello)
	routes.HandleFunc("/api/user/create", CreateUser).Methods("GET")
	routes.HandleFunc("/api/user/read/{id:[0-9]+}", GetUser)
	http.Handle("/", routes)
	http.ListenAndServe(":8080", nil)
}
