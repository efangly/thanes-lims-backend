package environment

import (
	"encoding/json"
	"sync"

	"github.com/efangly/thanes-lims-backend/internal/domain/environment"
	"github.com/fasthttp/websocket"
)

// Hub fans out newly created/escalated environment alerts to every
// connected WebSocket client. It implements
// ports/environment.AlertBroadcaster.
type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = struct{}{}
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
}

// Broadcast pushes the alert to every connected client, dropping any
// connection that fails to write (its read loop will unregister it).
func (h *Hub) Broadcast(a environment.EnvAlert) {
	payload, err := json.Marshal(toAlertResponse(a))
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			delete(h.clients, conn)
		}
	}
}
