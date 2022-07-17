package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"hourglass-socket/im"
	"hourglass-socket/socket"
	"net/http"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	wsService := socket.New()

	imService := im.New(wsService)

	imService.Init()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		wsService.HandleConn(conn)
	})

	err := http.ListenAndServe("0.0.0.0:3096", nil)

	if err != nil {
		fmt.Println(err)
	}
}
