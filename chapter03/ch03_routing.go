package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc"
	rjson "github.com/gorilla/rpc/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
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

type CreateResponse struct {
	Error     string `json:"error"`
	ErrorCode int    `json:"code"`
}

type ErrMsg struct {
	ErrCode    int
	StatusCode int
	Msg        string
}

func errorMessages(err int64) ErrMsg {
	var em ErrMsg
	errorMessage := ""
	statusCode := 200
	errorCode := 0
	switch err {
	case 1062:
		errorMessage = "Duplicated entry"
		errorCode = 10
		statusCode = 409
	}
	em.ErrCode = errorCode
	em.StatusCode = statusCode
	em.Msg = errorMessage
	return em
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
	output, err := json.Marshal(newUser)
	fmt.Println("output:", string(output))
	if err != nil {
		fmt.Println("Something went wrong", err.Error())
	}

	response := CreateResponse{}
	query := "INSERT INTO users set user_nickname='" + newUser.Name +
		"', user_first='" + newUser.First +
		"', user_last='" + newUser.Last + "', user_email='" + newUser.Email + "'"

	q, err := database.Exec(query)
	if err != nil {
		errorMessage, errorCode := dbErrorParse(err.Error())
		fmt.Println(errorMessage)
		errMsg := errorMessages(errorCode)
		response.Error = errMsg.Msg
		response.ErrorCode = errMsg.ErrCode
		http.Error(w, "Conflict", errMsg.StatusCode)
	}
	fmt.Println(q)
	createOutput, _ := json.Marshal(response)
	fmt.Fprintln(w, string(createOutput))
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

type RPCAPIArguments struct {
	Message string
}

type RPCAPIResponse struct {
	Message string
}

func getLength(message string) string {
	length := utf8.RuneCountInString(message)
	return strconv.FormatInt(int64(length), 10)
}

type StringService struct{}

func (h *StringService) Length(r *http.Request, arguments *RPCAPIArguments, reply *RPCAPIResponse) error {
	reply.Message = "Your string is " + getLength(arguments.Message) + " characters long"
	return nil
}

/* rpc test with following command:
curl -X POST -H "Content-Type: application/json" -d '{"method":"StringService.Length","params":[{"Message":"Testing the service, how long is it?"}], "id":"1"}' http://localhost:10000/rpc
*/

/* error test:
curl -d 'user=edu&email=edufinn&first=Edu&last=Finn' http://localhost:8080/api/users
same but more verbose
curl -v -d 'user=edu&email=edufinn&first=Edu&last=Finn' http://localhost:8080/api/users
*/
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

	fmt.Println("Starting service...")
	s := rpc.NewServer()
	s.RegisterCodec(rjson.NewCodec(), "application/json")
	s.RegisterService(new(StringService), "")
	http.Handle("/rpc", s)
	//http.ListenAndServe(":10000", nil)
	http.ListenAndServe(":8080", nil)
}
