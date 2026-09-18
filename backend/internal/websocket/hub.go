package websocket

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"live-polling-app/backend/internal/redisclient"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// Client represents a single connected browser socket, subscribed to one poll.
type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	pollID string
	hub    *Hub
}

// Hub keeps track of all connected clients grouped by poll, and bridges
// Redis Pub/Sub messages to the right set of WebSocket connections. Only
// one Redis subscription is opened per poll (shared by all its viewers),
// not one per client connection.
type Hub struct {
	mu          sync.Mutex
	rooms       map[string]map[*Client]bool
	redisCancel map[string]context.CancelFunc
	redisClient *redisclient.Client
}

func NewHub(rc *redisclient.Client) *Hub {
	return &Hub{
		rooms:       make(map[string]map[*Client]bool),
		redisCancel: make(map[string]context.CancelFunc),
		redisClient: rc,
	}
}

// Register adds a client to a poll's room, starting a Redis subscription
// for that poll if this is the first viewer.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[c.pollID]; !ok {
		h.rooms[c.pollID] = make(map[*Client]bool)
		h.startRedisBridge(c.pollID)
	}
	h.rooms[c.pollID][c] = true

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	n, _ := h.redisClient.IncrViewers(ctx, c.pollID)
	h.broadcastViewerCount(c.pollID, n)
}

// Unregister removes a client, closing the Redis subscription once a
// poll's room becomes empty so we don't leak subscriptions.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[c.pollID]; ok {
		if _, exists := clients[c]; exists {
			delete(clients, c)
			close(c.send)
		}
		if len(clients) == 0 {
			delete(h.rooms, c.pollID)
			if cancel, ok := h.redisCancel[c.pollID]; ok {
				cancel()
				delete(h.redisCancel, c.pollID)
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	n, _ := h.redisClient.DecrViewers(ctx, c.pollID)
	h.broadcastViewerCount(c.pollID, n)
}

// startRedisBridge opens exactly one Redis subscription for a poll and
// fans every message it receives out to all currently-registered clients
// for that poll. Must be called with h.mu held.
func (h *Hub) startRedisBridge(pollID string) {
	ctx, cancel := context.WithCancel(context.Background())
	h.redisCancel[pollID] = cancel

	pubsub := h.redisClient.Subscribe(ctx, pollID)
	ch := pubsub.Channel()

	go func() {
		defer pubsub.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				h.broadcastRaw(pollID, []byte(msg.Payload))
			}
		}
	}()
}

func (h *Hub) broadcastRaw(pollID string, payload []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.rooms[pollID] {
		select {
		case c.send <- payload:
		default:
			// slow consumer; drop connection rather than block the hub
			close(c.send)
			delete(h.rooms[pollID], c)
		}
	}
}

func (h *Hub) broadcastViewerCount(pollID string, count int64) {
	payload := []byte(`{"type":"viewer_count","pollId":"` + pollID + `","count":` + itoa(count) + `}`)
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.rooms[pollID] {
		select {
		case c.send <- payload:
		default:
		}
	}
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// ServeClient upgrades an HTTP connection and pumps messages for a single
// poll room until the connection closes.
func ServeClient(hub *Hub, conn *websocket.Conn, pollID string) {
	client := &Client{conn: conn, send: make(chan []byte, 32), pollID: pollID, hub: hub}
	hub.Register(client)

	go client.writePump()
	client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(1024)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("ws ping failed for poll %s: %v", c.pollID, err)
				return
			}
		}
	}
}
