package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected")
			break
		}

		command := string(msg)
		switch command {
		case "pull":
			result := simulateGacha()
			ws.WriteMessage(websocket.TextMessage, []byte(result))
		case "pull10":
			results := ""
			for i := 0; i < 10; i++ {
				results += simulateGacha() + " "
			}
			ws.WriteMessage(websocket.TextMessage, []byte(results))
		}
	}
}

func simulateGacha() string {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(100) + 1
	switch {
	case n <= 70:
		return "🎴 Common"
	case n <= 95:
		return "✨ Rare"
	default:
		return "🌟 Ultra Rare"
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	fmt.Println("Gacha server started on :8080")
	http.ListenAndServe(":8080", nil)
}
