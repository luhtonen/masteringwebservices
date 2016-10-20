package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var database *sql.DB
var format string

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

type UserResponse struct {
	Error     string `json:"error"`
	ErrorCode int    `json:"code"`
}

func errorMessages(err int64) (int, int, string) {
	errorMessage := ""
	statusCode := 200
	errorCode := 0
	switch err {
	case 1062:
		errorMessage = "Duplicated entry"
		errorCode = 10
		statusCode = 409
	default:
		errorMessage = http.StatusText(int(err))
		errorCode = 0
		statusCode = int(err)
	}
	return errorCode, statusCode, errorMessage
}

func getFormat(r *http.Request) {
	format = r.URL.Query()["format"][0]
}

func setFormat(data interface{}) []byte {
	var apiOutput []byte
	if format == "json" {
		output, _ := json.Marshal(data)
		apiOutput = output
	} else if format == "xml" {
		output, _ := xml.Marshal(data)
		apiOutput = output
	}
	return apiOutput
}

func dbErrorParse(err string) (string, int64) {
	parts := strings.Split(err, ":")
	errorMessage := parts[1]
	code := strings.Split(parts[0], "Error ")
	errorCode, _ := strconv.ParseInt(code[1], 10, 32)
	return errorMessage, errorCode
}

func UserCreate(w http.ResponseWriter, r *http.Request) {
	newUser := User{}
	newUser.Name = r.FormValue("user")
	newUser.Email = r.FormValue("email")
	newUser.First = r.FormValue("first")
	newUser.Last = r.FormValue("last")

	fileString := ""
	f, _, err := r.FormFile("image1")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fileData, _ := ioutil.ReadAll(f)
		fileString = base64.StdEncoding.EncodeToString(fileData)
	}

	output, err := json.Marshal(newUser)
	fmt.Println("output:", string(output))
	if err != nil {
		fmt.Println("Something went wrong", err.Error())
	}

	response := UserResponse{}
	query := "INSERT INTO users set user_nickname='" + newUser.Name +
		"', user_first='" + newUser.First +
		"', user_last='" + newUser.Last + "', user_email='" + newUser.Email +
		"', user_image='" + fileString + "'"

	q, err := database.Exec(query)
	if err != nil {
		errorMessage, errorCode := dbErrorParse(err.Error())
		fmt.Println(errorMessage)
		errCode, httpCode, errMsg := errorMessages(errorCode)
		response.Error = errMsg
		response.ErrorCode = errCode
		http.Error(w, "Conflict", httpCode)
	}
	fmt.Println(q)
	createOutput, _ := json.Marshal(response)
	fmt.Fprintln(w, string(createOutput))
}

func UsersRetrieve(w http.ResponseWriter, r *http.Request) {
	getFormat(r)
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
	output := setFormat(response)
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

func UsersUpdate(w http.ResponseWriter, r *http.Request) {
	response := UserResponse{}
	params := mux.Vars(r)
	uid := params["id"]
	email := r.FormValue("email")

	var userCount int
	err := database.QueryRow("SELECT COUNT(user_id) FROM users WHERE user_id=?", uid).
		Scan(&userCount)

	if userCount == 0 {
		errCode, httpCode, msg := errorMessages(404)
		log.Println(errCode)
		log.Println(w, msg, httpCode)
		response.Error = msg
		response.ErrorCode = errCode
		http.Error(w, msg, httpCode)
	} else if err != nil {
		log.Println(err.Error())
	} else {
		_, uperr := database.Exec("UPDATE users set user_email=? where user_id=?", email, uid)
		if uperr != nil {
			_, errorCode := dbErrorParse(uperr.Error())
			_, httpCode, msg := errorMessages(errorCode)

			response.Error = msg
			response.ErrorCode = httpCode
			http.Error(w, msg, httpCode)
		} else {
			response.Error = "success"
			response.ErrorCode = 0
			output := setFormat(response)
			fmt.Fprintln(w, string(output))
		}
	}
}

func StartServer() {
	db, err := sql.Open("mysql", "gosocnet:gosocnet@/social_network")
	if err != nil {
		fmt.Println("cannot connec to database", err.Error())
		os.Exit(1)
	}
	database = db

	routes := mux.NewRouter()
	routes.HandleFunc("/api/users", UserCreate).Methods("POST")
	routes.HandleFunc("/api/users", UsersRetrieve).Methods("GET")
	routes.HandleFunc("/api/users/{id:[0-9]+}", GetUser).Methods("GET")
	routes.HandleFunc("/api/users/{id:[0-9]+}", UsersUpdate).Methods("PUT")
	http.Handle("/", routes)

	fmt.Println("Starting service...")
	http.ListenAndServe(":8080", nil)
}
