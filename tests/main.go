package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Message represents a message structure
type Message struct {
	SenderID    string `json:"sender_id"`
	Content     string `json:"content"`
	RecipientID string `json:"recipient_id,omitempty"` // Added for private messages, optional
}

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	clients    map[string]*websocket.Conn // userID -> connection
	broadcast  chan Message
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin (for testing)
	},
}

var wsManager = &WebSocketManager{
	clients:    make(map[string]*websocket.Conn),
	broadcast:  make(chan Message),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}

func (ws *WebSocketManager) run() {
	for {
		select {
		case message := <-ws.broadcast:
			ws.mutex.RLock()
			if message.RecipientID != "" { // If RecipientID is set, it's a private message
				if conn, exists := ws.clients[message.RecipientID]; exists {
					if err := conn.WriteJSON(message); err != nil {
						log.Printf("Error sending message to %s: %v", message.RecipientID, err)
						conn.Close()
						delete(ws.clients, message.RecipientID)
					}
				} else {
					log.Printf("Recipient %s not found for direct message.", message.RecipientID)
				}
			} else { // No RecipientID, broadcast to all
				for userID, conn := range ws.clients {
					if err := conn.WriteJSON(message); err != nil {
						log.Printf("Error broadcasting message to %s: %v", userID, err)
						conn.Close()
						delete(ws.clients, userID)
					}
				}
			}
			ws.mutex.RUnlock()
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user ID from query parameter
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter required", http.StatusBadRequest)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Store the connection with user ID
	wsManager.mutex.Lock()
	wsManager.clients[userID] = conn
	wsManager.mutex.Unlock()

	log.Printf("User %s connected. Total connections: %d", userID, len(wsManager.clients))

	// Handle incoming messages
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error for user %s: %v", userID, err)
			break
		}

		// Set the sender ID from the current connection's user ID
		msg.SenderID = userID

		log.Printf("Received from %s: %s (to %s)", msg.SenderID, msg.Content, msg.RecipientID)

		// For broadcasting to all, ensure RecipientID is empty.
		// If you want to support private messages, the client would set RecipientID.
		// For this test, we want all messages to be broadcast.
		msg.RecipientID = ""

		// Send the message to the broadcast channel
		wsManager.broadcast <- msg
	}

	// Clean up when connection closes or loop breaks due to error
	wsManager.mutex.Lock()
	delete(wsManager.clients, userID)
	wsManager.mutex.Unlock()

	log.Printf("User %s disconnected. Total connections: %d", userID, len(wsManager.clients))
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	wsManager.mutex.RLock()
	connectionCount := len(wsManager.clients)
	wsManager.mutex.RUnlock()

	response := map[string]interface{}{
		"active_connections": connectionCount,
		"status":             "running",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	go wsManager.run()

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/status", handleStatus)

	// Serve a simple test page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Test</title>
</head>
<body>
    <div id="messages"></div>
    <input type="text" id="messageInput" placeholder="Enter message">
    <button onclick="sendMessage()">Send</button>
    <script>
        const userId = 'user_' + Math.random().toString(36).substr(2, 9);
        const ws = new WebSocket('ws://localhost:8081/ws?user_id=' + userId);
        
        ws.onmessage = function(event) {
            const data = JSON.parse(event.data);
            document.getElementById('messages').innerHTML += 
                '<div>From ' + data.sender_id + ': ' + data.content + '</div>';
        };
        
        function sendMessage() {
            const input = document.getElementById('messageInput');
            ws.send(JSON.stringify({
                content: input.value
            }));
            input.value = '';
        }
        
        document.getElementById('messageInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });
    </script>
</body>
</html>`)
	})

	fmt.Println("WebSocket server starting on :8081")
	fmt.Println("Visit http://localhost:8081 for a test page")
	fmt.Println("WebSocket endpoint: ws://localhost:8081/ws?user_id=YOUR_USER_ID")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
