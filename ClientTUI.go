package main

import (
    "flag"
    "fmt"
    "github.com/gizak/termui"
	"github.com/gorilla/websocket"
	"strconv"
	"time"
    "net/url"
)

type JoinMessage struct {
	Name string `json:"name"`
}

type ChatMessageOut struct {
	Message string `json:"message"`
}

type ChatMessageIn struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

func max(a, b int) int {
    if a >= b {
        return a
    }
    return b
}

func getUnixTimeStr() string {
	return strconv.Itoa(int(time.Now().Unix()))
}

var host = flag.String("host", "localhost:8080", "Server Hostname")
var name = flag.String("name", "", "Display Name")

func main() {
    flag.Parse()

    if len(*name) == 0 {
        fmt.Println("Please specify a display name")
        return
    }

    err := termui.Init()
    if err != nil {
        panic(err)
    }
    defer termui.Close()

    messages := []string{ }

    u := url.URL{ Scheme: "ws", Host: *host, Path: "/socket" }
    messages = append(messages, "Connecting to " + u.String())

    list := termui.NewList()
    list.Width = termui.TermWidth()
    list.Height = termui.TermHeight() - 3
	list.Items = messages
	list.BorderLabel = "Message Log"

    input := termui.NewPar("")
    input.Y = termui.TermHeight() - 3
    input.Width = termui.TermWidth()
	input.Height = 3

    conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
    if err == websocket.ErrBadHandshake {
         fmt.Printf("handshake failed with status %s\n", resp.Status)
    }
    if err != nil {
        panic(err)
    }

    join := JoinMessage{ *name }
    err = conn.WriteJSON(&join)

    go func() {
		defer conn.Close()
		for {
            msg := ChatMessageIn{}
    		err = conn.ReadJSON(&msg)
			if err != nil {
				break
			}

            messages = append(messages, msg.Name + ": " + msg.Message)
        	list.Items = messages[max(0, len(messages) - termui.TermHeight() + 5):]

            termui.Clear()
            termui.Render(list, input)
		}

        termui.StopLoop()
	}()

    termui.Handle("/sys/wnd/resize", func(termui.Event) {
        list.Width = termui.TermWidth()
        list.Height = termui.TermHeight() - 3
        input.Y = termui.TermHeight() - 3
        input.Width = termui.TermWidth()
        list.Items = messages[max(0, len(messages) - termui.TermHeight() + 5):]

        termui.Clear()
        termui.Render(list, input)
    })

    termui.Handle("/sys/kbd", func(evt termui.Event) {
        str := evt.Data.(termui.EvtKbd).KeyStr

        if str == "<enter>" {
            if len(input.Text) == 0 { return }

            msg := ChatMessageOut{ input.Text }
            err = conn.WriteJSON(&msg)

            messages = append(messages, *name + ": " + input.Text)
        	list.Items = messages[max(0, len(messages) - termui.TermHeight() + 5):]

            input.Text = ""
        }

        if str == "<space>" {
            str = " "
        } else if len(str) > 1 {
            str = ""
        }

        input.Text = input.Text + str
        termui.Clear()
        termui.Render(list, input)
    })

    termui.Handle("/sys/kbd/C-8", func(termui.Event) {
        if len(input.Text) == 0 { return }

        input.Text = input.Text[:len(input.Text) - 1]

        termui.Clear()
        termui.Render(list, input)
    })

    termui.Handle("/sys/kbd/C-c", func(termui.Event) {
        termui.StopLoop()
    })

    termui.Render(list, input)

    termui.Loop()
}
