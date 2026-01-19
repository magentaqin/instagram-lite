package realtime

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)


type Message struct {
	Type string      `json:"type"` // e.g. "post_created" "ping" "error"
	Data interface{} `json:"data"`
}

type PostItem struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	ImageURL  string   `json:"image_url"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
}

// server-side representation of a connected WebSocket peer
type Client struct {
	// underlying WebSocket connection for this client.
	conn *websocket.Conn
	//  buffered message queue.
	//  Later, a writer goroutine would drain it and write it to ws connection serially to avoid concurrently writing issues(data race issues).
	send chan []byte 
}

// creates a new WebSocket client.
func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, 128),
	}
}

// Hub is a pub/sub for WebSocket clients.
type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	clients    map[*Client]struct{}
}

func NewHub() *Hub {
	h := &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
	  // Small buffer to absorb short bursts of events (e.g. rapid post creation)
   // so HTTP handlers are not blocked by websocket fan-out.
		broadcast:  make(chan []byte, 128), 
		clients:    make(map[*Client]struct{}),
	}
	go h.run() // global goroutinme
	return h
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

// All mutations of client happen here.
func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
				_ = c.conn.Close()
			}

		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					// if the Client’s send queue is full; drop it to avoid blocking the hub.
					delete(h.clients, c)
					close(c.send)
					_ = c.conn.Close()
				}
			}
		}
	}
}

// BroadcastPostCreated encodes post and broadcasts a "post_created" event.
func (h *Hub) BroadcastPostCreated(post PostItem) {
	env := Message{Type: "post_created", Data: post}
	b, err := json.Marshal(env)
	if err != nil {
		log.Printf("ws marshal failed: %v", err)
		return
	}
	h.broadcast <- b
}


// Read and write goroutines for a single WebSocket client connection.
const (
	writeWait  = 10 * time.Second // Maximum time to write to the ws connection.
	pongWait   = 60 * time.Second // How long we wait for pong from the client(browser) before considering the connection dead.
	pingPeriod = (pongWait * 9) / 10 // Ping interval; should be < pongWait.
)

// Write to ws connection and send ping to the client to check whether the connection is alive
func (c *Client) writePump() {
	// set up a ticker to send ping to the client
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait)) // Set a write deadline to prevent blocking.
			if !ok {
				// Hub already closed the channel(de-register or too slow), 
				// so we need to send close Frame instead of c.conn.close() to indicate it's a normal close behavior.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("ws write failed: %v", err)
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// pong to the server to check whether the connection is alivce
func (c *Client) readPump(h *Hub) {
	defer func() { h.Unregister(c) }()

	c.conn.SetReadLimit(1024) // small limit as we don't accept real client payloads(in case someone sends big payload)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait)) // Close connection if we don't receive pong in time.
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait)) // Extend deadline on every pong.
		return nil
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

// Exported wrapper for writePump and readPump
func (c *Client) WritePump() { c.writePump() }
func (c *Client) ReadPump(h *Hub) { c.readPump(h) }