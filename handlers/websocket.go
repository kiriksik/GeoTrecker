package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex
var broadcast = make(chan []byte)

func WsHandler(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("err:", err)
		return err
	}
	defer ws.Close()

	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			log.Println("err:", err)
			break
		}
	}
	log.Println("Succes connect to WebSocket")
	return nil
}

func BroadcastLocations() {
	for {
		msg := <-broadcast
		log.Println("msg:", msg)

		clientsMu.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		clientsMu.Unlock()
	}
}
