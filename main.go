package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"hourglass-socket/distribution"
	"hourglass-socket/im"
	"hourglass-socket/socket"
	"log"
	"net/http"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const ListenAddr = "0.0.0.0:3096"

func main() {
	log.SetFlags(log.Llongfile | log.Ltime)

	ws := socket.New()

	service := im.New(distribution.Listen(ws))

	service.Init()

	//imService.Init()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ws.HandleConn(conn)
	})

	log.Println("Service started on " + ListenAddr)

	err := http.ListenAndServe(ListenAddr, nil)

	if err != nil {
		fmt.Println(err)
	}
}
