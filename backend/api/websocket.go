package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHub manages WebSocket connections
type WebSocketHub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	upgrader   websocket.Upgrader
}

// Client represents a WebSocket client
type Client struct {
	hub  *WebSocketHub
	conn *websocket.Conn
	send chan Message
}

// Message is the base interface for WebSocket messages
type Message interface{}

// StateUpdateMessage is sent when a pin state changes
type StateUpdateMessage struct {
	Type  string      `json:"type"`
	Node  string      `json:"node"`
	Pin   string      `json:"pin"`
	State interface{} `json:"state"`
}

// NodeStatusMessage is sent when a node connection status changes
type NodeStatusMessage struct {
	Type   string `json:"type"`
	Node   string `json:"node"`
	Status string `json:"status"`
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins (protected by Tailscale)
			},
		},
	}
}

// Run starts the hub's event loop
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clients := make(map[*Client]bool)
			for client := range h.clients {
				clients[client] = true
			}
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.send <- message:
				default:
					// Client send channel is full, close it
					h.mu.Lock()
					delete(h.clients, client)
					close(client.send)
					h.mu.Unlock()
				}
			}
		}
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *WebSocketHub) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan Message, 256),
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// Broadcast sends a message to all connected clients
func (h *WebSocketHub) Broadcast(message Message) {
	h.broadcast <- message
}

// ClientCount returns the number of connected clients
func (h *WebSocketHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// readPump handles incoming messages from the client
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close
			}
			break
		}
		// Handle incoming messages (subscriptions, etc.)
	}
}

// writePump handles outgoing messages to the client
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
