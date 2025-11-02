package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader converts HTTP to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow all origins
}

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan string)            // broadcast channel

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()

	// Register client
	clients[ws] = true

	for {
		// Read message from client
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected")
			delete(clients, ws)
			break
		}
		// Send message to broadcast channel
		broadcast <- string(msg)
	}
}

func handleMessages() {
	for {
		// Grab next message from broadcast channel
		msg := <-broadcast

		// Send to all clients
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				fmt.Println("Write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
