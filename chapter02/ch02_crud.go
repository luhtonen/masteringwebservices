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

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"username"`
	Email string `json:"email"`
	First string `json:"first"`
	Last  string `json:"last"`
}

type Users struct {
	Users []User `json:"users"`
}

func UserCreate(w http.ResponseWriter, r *http.Request) {
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

func UsersRetrieve(w http.ResponseWriter, r *http.Request) {
	start := 0
	limit := 10
	next := start + limit

	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Link", "<http://localhost:8080/api/users?start="+string(next)+
		"; rel=\"next\"")

	rows, _ := database.Query("select * from users LIMIT ?", limit)
	response := Users{}
	for rows.Next() {
		user := User{}
		rows.Scan(&user.ID, &user.Name, &user.First, &user.Last, &user.Email)
		response.Users = append(response.Users, user)
	}
	output, _ := json.Marshal(response)
	fmt.Fprint(w, string(output))
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Pragma", "no-cache")
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
	routes.HandleFunc("/api/users", UserCreate).Methods("POST")
	routes.HandleFunc("/api/users", UsersRetrieve).Methods("GET")
	routes.HandleFunc("/api/user/{id:[0-9]+}", GetUser).Methods("GET")
	http.Handle("/", routes)
	http.ListenAndServe(":8080", nil)
}
