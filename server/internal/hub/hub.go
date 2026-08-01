// Package hub tracks live WebSocket connections.
package hub

import (
	"context"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

const writeTimeout = 5 * time.Second

// Client wraps one participant's live WebSocket connection.
type Client struct {
	ParticipantID string
	Username      string
	conn          *websocket.Conn
	send          chan []byte
	closeOnce     sync.Once
	done          chan struct{}
}

func newClient(participantID, username string, conn *websocket.Conn) *Client {
	return &Client{
		ParticipantID: participantID,
		Username:      username,
		conn:          conn,
		send:          make(chan []byte, 32),
		done:          make(chan struct{}),
	}
}

// Enqueue queues a message for delivery without blocking the caller.
func (c *Client) Enqueue(msg []byte) bool {
	select {
	case c.send <- msg:
		return true
	default:
		return false
	}
}

// Close closes the underlying connection exactly once, unblocking the read
// pump so its cleanup path runs. Safe to call multiple times.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		_ = c.conn.Close(websocket.StatusNormalClosure, "closed")
	})
}

// Hub is the registry of currently-connected clients, keyed by participant ID.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

func New() *Hub {
	return &Hub{clients: make(map[string]*Client)}
}

// Register adds (or replaces) the live connection for a participant.
func (h *Hub) Register(participantID, username string, conn *websocket.Conn) *Client {
	c := newClient(participantID, username, conn)
	h.mu.Lock()
	old, existed := h.clients[participantID]
	h.clients[participantID] = c
	h.mu.Unlock()
	if existed {
		old.Close()
	}
	return c
}

func (h *Hub) Unregister(participantID string, c *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if current, ok := h.clients[participantID]; ok && current == c {
		delete(h.clients, participantID)
		return true
	}
	return false
}

func (h *Hub) Get(participantID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	c, ok := h.clients[participantID]
	return c, ok
}

func (h *Hub) IsConnected(participantID string) bool {
	_, ok := h.Get(participantID)
	return ok
}

func (h *Hub) Send(participantID string, msg []byte) bool {
	c, ok := h.Get(participantID)
	if !ok {
		return false
	}
	return c.Enqueue(msg)
}

// Broadcast delivers msg to every currently-connected client.
func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		c.Enqueue(msg)
	}
}

// Read blocks until a text frame arrives or the connection errors/closes.
func (c *Client) Read(ctx context.Context) ([]byte, error) {
	_, data, err := c.conn.Read(ctx)
	return data, err
}

// WritePump drains the client's send channel to the socket until the
// connection is closed.
func (c *Client) WritePump(ctx context.Context) {
	for {
		select {
		case <-c.done:
			return
		case <-ctx.Done():
			return
		case msg := <-c.send:
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				c.Close()
				return
			}
		}
	}
}
