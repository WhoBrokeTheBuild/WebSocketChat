package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"html/template"
	"strconv"
	"time"
	"sync"
)

type JoinMessage struct {
	name string
}

type ChatMessageIn struct {
	message string
}

type ChatMessageOut struct {
	name string
	message string
	time string
}

type ChatConn struct {
	name string
    conn *websocket.Conn
}

var allConnsMutex sync.RWMutex
var allConns []ChatConn

func getUnixTimeStr() string {
	return strconv.Itoa(int(time.Now().Unix()))
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

	go handleChat(conn)
}

func handleChat(conn *websocket.Conn) {
	defer conn.Close()

	join := JoinMessage{}
	err := conn.ReadJSON(&join)
	if err != nil {
		fmt.Println("Error reading json.", err)
		return
	}

	allConnsMutex.Lock()
	allConns = append(allConns, ChatConn{ join.name, conn })
    fmt.Printf("len = %d\n", len(allConns))

	for _, c := range allConns {
		if c.conn == conn { continue }
		if err = c.conn.WriteJSON(ChatMessageOut{ "Server", join.name + " has joined.", getUnixTimeStr() }); err != nil {
			fmt.Println(err)
		}
	}
	allConnsMutex.Unlock()

    for {
		msg := ChatMessageIn{}

		err = conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading json.", err)
            break
		}

		fmt.Printf("Got message: %#v\n", msg)

		out := ChatMessageOut{ join.name, msg.message, getUnixTimeStr() }

		allConnsMutex.Lock()
        for _, c := range allConns {
            if c.conn == conn { continue }
            if err = c.conn.WriteJSON(out); err != nil {
                fmt.Println(err)
            }
        }
		allConnsMutex.Unlock()
	}

	allConnsMutex.Lock()
    for i, c := range allConns {
        if c.conn == conn {
            allConns[i] = allConns[len(allConns) - 1]
            allConns = allConns[:len(allConns) - 1]
            break
        }
    }
	allConnsMutex.Unlock()

}
