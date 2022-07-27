package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"hourglass-socket/distribution"
	"hourglass-socket/socket"
	"net/http"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	ws := socket.New()

	_ = distribution.Listen(ws)

	//imService := im.New(wsService)

	//imService.Init()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ws.HandleConn(conn)
	})

	err := http.ListenAndServe("0.0.0.0:3096", nil)

	if err != nil {
		fmt.Println(err)
	}
}
