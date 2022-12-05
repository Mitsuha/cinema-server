package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
	"hourglass-socket/distribution"
	"hourglass-socket/model"
	"hourglass-socket/service"
	"io/ioutil"
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

	var config struct {
		DB            *model.DBConfig      `yaml:"DB"`
		ClientVersion *model.ClientVersion `yaml:"ClientVersion"`
	}

	if content, err := ioutil.ReadFile("./env.yaml"); err == nil {
		if err := yaml.Unmarshal(content, &config); err != nil {
			log.Fatalln(err)
		}

		if err := model.Boot(config.DB); err != nil {
			log.Fatalln(err)
		}
	} else {
		log.Fatalln(err)
	}

	var distributor = distribution.New(service.New)

	distributor.Enable(&distribution.ShowConnectLog{})
	distributor.Enable(distribution.DTracker())

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		go distributor.TakeOver(conn)
	})

	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		version, _ := json.Marshal(config.ClientVersion)

		_, _ = w.Write(version)
	})

	log.Println("Service started on " + ListenAddr)

	err := http.ListenAndServe(ListenAddr, nil)

	if err != nil {
		fmt.Println(err)
	}
}
