package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	"net/http"
	"strconv"
)

var addr = ":12345"

func echoLengthServer(ws *websocket.Conn) {
	var msg string

	for {
		websocket.Message.Receive(ws, &msg)
		fmt.Println("Got message", msg)
		length := len(msg)
		if err := websocket.Message.Send(ws, strconv.FormatInt(int64(length), 10)); err != nil {
			fmt.Println("Can't send message length", err.Error())
			break
		}
	}
}

func websocketListen() {
	http.Handle("/length", websocket.Handler(echoLengthServer))
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic("listenAndServe: " + err.Error())
	}
}

func main() {
	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websocket.html")

	})
	websocketListen()
}
