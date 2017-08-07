package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type JoinMessage struct {
	Name string `json:"name"`
}

type ChatMessageIn struct {
	Message string `json:"message"`
}

type ChatMessageOut struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type ChatConn struct {
	name string
	conn *websocket.Conn
}

var allConnsMutex sync.RWMutex
var allConns []ChatConn

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool { return true },
}

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
	conn, err := upgrader.Upgrade(w, r, w.Header())
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
	allConns = append(allConns, ChatConn{join.Name, conn})
	fmt.Printf("len = %d\n", len(allConns))

	for _, c := range allConns {
		if c.conn == conn {
			continue
		}
		if err = c.conn.WriteJSON(ChatMessageOut{"Server", join.Name + " has joined.", getUnixTimeStr()}); err != nil {
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

		out := ChatMessageOut{join.Name, msg.Message, getUnixTimeStr()}

		allConnsMutex.Lock()
		for _, c := range allConns {
			if c.conn == conn {
				continue
			}
			if err = c.conn.WriteJSON(out); err != nil {
				fmt.Println(err)
			}
		}
		allConnsMutex.Unlock()
	}

	allConnsMutex.Lock()
	for _, c := range allConns {
		if c.conn == conn {
			continue
		}
		if err = c.conn.WriteJSON(ChatMessageOut{"Server", join.Name + " has left.", getUnixTimeStr()}); err != nil {
			fmt.Println(err)
		}
	}
	for i, c := range allConns {
		if c.conn == conn {
			allConns[i] = allConns[len(allConns)-1]
			allConns = allConns[:len(allConns)-1]
			break
		}
	}
	allConnsMutex.Unlock()

}
