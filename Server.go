package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"html/template"
)

type msg struct {
	Num int
}

func main() {
	http.HandleFunc("/socket", wsHandler)
	http.HandleFunc("/", rootHandler)

	fmt.Println("Starting server")
	panic(http.ListenAndServe(":8080", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("index.html")
	if err != nil {
		fmt.Println("Could not open file.", err)
	}
	err = template.ExecuteTemplate(w, "index.html", r)
	if err != nil {
		fmt.Println("Could not parse template.", err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "http://"+r.Host {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	go echo(conn)
}

func echo(conn *websocket.Conn) {
	for {
		m := msg{}

		err := conn.ReadJSON(&m)
		if err != nil {
			fmt.Println("Error reading json.", err)
		}

		fmt.Printf("Got message: %#v\n", m)

		if err = conn.WriteJSON(m); err != nil {
			fmt.Println(err)
		}
	}
}
