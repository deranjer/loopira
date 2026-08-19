// Package ws implements an in-process websocket broadcast hub. A single
// running instance is assumed (the self-hosted target for this app); if the
// app ever needs to scale to multiple instances, this hub is the piece that
// would move to Postgres LISTEN/NOTIFY or Redis pub/sub.
package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

// Event is a broadcast message describing a mutation elsewhere in the app,
// e.g. {Type: "issue.updated", TeamID: "...", Payload: <issue json>}.
type Event struct {
	Type    string `json:"type"`
	TeamID  string `json:"teamId"`
	Payload any    `json:"payload"`
}

type client struct {
	conn   *websocket.Conn
	teamID string
	send   chan Event
}

// Hub fans out events to connected clients scoped by team.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*client]struct{})}
}

// Broadcast delivers an event to every connected client subscribed to the
// event's team.
func (h *Hub) Broadcast(ev Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.teamID != ev.TeamID {
			continue
		}
		select {
		case c.send <- ev:
		default:
			slog.Warn("ws client send buffer full, dropping event", "team", ev.TeamID)
		}
	}
}

// ClientCount reports the number of currently connected clients (for the
// health/debug endpoint).
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ServeHTTP upgrades the connection and registers the client for the team
// given in the "team" query parameter until the connection closes.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		slog.Error("ws accept failed", "err", err)
		return
	}

	c := &client{
		conn:   conn,
		teamID: r.URL.Query().Get("team"),
		send:   make(chan Event, 16),
	}

	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.mu.Unlock()
		close(c.send)
	}()

	ctx := r.Context()
	for {
		select {
		case ev, ok := <-c.send:
			if !ok {
				conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			b, err := json.Marshal(ev)
			if err != nil {
				continue
			}
			if err := conn.Write(ctx, websocket.MessageText, b); err != nil {
				return
			}
		case <-ctx.Done():
			conn.Close(websocket.StatusNormalClosure, "server shutting down")
			return
		}
	}
}
