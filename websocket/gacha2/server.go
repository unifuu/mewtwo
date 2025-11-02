package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client data
type ClientData struct {
	PullsLeft     int
	PendingResult []string
}

var clientsData = struct {
	sync.Mutex
	data map[string]*ClientData
}{data: make(map[string]*ClientData)}

// Initialize rand
func init() {
	rand.Seed(time.Now().UnixNano())
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer ws.Close()

	// Register client data if new
	clientsData.Lock()
	if _, ok := clientsData.data[sessionID]; !ok {
		clientsData.data[sessionID] = &ClientData{PullsLeft: 20}
	}
	client := clientsData.data[sessionID]
	clientsData.Unlock()

	// Send pending results if any
	for _, r := range client.PendingResult {
		ws.WriteMessage(websocket.TextMessage, []byte(r))
	}
	client.PendingResult = nil

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Client disconnected:", sessionID)
			break
		}

		command := string(msg)
		numPulls := 0
		if command == "pull1" {
			numPulls = 1
		} else if command == "pull10" {
			numPulls = 10
		} else {
			continue
		}

		clientsData.Lock()
		if client.PullsLeft < numPulls {
			numPulls = client.PullsLeft // cannot pull more than remaining
		}
		client.PullsLeft -= numPulls
		clientsData.Unlock()

		results := make([]string, numPulls)
		for i := 0; i < numPulls; i++ {
			results[i] = simulateGacha()
		}

		// Try sending results
		for i, r := range results {
			err := ws.WriteMessage(websocket.TextMessage, []byte(r))
			if err != nil {
				fmt.Println("Cannot send to client, storing remaining results:", sessionID)
				clientsData.Lock()
				client.PendingResult = append(client.PendingResult, results[i:]...)
				clientsData.Unlock()
				return
			}
		}

		// Optional: send pulls left info
		info := fmt.Sprintf("Pulls left: %d", client.PullsLeft)
		ws.WriteMessage(websocket.TextMessage, []byte(info))
	}
}

func simulateGacha() string {
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
